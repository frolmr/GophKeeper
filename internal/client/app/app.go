package app

import (
	"crypto/rsa"

	"github.com/frolmr/GophKeeper/internal/client/config"
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

type EncryptorDecryptor interface {
	Encryptor
	Decryptor
	SecretsGenerator
}

type GophKeeper struct {
	Config        *config.Config
	LocalStorage  LocalStorager
	CryptoService EncryptorDecryptor
}

func NewApplication(cfg *config.Config, ls LocalStorager, cs EncryptorDecryptor) *GophKeeper {
	return &GophKeeper{
		Config:        cfg,
		LocalStorage:  ls,
		CryptoService: cs,
	}
}
