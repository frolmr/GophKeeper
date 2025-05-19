package service

import (
	"crypto/rsa"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/crypto"
)

type LocalStorager interface {
	SavePrivateKey([]byte) error
	ReadPrivateKey() ([]byte, error)
}

type Encryptor interface {
	EncryptWithPublicKey([]byte, *rsa.PublicKey) ([]byte, error)
	EncryptWithPassword([]byte, string) ([]byte, error)
}

type Decryptor interface {
	DecryptWithPrivateKey([]byte, *rsa.PrivateKey) ([]byte, error)
	DecryptWithPassword([]byte, string) ([]byte, error)
}

type SecretsGenerator interface {
	GenerateStrongPassword() (string, error)
	GenerateDeviceKey() (*crypto.DeviceKey, error)
	GenerateMasterKey() ([]byte, error)
	PublicKeyToBytes(*rsa.PublicKey) ([]byte, error)
	PrivateKeyToBytes(*rsa.PrivateKey) ([]byte, error)
}

type UserServiceConnector interface {
	SendRegisterRequest(email, password string, mk []byte) error
	SendLoginRequest()
}

type EncryptorDecryptor interface {
	Encryptor
	Decryptor
	SecretsGenerator
}

type UserService struct {
	LocalStorage  LocalStorager
	CryptoService EncryptorDecryptor
	Connector     UserServiceConnector
}

func NewUserService(ls LocalStorager, cs EncryptorDecryptor, cn UserServiceConnector) *UserService {
	return &UserService{
		LocalStorage:  ls,
		CryptoService: cs,
		Connector:     cn,
	}
}

func (us *UserService) Register(email, password string) error {
	dk, err := us.CryptoService.GenerateDeviceKey()
	if err != nil {
		return fmt.Errorf("device key generation error: %w", err)
	}

	privateBytes, err := us.CryptoService.PrivateKeyToBytes(dk.PrivateKey)
	if err != nil {
		return fmt.Errorf("error private key conversion: %w", err)
	}

	if err := us.LocalStorage.SavePrivateKey(privateBytes); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	mk, err := us.CryptoService.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("master key generation error: %w", err)
	}

	mp, err := us.CryptoService.GenerateStrongPassword()
	if err != nil {
		return fmt.Errorf("master password generation error: %w", err)
	}
	fmt.Println("\nIMPORTANT! Please write down your Master Password and store it securely:")
	fmt.Printf("\nMaster Password: %s\n\n", mp)
	fmt.Println("You will need this password to access your data from other devices.")

	_, err = us.CryptoService.EncryptWithPublicKey(mk, dk.PublicKey)
	if err != nil {
		return fmt.Errorf("master key encryption with pk error: %w", err)
	}

	encryptedWithMP, err := us.CryptoService.EncryptWithPassword(mk, mp)
	if err != nil {
		return fmt.Errorf("master key encryption with pass error: %w", err)
	}

	connErr := us.Connector.SendRegisterRequest(email, password, encryptedWithMP)
	if connErr != nil {
		fmt.Println("GOVNO: ", connErr.Error())
	}

	return nil
}

func (us *UserService) Login(email, password string) error {
	return nil
}
