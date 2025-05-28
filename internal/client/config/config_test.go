package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServerAddressEnv(t *testing.T) {
	type want struct {
		ServerAddress string
		expectError   bool
	}
	tests := []struct {
		name string
		envs map[string]string
		want want
	}{
		{
			name: "when address is set",
			envs: map[string]string{
				"GOPHKEEPER_SERVER": "localhost:5050",
			},
			want: want{
				ServerAddress: "localhost:5050",
				expectError:   false,
			},
		},
		{
			name: "when address is not set",
			envs: map[string]string{},
			want: want{
				expectError: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range test.envs {
				os.Setenv(k, v)
			}

			config, err := NewConfig()

			if test.want.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, test.want.ServerAddress, config.ServerAddress)
			}
		})
	}
}
