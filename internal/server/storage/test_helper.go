package storage

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func NewMock() (*sql.DB, sqlmock.Sqlmock, *zap.SugaredLogger) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	logger := zap.NewNop().Sugar()
	return db, mock, logger
}
