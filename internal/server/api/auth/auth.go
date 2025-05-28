package auth

import (
	"fmt"
	"time"

	"github.com/frolmr/GophKeeper/internal/server/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT token claims containing:
// - UserUUID: Unique user identifier
// - Standard JWT registered claims
type Claims struct {
	UserUUID uuid.UUID `json:"uuid"`
	jwt.RegisteredClaims
}

// AuthService handles JWT token operations:
// - Generation
// - Validation
// - Expiration management
type AuthService struct {
	authConfig *config.AuthConfig
}

// NewAuthService creates a new auth service instance.
//
// Parameters:
//
//	cfg - Auth configuration (secret key, expiration)
func NewAuthService(cfg *config.AuthConfig) *AuthService {
	return &AuthService{
		authConfig: cfg,
	}
}

// GenerateAccessToken creates a new JWT for a user.
//
// Parameters:
//
//	userUUID - Subject of the token
//
// Returns:
//
//	string - Signed JWT
//	error - Token generation errors
func (as *AuthService) GenerateAccessToken(userUUID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(as.authConfig.JWTAccessTokenExpiresIn)

	return generateToken(userUUID, expirationTime, as.authConfig.JWTKey)
}

func generateToken(userUUID uuid.UUID, expirationTime time.Time, key []byte) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("empty secret key")
	}

	claims := &Claims{
		UserUUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error signing: %w", err)
	}

	return tokenString, nil
}

// VerifyAccessToken validates a JWT and extracts claims.
//
// Parameters:
//
//	tokenString - JWT to validate
//
// Returns:
//
//	*Claims - Validated token claims
//	error - Validation errors
func (as *AuthService) VerifyAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return as.authConfig.JWTKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
