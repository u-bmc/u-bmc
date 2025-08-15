// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"u-bmc.org/u-bmc/api/gen/schema/v1alpha1/protov1alpha1connect"
	"u-bmc.org/u-bmc/pkg/cert"
	"u-bmc.org/u-bmc/pkg/log"
)

type ProtoServer struct {
	protov1alpha1connect.UnimplementedChassisServiceHandler
	protov1alpha1connect.UnimplementedCoolingDeviceServiceHandler
	protov1alpha1connect.UnimplementedHostManagementServiceHandler
	protov1alpha1connect.UnimplementedHostServiceHandler
	protov1alpha1connect.UnimplementedManagementControllerServiceHandler
	protov1alpha1connect.UnimplementedSensorServiceHandler
	protov1alpha1connect.UnimplementedThermalManagementServiceHandler
	protov1alpha1connect.UnimplementedThermalZoneServiceHandler
	protov1alpha1connect.UnimplementedUserServiceHandler
}

type WebSrv struct {
	config
}

func New(opts ...Option) *WebSrv {
	cfg := &config{
		name:  "websrv",
		addr:  ":443",
		webui: false,
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &WebSrv{
		config: *cfg,
	}
}

func (s *WebSrv) Name() string {
	return s.name
}

func (s *WebSrv) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) (err error) {
	l := log.GetGlobalLogger()

	l.InfoContext(ctx, "Starting web server", "service", s.name)

	// Those are needed for QUIC/HTTP3 but we can proceed without them
	if err := sysctl.Set("net.core.rmem_max", "7500000"); err != nil {
		l.ErrorContext(ctx, "Failed to update RMEM for QUIC usage", "error", err)
	}
	if err := sysctl.Set("net.core.wmem_max", "7500000"); err != nil {
		l.ErrorContext(ctx, "Failed to update WMEM for QUIC usage", "error", err)
	}

	router, err := setupRouter(s.webui)
	if err != nil {
		return err
	}

	certPem, keyPem, err := cert.GenerateSelfsigned("localhost")
	if err != nil {
		return err
	}
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return err
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
	}

	// Create TCP and UDP listeners
	tcpListener, err := net.Listen("tcp", "localhost:443")
	if err != nil {
		return fmt.Errorf("error creating TCP listener: %w", err)
	}
	defer tcpListener.Close()

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:443")
	if err != nil {
		return fmt.Errorf("error resolving UDP address: %w", err)
	}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("error creating UDP listener: %w", err)
	}
	defer udpListener.Close()

	// Create a QUIC/HTTP3 server with the mux
	http3Server := http3.Server{
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
		TLSConfig: http3.ConfigureTLSConfig(tlsConf),
	}

	// Create an HTTP/2 server for fallback
	http2Server := &http.Server{
		Handler:      router,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
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

	return nursery.RunConcurrentlyWithContext(
		ctx,
		func(ctx context.Context, c chan error) {
			l.InfoContext(ctx, "Starting HTTP3 server", "service", s.name, "addr", s.addr)
			if err := http3Server.Serve(udpListener); err != nil {
				c <- fmt.Errorf("error serving HTTP3: %w", err)
			}
		},
		func(ctx context.Context, c chan error) {
			l.InfoContext(ctx, "Starting HTTP2 fallback server", "service", s.name, "addr", s.addr)
			if err := http2Server.Serve(tcpListener); err != nil && err != http.ErrServerClosed {
				c <- fmt.Errorf("error serving HTTP2: %w", err)
			}
		})
}

func setupRouter(webui bool) (http.Handler, error) {
	mux := http.NewServeMux()

	if webui {
		fileServer := http.FileServer(http.Dir("/usr/share/webui"))
		mux.Handle("/", fileServer)
	}

	validatorInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator interceptor: %w", err)
	}

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTelemetry interceptor: %w", err)
	}

	protoServer := &ProtoServer{}

	chassisService := vanguard.NewService(
		protov1alpha1connect.NewChassisServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	coolingDeviceService := vanguard.NewService(
		protov1alpha1connect.NewCoolingDeviceServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	hostManagementService := vanguard.NewService(
		protov1alpha1connect.NewHostManagementServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	hostService := vanguard.NewService(
		protov1alpha1connect.NewHostServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	managementControllerService := vanguard.NewService(
		protov1alpha1connect.NewManagementControllerServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	sensorService := vanguard.NewService(
		protov1alpha1connect.NewSensorServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	thermalManagementService := vanguard.NewService(
		protov1alpha1connect.NewThermalManagementServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	thermalZoneService := vanguard.NewService(
		protov1alpha1connect.NewThermalZoneServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	userService := vanguard.NewService(
		protov1alpha1connect.NewUserServiceHandler(
			protoServer,
			connect.WithInterceptors(validatorInterceptor, otelInterceptor),
		),
	)

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
		chassisService,
		coolingDeviceService,
		hostManagementService,
		hostService,
		managementControllerService,
		sensorService,
		thermalManagementService,
		thermalZoneService,
		userService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transcoder: %w", err)
	}

	// Serve the Connect API on the /api route
	mux.Handle("/api", transcoder)

	// Extra routes helpful
	healthCheck := grpchealth.NewStaticChecker(
		protov1alpha1connect.ChassisServiceName,
		protov1alpha1connect.CoolingDeviceServiceName,
		protov1alpha1connect.HostManagementServiceName,
		protov1alpha1connect.HostServiceName,
		protov1alpha1connect.ManagementControllerServiceName,
		protov1alpha1connect.SensorServiceName,
		protov1alpha1connect.ThermalManagementServiceName,
		protov1alpha1connect.ThermalZoneServiceName,
		protov1alpha1connect.UserServiceName,
	)
	reflector := grpcreflect.NewStaticReflector(
		protov1alpha1connect.ChassisServiceName,
		protov1alpha1connect.CoolingDeviceServiceName,
		protov1alpha1connect.HostManagementServiceName,
		protov1alpha1connect.HostServiceName,
		protov1alpha1connect.ManagementControllerServiceName,
		protov1alpha1connect.SensorServiceName,
		protov1alpha1connect.ThermalManagementServiceName,
		protov1alpha1connect.ThermalZoneServiceName,
		protov1alpha1connect.UserServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(grpchealth.NewHandler(healthCheck))

	// Connect CORS rules on all routes
	corsMiddleware := cors.New(cors.Options{
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	handler := corsMiddleware.Handler(mux)

	handler = otelhttp.NewHandler(handler, "websrv")

	return handler, nil
}
