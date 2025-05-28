// Package storage handles secure persistence of application data including:
// - Authentication tokens
// - Cryptographic keys
// - Application configuration
//
// The package provides:
// - Encrypted local storage using bbolt DB
// - Secure file permissions
// - Standardized storage locations
//
// Security Features:
// - Strict file permissions (0600 for files, 0700 for directories)
// - Secure token storage
// - Proper key material handling
package storage

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"go.etcd.io/bbolt"
)

// Permission constants for secure file handling
const (
	keyFilePermissionsMask = 0600 // rw------- for private key files
	appDirPermissionsMask  = 0700 // rwx------ for application directory
	dbPersmissionsMask     = 0600 // rw------- for database file

	dbFileName = "gk.db" // Database filename
	tokenKey   = "jwt"   // Key for JWT token storage
	authBucket = "auth"  // Bucket name for authentication data
)

// Storage manages persistent application data including:
// - Authentication tokens (in bbolt DB)
// - Private keys (in encrypted files)
type Storage struct {
	appDir string
	db     *bbolt.DB
}

// NewStorage initializes the storage system by:
// 1. Creating the application config directory
// 2. Opening/creating the database file
// 3. Setting up required buckets
//
// Returns:
//
//	*Storage - Initialized storage instance
//	error    - Directory or database initialization errors
func NewStorage() (*Storage, error) {
	appDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get app directory: %w", err)
	}

	dbFilePath := filepath.Join(appDir, dbFileName)
	db, err := bbolt.Open(dbFilePath, dbPersmissionsMask, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(authBucket))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user bucket: %w", err)
	}

	return &Storage{
		appDir: appDir,
		db:     db,
	}, nil
}

// Close cleanly shuts down the storage system by closing the database.
// Should be called when the application exits.
//
// Returns:
//
//	error - Database close errors
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveToken securely stores a JWT token in the database.
//
// Parameters:
//
//	token - JWT token string to store
//
// Returns:
//
//	error - Storage errors or empty token validation
//
// Security:
// - Token is stored in encrypted BoltDB bucket
// - Validates token is not empty
func (s *Storage) SaveToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	err := s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(authBucket))
		return bucket.Put([]byte(tokenKey), []byte(token))
	})

	if err != nil {
		log.Printf("failed to save token: %v", err)
		return err
	}
	return nil
}

// GetToken retrieves the stored JWT token from the database.
//
// Returns:
//
//	string - The stored token
//	error  - Retrieval errors or missing token
func (s *Storage) GetToken() (string, error) {
	var token string

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(authBucket))
		if bucket == nil {
			return errors.New("auth bucket not found")
		}

		data := bucket.Get([]byte(tokenKey))
		if data == nil {
			return errors.New("token not found")
		}

		token = string(data)
		return nil
	})

	if err != nil {
		log.Printf("failed to get token: %v", err)
		return "", err
	}

	return token, nil
}

// SavePrivateKey writes RSA private key data to a secure file.
//
// Parameters:
//
//	data - PEM encoded private key bytes
//
// Returns:
//
//	error - File operation errors
//
// Security:
// - File is created with 0600 permissions
// - Existing file is truncated
func (s *Storage) SavePrivateKey(data []byte) error {
	filePath := filepath.Join(s.appDir, domain.PrivKeyFileName)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, keyFilePermissionsMask)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// ReadPrivateKey reads the stored private key from disk.
//
// Returns:
//
//	[]byte - PEM encoded private key
//	error  - File read errors
func (s *Storage) ReadPrivateKey() ([]byte, error) {
	filePath := filepath.Join(s.appDir, domain.PrivKeyFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	return data, nil
}

func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, domain.AppName)
	if err := os.MkdirAll(appDir, appDirPermissionsMask); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return appDir, nil
}
