package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

func (cs *CryptoService) EncryptWithPublicKey(data []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	if pubKey == nil {
		return nil, errors.New("public key cannot be nil")
	}

	maxChunkSize := pubKey.Size() - 2*sha256.Size - 2
	if len(data) <= maxChunkSize {
		ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, data, nil)
		if err != nil {
			return nil, err
		}
		return ciphertext, nil
	}

	var encryptedChunks [][]byte

	for offset := 0; offset < len(data); offset += maxChunkSize {
		end := offset + maxChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[offset:end]
		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, chunk, nil)
		if err != nil {
			return nil, err
		}

		encryptedChunks = append(encryptedChunks, encryptedChunk)
	}

	return combineChunks(encryptedChunks), nil
}

func combineChunks(chunks [][]byte) []byte {
	var totalLen int
	for _, chunk := range chunks {
		totalLen += len(chunk)
	}

	result := make([]byte, 0, totalLen)
	for _, chunk := range chunks {
		result = append(result, chunk...)
	}

	return result
}

func (cs *CryptoService) EncryptWithPassword(data []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	salt := make([]byte, DefaultParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
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
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encrypted := gcm.Seal(nil, nonce, data, nil)

	combined := salt
	combined = append(combined, nonce...)
	combined = append(combined, encrypted...)

	return []byte(base64.StdEncoding.EncodeToString(combined)), nil
}
