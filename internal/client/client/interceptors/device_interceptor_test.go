package interceptors

import (
	"context"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestDeviceInterceptor_Unary(t *testing.T) {
	testDevice := &domain.Device{
		ID:   "test-device-id-123",
		Name: "test-device-name",
	}

	tests := []struct {
		name         string
		device       *domain.Device
		expectDevice bool
		expectedErr  error
	}{
		{
			name:         "valid device adds metadata",
			device:       testDevice,
			expectDevice: true,
		},
		{
			name:         "nil device creates nil interceptor",
			device:       nil,
			expectDevice: false,
		},
		{
			name:         "propagates invoker error",
			device:       testDevice,
			expectDevice: true,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewDeviceInterceptor(tt.device)

			if tt.device == nil {
				assert.Nil(t, interceptor, "interceptor should be nil for nil device")
				return
			}

			mockInv := &mockInvoker{err: tt.expectedErr}
			ctx := context.Background()

			err := interceptor.Unary()(
				ctx,
				"test-method",
				nil, // req
				nil, // reply
				nil, // cc
				mockInv.invoker,
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, mockInv.called, "invoker should have been called")

			if tt.expectDevice {
				md, ok := metadata.FromOutgoingContext(mockInv.lastContext)
				require.True(t, ok, "metadata should exist")
				assert.Equal(t, []string{testDevice.ID}, md.Get("device_id"))
				assert.Equal(t, []string{testDevice.Name}, md.Get("device_name"))
			} else {
				assert.Equal(t, ctx, mockInv.lastContext, "context should be unchanged")
			}
		})
	}
}

func TestNewDeviceInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		device  *domain.Device
		wantNil bool
	}{
		{
			name: "creates interceptor with valid device",
			device: &domain.Device{
				ID:   "valid-id",
				Name: "valid-name",
			},
			wantNil: false,
		},
		{
			name:    "nil device creates nil interceptor",
			device:  nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewDeviceInterceptor(tt.device)

			if tt.wantNil {
				assert.Nil(t, interceptor)
			} else {
				require.NotNil(t, interceptor)
				assert.Equal(t, tt.device, interceptor.device)
			}
		})
	}
}
