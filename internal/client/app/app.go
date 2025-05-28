// Package app provides the core application structure and initialization
// for the GophKeeper client.
//
// The package:
// - Manages application lifecycle and dependencies
// - Coordinates between storage, network, and crypto components
// - Provides the main entry point for service initialization
//
// Usage:
//
//	app, err := app.NewApp(true, true) // needs auth and crypto
//	defer app.Close()
//	// Use app services...
package app

import (
	"fmt"

	adapter "github.com/frolmr/GophKeeper/internal/client/adapter/grpc"
	"github.com/frolmr/GophKeeper/internal/client/client"
	"github.com/frolmr/GophKeeper/internal/client/config"
	"github.com/frolmr/GophKeeper/internal/client/crypto"
	"github.com/frolmr/GophKeeper/internal/client/storage"
	"google.golang.org/grpc"
)

// App is the central application container that holds all service dependencies.
// It manages:
// - Local storage (credentials, keys)
// - gRPC network connections
// - Cryptographic operations
type App struct {
	Storage       *storage.Storage
	conn          *grpc.ClientConn
	ConnAdapter   *adapter.GRPCAdapter
	CryptoService *crypto.CryptoService
}

// NewApp initializes the application with required services.
//
// Parameters:
//
//	needAuth   - If true, requires valid authentication token
//	needCrypto - If true, initializes cryptographic services
//
// Returns:
//
//	*App - Configured application instance
//	error - Initialization errors (config, storage, or network)
//
// Note:
// The connection is established immediately but services are lazy-loaded.
func NewApp(needAuth, needCrypto bool) (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	stor, err := storage.NewStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	token := ""
	if needAuth {
		token, err = stor.GetToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get token! Assume need to login first")
		}
	}

	clientConn, err := client.NewGRPCClient(cfg.ServerAddress, token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize network client: %v", err)
	}

	connAdapter := adapter.NewGRPCAdapter(clientConn)

	var crps *crypto.CryptoService
	if needCrypto {
		crps = crypto.NewCryptoService()
	}

	return &App{
		Storage:       stor,
		conn:          clientConn,
		ConnAdapter:   connAdapter,
		CryptoService: crps,
	}, nil
}

// Close gracefully shuts down application resources including:
// - Network connections
// - Storage handles
//
// Returns:
//
//	error - Any errors encountered during shutdown
func (a *App) Close() error {
	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			return err
		}
	}
	if a.Storage != nil {
		if err := a.Storage.Close(); err != nil {
			return err
		}
	}
	return nil
}
