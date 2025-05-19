package app

import (
	"fmt"

	adapter "github.com/frolmr/GophKeeper/internal/client/adapter/grpc"
	"github.com/frolmr/GophKeeper/internal/client/client"
	"github.com/frolmr/GophKeeper/internal/client/config"
	"github.com/frolmr/GophKeeper/internal/client/crypto"
	"github.com/frolmr/GophKeeper/internal/client/service"
	"github.com/frolmr/GophKeeper/internal/client/storage"
	"google.golang.org/grpc"
)

type GophKeeper struct {
	Config      *config.Config
	UserService *service.UserService
	clientConn  *grpc.ClientConn
}

func NewApplication() (*GophKeeper, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cryptoService := crypto.NewCryptoService()

	localStorage, err := storage.NewLocalStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application dir: %w", err)
	}

	clientConn, err := client.NewGRPCClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network client: %w", err)
	}

	us := service.NewUserService(localStorage, cryptoService, adapter.NewUserGRPCAdapter(clientConn))

	return &GophKeeper{
		Config:      cfg,
		UserService: us,
	}, nil
}

func (gk *GophKeeper) Close() error {
	if gk.clientConn != nil {
		return gk.clientConn.Close()
	}
	return nil
}
