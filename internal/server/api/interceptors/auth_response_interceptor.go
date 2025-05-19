package interceptors

import (
	"context"

	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthResponseInterceptor struct {
	jwtManager JWTManager
}

func NewAuthResponseInterceptor(jwtManager JWTManager) *AuthResponseInterceptor {
	return &AuthResponseInterceptor{
		jwtManager: jwtManager,
	}
}

func (i *AuthResponseInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		if err == nil && (info.FullMethod == "/users.Users/RegisterUser" || info.FullMethod == "/users.Users/LoginUser") {
			var email string
			switch r := req.(type) {
			case *pb.RegisterUserRequest:
				email = *r.Email
			case *pb.LoginUserRequest:
				email = *r.Email
			}

			if email != "" {
				token, err := i.jwtManager.GenerateAccessToken(email)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "can't generate auth token: %v", err)
				}
				header := metadata.Pairs("authorization", "Bearer "+token)
				if err := grpc.SetHeader(ctx, header); err != nil {
					return nil, status.Errorf(codes.Internal, "can't set auth header: %v", err)
				}
			} else {
				return nil, status.Error(codes.Unauthenticated, "invalid email")
			}
		}

		return resp, err
	}
}
