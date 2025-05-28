package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	pb "github.com/frolmr/GophKeeper/pkg/proto/devices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendAddDeviceRequest registers a new device with the server.
//
// Parameters:
//
//	pk - Device public key to register
//
// Returns:
//
//	error - nil on success, or:
//	  - ErrDeviceExists if device already registered
//	  - wrapped gRPC error for other failures
func (a *GRPCAdapter) SendAddDeviceRequest(pk []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.AddDeviceRequest{
		Pk: pk,
	}

	_, err := a.deviceClient.AddDevice(ctx, req)
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.AlreadyExists {
			return errors.New("device already registered")
		}
	}
	if err != nil {
		return fmt.Errorf("add device request failed: %w", err)
	}

	return nil
}

// SendListDevicesRequest retrieves all devices associated with the current user.
//
// Returns:
//
//	[]domain.DeviceOnServer - List of registered devices
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendListDevicesRequest() ([]domain.DeviceOnServer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.ListDevicesRequest{}
	resp, err := a.deviceClient.ListDevices(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list devices response error: %v", err)
	}

	var devices []domain.DeviceOnServer
	for _, respDevice := range resp.Devices {
		device := domain.DeviceOnServer{
			ID:        *respDevice.Uuid,
			Name:      *respDevice.Name,
			Confirmed: *respDevice.Confirmed,
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// SendGetDevicePKRequest retrieves a device's public key by ID.
//
// Parameters:
//
//	id - UUID of the target device
//
// Returns:
//
//	[]byte - The device's public key
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendGetDevicePKRequest(id string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.GetDevicePKRequest{
		Uuid: &id,
	}

	resp, err := a.deviceClient.GetDevicePK(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get devices PK response error: %v", err)
	}

	return resp.GetPk(), nil
}

// SendApproveDeviceRequest confirms a new device registration.
//
// Parameters:
//
//	id - UUID of the device to approve
//	mk - Master key for encryption operations
//
// Returns:
//
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendApproveDeviceRequest(id string, mk []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.ApproveDeviceRequest{
		Uuid: &id,
		Mk:   mk,
	}

	_, err := a.deviceClient.ApproveDevice(ctx, req)
	if err != nil {
		return fmt.Errorf("approve devices response error: %v", err)
	}

	return nil
}

// SendGetDeviceMKRequest retrieves the master key for the current device.
//
// Returns:
//
//	[]byte - The device's master key
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendGetDeviceMKRequest() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := &pb.GetDeviceMKRequest{}

	resp, err := a.deviceClient.GetDeviceMK(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	return resp.GetMk(), nil
}
