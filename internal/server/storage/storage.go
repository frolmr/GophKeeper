// Package storage implements persistent data storage for the GophKeeper server.
//
// The package provides:
// - Database operations for users, devices and records
// - Transactional safety
// - Secure credential handling
// - Error logging and wrapping
//
// Storage Features:
// - User authentication data
// - Device registration and confirmation
// - Encrypted record storage
// - Atomic operations
package storage

import (
	"database/sql"

	"go.uber.org/zap"
)

// Storage is the central data access layer that manages:
// - Database connections
// - Query execution
// - Transaction handling
// - Error logging
type Storage struct {
	db     *sql.DB            // Database connection pool
	logger *zap.SugaredLogger // Structured logger
}

// NewStorage creates a new Storage instance with the given dependencies.
//
// Parameters:
//
//	db  - Active database connection
//	lgr - Logger instance
//
// Returns:
//
//	*Storage - Initialized storage layer
func NewStorage(db *sql.DB, lgr *zap.SugaredLogger) *Storage {
	return &Storage{
		db:     db,
		logger: lgr,
	}
}
