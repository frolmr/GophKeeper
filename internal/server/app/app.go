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

type App struct {
	config *config.AppConfig
	logger *zap.SugaredLogger
	api    *api.API
}

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

func setupLogger() (*zap.SugaredLogger, error) {
	l, err := zap.NewDevelopment()

	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

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
