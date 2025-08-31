// SPDX-License-Identifier: BSD-3-Clause

// Package websrv provides a web server implementation that supports both HTTP/2 and HTTP/3
// protocols with TLS encryption. It serves Connect RPC APIs and optionally a web UI.
package websrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"github.com/arunsworld/nursery"
	"github.com/lorenzosaino/go-sysctl"
	"github.com/nats-io/nats.go"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"github.com/rs/cors"
	"github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1/schemav1alpha1connect"
	"github.com/u-bmc/u-bmc/pkg/cert"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ProtoServer implements all the Connect RPC service handlers for the BMC API.
type ProtoServer struct {
	schemav1alpha1connect.UnimplementedBMCServiceHandler
}

// Compile-time assertion that WebSrv implements service.Service.
var _ service.Service = (*WebSrv)(nil)

// WebSrv is a web server that provides HTTP/2 and HTTP/3 endpoints for BMC operations.
type WebSrv struct {
	config
}

// New creates a new WebSrv instance with the provided options.
func New(opts ...Option) *WebSrv {
	cfg := &config{
		name:         "websrv",
		addr:         ":443",
		webui:        false,
		hostname:     "localhost",
		certPath:     "/var/cache/selfsigned/cert.pem",
		keyPath:      "/var/cache/selfsigned/key.pem",
		webuiPath:    "/usr/share/webui",
		readTimeout:  5 * time.Second,
		writeTimeout: 5 * time.Second,
		idleTimeout:  120 * time.Second,
		rmemMax:      "7500000",
		wmemMax:      "7500000",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &WebSrv{
		config: *cfg,
	}
}

// Name returns the service name.
func (s *WebSrv) Name() string {
	return s.name
}

// Run starts the web server and blocks until the context is canceled.
func (s *WebSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting web server", "service", s.name)

	if err := s.configureSysctl(ctx); err != nil {
		l.WarnContext(ctx, "Failed to configure sysctls for QUIC", "error", err)
	}

	router, err := s.setupRouter()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSetupRouter, err)
	}

	tlsConfig, err := s.setupTLS()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSetupTLS, err)
	}

	lc := &net.ListenConfig{}
	tcpListener, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateTCPListener, err)
	}
	defer tcpListener.Close()

	// Create HTTP redirect listener on port 80
	httpListener, err := lc.Listen(ctx, "tcp", ":80")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateHTTPListener, err)
	}
	defer httpListener.Close()

	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResolveUDPAddress, err)
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateUDPListener, err)
	}
	defer udpListener.Close()

	http3Server := &http3.Server{
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
		TLSConfig: http3.ConfigureTLSConfig(tlsConfig),
	}

	http2Server := &http.Server{
		Handler:      router,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
		TLSConfig:    tlsConfig,
		ErrorLog:     log.NewStdLoggerAt(l, slog.LevelWarn),
	}

	// HTTP redirect server that just forwards port 80 to 443
	redirectServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.Host
			if strings.Contains(host, ":") {
				host, _, _ = net.SplitHostPort(host)
			}
			httpsURL := "https://" + host + r.RequestURI
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
		}),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
		ErrorLog:     log.NewStdLoggerAt(l, slog.LevelWarn),
	}

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

// configureSysctl sets kernel parameters needed for optimal QUIC performance.
func (s *WebSrv) configureSysctl(ctx context.Context) error {
	if err := sysctl.Set("net.core.rmem_max", s.rmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetRmemMax, err)
	}
	if err := sysctl.Set("net.core.wmem_max", s.wmemMax); err != nil {
		return fmt.Errorf("%w: %w", ErrSetWmemMax, err)
	}
	return nil
}

// setupTLS configures TLS settings and loads or generates certificates.
func (s *WebSrv) setupTLS() (*tls.Config, error) {
	certOpts := cert.CertificateOptions{
		Hostname: s.hostname,
	}

	certPem, keyPem, err := cert.LoadOrGenerateCertificate(s.certPath, s.keyPath, certOpts)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadOrGenerateCertificate, err)
	}

	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseCertificate, err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// setupRouter configures the HTTP router with all endpoints and middleware.
func (s *WebSrv) setupRouter() (http.Handler, error) {
	mux := http.NewServeMux()

	if s.webui {
		fileServer := http.FileServer(http.Dir(s.webuiPath))
		mux.Handle("/", fileServer)
	}

	validatorInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateValidatorInterceptor, err)
	}

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOpenTelemetryInterceptor, err)
	}

	protoServer := &ProtoServer{}

	services := []*vanguard.Service{
		vanguard.NewService(
			schemav1alpha1connect.NewBMCServiceHandler(
				protoServer,
				connect.WithInterceptors(validatorInterceptor, otelInterceptor),
			),
		),
	}

	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTranscoder, err)
	}

	mux.Handle("/api/", http.StripPrefix("/api", transcoder))

	healthCheck := grpchealth.NewStaticChecker(
		schemav1alpha1connect.BMCServiceName,
	)
	reflector := grpcreflect.NewStaticReflector(
		schemav1alpha1connect.BMCServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(grpchealth.NewHandler(healthCheck))

	corsMiddleware := cors.New(cors.Options{
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	handler := corsMiddleware.Handler(mux)

	handler = otelhttp.NewHandler(handler, "websrv")

	return handler, nil
}
