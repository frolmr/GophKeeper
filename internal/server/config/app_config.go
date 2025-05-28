// Package config handles application configuration management for the GophKeeper server.
//
// The package provides:
// - Centralized configuration loading from multiple sources (env vars, flags)
// - Type-safe configuration access
// - Validation of required settings
// - Secure credential generation
//
// Configuration Sources:
// 1. Environment variables (primary)
// 2. Command-line flags (secondary)
// 3. Secure defaults for sensitive values
package config

import (
	"errors"
	"flag"
	"os"
)

// AppConfig contains core application configuration parameters.
type AppConfig struct {
	RunAddress  string // Server listen address (host:port)
	DatabaseURI string // Database connection URI
	TLSCertFile string // Path to TLS certificate file
	TLSKeyFile  string // Path to TLS private key file
}

// Environment variable names
const (
	runAddressEnvName    = "RUN_ADDRESS"  // Env var for server address
	dadatabaseURIEnvName = "DATABASE_URI" // Env var for database URI
)

// Configuration validation errors
var (
	ErrMissingAddress  = errors.New("missing run address")
	ErrMissingDDURI    = errors.New("missing database URI")
	ErrMissingCertFile = errors.New("missing TLS certificate file")
	ErrMissingKeyFile  = errors.New("missing TLS key file")
)

// NewAppConfig loads and validates the application configuration from:
// 1. Command-line flags (-a, -d, -c, -k)
// 2. Environment variables (fallback)
//
// Returns:
//
//	*AppConfig - Validated configuration
//	error - Aggregated validation errors
//
// Example:
//
//	cfg, err := config.NewAppConfig()
//	if err != nil {
//	    log.Fatal("Configuration error:", err)
//	}
func NewAppConfig() (*AppConfig, error) {
	var (
		runAddress  string
		databaseURI string
		certFile    string
		keyFile     string
		errs        []error
	)

	flag.StringVar(&runAddress, "a", runAddress, "sets host and port to run")
	flag.StringVar(&databaseURI, "d", databaseURI, "set database URI to use")
	flag.StringVar(&certFile, "c", certFile, "sets TLS certificate file")
	flag.StringVar(&keyFile, "k", keyFile, "set TLS key file")
	flag.Parse()

	if runAddressEnv := os.Getenv(runAddressEnvName); runAddressEnv != "" && runAddress == "" {
		runAddress = runAddressEnv
	}

	if databaseURIEnv := os.Getenv(dadatabaseURIEnvName); databaseURIEnv != "" && databaseURI == "" {
		databaseURI = databaseURIEnv
	}

	if runAddress == "" {
		errs = append(errs, ErrMissingAddress)
	}

	if databaseURI == "" {
		errs = append(errs, ErrMissingDDURI)
	}

	if certFile == "" {
		errs = append(errs, ErrMissingCertFile)
	}

	if keyFile == "" {
		errs = append(errs, ErrMissingKeyFile)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return &AppConfig{
		RunAddress:  runAddress,
		DatabaseURI: databaseURI,
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
	}, nil
}
