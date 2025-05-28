package device

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/frolmr/GophKeeper/internal/client/crypto"
	"github.com/frolmr/GophKeeper/internal/client/domain"
)

// Connector defines the interface for device-related network operations.
type Connector interface {
	SendGetDevicePKRequest(id string) ([]byte, error)
	SendGetDeviceMKRequest() ([]byte, error)
	SendAddDeviceRequest([]byte) error
	SendApproveDeviceRequest(id string, mk []byte) error
	SendListDevicesRequest() ([]domain.DeviceOnServer, error)
}

// Repository defines persistence operations for device keys.
type Repository interface {
	SavePrivateKey([]byte) error
	ReadPrivateKey() ([]byte, error)
}

// CryptoProcessor defines cryptographic operations needed for device management.
type CryptoProcessor interface {
	GenerateDeviceKey() (*crypto.DeviceKey, error)
	PrivateKeyToBytes(*rsa.PrivateKey) ([]byte, error)
	PublicKeyToBytes(*rsa.PublicKey) ([]byte, error)
	BytesToPrivateKey([]byte) (*rsa.PrivateKey, error)
	DecryptWithPrivateKey([]byte, *rsa.PrivateKey) ([]byte, error)
	BytesToPublicKey([]byte) (*rsa.PublicKey, error)
	EncryptWithPublicKey([]byte, *rsa.PublicKey) ([]byte, error)
}

// DeviceService handles device registration and management operations.
//
// Security Considerations:
// - Manages device public/private key pairs
// - Handles master key distribution between devices
type DeviceService struct {
	repository    Repository
	cryptoService CryptoProcessor
	connector     Connector
}

// NewDeviceService creates a new device service instance.
//
// Parameters:
//
//	repo - Storage repository for device keys
//	cp   - Cryptographic processor
//	dc   - Network connector
func NewDeviceService(repo Repository, cp CryptoProcessor, dc Connector) *DeviceService {
	return &DeviceService{
		repository:    repo,
		cryptoService: cp,
		connector:     dc,
	}
}

// AddDevice registers a new device with the server.
//
// Generates or loads device keys and registers the public key.
//
// Returns:
//
//	error - Key generation, storage, or network errors
func (s *DeviceService) AddDevice() error {
	privKey, err := s.getOrGenPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get pk: %w", err)
	}

	pubKey, err := s.cryptoService.PublicKeyToBytes(&privKey.PublicKey)
	if err != nil {
		return fmt.Errorf("error getting public key from private: %w", err)
	}

	if err := s.connector.SendAddDeviceRequest(pubKey); err != nil {
		return fmt.Errorf("failed to add device: %w", err)
	}

	return nil
}

// ConfirmDevice completes device registration by:
// 1. Retrieving the master key
// 2. Re-encrypting it for the new device
// 3. Sending approval to server
//
// Parameters:
//
//	id - Device UUID to confirm
//
// Returns:
//
//	error - Key, network, or encryption errors
func (s *DeviceService) ConfirmDevice(id string) error {
	privKeyBytes, err := s.repository.ReadPrivateKey()
	if err != nil {
		return fmt.Errorf("error reading PK: %w", err)
	}

	privKey, err := s.cryptoService.BytesToPrivateKey(privKeyBytes)
	if err != nil {
		return fmt.Errorf("error private key convertation: %w", err)
	}

	mk, err := s.connector.SendGetDeviceMKRequest()
	if err != nil {
		return fmt.Errorf("error getting users MK: %w", err)
	}

	decryptedMK, err := s.cryptoService.DecryptWithPrivateKey(mk, privKey)
	if err != nil {
		return fmt.Errorf("MK decryption failed: %w", err)
	}

	pubKeyBytes, err := s.connector.SendGetDevicePKRequest(id)
	if err != nil {
		return fmt.Errorf("error getting device by id: %w", err)
	}

	pubKey, err := s.cryptoService.BytesToPublicKey(pubKeyBytes)
	if err != nil {
		return fmt.Errorf("can't parse PK from bytes: %w", err)
	}

	encryptedMK, err := s.cryptoService.EncryptWithPublicKey(decryptedMK, pubKey)
	if err != nil {
		return fmt.Errorf("MK encryption failed: %w", err)
	}

	return s.connector.SendApproveDeviceRequest(id, encryptedMK)
}

// ListDevices retrieves all devices registered for the current user.
//
// Returns:
//
//	[]domain.DeviceOnServer - List of registered devices
//	error - Network errors
func (s *DeviceService) ListDevices() ([]domain.DeviceOnServer, error) {
	devices, err := s.connector.SendListDevicesRequest()
	if err != nil {
		return nil, fmt.Errorf("list devices request failed: %w", err)
	}

	return devices, nil
}

// getOrGenPrivateKey loads existing device key or generates a new one.
//
// Returns:
//
//	*rsa.PrivateKey - Device private key
//	error - Key generation or storage errors
func (s *DeviceService) getOrGenPrivateKey() (*rsa.PrivateKey, error) {
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
