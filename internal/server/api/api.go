// Package api implements the gRPC API server for GophKeeper with:
// - Authentication and authorization
// - Device management
// - Secure record storage
// - Request validation and interceptors
//
// Security Features:
// - JWT authentication
// - Device verification
// - TLS encryption
// - Role-based access control
//
// Architecture:
// - Handlers: Implement gRPC service interfaces
// - Interceptors: Handle auth and device verification
// - Auth: JWT token generation/validation
package api

import (
	"context"
	"fmt"
	"net"

	"github.com/frolmr/GophKeeper/internal/server/api/auth"
	"github.com/frolmr/GophKeeper/internal/server/api/handlers"
	"github.com/frolmr/GophKeeper/internal/server/api/interceptors"
	"github.com/frolmr/GophKeeper/internal/server/config"
	"github.com/frolmr/GophKeeper/internal/server/storage"
	devices "github.com/frolmr/GophKeeper/pkg/proto/devices"
	records "github.com/frolmr/GophKeeper/pkg/proto/records"
	users "github.com/frolmr/GophKeeper/pkg/proto/users"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// API is the central gRPC server component that manages:
// - Server configuration
// - Storage access
// - Request handling
// - Interceptor chain
type API struct {
	config  *config.AppConfig  // Server configuration
	storage *storage.Storage   // Data storage backend
	logger  *zap.SugaredLogger // Structured logger
	server  *grpc.Server       // gRPC server instance
}

// NewAPI initializes and configures the gRPC server with:
// 1. TLS credentials
// 2. Authentication service
// 3. Interceptors (auth + device)
// 4. Service handlers
//
// Parameters:
//
//	cfg  - Application configuration
//	stor - Storage backend
//	lgr  - Logger
//
// Returns:
//
//	*API - Configured server instance
//	error - Initialization errors
func NewAPI(cfg *config.AppConfig, stor *storage.Storage, lgr *zap.SugaredLogger) (*API, error) {
	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS credentials: %w", err)
	}

	authConfig, err := config.NewAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch auth config: %w", err)
	}
	authService := auth.NewAuthService(authConfig)
	authInterceptor := interceptors.NewAuthInterceptor(authService, stor)
	deviceInterceptor := interceptors.NewDeviceInterceptor(stor)

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary(),
			deviceInterceptor.Unary(),
		),
	}

	s := grpc.NewServer(opts...)
	users.RegisterUsersServer(s, handlers.NewUsersService(stor, lgr, authService))
	devices.RegisterDevicesServer(s, handlers.NewDevicesService(stor, lgr))
	records.RegisterRecordsServer(s, handlers.NewRecordsService(stor, lgr))

	return &API{
		config:  cfg,
		storage: stor,
		logger:  lgr,
		server:  s,
	}, nil
}

// Run starts the gRPC server and manages its lifecycle:
// - Binds to configured address
// - Handles graceful shutdown
// - Manages error propagation
//
// Parameters:
//
//	ctx - Context for graceful shutdown
//
// Returns:
//
//	error - Server runtime errors
func (api *API) Run(ctx context.Context) error {
	listen, err := net.Listen("tcp", api.config.RunAddress)
	if err != nil {
		return err
	}

	serveErr := make(chan error, 1)
	go func() {
		api.logger.Infof("Starting gRPC server on %s", api.config.RunAddress)
		if err := api.server.Serve(listen); err != nil {
			serveErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		api.logger.Info("Gracefully stopping gRPC server")
		api.server.GracefulStop()
		return nil
	case err := <-serveErr:
		return fmt.Errorf("gRPC server error: %w", err)
	}
}
