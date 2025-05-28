// Package crypto provides cryptographic operations for the GophKeeper client.
//
// The package implements:
// - Asymmetric encryption (RSA-OAEP with SHA-256)
// - Symmetric encryption (XChaCha20-Poly1305)
// - Key generation and management
// - PEM encoding/decoding of keys
//
// Security Features:
// - Secure random number generation
// - Authenticated encryption
// - Key size validation
// - Chunked encryption for large data
//
// Note: All cryptographic operations use cryptographically secure random sources.
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	privateKeySize = 2048 // RSA key size in bits
	masterKeySize  = 256  // XChaCha20-Poly1305 key size in bits
)

// DeviceKey holds an RSA key pair for device identity and encryption.
type DeviceKey struct {
	PrivateKey *rsa.PrivateKey // Device's private key
	PublicKey  *rsa.PublicKey  // Corresponding public key
}

// CryptoService provides cryptographic operations for the application.
// All methods are safe for concurrent use.
type CryptoService struct{}

// NewCryptoService creates a new CryptoService instance.
func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

// EncryptWithPublicKey encrypts data using RSA-OAEP with SHA-256.
// Handles large data by splitting into chunks automatically.
//
// Parameters:
//
//	data    - Plaintext to encrypt
//	pubKey  - Recipient's RSA public key
//
// Returns:
//
//	[]byte  - Encrypted ciphertext
//	error   - Encryption or validation errors
//
// Security:
// - Uses OAEP padding with SHA-256
// - Validates key before use
// - Each chunk encrypted independently
func (cs *CryptoService) EncryptWithPublicKey(data []byte, pubKey *rsa.PublicKey) ([]byte, error) {
	if pubKey == nil {
		return nil, errors.New("public key cannot be nil")
	}

	if pubKey.N == nil || pubKey.E <= 1 {
		return nil, errors.New("invalid public key")
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

// EncryptWithMasterKey encrypts data using XChaCha20-Poly1305.
// Returns ciphertext with prepended nonce.
//
// Parameters:
//
//	plaintext  - Data to encrypt
//	masterKey  - 32-byte symmetric key
//
// Returns:
//
//	[]byte     - nonce || ciphertext
//	error      - Encryption or validation errors
//
// Security:
// - Uses authenticated encryption
// - Generates random nonce for each operation
// - Validates key size
func (cs *CryptoService) EncryptWithMasterKey(plaintext, masterKey []byte) ([]byte, error) {
	if len(masterKey) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size, must be 32 bytes")
	}

	aead, err := chacha20poly1305.NewX(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

// DecryptWithMasterKey decrypts data encrypted with EncryptWithMasterKey.
//
// Parameters:
//
//	ciphertext - nonce || ciphertext
//	masterKey  - Same 32-byte key used for encryption
//
// Returns:
//
//	[]byte     - Decrypted plaintext
//	error      - Decryption or validation errors
func (cs *CryptoService) DecryptWithMasterKey(ciphertext, masterKey []byte) ([]byte, error) {
	if len(masterKey) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size, must be 32 bytes")
	}

	aead, err := chacha20poly1305.NewX(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// DecryptWithPrivateKey decrypts RSA-OAEP encrypted data.
// Handles chunked ciphertext automatically.
//
// Parameters:
//
//	ciphertext - Data to decrypt
//	privKey    - Recipient's RSA private key
//
// Returns:
//
//	[]byte     - Decrypted plaintext
//	error      - Decryption or validation errors
func (cs *CryptoService) DecryptWithPrivateKey(ciphertext []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	if privKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	chunkSize := privKey.Size()

	if len(ciphertext) <= chunkSize {
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, ciphertext, nil)
	}

	var decryptedChunks [][]byte

	for offset := 0; offset < len(ciphertext); offset += chunkSize {
		end := offset + chunkSize
		if end > len(ciphertext) {
			end = len(ciphertext)
		}

		chunk := ciphertext[offset:end]
		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt: %w", err)
		}

		decryptedChunks = append(decryptedChunks, decryptedChunk)
	}

	return combineChunks(decryptedChunks), nil
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

// GenerateDeviceKey creates a new RSA key pair for device identity.
//
// Returns:
//
//	*DeviceKey - Generated key pair
//	error      - Key generation errors
//
// Note:
// Currently generates 2048-bit RSA keys
func (cs *CryptoService) GenerateDeviceKey() (*DeviceKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, privateKeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// TODO: protect private key with password
	return &DeviceKey{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// GenerateMasterKey creates a new random symmetric key.
//
// Returns:
//
//	[]byte - 32-byte key
//	error  - Random number generation errors
func (cs *CryptoService) GenerateMasterKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize) // 32 bytes
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("key generation failed: %w", err)
	}
	return key, nil
}

// PublicKeyToBytes converts RSA public key to PEM format.
//
// Parameters:
//
//	pub - Public key to serialize
//
// Returns:
//
//	[]byte - PEM encoded key
//	error  - Encoding errors
func (cs *CryptoService) PublicKeyToBytes(pub *rsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key is nil")
	}

	pubASN, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN,
	})

	return pubBytes, nil
}

// PrivateKeyToBytes converts RSA private key to PEM format.
//
// Parameters:
//
//	priv - Private key to serialize
//
// Returns:
//
//	[]byte - PEM encoded key
//	error  - Encoding errors
//
// Security Note:
// Does not encrypt the private key - handle with care
func (cs *CryptoService) PrivateKeyToBytes(priv *rsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key is nil")
	}

	privBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	return privBytes, nil
}

// BytesToPrivateKey parses PEM encoded RSA private key.
// Supports both PKCS#1 and PKCS#8 formats.
//
// Parameters:
//
//	pkb - PEM encoded private key
//
// Returns:
//
//	*rsa.PrivateKey - Parsed key
//	error           - Decoding/parsing errors
func (cs *CryptoService) BytesToPrivateKey(pkb []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pkb)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}

		rsaPriv, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaPriv, nil
	}

	return priv, nil
}

// BytesToPublicKey parses RSA public key from PEM or DER format.
//
// Parameters:
//
//	pkb - Key data (PEM or DER)
//
// Returns:
//
//	*rsa.PublicKey - Parsed key
//	error          - Decoding/parsing errors
func (cs *CryptoService) BytesToPublicKey(pkb []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pkb)
	if block != nil {
		pkb = block.Bytes
	}

	pub, err := x509.ParsePKIXPublicKey(pkb)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}
