// Package config handles application configuration management for the GophKeeper client.
//
// The package:
// - Manages environment-based configuration
// - Validates required settings
// - Provides type-safe access to configuration values
//
// Configuration Sources:
// - Environment variables (primary)
// - Future expansion for config files supported
package config

import (
	"errors"
	"os"
)

// serverAddressEnvName defines the environment variable name for the server address.
const (
	serverAddressEnvName = "GOPHKEEPER_SERVER"
)

// Config holds all application configuration values.
// Currently manages:
// - ServerAddress: The gRPC server endpoint (host:port)
type Config struct {
	ServerAddress string
}

// NewConfig creates and validates a new configuration instance.
//
// Reads configuration from environment variables:
// - GOPHKEEPER_SERVER: Required server address (e.g., "localhost:50051")
//
// Returns:
//
//	*Config - Validated configuration
//	error   - If required configuration is missing, with clear error message
//
// Example:
//
//	cfg, err := config.NewConfig()
//	if err != nil {
//	    log.Fatal("Configuration error:", err)
//	}
func NewConfig() (*Config, error) {
	serverAddress := os.Getenv(serverAddressEnvName)
	if serverAddress == "" {
		return nil, errors.New("missing server address")
	}

	return &Config{
		ServerAddress: serverAddress,
	}, nil
}
