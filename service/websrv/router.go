// SPDX-License-Identifier: BSD-3-Clause

package websrv

import (
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"connectrpc.com/vanguard"
	"github.com/rs/cors"
	"github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1/schemav1alpha1connect"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func (s *WebSrv) setupRouter() (http.Handler, error) {
	mux := http.NewServeMux()

	// Create interceptors
	validatorInterceptor, err := validate.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateValidatorInterceptor, err)
	}

	otelInterceptor, err := otelconnect.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOpenTelemetryInterceptor, err)
	}

	// Create the main proto server
	protoServer := &ProtoServer{}

	// Setup gRPC/Connect services
	services := []*vanguard.Service{
		vanguard.NewService(
			schemav1alpha1connect.NewBMCServiceHandler(
				protoServer,
				connect.WithInterceptors(validatorInterceptor, otelInterceptor),
			),
		),
	}

	// Create transcoder for protocol conversion
	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTranscoder, err)
	}

	// Mount routes based on webui flag
	if s.webui {
		fileServer := http.FileServer(http.Dir(s.webuiPath))
		mux.Handle("/", combinedRouter(fileServer, transcoder))
	} else {
		mux.Handle("/", transcoder)
	}

	// Setup health check and reflection services
	healthCheck := grpchealth.NewStaticChecker(
		schemav1alpha1connect.BMCServiceName,
	)
	reflector := grpcreflect.NewStaticReflector(
		schemav1alpha1connect.BMCServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	mux.Handle(grpchealth.NewHandler(healthCheck))

	// Apply CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	handler := corsMiddleware.Handler(mux)

	// Apply OpenTelemetry HTTP instrumentation
	handler = otelhttp.NewHandler(handler, "websrv")

	return handler, nil
}

func combinedRouter(htmlHandler, apiHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Type"), "application") {
			apiHandler.ServeHTTP(w, r)
		} else {
			htmlHandler.ServeHTTP(w, r)
		}
	})
}
