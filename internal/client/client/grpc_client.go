package client

import (
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGRPCClient(cfg *config.Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	creds := credentials.NewClientTLSFromCert(nil, "")
	opts = append(opts, grpc.WithTransportCredentials(creds))

	clientConn, err := grpc.NewClient(cfg.ServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return clientConn, nil
}
