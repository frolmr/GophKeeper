// Package client provides gRPC client initialization and interceptors
// for the GophKeeper application.
//
// Key Components:
// - gRPC connection establishment with TLS
// - Authentication interceptors
// - Device identification interceptors
//
// Security Features:
// - Mandatory TLS encryption
// - JWT token handling
// - Device identification headers
package client

import (
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/client/interceptors"
	"github.com/frolmr/GophKeeper/internal/client/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewGRPCClient creates a secure gRPC client connection with interceptors.
//
// Parameters:
//
//	serverAddress - gRPC server endpoint (host:port)
//	token         - JWT authentication token (empty for no auth)
//
// Returns:
//
//	*grpc.ClientConn - Established gRPC connection
//	error            - Connection or initialization errors
//
// Features:
// - Enforces TLS encryption
// - Adds device identification headers
// - Conditionally adds auth token if provided
// - Excludes auth for registration/login endpoints
func NewGRPCClient(serverAddress, token string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	creds := credentials.NewClientTLSFromCert(nil, "")
	opts = append(opts, grpc.WithTransportCredentials(creds))

	thisDevice, err := domain.ThisDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch device information: %w", err)
	}

	deviceInterceptor := interceptors.NewDeviceInterceptor(thisDevice)
	opts = append(opts, grpc.WithUnaryInterceptor(deviceInterceptor.Unary()))

	if token != "" {
		authInterceptor := interceptors.NewAuthInterceptor(token)
		opts = append(opts, grpc.WithChainUnaryInterceptor(authInterceptor.Unary()))
	}

	clientConn, err := grpc.NewClient(serverAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to init connect to the server: %w", err)
	}

	return clientConn, nil
}
