package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	type want struct {
		address  string
		dbURI    string
		certFile string
		keyFile  string
	}
	tests := []struct {
		args []string
		envs map[string]string
		want want
	}{
		{
			args: []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta", "-c", "cert.pem", "-k", "key.pem"},
			envs: map[string]string{
				"RUN_ADDRESS":  "localhost:9090",
				"DATABASE_URI": "postgres:ko-ko-ko",
			},
			want: want{
				address:  "localhost:8080",
				dbURI:    "postgres:tra-ta-ta",
				certFile: "cert.pem",
				keyFile:  "key.pem",
			},
		},
		{
			args: []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta", "-c", "cert.pem", "-k", "key.pem"},
			envs: map[string]string{},
			want: want{
				address:  "localhost:8080",
				dbURI:    "postgres:tra-ta-ta",
				certFile: "cert.pem",
				keyFile:  "key.pem",
			},
		},
		{
			args: []string{"-c", "cert.pem", "-k", "key.pem"},
			envs: map[string]string{
				"RUN_ADDRESS":  "localhost:9090",
				"DATABASE_URI": "postgres:ko-ko-ko",
			},
			want: want{
				address:  "localhost:9090",
				dbURI:    "postgres:ko-ko-ko",
				certFile: "cert.pem",
				keyFile:  "key.pem",
			},
		},
	}

	for _, test := range tests {
		for k, v := range test.envs {
			os.Setenv(k, v)
		}
		os.Args = append([]string{"cmd"}, test.args...)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		config, err := NewAppConfig()
		os.Clearenv()

		assert.Equal(t, test.want.address, config.RunAddress)
		assert.NoError(t, err)
	}
}

func TestEmptyFlags(t *testing.T) {
	tests := []struct {
		args           []string
		expectedErrors []error
		expectedMsg    string
	}{
		{
			args:           []string{},
			expectedErrors: []error{ErrMissingAddress, ErrMissingDDURI, ErrMissingCertFile, ErrMissingKeyFile},
			expectedMsg:    "missing run address\nmissing database URI\nmissing TLS certificate file\nmissing TLS key file",
		},
		{
			args:           []string{"-a", "localhost:8080"},
			expectedErrors: []error{ErrMissingDDURI, ErrMissingCertFile, ErrMissingKeyFile},
			expectedMsg:    "missing database URI\nmissing TLS certificate file\nmissing TLS key file",
		},
		{
			args:           []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta"},
			expectedErrors: []error{ErrMissingCertFile, ErrMissingKeyFile},
			expectedMsg:    "missing TLS certificate file\nmissing TLS key file",
		},
		{
			args:           []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta", "-c", "cert.pem"},
			expectedErrors: []error{ErrMissingKeyFile},
			expectedMsg:    "missing TLS key file",
		},
	}

	for _, test := range tests {
		os.Args = append([]string{"cmd"}, test.args...)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		_, err := NewAppConfig()
		os.Clearenv()

		for _, expectedError := range test.expectedErrors {
			assert.ErrorIs(t, err, expectedError)
		}
		assert.EqualError(t, err, test.expectedMsg)
	}
}
