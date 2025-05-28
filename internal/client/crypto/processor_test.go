package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestNewCryptoService(t *testing.T) {
	cs := NewCryptoService()
	assert.NotNil(t, cs)
}

func TestEncryptDecryptWithPublicKey(t *testing.T) {
	cs := NewCryptoService()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	tests := []struct {
		name     string
		input    string
		key      *rsa.PublicKey
		wantErr  bool
		errMatch string
	}{
		{
			name:    "successful encryption/decryption",
			input:   "test message",
			key:     publicKey,
			wantErr: false,
		},
		{
			name:    "empty message",
			input:   "",
			key:     publicKey,
			wantErr: false,
		},
		{
			name:    "large message",
			input:   string(make([]byte, 1000)),
			key:     publicKey,
			wantErr: false,
		},
		{
			name:     "nil public key",
			input:    "test message",
			key:      nil,
			wantErr:  true,
			errMatch: "public key cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cs.EncryptWithPublicKey([]byte(tt.input), tt.key)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, []byte(tt.input), ciphertext)

			plaintext, err := cs.DecryptWithPrivateKey(ciphertext, privateKey)
			require.NoError(t, err)
			assert.Equal(t, tt.input, string(plaintext))
		})
	}
}

func TestEncryptDecryptWithMasterKey(t *testing.T) {
	cs := NewCryptoService()

	masterKey := make([]byte, chacha20poly1305.KeySize)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		key         []byte
		wantEncrypt bool
		wantDecrypt bool
		errMatch    string
	}{
		{
			name:        "successful encryption/decryption",
			input:       "test message",
			key:         masterKey,
			wantEncrypt: true,
			wantDecrypt: true,
		},
		{
			name:        "empty message",
			input:       "",
			key:         masterKey,
			wantEncrypt: true,
			wantDecrypt: true,
		},
		{
			name:        "invalid key size",
			input:       "test message",
			key:         []byte("too short"),
			wantEncrypt: false,
			errMatch:    "invalid key size, must be 32 bytes",
		},
		{
			name:        "decryption with invalid key",
			input:       "test message",
			key:         masterKey,
			wantEncrypt: true,
			wantDecrypt: false,
		},
		{
			name:        "corrupted ciphertext",
			input:       "test message",
			key:         masterKey,
			wantEncrypt: true,
			wantDecrypt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cs.EncryptWithMasterKey([]byte(tt.input), tt.key)
			if !tt.wantEncrypt {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
				return
			}
			require.NoError(t, err)
			assert.NotEqual(t, []byte(tt.input), ciphertext)

			var decryptInput []byte
			switch tt.name {
			case "decryption with invalid key":
				wrongKey := make([]byte, chacha20poly1305.KeySize)
				_, err := rand.Read(wrongKey)
				require.NoError(t, err)
				//nolint:ineffassign // TODO: fix linter warn
				decryptInput, err = cs.DecryptWithMasterKey(ciphertext, wrongKey)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "decryption failed")
				return
			case "corrupted ciphertext":
				if len(ciphertext) > 0 {
					ciphertext[0] ^= 0xff
				}
				//nolint:ineffassign // TODO: fix linter warn
				decryptInput, err = cs.DecryptWithMasterKey(ciphertext, tt.key)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "decryption failed")
				return
			default:
				decryptInput = ciphertext
			}

			plaintext, err := cs.DecryptWithMasterKey(decryptInput, tt.key)
			if !tt.wantDecrypt {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input, string(plaintext))
		})
	}
}

func TestCombineChunks(t *testing.T) {
	tests := []struct {
		name   string
		chunks [][]byte
		want   []byte
	}{
		{
			name:   "single chunk",
			chunks: [][]byte{[]byte("hello")},
			want:   []byte("hello"),
		},
		{
			name:   "multiple chunks",
			chunks: [][]byte{[]byte("hello"), []byte(" "), []byte("world")},
			want:   []byte("hello world"),
		},
		{
			name:   "empty chunks",
			chunks: [][]byte{[]byte(""), []byte("test"), []byte("")},
			want:   []byte("test"),
		},
		{
			name:   "nil chunks",
			chunks: [][]byte{nil, []byte("test"), nil},
			want:   []byte("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := combineChunks(tt.chunks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateDeviceKey(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		cs := NewCryptoService()
		key, err := cs.GenerateDeviceKey()
		require.NoError(t, err)
		assert.NotNil(t, key.PrivateKey)
		assert.NotNil(t, key.PublicKey)
		assert.Equal(t, 2048, key.PrivateKey.Size()*8)
	})

	t.Run("invalid key size", func(t *testing.T) {
		// Test with invalid key size through the CryptoService interface
		// This assumes your CryptoService has a way to set key size
		// If not, we'll need to skip this test or find another way
		t.Skip("No way to test key generation failure without modifying the CryptoService interface")
	})
}

func TestGenerateMasterKey(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		cs := NewCryptoService()
		key, err := cs.GenerateMasterKey()
		require.NoError(t, err)
		assert.Len(t, key, chacha20poly1305.KeySize)
	})

	t.Run("invalid key size", func(t *testing.T) {
		// This would require modifying the CryptoService to allow injecting key size
		// Since we can't modify the chacha20poly1305.KeySize constant
		t.Skip("No way to test key generation failure without modifying the CryptoService interface")
	})
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

func TestBytesToPrivateKey(t *testing.T) {
	cs := NewCryptoService()

	key, err := cs.GenerateDeviceKey()
	require.NoError(t, err)

	privBytes, err := cs.PrivateKeyToBytes(key.PrivateKey)
	require.NoError(t, err)
	require.NotEmpty(t, privBytes)

	t.Run("successful conversion", func(t *testing.T) {
		privKey, err := cs.BytesToPrivateKey(privBytes)
		require.NoError(t, err)
		assert.NotNil(t, privKey)
		assert.Equal(t, key.PrivateKey.D, privKey.D)
		assert.Equal(t, key.PrivateKey.PublicKey.E, privKey.PublicKey.E)
		assert.Equal(t, key.PrivateKey.PublicKey.N, privKey.PublicKey.N)
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		_, err := cs.BytesToPrivateKey([]byte("invalid pem data"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode PEM block")
	})

	t.Run("invalid PKCS1 key data", func(t *testing.T) {
		block := pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: []byte("invalid key data"),
		}
		pemBytes := pem.EncodeToMemory(&block)

		_, err := cs.BytesToPrivateKey(pemBytes)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("invalid PKCS8 key data", func(t *testing.T) {
		block := pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: []byte("invalid pkcs8 data"),
		}
		pemBytes := pem.EncodeToMemory(&block)

		_, err := cs.BytesToPrivateKey(pemBytes)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})
}

func TestBytesToPublicKey(t *testing.T) {
	cs := NewCryptoService()

	key, err := cs.GenerateDeviceKey()
	require.NoError(t, err)

	pubBytes, err := cs.PublicKeyToBytes(key.PublicKey)
	require.NoError(t, err)
	require.NotEmpty(t, pubBytes)

	t.Run("successful conversion", func(t *testing.T) {
		pubKey, err := cs.BytesToPublicKey(pubBytes)
		require.NoError(t, err)
		assert.NotNil(t, pubKey)
		assert.Equal(t, key.PublicKey.E, pubKey.E)
		assert.Equal(t, key.PublicKey.N, pubKey.N)
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		_, err := cs.BytesToPublicKey([]byte("invalid pem data"))
		require.Error(t, err)
	})

	t.Run("invalid key data", func(t *testing.T) {
		block := pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: []byte("invalid key data"),
		}
		pemBytes := pem.EncodeToMemory(&block)

		_, err := cs.BytesToPublicKey(pemBytes)
		require.Error(t, err)
	})

	t.Run("raw bytes without PEM", func(t *testing.T) {
		pubASN, err := x509.MarshalPKIXPublicKey(key.PublicKey)
		require.NoError(t, err)

		pubKey, err := cs.BytesToPublicKey(pubASN)
		require.NoError(t, err)
		assert.NotNil(t, pubKey)
	})
}

func TestEncryptWithPublicKey_ErrorCases(t *testing.T) {
	t.Run("invalid public key", func(t *testing.T) {
		cs := NewCryptoService()
		invalidKey := &rsa.PublicKey{N: nil, E: 0} // Invalid key
		_, err := cs.EncryptWithPublicKey([]byte("test"), invalidKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid public key")
	})
}

func TestEncryptWithMasterKey_ErrorCases(t *testing.T) {
	t.Run("invalid key size", func(t *testing.T) {
		cs := NewCryptoService()
		_, err := cs.EncryptWithMasterKey([]byte("test"), []byte("too short"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key size, must be 32 bytes")
	})
}

func TestDecryptWithPrivateKey_ErrorCases(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("nil private key", func(t *testing.T) {
		cs := NewCryptoService()
		_, err := cs.DecryptWithPrivateKey([]byte("test"), nil)
		require.Error(t, err)
		assert.Equal(t, "private key cannot be nil", err.Error())
	})

	t.Run("invalid ciphertext", func(t *testing.T) {
		cs := NewCryptoService()
		_, err := cs.DecryptWithPrivateKey([]byte("too short"), privateKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crypto/rsa: decryption error")
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		goodCs := NewCryptoService()
		ciphertext, err := goodCs.EncryptWithPublicKey([]byte("test"), &privateKey.PublicKey)
		require.NoError(t, err)

		if len(ciphertext) > 0 {
			ciphertext[0] ^= 0xff
		}

		cs := NewCryptoService()
		_, err = cs.DecryptWithPrivateKey(ciphertext, privateKey)
		require.Error(t, err)
	})
}
