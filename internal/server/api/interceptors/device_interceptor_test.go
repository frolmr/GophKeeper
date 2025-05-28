package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	mocks "github.com/frolmr/GophKeeper/internal/server/mocks/interceptors/device"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestDeviceInterceptor_Unary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceID := "device-123"
	deviceName := "My Device"
	userUUID := uuid.New()
	testUser := &domain.User{UUID: userUUID}
	testDevice := &domain.Device{
		ID:        deviceID,
		Name:      deviceName,
		UserUUID:  userUUID,
		Confirmed: true,
	}
	unconfirmedDevice := &domain.Device{
		ID:        deviceID,
		Name:      deviceName,
		UserUUID:  userUUID,
		Confirmed: false,
	}

	mockRepo := mocks.NewMockDeviceRepository(ctrl)

	successHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	tests := []struct {
		name        string
		fullMethod  string
		metadata    metadata.MD
		userInCtx   *domain.User
		setupMocks  func()
		expectedRes interface{}
		expectedErr error
	}{
		{
			name:       "RegisterUser sets device info in context",
			fullMethod: "/users.Users/RegisterUser",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			setupMocks:  func() {},
			expectedRes: "success",
		},
		{
			name:       "AddDevice sets device info in context",
			fullMethod: "/devices.Devices/AddDevice",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			setupMocks:  func() {},
			expectedRes: "success",
		},
		{
			name:        "Missing metadata",
			fullMethod:  "/secure.Method",
			metadata:    nil,
			setupMocks:  func() {},
			expectedErr: status.Error(codes.Unauthenticated, "metadata is not provided"),
		},
		{
			name:       "Missing device_id",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_name": []string{deviceName},
			},
			userInCtx:   testUser,
			setupMocks:  func() {},
			expectedErr: status.Error(codes.Unauthenticated, "device information is not provided"),
		},
		{
			name:       "Missing device_name",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_id": []string{deviceID},
			},
			userInCtx:   testUser,
			setupMocks:  func() {},
			expectedErr: status.Error(codes.Unauthenticated, "device information is not provided"),
		},
		{
			name:       "User not in context",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			setupMocks:  func() {},
			expectedErr: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "Device not found",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			userInCtx: testUser,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetDeviceByUserUUIDAndID(gomock.Any(), deviceID, userUUID).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedErr: status.Error(codes.Unauthenticated, "reqesting device is unothorized"),
		},
		{
			name:       "Device not confirmed",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			userInCtx: testUser,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetDeviceByUserUUIDAndID(gomock.Any(), deviceID, userUUID).
					Return(unconfirmedDevice, nil).
					Times(1)
			},
			expectedErr: status.Error(codes.Unauthenticated, "reqesting device is unothorized"),
		},
		{
			name:       "Device confirmed and valid",
			fullMethod: "/secure.Method",
			metadata: metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			},
			userInCtx: testUser,
			setupMocks: func() {
				mockRepo.EXPECT().
					GetDeviceByUserUUIDAndID(gomock.Any(), deviceID, userUUID).
					Return(testDevice, nil).
					Times(1)
			},
			expectedRes: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			interceptor := NewDeviceInterceptor(mockRepo)
			unaryInterceptor := interceptor.Unary()

			ctx := context.Background()
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}
			if tt.userInCtx != nil {
				ctx = context.WithValue(ctx, contextkeys.UserKey, tt.userInCtx)
			}

			res, err := unaryInterceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: tt.fullMethod},
				successHandler,
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				statusErr, _ := status.FromError(err)
				expectedStatus, _ := status.FromError(tt.expectedErr)
				assert.Equal(t, expectedStatus.Code(), statusErr.Code())
				assert.Contains(t, statusErr.Message(), expectedStatus.Message())
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestDeviceInterceptor_ContextValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceID := "device-123"
	deviceName := "My Device"
	userUUID := uuid.New()
	testUser := &domain.User{UUID: userUUID}
	testDevice := &domain.Device{
		ID:        deviceID,
		Name:      deviceName,
		UserUUID:  userUUID,
		Confirmed: true,
	}

	mockRepo := mocks.NewMockDeviceRepository(ctrl)
	mockRepo.EXPECT().
		GetDeviceByUserUUIDAndID(gomock.Any(), deviceID, userUUID).
		Return(testDevice, nil).
		Times(1)

	interceptor := NewDeviceInterceptor(mockRepo)
	unaryInterceptor := interceptor.Unary()

	md := metadata.MD{
		"device_id":   []string{deviceID},
		"device_name": []string{deviceName},
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, contextkeys.UserKey, testUser)

	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		device, ok := ctx.Value(contextkeys.DeviceKey).(*domain.Device)
		if !ok || device == nil {
			return nil, status.Error(codes.Internal, "device not in context")
		}
		assert.Equal(t, deviceID, device.ID)
		assert.Equal(t, deviceName, device.Name)
		return "ok", nil
	}

	res, err := unaryInterceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/secure.Method"},
		testHandler,
	)

	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestDeviceInterceptor_PublicEndpointsSetContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deviceID := "device-123"
	deviceName := "My Device"

	mockRepo := mocks.NewMockDeviceRepository(ctrl)

	tests := []struct {
		name       string
		fullMethod string
	}{
		{
			name:       "RegisterUser sets device ID and name",
			fullMethod: "/users.Users/RegisterUser",
		},
		{
			name:       "AddDevice sets device ID and name",
			fullMethod: "/devices.Devices/AddDevice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewDeviceInterceptor(mockRepo)
			unaryInterceptor := interceptor.Unary()

			md := metadata.MD{
				"device_id":   []string{deviceID},
				"device_name": []string{deviceName},
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)

			testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				id, ok := ctx.Value(contextkeys.DeviceIDKey).(string)
				if !ok || id != deviceID {
					return nil, status.Error(codes.Internal, "device ID not in context")
				}

				name, ok := ctx.Value(contextkeys.DeviceNameKey).(string)
				if !ok || name != deviceName {
					return nil, status.Error(codes.Internal, "device name not in context")
				}

				return "ok", nil
			}

			res, err := unaryInterceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: tt.fullMethod},
				testHandler,
			)

			assert.NoError(t, err)
			assert.Equal(t, "ok", res)
		})
	}
}
