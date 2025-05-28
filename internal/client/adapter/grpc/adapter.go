// Package adapter provides gRPC client implementations for interacting with
// GophKeeper's remote services.
//
// The package implements:
// - User authentication and registration
// - Device management and approval
// - Secure record operations (CRUD)
//
// All operations are timeout-bound (5 seconds by default) and handle
// gRPC-specific error conversion to domain errors.
//
// Usage:
//
//	adapter := NewGRPCAdapter(conn) // conn is *grpc.ClientConn
//	err := adapter.SendLoginRequest(email, password)
//
// Security Note:
// All methods handle sensitive data and should only be used with secure
// connections (TLS). The adapter doesn't implement retry logic - that should
// be handled at a higher layer.
package adapter

import (
	"time"

	devices "github.com/frolmr/GophKeeper/pkg/proto/devices"
	records "github.com/frolmr/GophKeeper/pkg/proto/records"
	users "github.com/frolmr/GophKeeper/pkg/proto/users"
	"google.golang.org/grpc"
)

const (
	// requestTimeout defines the maximum duration for gRPC requests
	requestTimeout = 5 * time.Second
)

// GRPCAdapter implements client-side gRPC communication with GophKeeper services.
// It maintains separate clients for each service domain (users, devices, records).
//
// The adapter converts between domain types and protocol buffer messages,
// handles timeouts, and transforms gRPC errors to domain-appropriate errors.
type GRPCAdapter struct {
	userClient   users.UsersClient
	deviceClient devices.DevicesClient
	recordClient records.RecordsClient
}

// NewGRPCAdapter creates a new adapter instance with initialized gRPC clients.
//
// Parameters:
//
//	clientConn - Established gRPC connection (must be non-nil)
//
// Returns:
//
//	*GRPCAdapter ready for service operations
func NewGRPCAdapter(clientConn *grpc.ClientConn) *GRPCAdapter {
	uc := users.NewUsersClient(clientConn)
	dc := devices.NewDevicesClient(clientConn)
	rc := records.NewRecordsClient(clientConn)

	return &GRPCAdapter{
		userClient:   uc,
		deviceClient: dc,
		recordClient: rc,
	}
}
