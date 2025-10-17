// SPDX-License-Identifier: BSD-3-Clause

package usermgr

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	schemav1alpha1 "github.com/u-bmc/u-bmc/api/gen/schema/v1alpha1"
	"github.com/u-bmc/u-bmc/pkg/ipc"
	"github.com/u-bmc/u-bmc/pkg/log"
	"github.com/u-bmc/u-bmc/pkg/telemetry"
	"github.com/u-bmc/u-bmc/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Compile-time assertion that UserMgr implements service.Service.
var _ service.Service = (*UserMgr)(nil)

// UserMgr provides user management functionality for BMC authentication and authorization.
type UserMgr struct {
	config       config
	nc           *nats.Conn
	microService micro.Service
	logger       *slog.Logger
	tracer       trace.Tracer
}

// New creates a new UserMgr instance with the provided options.
func New(opts ...Option) *UserMgr {
	cfg := &config{
		name: "usermgr",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return &UserMgr{
		config: *cfg,
	}
}

func (s *UserMgr) Name() string {
	return s.config.name
}

func (s *UserMgr) Run(ctx context.Context, ipcConn nats.InProcessConnProvider) error {
	s.tracer = otel.Tracer(s.config.name)
	s.logger = log.GetGlobalLogger().With("service", s.config.name)

	s.logger.InfoContext(ctx, "Starting user manager", "service", s.config.name)

	var err error
	s.nc, err = nats.Connect("", nats.InProcessServer(ipcConn))
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer s.nc.Drain()

	s.microService, err = micro.AddService(s.nc, micro.Config{
		Name:        s.config.name,
		Description: "User management service",
		Version:     "1.0.0",
	})
	if err != nil {
		return fmt.Errorf("failed to create micro service: %w", err)
	}

	if err := s.registerEndpoints(ctx); err != nil {
		return fmt.Errorf("failed to register endpoints: %w", err)
	}

	s.logger.InfoContext(ctx, "User manager started successfully")

	<-ctx.Done()
	s.logger.InfoContext(ctx, "Stopping user manager", "service", s.config.name, "reason", ctx.Err())

	return ctx.Err()
}

func (s *UserMgr) registerEndpoints(ctx context.Context) error {
	groups := make(map[string]micro.Group)

	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserCreate,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserCreate)), groups); err != nil {
		return fmt.Errorf("failed to register user create endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserInfo,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserInfo)), groups); err != nil {
		return fmt.Errorf("failed to register user info endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserUpdate,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserUpdate)), groups); err != nil {
		return fmt.Errorf("failed to register user update endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserDelete,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserDelete)), groups); err != nil {
		return fmt.Errorf("failed to register user delete endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserList,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserList)), groups); err != nil {
		return fmt.Errorf("failed to register user list endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserChangePassword,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserChangePassword)), groups); err != nil {
		return fmt.Errorf("failed to register user change password endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserResetPassword,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserResetPassword)), groups); err != nil {
		return fmt.Errorf("failed to register user reset password endpoint: %w", err)
	}
	if err := ipc.RegisterEndpointWithGroupCache(s.microService, ipc.SubjectUserAuthenticate,
		micro.HandlerFunc(s.createRequestHandler(ctx, s.handleUserAuthenticate)), groups); err != nil {
		return fmt.Errorf("failed to register user authenticate endpoint: %w", err)
	}

	return nil
}

func (s *UserMgr) createRequestHandler(parentCtx context.Context, handler func(context.Context, micro.Request)) micro.HandlerFunc {
	return func(req micro.Request) {
		ctx := telemetry.GetCtxFromReq(req)
		ctx = context.WithoutCancel(ctx)

		if s.tracer != nil {
			_, span := s.tracer.Start(ctx, "usermgr.handleRequest")
			span.SetAttributes(
				attribute.String("subject", req.Subject()),
				attribute.String("service", s.config.name),
			)
			defer span.End()
		}

		handler(ctx, req)
	}
}

func (s *UserMgr) handleUserCreate(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserCreate")
		defer span.End()
	}

	var request schemav1alpha1.CreateUserRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.CreateUserResponse{
		User: &schemav1alpha1.User{
			Username: request.User.Username,
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserInfo(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserInfo")
		defer span.End()
	}

	var request schemav1alpha1.GetUserRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.GetUserResponse{
		User: &schemav1alpha1.User{
			Username: "mockuser",
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserUpdate(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserUpdate")
		defer span.End()
	}

	var request schemav1alpha1.UpdateUserRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.UpdateUserResponse{
		User: request.User,
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserDelete(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserDelete")
		defer span.End()
	}

	var request schemav1alpha1.DeleteUserRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.DeleteUserResponse{}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserList(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserList")
		defer span.End()
	}

	var request schemav1alpha1.ListUsersRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.ListUsersResponse{
		Users: []*schemav1alpha1.User{
			{Username: "admin"},
			{Username: "operator"},
		},
	}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserChangePassword(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserChangePassword")
		defer span.End()
	}

	var request schemav1alpha1.ChangePasswordRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.ChangePasswordResponse{}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserResetPassword(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserResetPassword")
		defer span.End()
	}

	var request schemav1alpha1.ResetPasswordRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.ResetPasswordResponse{}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}

func (s *UserMgr) handleUserAuthenticate(ctx context.Context, req micro.Request) {
	if s.tracer != nil {
		_, span := s.tracer.Start(ctx, "usermgr.handleUserAuthenticate")
		defer span.End()
	}

	var request schemav1alpha1.AuthenticateUserRequest
	if err := request.UnmarshalVT(req.Data()); err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrUnmarshalingFailed, err.Error())
		return
	}

	// Mock response for now
	response := &schemav1alpha1.AuthenticateUserResponse{}

	data, err := response.MarshalVT()
	if err != nil {
		ipc.RespondWithError(ctx, req, ipc.ErrMarshalingFailed, err.Error())
		return
	}

	if err := req.Respond(data); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "Failed to respond to request", "error", err)
	}
}
