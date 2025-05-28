package migrator

import (
	"embed"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const migrationsFolder = "migrations"

type Migrator struct {
	dbURI string
}

func NewMigrator(dbURI string) *Migrator {
	return &Migrator{
		dbURI: dbURI,
	}
}

func (m *Migrator) RunMigrations() error {
	goose.SetBaseFS(embedMigrations)

	db, err := goose.OpenDBWithDriver("pgx", m.dbURI)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, migrationsFolder); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
