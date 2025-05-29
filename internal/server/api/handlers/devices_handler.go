package handlers

import (
	"context"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/devices"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type DevicesRepository interface {
	GetDeviceByUserUUIDAndID(ctx context.Context, id string, userUUID uuid.UUID) (*domain.Device, error)
	GetDeviceByUserUUIDAndUUID(ctx context.Context, deviceUUID, userUUID uuid.UUID) (*domain.Device, error)
	AddDevice(ctx context.Context, name, id string, pk []byte, user *domain.User) error
	GetAllUserDevices(ctx context.Context, userUUID uuid.UUID) ([]*domain.Device, error)
	ConfirmDevice(ctx context.Context, deviceUUID, userUUID uuid.UUID, mk []byte) error
}

// DevicesService implements the gRPC Devices service for:
// - Device registration
// - Device management
// - Master key distribution
type DevicesService struct {
	pb.UnimplementedDevicesServer
	repo   DevicesRepository
	logger *zap.SugaredLogger
}

// NewDevicesService creates a new devices service handler.
func NewDevicesService(repo DevicesRepository, lgr *zap.SugaredLogger) *DevicesService {
	return &DevicesService{
		repo:   repo,
		logger: lgr,
	}
}

// AddDevice registers a new device for the authenticated user.
// Validates device ID uniqueness before storage.
func (ds *DevicesService) AddDevice(ctx context.Context, in *pb.AddDeviceRequest) (*emptypb.Empty, error) {
	name, ok := ctx.Value(contextkeys.DeviceNameKey).(string)
	if !ok {
		return nil, status.Error(codes.NotFound, "device data doesn't provided")
	}

	id, ok := ctx.Value(contextkeys.DeviceIDKey).(string)
	if !ok {
		return nil, status.Error(codes.NotFound, "device data doesn't provided")
	}

	pk := in.GetPk()

	if name == "" || id == "" || pk == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid device name or pk")
	}

	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	existingDevice, err := ds.repo.GetDeviceByUserUUIDAndID(ctx, id, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "get device error: %v", err)
	}

	if existingDevice != nil {
		return nil, status.Error(codes.AlreadyExists, "device already registered")
	}

	if err := ds.repo.AddDevice(ctx, name, id, pk, user); err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't add device: %v", err)
	}

	return nil, nil
}

// ListDevices retrieves all devices for the authenticated user.
func (ds *DevicesService) ListDevices(ctx context.Context, in *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	devices, err := ds.repo.GetAllUserDevices(ctx, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't list devices: %v", err)
	}

	var respDevices []*pb.Device
	for _, device := range devices {
		deviceUUID := device.UUID.String()
		respDevice := pb.Device{
			Uuid:      &deviceUUID,
			Name:      &device.Name,
			Confirmed: &device.Confirmed,
		}
		respDevices = append(respDevices, &respDevice)
	}

	return &pb.ListDevicesResponse{Devices: respDevices}, nil
}

// GetDeviceMK retrieves the master key for the requesting device.
// Requires device to be confirmed.
func (ds *DevicesService) GetDeviceMK(ctx context.Context, in *pb.GetDeviceMKRequest) (*pb.GetDeviceMKResponse, error) {
	device, ok := ctx.Value(contextkeys.DeviceKey).(*domain.Device)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetDeviceMKResponse{Mk: device.MK}, nil
}

// GetDevicePK retrieves the public key of another device.
// Used during device confirmation.
func (ds *DevicesService) GetDevicePK(ctx context.Context, in *pb.GetDevicePKRequest) (*pb.GetDevicePKResponse, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	requestedDeviceUUID, err := uuid.Parse(in.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "can't parse device uuid")
	}

	requestedDevice, err := ds.repo.GetDeviceByUserUUIDAndUUID(ctx, requestedDeviceUUID, user.UUID)
	if err != nil || requestedDevice == nil {
		return nil, status.Error(codes.NotFound, "no devices to confirm")
	}

	if requestedDevice.PK == nil {
		return nil, status.Error(codes.NotFound, "device has no PK")
	}

	return &pb.GetDevicePKResponse{Pk: requestedDevice.PK}, nil
}

// ApproveDevice confirms a new device and shares the master key.
func (ds *DevicesService) ApproveDevice(ctx context.Context, in *pb.ApproveDeviceRequest) (*emptypb.Empty, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	deviceToUpdateUUID, err := uuid.Parse(in.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "can't parse device uuid")
	}

	deviceToUpdate, err := ds.repo.GetDeviceByUserUUIDAndUUID(ctx, deviceToUpdateUUID, user.UUID)
	if err != nil || deviceToUpdate == nil {
		return nil, status.Error(codes.NotFound, "no devices to confirm")
	}

	if err := ds.repo.ConfirmDevice(ctx, deviceToUpdate.UUID, user.UUID, in.Mk); err != nil {
		return nil, status.Error(codes.Internal, "device update failure")
	}

	return nil, nil
}
