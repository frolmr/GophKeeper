// Package app provides the core server application setup and lifecycle management
// for the GophKeeper server.
//
// Responsibilities:
// - Application initialization (config, logging, database)
// - Dependency wiring (storage, API)
// - Graceful shutdown handling
// - Error management
//
// Usage:
//
//	ctx := context.Background()
//	app, err := app.NewServerApp()
//	if err != nil {
//	    log.Fatal("Failed to initialize app:", err)
//	}
//	if err := app.Run(ctx); err != nil {
//	    log.Fatal("Server error:", err)
//	}
package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/server/api"
	"github.com/frolmr/GophKeeper/internal/server/config"
	migrator "github.com/frolmr/GophKeeper/internal/server/db"
	"github.com/frolmr/GophKeeper/internal/server/storage"
	"go.uber.org/zap"
)

// App is the central server application container that manages:
// - Configuration
// - Logging
// - Database connection
// - API server
// - Application lifecycle
type App struct {
	config *config.AppConfig
	logger *zap.SugaredLogger
	api    *api.API
}

// NewServerApp initializes and wires all application components:
// 1. Loads configuration
// 2. Sets up logging
// 3. Initializes database connection and runs migrations
// 4. Creates storage layer
// 5. Configures API server
//
// Returns:
//
//	*App - Fully initialized application instance
//	error - Initialization errors with context
func NewServerApp() (*App, error) {
	cfg, err := config.NewAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to setup application config: %w", err)
	}

	lgr, err := setupLogger()
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	db, err := setupDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	stor := storage.NewStorage(db, lgr)

	srv, err := api.NewAPI(cfg, stor, lgr)
	if err != nil {
		return nil, fmt.Errorf("failed to setup api: %w", err)
	}

	return &App{
		config: cfg,
		logger: lgr,
		api:    srv,
	}, nil
}

// Run starts the application and manages its lifecycle:
// - Launches the API server in a goroutine
// - Handles graceful shutdown on context cancellation
// - Manages error propagation from components
//
// Parameters:
//
//	ctx - Context for graceful shutdown handling
//
// Returns:
//
//	error - Runtime errors from server components
func (app *App) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := app.api.Run(ctx); err != nil {
			errChan <- fmt.Errorf("API server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		app.logger.Info("Shutting down due to interrupt signal")
		return nil
	case err := <-errChan:
		return err
	}
}

// setupLogger configures the application logger with development settings.
// Production deployments should use NewProduction() instead.
//
// Returns:
//
//	*zap.SugaredLogger - Configured logger
//	error - Logger initialization errors
func setupLogger() (*zap.SugaredLogger, error) {
	l, err := zap.NewDevelopment()

	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

// setupDB initializes the database connection and runs migrations:
// 1. Opens connection to PostgreSQL using pgx driver
// 2. Applies database migrations
// 3. Verifies connection
//
// Parameters:
//
//	conf - Application configuration containing database URI
//
// Returns:
//
//	*sql.DB - Database connection handle
//	error - Connection or migration errors
func setupDB(conf *config.AppConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", conf.DatabaseURI)
	if err != nil {
		return nil, err
	}

	m := migrator.NewMigrator(conf.DatabaseURI)
	if err := m.RunMigrations(); err != nil {
		return nil, err
	}

	return db, nil
}
