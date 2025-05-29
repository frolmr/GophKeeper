package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
)

// AuthConfig contains authentication-related configuration.
type AuthConfig struct {
	JWTKey                  []byte        // Secret key for JWT signing
	JWTAccessTokenExpiresIn time.Duration // Token expiration duration
}

// Constants for JWT configuration
const (
	jwtKeyEnvName        = "JWT_SECRET"     // Env var for JWT secret
	jwtAccessTokenExpiry = 15 * time.Minute // Default token lifetime
	jwtSecureLength      = 32               // Bytes for generated secrets
)

var (
	ErrMissingJwtKey = errors.New("missing jwt secret")
)

// NewAuthConfig loads JWT configuration with secure defaults:
// 1. Checks for JWT_SECRET environment variable
// 2. Generates secure random secret if not provided
//
// Security Note:
// - Generated secrets are cryptographically random
// - Secrets are base64-encoded for storage
//
// Returns:
//
//	*AuthConfig - Initialized auth configuration
//	error - Secret generation errors
func NewAuthConfig() (*AuthConfig, error) {
	jwtSecret := os.Getenv(jwtKeyEnvName)

	if jwtSecret == "" {
		var err error
		jwtSecret, err = generateSecureSecret()
		if err != nil {
			return nil, ErrMissingJwtKey
		}
	}

	return &AuthConfig{
		JWTKey:                  []byte(jwtSecret),
		JWTAccessTokenExpiresIn: jwtAccessTokenExpiry,
	}, nil
}

// generateSecureSecret creates a cryptographically secure random secret.
// Used when no JWT secret is provided via environment.
//
// Returns:
//
//	string - Base64-encoded random bytes
//	error - Random number generation errors
func generateSecureSecret() (string, error) {
	secret := make([]byte, jwtSecureLength)

	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encodedSecret := base64.URLEncoding.EncodeToString(secret)

	return encodedSecret, nil
}
