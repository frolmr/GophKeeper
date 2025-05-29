package adapter

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// SendRegisterRequest creates a new user account.
//
// Parameters:
//
//	email - User's email address
//	password - Plaintext password (will be hashed server-side)
//	mk - Master key for encryption operations
//
// Returns:
//
//	error - Wrapped gRPC error if registration fails
func (a *GRPCAdapter) SendRegisterRequest(email, password string, mk []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.RegisterUserRequest{
		Email:    &email,
		Password: &password,
		Mk:       mk,
	}

	_, err := a.userClient.RegisterUser(ctx, req)
	if err != nil {
		return fmt.Errorf("register request failed: %w", err)
	}

	return nil
}

// SendLoginRequest authenticates a user and retrieves an auth token.
//
// Parameters:
//
//	email - User's email address
//	password - Plaintext password
//
// Returns:
//
//	string - Authorization token for subsequent requests
//	error - Wrapped gRPC error if authentication fails
func (a *GRPCAdapter) SendLoginRequest(email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	var header metadata.MD
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs())

	req := &pb.LoginUserRequest{
		Email:    &email,
		Password: &password,
	}

	_, err := a.userClient.LoginUser(ctx, req, grpc.Header(&header))
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}

	authHeaders := header.Get("authorization")
	if len(authHeaders) > 0 {
		return authHeaders[0], nil
	} else {
		return "", errors.New("server did not return authorization token")
	}
}
