package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	mocks "github.com/frolmr/GophKeeper/internal/server/mocks/handlers/devices"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/devices"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDevicesService_AddDevice(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDevicesRepository)
		ctx           context.Context
		input         *pb.AddDeviceRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndID(gomock.Any(), "device-id", userUUID).Return(nil, nil)
				repo.EXPECT().AddDevice(
					gomock.Any(),
					"test-device",
					"device-id",
					[]byte("public-key"),
					user,
				).Return(nil)
			},
			ctx: context.WithValue(
				context.WithValue(
					context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
					contextkeys.DeviceIDKey, "device-id",
				),
				contextkeys.UserKey, user,
			),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: nil,
		},
		{
			name:       "MissingDeviceData",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx:        context.Background(),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: status.Error(codes.NotFound, "device data doesn't provided"),
		},
		{
			name:       "MissingUser",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx: context.WithValue(
				context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
				contextkeys.DeviceIDKey, "device-id",
			),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "InvalidArguments",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx: context.WithValue(
				context.WithValue(
					context.WithValue(context.Background(), contextkeys.DeviceNameKey, ""),
					contextkeys.DeviceIDKey, "",
				),
				contextkeys.UserKey, user,
			),
			input: &pb.AddDeviceRequest{
				Pk: nil,
			},
			expectedError: status.Error(codes.InvalidArgument, "invalid device name or pk"),
		},
		{
			name: "DeviceAlreadyExists",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndID(gomock.Any(), "device-id", userUUID).Return(&domain.Device{}, nil)
			},
			ctx: context.WithValue(
				context.WithValue(
					context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
					contextkeys.DeviceIDKey, "device-id",
				),
				contextkeys.UserKey, user,
			),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: status.Error(codes.AlreadyExists, "device already registered"),
		},
		{
			name: "DatabaseErrorOnGetDevice",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndID(gomock.Any(), "device-id", userUUID).Return(nil, errors.New("db error"))
			},
			ctx: context.WithValue(
				context.WithValue(
					context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
					contextkeys.DeviceIDKey, "device-id",
				),
				contextkeys.UserKey, user,
			),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: status.Errorf(codes.Internal, "get device error: %v", errors.New("db error")),
		},
		{
			name: "DatabaseErrorOnAddDevice",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndID(gomock.Any(), "device-id", userUUID).Return(nil, nil)
				repo.EXPECT().AddDevice(
					gomock.Any(),
					"test-device",
					"device-id",
					[]byte("public-key"),
					user,
				).Return(errors.New("db error"))
			},
			ctx: context.WithValue(
				context.WithValue(
					context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
					contextkeys.DeviceIDKey, "device-id",
				),
				contextkeys.UserKey, user,
			),
			input: &pb.AddDeviceRequest{
				Pk: []byte("public-key"),
			},
			expectedError: status.Errorf(codes.Internal, "can't add device: %v", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDevicesRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewDevicesService(repo, logger)
			_, err := service.AddDevice(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevicesService_ListDevices(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	deviceUUID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDevicesRepository)
		ctx           context.Context
		expectedError error
		expectedCount int
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetAllUserDevices(gomock.Any(), userUUID).Return([]*domain.Device{
					{
						UUID:      deviceUUID,
						Name:      "device-1",
						Confirmed: true,
					},
					{
						UUID:      uuid.New(),
						Name:      "device-2",
						Confirmed: false,
					},
				}, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name:          "UserNotFound",
			setupMocks:    func(repo *mocks.MockDevicesRepository) {},
			ctx:           context.Background(),
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetAllUserDevices(gomock.Any(), userUUID).Return(nil, errors.New("db error"))
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: status.Errorf(codes.Internal, "can't list devices: %v", errors.New("db error")),
		},
		{
			name: "EmptyList",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetAllUserDevices(gomock.Any(), userUUID).Return([]*domain.Device{}, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDevicesRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewDevicesService(repo, logger)
			resp, err := service.ListDevices(tt.ctx, &pb.ListDevicesRequest{})

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.GetDevices(), tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, deviceUUID.String(), *resp.Devices[0].Uuid)
					assert.Equal(t, "device-1", *resp.Devices[0].Name)
					assert.True(t, *resp.Devices[0].Confirmed)
				}
			}
		})
	}
}

func TestDevicesService_GetDeviceMK(t *testing.T) {
	deviceUUID := uuid.New()
	device := &domain.Device{
		UUID: deviceUUID,
		MK:   []byte("master-key"),
	}

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDevicesRepository)
		ctx           context.Context
		expectedError error
		expectedMK    []byte
	}{
		{
			name:       "Success",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx:        context.WithValue(context.Background(), contextkeys.DeviceKey, device),
			expectedMK: []byte("master-key"),
		},
		{
			name:          "DeviceNotFound",
			setupMocks:    func(repo *mocks.MockDevicesRepository) {},
			ctx:           context.Background(),
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDevicesRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewDevicesService(repo, logger)
			resp, err := service.GetDeviceMK(tt.ctx, &pb.GetDeviceMKRequest{})

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMK, resp.Mk)
			}
		})
	}
}

func TestDevicesService_GetDevicePK(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	deviceUUID := uuid.New()
	deviceUUIDStr := deviceUUID.String()
	invalidUUID := "invalid-uuid"

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDevicesRepository)
		ctx           context.Context
		input         *pb.GetDevicePKRequest
		expectedError error
		expectedPK    []byte
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(&domain.Device{
					UUID: deviceUUID,
					PK:   []byte("public-key"),
				}, nil)
			},
			ctx:        context.WithValue(context.Background(), contextkeys.UserKey, user),
			input:      &pb.GetDevicePKRequest{Uuid: &deviceUUIDStr},
			expectedPK: []byte("public-key"),
		},
		{
			name:          "UserNotFound",
			setupMocks:    func(repo *mocks.MockDevicesRepository) {},
			ctx:           context.Background(),
			input:         &pb.GetDevicePKRequest{Uuid: &deviceUUIDStr},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:          "InvalidUUID",
			setupMocks:    func(repo *mocks.MockDevicesRepository) {},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			input:         &pb.GetDevicePKRequest{Uuid: &invalidUUID},
			expectedError: status.Error(codes.InvalidArgument, "can't parse device uuid"),
		},
		{
			name: "DeviceNotFound",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(nil, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			input:         &pb.GetDevicePKRequest{Uuid: &deviceUUIDStr},
			expectedError: status.Error(codes.NotFound, "no devices to confirm"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(nil, errors.New("db error"))
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			input:         &pb.GetDevicePKRequest{Uuid: &deviceUUIDStr},
			expectedError: status.Error(codes.NotFound, "no devices to confirm"),
		},
		{
			name: "NoPublicKey",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(&domain.Device{
					UUID: deviceUUID,
					PK:   nil,
				}, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			input:         &pb.GetDevicePKRequest{Uuid: &deviceUUIDStr},
			expectedError: status.Error(codes.NotFound, "device has no PK"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDevicesRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewDevicesService(repo, logger)
			resp, err := service.GetDevicePK(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPK, resp.Pk)
			}
		})
	}
}

func TestDevicesService_ApproveDevice(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	deviceUUID := uuid.New()
	deviceUUIDStr := deviceUUID.String()
	invalidUUID := "invalid-uuid"

	// GetDeviceByUserUUIDAndUUID retruns default UUID value
	var mockUUID uuid.UUID

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockDevicesRepository)
		ctx           context.Context
		input         *pb.ApproveDeviceRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(&domain.Device{}, nil)
				repo.EXPECT().ConfirmDevice(gomock.Any(), mockUUID, userUUID, []byte("master-key")).Return(nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.ApproveDeviceRequest{
				Uuid: &deviceUUIDStr,
				Mk:   []byte("master-key"),
			},
		},
		{
			name:       "UserNotFound",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx:        context.Background(),
			input: &pb.ApproveDeviceRequest{
				Uuid: &deviceUUIDStr,
				Mk:   []byte("master-key"),
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "InvalidUUID",
			setupMocks: func(repo *mocks.MockDevicesRepository) {},
			ctx:        context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.ApproveDeviceRequest{
				Uuid: &invalidUUID,
				Mk:   []byte("master-key"),
			},
			expectedError: status.Error(codes.InvalidArgument, "can't parse device uuid"),
		},
		{
			name: "DeviceNotFound",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(nil, nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.ApproveDeviceRequest{
				Uuid: &deviceUUIDStr,
				Mk:   []byte("master-key"),
			},
			expectedError: status.Error(codes.NotFound, "no devices to confirm"),
		},
		{
			name: "DatabaseErrorOnGetDevice",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(nil, errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.ApproveDeviceRequest{
				Uuid: &deviceUUIDStr,
				Mk:   []byte("master-key"),
			},
			expectedError: status.Error(codes.NotFound, "no devices to confirm"),
		},
		{
			name: "DatabaseErrorOnConfirm",
			setupMocks: func(repo *mocks.MockDevicesRepository) {
				repo.EXPECT().GetDeviceByUserUUIDAndUUID(gomock.Any(), deviceUUID, userUUID).Return(&domain.Device{}, nil)
				repo.EXPECT().ConfirmDevice(gomock.Any(), mockUUID, userUUID, []byte("master-key")).Return(errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.ApproveDeviceRequest{
				Uuid: &deviceUUIDStr,
				Mk:   []byte("master-key"),
			},
			expectedError: status.Error(codes.Internal, "device update failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockDevicesRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewDevicesService(repo, logger)
			_, err := service.ApproveDevice(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
