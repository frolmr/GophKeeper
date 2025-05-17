package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestDecryptWithPrivateKey(t *testing.T) {
	cs := NewCryptoService()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	testMessage := []byte("test message")
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, testMessage, nil)
	require.NoError(t, err)

	tests := []struct {
		name        string
		ciphertext  []byte
		privKey     *rsa.PrivateKey
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil private key",
			ciphertext:  ciphertext,
			privKey:     nil,
			wantErr:     true,
			errContains: "private key cannot be nil",
		},
		{
			name:        "empty ciphertext",
			ciphertext:  []byte{},
			privKey:     privateKey,
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:        "invalid ciphertext",
			ciphertext:  []byte("invalid"),
			privKey:     privateKey,
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:       "successful decryption",
			ciphertext: ciphertext,
			privKey:    privateKey,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := cs.DecryptWithPrivateKey(tt.ciphertext, tt.privKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testMessage, decrypted)
		})
	}
}

func TestDecryptWithPassword(t *testing.T) {
	cs := NewCryptoService()
	password := "test-password"
	plaintext := []byte("secret message")

	encrypted, err := func() ([]byte, error) {
		salt := make([]byte, DefaultParams.SaltLength)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}

		key := argon2.IDKey(
			[]byte(password),
			salt,
			DefaultParams.Time,
			DefaultParams.Memory,
			DefaultParams.Threads,
			DefaultParams.KeyLength,
		)

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}

		ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
		combined := append(salt, nonce...)
		combined = append(combined, ciphertext...)
		return []byte(base64.StdEncoding.EncodeToString(combined)), nil
	}()
	require.NoError(t, err)

	tests := []struct {
		name        string
		encrypted   []byte
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty password",
			encrypted:   encrypted,
			password:    "",
			wantErr:     true,
			errContains: "password cannot be empty",
		},
		{
			name:        "invalid base64",
			encrypted:   []byte("not base64"),
			password:    password,
			wantErr:     true,
			errContains: "failed to decode base64",
		},
		{
			name:        "wrong password",
			encrypted:   encrypted,
			password:    "wrong-password",
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:      "successful decryption",
			encrypted: encrypted,
			password:  password,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := cs.DecryptWithPassword(tt.encrypted, tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}
