package interceptors

import (
	"context"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DeviceInterceptor injects device identification into gRPC requests.
//
// Adds headers:
// - device_id: Unique device identifier
// - device_name: Human-readable device name
type DeviceInterceptor struct {
	device *domain.Device
}

// NewDeviceInterceptor creates a new device interceptor instance.
//
// Parameters:
//
//	device - Device information to inject (nil-safe)
func NewDeviceInterceptor(device *domain.Device) *DeviceInterceptor {
	if device == nil {
		return nil
	}
	return &DeviceInterceptor{
		device: device,
	}
}

// Unary returns a unary client interceptor that injects device headers.
//
// The interceptor:
// 1. Adds device_id and device_name headers to all requests
// 2. Propagates the context with device metadata
func (i *DeviceInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"device_id", i.device.ID,
			"device_name", i.device.Name,
		)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
