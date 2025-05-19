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
	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type API struct {
	config  *config.AppConfig
	storage *storage.Storage
	logger  *zap.SugaredLogger
	server  *grpc.Server
}

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
	authInterceptor := interceptors.NewAuthInterceptor(authService)
	responseInterceptor := interceptors.NewAuthResponseInterceptor(authService)

	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary(),
			responseInterceptor.Unary(),
		),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterUsersServer(s, handlers.NewUserService(stor, lgr))

	return &API{
		config:  cfg,
		storage: stor,
		logger:  lgr,
		server:  s,
	}, nil
}

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
