package config

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSecretEnv(t *testing.T) {
	type want struct {
		JWTKey      []byte
		expectError bool
		expectedExp time.Duration
		checkLength bool
	}
	tests := []struct {
		name string
		envs map[string]string
		want want
	}{
		{
			name: "when JWT_SECRET is set",
			envs: map[string]string{
				"JWT_SECRET": "zna40k_dura40k",
			},
			want: want{
				JWTKey:      []byte("zna40k_dura40k"),
				expectError: false,
				expectedExp: 15 * time.Minute,
			},
		},
		{
			name: "when JWT_SECRET is not set",
			envs: map[string]string{},
			want: want{
				expectError: false,
				expectedExp: 15 * time.Minute,
				checkLength: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range test.envs {
				os.Setenv(k, v)
			}

			config, err := NewAuthConfig()

			if test.want.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, test.want.expectedExp, config.JWTAccessTokenExpiresIn)

				if len(test.want.JWTKey) > 0 {
					assert.Equal(t, test.want.JWTKey, config.JWTKey)
				}
				if test.want.checkLength {
					decoded, err := base64.URLEncoding.DecodeString(string(config.JWTKey))
					assert.NoError(t, err)
					assert.Len(t, decoded, jwtSecureLength)
				}
			}
		})
	}
}
