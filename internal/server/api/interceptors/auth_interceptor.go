package interceptors

import (
	"context"
	"strings"

	"github.com/frolmr/GophKeeper/internal/server/api/auth"
	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTManager interface {
	VerifyAccessToken(tokenString string) (*auth.Claims, error)
}

type UsersRepository interface {
	GetUserByUUID(ctx context.Context, uuid uuid.UUID) (*domain.User, error)
}

// AuthInterceptor validates JWT tokens and injects user context.
// Exempts auth for registration/login endpoints.
type AuthInterceptor struct {
	jwtManager JWTManager
	repo       UsersRepository
}

// NewAuthInterceptor creates a new auth interceptor.
func NewAuthInterceptor(jwtManager JWTManager, repo UsersRepository) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
		repo:       repo,
	}
}

// Unary implements gRPC unary interceptor for JWT validation:
// 1. Checks for exempted methods
// 2. Extracts token from metadata
// 3. Validates token
// 4. Looks up user
// 5. Injects user into context
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

		user, err := i.repo.GetUserByUUID(ctx, claims.UserUUID)
		if err != nil || user == nil {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		ctx = context.WithValue(ctx, contextkeys.UserKey, user)

		return handler(ctx, req)
	}
}
