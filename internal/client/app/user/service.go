package user

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/crypto"
)

// Connector defines user authentication operations.
type Connector interface {
	SendRegisterRequest(email, password string, mk []byte) error
	SendLoginRequest(email, password string) (string, error)
}

// Repository defines persistence for user credentials and keys.
type Repository interface {
	SaveToken(string) error
	ReadPrivateKey() ([]byte, error)
	SavePrivateKey([]byte) error
}

// CryptoProcessor defines cryptographic operations for user management.
type CryptoProcessor interface {
	GenerateDeviceKey() (*crypto.DeviceKey, error)
	GenerateMasterKey() ([]byte, error)
	EncryptWithPublicKey([]byte, *rsa.PublicKey) ([]byte, error)
	PrivateKeyToBytes(*rsa.PrivateKey) ([]byte, error)
	BytesToPrivateKey([]byte) (*rsa.PrivateKey, error)
}

// UserService handles user authentication and registration.
//
// Manages:
// - User credentials
// - Authentication tokens
// - Initial key generation
type UserService struct {
	repository    Repository
	cryptoService CryptoProcessor
	connector     Connector
}

// NewUserService creates a new user service instance.
//
// Parameters:
//
//	repo - Storage repository for device keys
//	cp   - Cryptographic processor
//	uc   - Network connector
func NewUserService(repo Repository, cp CryptoProcessor, uc Connector) *UserService {
	return &UserService{
		repository:    repo,
		cryptoService: cp,
		connector:     uc,
	}
}

// Login authenticates the user and stores the session token.
//
// Parameters:
//
//	email    - User email
//	password - Plaintext password
//
// Returns:
//
//	error - Authentication or storage errors
func (s *UserService) Login(email, password string) error {
	token, err := s.connector.SendLoginRequest(email, password)
	if err != nil {
		return fmt.Errorf("response token missing: %w", err)
	}

	tokenSaveErr := s.repository.SaveToken(token)
	if tokenSaveErr != nil {
		return fmt.Errorf("token save error: %w", tokenSaveErr)
	}

	return nil
}

// Register creates a new user account with generated keys.
//
// Parameters:
//
//	email    - User email
//	password - Plaintext password
//
// Returns:
//
//	error - Key generation, encryption, or network errors
func (s *UserService) Register(email, password string) error {
	privKey, err := s.getOrGenPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get pk: %w", err)
	}

	mk, err := s.cryptoService.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("master key generation error: %w", err)
	}

	encryptedMK, err := s.cryptoService.EncryptWithPublicKey(mk, &privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("master key encryption with pk error: %w", err)
	}

	if err := s.connector.SendRegisterRequest(email, password, encryptedMK); err != nil {
		return fmt.Errorf("response token missing: %w", err)
	}

	return nil
}

func (s *UserService) getOrGenPrivateKey() (*rsa.PrivateKey, error) {
	var privKey *rsa.PrivateKey
	pkBytes, err := s.repository.ReadPrivateKey()
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && os.IsNotExist(pathErr) {
			dk, err := s.cryptoService.GenerateDeviceKey()
			if err != nil {
				return nil, fmt.Errorf("device key generation error: %w", err)
			}

			privateBytes, err := s.cryptoService.PrivateKeyToBytes(dk.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("error private key conversion: %w", err)
			}

			if err := s.repository.SavePrivateKey(privateBytes); err != nil {
				return nil, fmt.Errorf("failed to save private key: %w", err)
			}

			privKey = dk.PrivateKey
		} else {
			return nil, fmt.Errorf("error private key reading: %w", err)
		}
	} else {
		privKey, err = s.cryptoService.BytesToPrivateKey(pkBytes)
		if err != nil {
			return nil, fmt.Errorf("error private key convertation: %w", err)
		}
	}

	return privKey, nil
}
