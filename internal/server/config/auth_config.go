package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
)

type AuthConfig struct {
	JWTKey                  []byte
	JWTAccessTokenExpiresIn time.Duration
}

const (
	jwtKeyEnvName        = "JWT_SECRET"
	jwtAccessTokenExpiry = 15 * time.Minute
	jwtSecureLength      = 32
)

var (
	ErrMissingJwtKey = errors.New("missing jwt secret")
)

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

func generateSecureSecret() (string, error) {
	secret := make([]byte, jwtSecureLength)

	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encodedSecret := base64.URLEncoding.EncodeToString(secret)

	return encodedSecret, nil
}
