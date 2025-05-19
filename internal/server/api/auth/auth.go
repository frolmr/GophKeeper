package auth

import (
	"fmt"
	"time"

	"github.com/frolmr/GophKeeper/internal/server/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserEmail string `json:"email"`
	jwt.RegisteredClaims
}

type AuthService struct {
	authConfig *config.AuthConfig
}

func NewAuthService(cfg *config.AuthConfig) *AuthService {
	return &AuthService{
		authConfig: cfg,
	}
}

func (as *AuthService) GenerateAccessToken(userEmail string) (string, error) {
	expirationTime := time.Now().Add(as.authConfig.JWTAccessTokenExpiresIn)

	return generateToken(userEmail, expirationTime, as.authConfig.JWTKey)
}

func generateToken(userEmail string, expirationTime time.Time, key []byte) (string, error) {
	claims := &Claims{
		UserEmail: userEmail,
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
