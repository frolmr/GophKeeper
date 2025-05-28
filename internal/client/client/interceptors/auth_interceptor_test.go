package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuthInterceptor_Unary(t *testing.T) {
	const testToken = "test-token-123"

	tests := []struct {
		name        string
		method      string
		expectAuth  bool
		expectedErr error
	}{
		{
			name:       "unprotected method - RegisterUser",
			method:     "/users.Users/RegisterUser",
			expectAuth: false,
		},
		{
			name:       "unprotected method - LoginUser",
			method:     "/users.Users/LoginUser",
			expectAuth: false,
		},
		{
			name:       "protected method",
			method:     "/users.Users/GetUser",
			expectAuth: true,
		},
		{
			name:        "protected method with error",
			method:      "/users.Users/GetUser",
			expectAuth:  true,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewAuthInterceptor(testToken)
			mockInv := &mockInvoker{err: tt.expectedErr}

			ctx := context.Background()

			err := interceptor.Unary()(
				ctx,
				tt.method,
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

			if tt.expectAuth {
				md, ok := metadata.FromOutgoingContext(mockInv.lastContext)
				require.True(t, ok, "metadata should exist")
				assert.Equal(t, []string{testToken}, md.Get("authorization"))
			} else {
				assert.Equal(t, ctx, mockInv.lastContext)
			}
		})
	}
}

func TestNewAuthInterceptor(t *testing.T) {
	t.Run("creates interceptor with token", func(t *testing.T) {
		const testToken = "new-test-token"
		interceptor := NewAuthInterceptor(testToken)
		assert.Equal(t, testToken, interceptor.token)
	})
}
