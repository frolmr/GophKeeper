package interceptors

import (
	"context"
	"strings"

	"github.com/frolmr/GophKeeper/internal/server/api/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTManager interface {
	GenerateAccessToken(userID string) (string, error)
	VerifyAccessToken(tokenString string) (*auth.Claims, error)
}

type AuthInterceptor struct {
	jwtManager JWTManager
}

func NewAuthInterceptor(jwtManager JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
	}
}

type claimsKey string

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if info.FullMethod == "/users.Users/RegisterUser" || info.FullMethod == "/users.Users/LoginUser" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := strings.TrimPrefix(values[0], "Bearer ")
		claims, err := i.jwtManager.VerifyAccessToken(accessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
		}

		ck := claimsKey("claims")
		ctx = context.WithValue(ctx, ck, claims)

		return handler(ctx, req)
	}
}
