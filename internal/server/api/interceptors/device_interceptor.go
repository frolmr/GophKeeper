package interceptors

import (
	"context"
	"strings"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type DeviceRepository interface {
	GetDeviceByUserUUIDAndID(ctx context.Context, id string, userUUID uuid.UUID) (*domain.Device, error)
}

// DeviceInterceptor validates and manages device context.
// Handles three scenarios:
// 1. Registration: Stores new device info
// 2. Login: Bypasses device check
// 3. Other operations: Validates confirmed device
type DeviceInterceptor struct {
	repo DeviceRepository
}

// NewDeviceInterceptor creates a new device interceptor.
func NewDeviceInterceptor(repo DeviceRepository) *DeviceInterceptor {
	return &DeviceInterceptor{
		repo: repo,
	}
}

// Unary implements gRPC unary interceptor for device verification.
func (i *DeviceInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		deviceIDs := md.Get("device_id")
		deviceNames := md.Get("device_name")
		if len(deviceIDs) == 0 || len(deviceNames) == 0 {
			return nil, status.Error(codes.Unauthenticated, "device information is not provided")
		}

		switch info.FullMethod {
		case "/users.Users/RegisterUser", "/devices.Devices/AddDevice":
			ctx = context.WithValue(ctx, contextkeys.DeviceIDKey, strings.TrimSpace(deviceIDs[0]))
			ctx = context.WithValue(ctx, contextkeys.DeviceNameKey, strings.TrimSpace(deviceNames[0]))
		case "/users.Users/LoginUser":
			return handler(ctx, req)
		default:
			user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
			if !ok {
				return nil, status.Error(codes.NotFound, "user not found")
			}
			device, err := i.repo.GetDeviceByUserUUIDAndID(ctx, strings.TrimSpace(deviceIDs[0]), user.UUID)
			if err != nil || device == nil || !device.Confirmed {
				return nil, status.Error(codes.Unauthenticated, "reqesting device is unothorized")
			}
			ctx = context.WithValue(ctx, contextkeys.DeviceKey, device)
		}

		return handler(ctx, req)
	}
}
