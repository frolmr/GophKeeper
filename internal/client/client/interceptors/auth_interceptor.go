// Package interceptors implements gRPC client interceptors for:
// - Authentication (JWT tokens)
// - Device identification
//
// Interceptors automatically modify outgoing requests with required headers
// and handle special cases for auth-exempt endpoints.
package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor injects JWT tokens into gRPC requests.
//
// Exempts authentication for:
// - User registration (/users.Users/RegisterUser)
// - User login (/users.Users/LoginUser)
type AuthInterceptor struct {
	token string
}

// NewAuthInterceptor creates a new auth interceptor instance.
//
// Parameters:
//
//	token - JWT token to inject into requests
func NewAuthInterceptor(token string) *AuthInterceptor {
	return &AuthInterceptor{
		token: token,
	}
}

// Unary returns a unary client interceptor that injects the auth token.
//
// The interceptor:
// 1. Skips auth for whitelisted methods
// 2. Adds Authorization header to all other requests
// 3. Propagates the context with auth metadata
func (i *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if method == "/users.Users/RegisterUser" || method == "/users.Users/LoginUser" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"authorization", i.token,
		)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
