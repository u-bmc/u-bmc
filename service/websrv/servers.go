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
		if err := http3Server.Shutdown(ctx); err != nil && ctx.Err() == nil {
			s.logger.ErrorContext(ctx, "Error shutting down HTTP3 server", "error", err)
		}
	}()

	defer func() {
		if err := http2Server.Shutdown(ctx); err != nil && ctx.Err() == nil {
			s.logger.ErrorContext(ctx, "Error shutting down HTTP2 server", "error", err)
		}
	}()

	defer func() {
		if err := redirectServer.Shutdown(ctx); err != nil && ctx.Err() == nil {
			s.logger.ErrorContext(ctx, "Error shutting down HTTP redirect server", "error", err)
		}
	}()

	// Start all servers concurrently
	return nursery.RunConcurrentlyWithContext(
		ctx,
		func(ctx context.Context, c chan error) {
			s.logger.InfoContext(ctx, "Starting HTTP3 server", "addr", s.config.addr)
			if err := http3Server.Serve(udpListener); err != nil {
				c <- fmt.Errorf("%w: %w", ErrHTTP3Server, err)
			}
		},
		func(ctx context.Context, c chan error) {
			s.logger.InfoContext(ctx, "Starting HTTP2 fallback server", "addr", s.config.addr)
			if err := http2Server.ServeTLS(tcpListener, "", ""); err != nil && err != http.ErrServerClosed {
				c <- fmt.Errorf("%w: %w", ErrHTTP2Server, err)
			}
		},
		func(ctx context.Context, c chan error) {
			s.logger.InfoContext(ctx, "Starting HTTP redirect server", "addr", ":80")
			if err := redirectServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
				c <- fmt.Errorf("%w: %w", ErrHTTPRedirectServer, err)
			}
		})
}

func (s *WebSrv) createTCPListener(ctx context.Context) (net.Listener, error) {
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.config.addr)
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
	udpAddr, err := net.ResolveUDPAddr("udp", s.config.addr)
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
					log.NewQLogger(s.logger.With("role", role)),
					perspective,
					id,
				)
			},
		},
		TLSConfig: configureTLSForHTTP3(tlsConfig),
	}
}

func (s *WebSrv) createHTTP2Server(ctx context.Context, router http.Handler, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Handler:      router,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  s.config.readTimeout,
		WriteTimeout: s.config.writeTimeout,
		IdleTimeout:  s.config.idleTimeout,
		TLSConfig:    configureTLSForHTTP2(tlsConfig),
		ErrorLog:     log.NewStdLoggerAt(s.logger, slog.LevelWarn),
	}
}

func (s *WebSrv) createRedirectServer(ctx context.Context, httpHandler http.Handler) *http.Server {
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
		ReadTimeout:  s.config.readTimeout,
		WriteTimeout: s.config.writeTimeout,
		IdleTimeout:  s.config.idleTimeout,
		ErrorLog:     log.NewStdLoggerAt(s.logger, slog.LevelWarn),
	}
}

func (s *WebSrv) redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Extract host without port (supporting bracketed IPv6)
	host := r.Host
	hostNoPort := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostNoPort = h
	} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		// Bracketed IPv6 without port
		hostNoPort = strings.Trim(host, "[]")
	}

	// Parse s.config.addr for an explicit port only when present and not equal to "443"
	var port string
	if strings.HasPrefix(s.config.addr, ":") {
		configPort := strings.TrimPrefix(s.config.addr, ":")
		if configPort != "443" {
			port = configPort
		}
	} else if _, p, err := net.SplitHostPort(s.config.addr); err == nil {
		if p != "443" {
			port = p
		}
	}

	// Build target host, only appending port when non-empty
	targetHost := hostNoPort
	if port != "" {
		targetHost = net.JoinHostPort(hostNoPort, port)
	}

	httpsURL := "https://" + targetHost + r.RequestURI
	http.Redirect(w, r, httpsURL, http.StatusPermanentRedirect)
}

// GetServerInfo returns information about the configured servers.
func (s *WebSrv) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":          s.config.name,
		"addr":          s.config.addr,
		"webui":         s.config.webui,
		"webui_path":    s.config.webuiPath,
		"read_timeout":  s.config.readTimeout,
		"write_timeout": s.config.writeTimeout,
		"idle_timeout":  s.config.idleTimeout,
		"rmem_max":      s.config.rmemMax,
		"wmem_max":      s.config.wmemMax,
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
	if s.config.name == "" {
		return fmt.Errorf("server name cannot be empty")
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
	if s.config.readTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}

	if s.config.writeTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}

	if s.config.idleTimeout < 0 {
		return fmt.Errorf("idle timeout cannot be negative")
	}

	return nil
}

// GetListenAddresses returns the addresses the servers will listen on.
func (s *WebSrv) GetListenAddresses() map[string]string {
	return map[string]string{
		"http3":    s.config.addr + " (UDP)",
		"http2":    s.config.addr + " (TCP)",
		"redirect": ":80 (TCP)",
	}
}
