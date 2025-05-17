package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDeviceKey(t *testing.T) {
	cs := NewCryptoService()

	tests := []struct {
		name        string
		mockReader  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful key generation",
			mockReader: false,
			wantErr:    false,
		},
		{
			name:        "handles random number generator error",
			mockReader:  true,
			wantErr:     true,
			errContains: "failed to generate RSA key pair",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockReader {
				oldReader := rand.Reader
				defer func() { rand.Reader = oldReader }()
				rand.Reader = &errorReader{}
			}

			key, err := cs.GenerateDeviceKey()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, key)
			require.NotNil(t, key.PrivateKey)
			require.NotNil(t, key.PublicKey)
			assert.Equal(t, privateKeySize, key.PrivateKey.Size()*8)
		})
	}
}

func TestGenerateMasterKey(t *testing.T) {
	cs := NewCryptoService()

	tests := []struct {
		name        string
		mockReader  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful key generation",
			mockReader: false,
			wantErr:    false,
		},
		{
			name:        "handles random number generator error",
			mockReader:  true,
			wantErr:     true,
			errContains: "could not generate random key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockReader {
				oldReader := rand.Reader
				defer func() { rand.Reader = oldReader }()
				rand.Reader = &errorReader{}
			}

			key, err := cs.GenerateMasterKey()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, key)
			assert.Len(t, key, masterKeySize)
		})
	}
}

func TestPublicKeyToBytes(t *testing.T) {
	cs := NewCryptoService()
	key, err := cs.GenerateDeviceKey()
	require.NoError(t, err)

	tests := []struct {
		name        string
		pubKey      *rsa.PublicKey
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful conversion",
			pubKey:  key.PublicKey,
			wantErr: false,
		},
		{
			name:        "handles nil public key",
			pubKey:      nil,
			wantErr:     true,
			errContains: "public key is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pubBytes, err := cs.PublicKeyToBytes(tt.pubKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, pubBytes)

			block, _ := pem.Decode(pubBytes)
			require.NotNil(t, block)
			assert.Equal(t, "RSA PUBLIC KEY", block.Type)

			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			require.NoError(t, err)
			assert.IsType(t, &rsa.PublicKey{}, pubKey)
		})
	}
}

func TestPrivateKeyToBytes(t *testing.T) {
	cs := NewCryptoService()
	key, err := cs.GenerateDeviceKey()
	require.NoError(t, err)

	tests := []struct {
		name        string
		privKey     *rsa.PrivateKey
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful conversion",
			privKey: key.PrivateKey,
			wantErr: false,
		},
		{
			name:        "handles nil private key",
			privKey:     nil,
			wantErr:     true,
			errContains: "private key is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privBytes, err := cs.PrivateKeyToBytes(tt.privKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, privBytes)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, privBytes)

			block, _ := pem.Decode(privBytes)
			require.NotNil(t, block)
			assert.Equal(t, "RSA PRIVATE KEY", block.Type)

			privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			require.NoError(t, err)
			assert.NotNil(t, privKey)
		})
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}
