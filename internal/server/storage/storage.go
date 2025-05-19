package storage

import (
	"database/sql"

	"go.uber.org/zap"
)

type Storage struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewStorage(db *sql.DB, lgr *zap.SugaredLogger) *Storage {
	return &Storage{
		db:     db,
		logger: lgr,
	}
}
