// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/arunsworld/nursery"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"github.com/u-bmc/u-bmc/pkg/log"
)

// StartServers starts all HTTP servers concurrently.
func (s *WebSrv) StartServers(ctx context.Context, router http.Handler, tlsConfig *tls.Config, httpHandler http.Handler) error {
	l := log.GetGlobalLogger()

	// Create listeners
	tcpListener, err := s.createTCPListener(ctx)
	if err != nil {
		return err
	}
	defer tcpListener.Close()

	httpListener, err := s.createHTTPListener(ctx)
	if err != nil {
		return err
	}
	defer httpListener.Close()

	udpListener, err := s.createUDPListener()
	if err != nil {
		return err
	}
	defer udpListener.Close()

	// Create servers
	http3Server := s.createHTTP3Server(router, tlsConfig)
	http2Server := s.createHTTP2Server(ctx, router, tlsConfig)
	redirectServer := s.createRedirectServer(ctx, httpHandler)

	// Setup graceful shutdown
	defer func() {
		if err := http3Server.Shutdown(ctx); err != nil {
			l.ErrorContext(ctx, "Error shutting down HTTP3 server", "service", s.name, "error", err)
		}
	}()

	defer func() {
		if err := http2Server.Shutdown(ctx); err != nil {
			l.ErrorContext(ctx, "Error shutting down HTTP2 server", "service", s.name, "error", err)
		}
	}()

	defer func() {
		if err := redirectServer.Shutdown(ctx); err != nil {
			l.ErrorContext(ctx, "Error shutting down HTTP redirect server", "service", s.name, "error", err)
		}
	}()

	// Start all servers concurrently
	return nursery.RunConcurrentlyWithContext(
		ctx,
		func(ctx context.Context, c chan error) {
			l.InfoContext(ctx, "Starting HTTP3 server", "service", s.name, "addr", s.addr)
			if err := http3Server.Serve(udpListener); err != nil {
				c <- fmt.Errorf("%w: %w", ErrHTTP3Server, err)
			}
		},
		func(ctx context.Context, c chan error) {
			l.InfoContext(ctx, "Starting HTTP2 fallback server", "service", s.name, "addr", s.addr)
			if err := http2Server.ServeTLS(tcpListener, "", ""); err != nil && err != http.ErrServerClosed {
				c <- fmt.Errorf("%w: %w", ErrHTTP2Server, err)
			}
		},
		func(ctx context.Context, c chan error) {
			l.InfoContext(ctx, "Starting HTTP redirect server", "service", s.name, "addr", ":80")
			if err := redirectServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
				c <- fmt.Errorf("%w: %w", ErrHTTPRedirectServer, err)
			}
		})
}

func (s *WebSrv) createTCPListener(ctx context.Context) (net.Listener, error) {
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTCPListener, err)
	}
	return listener, nil
}

func (s *WebSrv) createHTTPListener(ctx context.Context) (net.Listener, error) {
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", ":80")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateHTTPListener, err)
	}
	return listener, nil
}

func (s *WebSrv) createUDPListener() (net.PacketConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrResolveUDPAddress, err)
	}

	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateUDPListener, err)
	}

	return udpListener, nil
}

func (s *WebSrv) createHTTP3Server(router http.Handler, tlsConfig *tls.Config) *http3.Server {
	l := log.GetGlobalLogger()

	return &http3.Server{
		Handler: router,
		QUICConfig: &quic.Config{
			Tracer: func(ctx context.Context, perspective logging.Perspective, id quic.ConnectionID) *logging.ConnectionTracer {
				var role string
				switch perspective {
				case logging.PerspectiveServer:
					role = "server"
				case logging.PerspectiveClient:
					role = "client"
				}

				return qlog.NewConnectionTracer(
					log.NewQLogger(l.With("role", role)),
					perspective,
					id,
				)
			},
		},
		TLSConfig: configureTLSForHTTP3(tlsConfig),
	}
}

func (s *WebSrv) createHTTP2Server(ctx context.Context, router http.Handler, tlsConfig *tls.Config) *http.Server {
	l := log.GetGlobalLogger()

	return &http.Server{
		Handler:      router,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
		TLSConfig:    configureTLSForHTTP2(tlsConfig),
		ErrorLog:     log.NewStdLoggerAt(l, slog.LevelWarn),
	}
}

func (s *WebSrv) createRedirectServer(ctx context.Context, httpHandler http.Handler) *http.Server {
	l := log.GetGlobalLogger()

	var handler http.Handler

	// If we have an HTTP handler (e.g., for Let's Encrypt), use a mux to handle both
	if httpHandler != nil {
		mux := http.NewServeMux()

		// Handle Let's Encrypt HTTP-01 challenge
		mux.Handle("/.well-known/acme-challenge/", httpHandler)

		// Handle all other requests with HTTPS redirect
		mux.HandleFunc("/", s.redirectToHTTPS)

		handler = mux
	} else {
		// Just redirect everything to HTTPS
		handler = http.HandlerFunc(s.redirectToHTTPS)
	}

	return &http.Server{
		Handler:      handler,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
		ErrorLog:     log.NewStdLoggerAt(l, slog.LevelWarn),
	}
}

func (s *WebSrv) redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
	}

	// Construct HTTPS URL
	httpsURL := "https://" + host

	// Add port if not using standard HTTPS port
	if s.addr != ":443" && !strings.HasPrefix(s.addr, ":443") {
		if strings.HasPrefix(s.addr, ":") {
			httpsURL += s.addr
		} else {
			// Extract port from addr
			if _, port, err := net.SplitHostPort(s.addr); err == nil {
				httpsURL += ":" + port
			}
		}
	}

	httpsURL += r.RequestURI

	// Use 301 (Moved Permanently) for better SEO and caching
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

// GetServerInfo returns information about the configured servers.
func (s *WebSrv) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":          s.name,
		"addr":          s.addr,
		"webui":         s.webui,
		"webui_path":    s.webuiPath,
		"read_timeout":  s.readTimeout,
		"write_timeout": s.writeTimeout,
		"idle_timeout":  s.idleTimeout,
		"rmem_max":      s.rmemMax,
		"wmem_max":      s.wmemMax,
		"protocols":     []string{"HTTP/3", "HTTP/2", "HTTP/1.1"},
		"features": map[string]bool{
			"tls":                true,
			"http3":              true,
			"http2":              true,
			"automatic_redirect": true,
			"lets_encrypt":       false, // Will be determined when httpHandler is available
		},
	}
}

// HealthCheck performs a basic health check on the server configuration.
func (s *WebSrv) HealthCheck(router http.Handler, tlsConfig *tls.Config) error {
	if s.name == "" {
		return fmt.Errorf("server configuration is nil")
	}

	if router == nil {
		return fmt.Errorf("HTTP router is nil")
	}

	if tlsConfig == nil {
		return fmt.Errorf("TLS configuration is nil")
	}

	if err := validateTLSConfig(tlsConfig); err != nil {
		return fmt.Errorf("invalid TLS configuration: %w", err)
	}

	// Validate timeouts
	if s.readTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}

	if s.writeTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}

	if s.idleTimeout < 0 {
		return fmt.Errorf("idle timeout cannot be negative")
	}

	return nil
}

// GetListenAddresses returns the addresses the servers will listen on.
func (s *WebSrv) GetListenAddresses() map[string]string {
	return map[string]string{
		"http3":    s.addr + " (UDP)",
		"http2":    s.addr + " (TCP)",
		"redirect": ":80 (TCP)",
	}
}
