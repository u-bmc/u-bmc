// SPDX-License-Identifier: BSD-3-Clause

package websrv

import "errors"

var (
	// ErrSetupRouter indicates a failure during HTTP router configuration and middleware setup.
	ErrSetupRouter = errors.New("failed to setup router")
	// ErrSetupTLS indicates a failure during TLS configuration or certificate loading.
	ErrSetupTLS = errors.New("failed to setup TLS")
	// ErrCreateTCPListener indicates a failure to create a TCP listener for HTTP/2 connections.
	ErrCreateTCPListener = errors.New("failed to create TCP listener")
	// ErrResolveUDPAddress indicates a failure to resolve the UDP address for HTTP/3 connections.
	ErrResolveUDPAddress = errors.New("failed to resolve UDP address")
	// ErrCreateUDPListener indicates a failure to create a UDP listener for HTTP/3 connections.
	ErrCreateUDPListener = errors.New("failed to create UDP listener")
	// ErrHTTP3Server indicates an error occurred while running the HTTP/3 server.
	ErrHTTP3Server = errors.New("HTTP3 server error")
	// ErrHTTP2Server indicates an error occurred while running the HTTP/2 server.
	ErrHTTP2Server = errors.New("HTTP2 server error")
	// ErrSetRmemMax indicates a failure to configure the kernel's maximum receive buffer size.
	ErrSetRmemMax = errors.New("failed to set rmem_max")
	// ErrSetWmemMax indicates a failure to configure the kernel's maximum send buffer size.
	ErrSetWmemMax = errors.New("failed to set wmem_max")
	// ErrLoadOrGenerateCertificate indicates a failure to load existing or generate new TLS certificates.
	ErrLoadOrGenerateCertificate = errors.New("failed to load or generate certificate")
	// ErrParseCertificate indicates a failure to parse the loaded TLS certificate and key pair.
	ErrParseCertificate = errors.New("failed to parse certificate")
	// ErrCreateValidatorInterceptor indicates a failure to create the request validation interceptor.
	ErrCreateValidatorInterceptor = errors.New("failed to create validator interceptor")
	// ErrCreateOpenTelemetryInterceptor indicates a failure to create the OpenTelemetry tracing interceptor.
	ErrCreateOpenTelemetryInterceptor = errors.New("failed to create OpenTelemetry interceptor")
	// ErrCreateTranscoder indicates a failure to create the protocol transcoder for gRPC/Connect services.
	ErrCreateTranscoder = errors.New("failed to create transcoder")
)
