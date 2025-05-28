package auth

import (
	"testing"
	"time"

	"github.com/frolmr/GophKeeper/internal/server/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTKey:                  []byte("test-secret-key"),
		JWTAccessTokenExpiresIn: 15 * time.Minute,
	}

	service := NewAuthService(cfg)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.authConfig)
}

func TestGenerateAccessToken(t *testing.T) {
	testCases := []struct {
		name      string
		userUUID  uuid.UUID
		expiry    time.Duration
		expectErr bool
	}{
		{
			name:      "Valid token generation",
			userUUID:  uuid.New(),
			expiry:    15 * time.Minute,
			expectErr: false,
		},
		{
			name:      "Empty UUID",
			userUUID:  uuid.Nil,
			expiry:    15 * time.Minute,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.AuthConfig{
				JWTKey:                  []byte("test-secret-key"),
				JWTAccessTokenExpiresIn: tc.expiry,
			}
			service := NewAuthService(cfg)

			token, err := service.GenerateAccessToken(tc.userUUID)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestVerifyAccessToken(t *testing.T) {
	validUserUUID := uuid.New()
	validKey := []byte("valid-secret-key")
	invalidKey := []byte("invalid-secret-key")

	validToken, err := generateToken(validUserUUID, time.Now().Add(15*time.Minute), validKey)
	require.NoError(t, err)

	invalidToken, err := generateToken(validUserUUID, time.Now().Add(15*time.Minute), invalidKey)
	require.NoError(t, err)

	expiredToken, err := generateToken(validUserUUID, time.Now().Add(-15*time.Minute), validKey)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		tokenString string
		serviceKey  []byte
		expectErr   bool
		expectClaim *Claims
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			serviceKey:  validKey,
			expectErr:   false,
			expectClaim: &Claims{UserUUID: validUserUUID},
		},
		{
			name:        "Invalid signature",
			tokenString: invalidToken,
			serviceKey:  validKey,
			expectErr:   true,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			serviceKey:  validKey,
			expectErr:   true,
		},
		{
			name:        "Malformed token",
			tokenString: "malformed.token.string",
			serviceKey:  validKey,
			expectErr:   true,
		},
		{
			name:        "Empty token",
			tokenString: "",
			serviceKey:  validKey,
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.AuthConfig{
				JWTKey:                  tc.serviceKey,
				JWTAccessTokenExpiresIn: 15 * time.Minute,
			}
			service := NewAuthService(cfg)

			claims, err := service.VerifyAccessToken(tc.tokenString)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tc.expectClaim.UserUUID, claims.UserUUID)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	testUserUUID := uuid.New()
	validKey := []byte("test-secret-key")
	emptyKey := []byte("")

	testCases := []struct {
		name      string
		userUUID  uuid.UUID
		expiry    time.Time
		key       []byte
		expectErr bool
	}{
		{
			name:      "Valid token generation",
			userUUID:  testUserUUID,
			expiry:    time.Now().Add(15 * time.Minute),
			key:       validKey,
			expectErr: false,
		},
		{
			name:      "Empty secret key",
			userUUID:  testUserUUID,
			expiry:    time.Now().Add(15 * time.Minute),
			key:       emptyKey,
			expectErr: true,
		},
		{
			name:      "Expired token generation",
			userUUID:  testUserUUID,
			expiry:    time.Now().Add(-15 * time.Minute),
			key:       validKey,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := generateToken(tc.userUUID, tc.expiry, tc.key)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
