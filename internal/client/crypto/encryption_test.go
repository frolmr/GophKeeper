package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptWithPublicKey(t *testing.T) {
	cs := NewCryptoService()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	tests := []struct {
		name        string
		data        []byte
		pubKey      *rsa.PublicKey
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil public key",
			data:        []byte("data"),
			pubKey:      nil,
			wantErr:     true,
			errContains: "public key cannot be nil",
		},
		{
			name:    "small data",
			data:    []byte("small data"),
			pubKey:  publicKey,
			wantErr: false,
		},
		{
			name:    "large data",
			data:    make([]byte, publicKey.Size()*2),
			pubKey:  publicKey,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.data) > 0 && tt.data[0] == 0 { // Initialize large data
				_, err := rand.Read(tt.data)
				require.NoError(t, err)
			}

			encrypted, err := cs.EncryptWithPublicKey(tt.data, tt.pubKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, tt.data, encrypted)
		})
	}
}

func TestEncryptWithPassword(t *testing.T) {
	cs := NewCryptoService()

	tests := []struct {
		name        string
		data        []byte
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty password",
			data:        []byte("data"),
			password:    "",
			wantErr:     true,
			errContains: "password cannot be empty",
		},
		{
			name:     "successful encryption",
			data:     []byte("data"),
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty data",
			data:     []byte{},
			password: "password123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := cs.EncryptWithPassword(tt.data, tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
		})
	}
}

func TestCombineChunks(t *testing.T) {
	tests := []struct {
		name     string
		chunks   [][]byte
		expected []byte
	}{
		{
			name:     "multiple chunks",
			chunks:   [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			expected: []byte("abc"),
		},
		{
			name:     "empty chunks",
			chunks:   [][]byte{},
			expected: []byte{},
		},
		{
			name:     "nil chunks",
			chunks:   nil,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineChunks(tt.chunks)
			assert.Equal(t, tt.expected, result)
		})
	}
}
