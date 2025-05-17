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

	"golang.org/x/crypto/argon2"
)

func (cs *CryptoService) DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key cannot be nil")
	}

	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		priv,
		ciphertext,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with private key: %w", err)
	}
	return plaintext, nil
}

func (cs *CryptoService) DecryptWithPassword(encrypted []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	salt := decoded[:DefaultParams.SaltLength]
	nonceStart := DefaultParams.SaltLength
	nonceEnd := nonceStart + 12
	nonce := decoded[nonceStart:nonceEnd]
	ciphertext := decoded[nonceEnd:]

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

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}
