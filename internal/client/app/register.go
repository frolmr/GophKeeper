package app

import (
	"fmt"
)

func (gk *GophKeeper) Register(email, password string) error {
	dk, err := gk.CryptoService.GenerateDeviceKey()
	if err != nil {
		return fmt.Errorf("device key generation error: %w", err)
	}

	privateBytes, err := gk.CryptoService.PrivateKeyToBytes(dk.PrivateKey)
	if err != nil {
		return fmt.Errorf("error private key conversion: %w", err)
	}

	if err := gk.LocalStorage.SavePrivateKey(privateBytes); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	mk, err := gk.CryptoService.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("master key generation error: %w", err)
	}

	mp, err := gk.CryptoService.GenerateStrongPassword()
	if err != nil {
		return fmt.Errorf("master password generation error: %w", err)
	}
	fmt.Println("\nIMPORTANT! Please write down your Master Password and store it securely:")
	fmt.Printf("\nMaster Password: %s\n\n", mp)
	fmt.Println("You will need this password to access your data from other devices.")

	encryptedWithPK, err := gk.CryptoService.EncryptWithPublicKey(mk, dk.PublicKey)
	if err != nil {
		return fmt.Errorf("master key encryption with pk error: %w", err)
	}

	encryptedWithMP, err := gk.CryptoService.EncryptWithPassword(mk, mp)
	if err != nil {
		return fmt.Errorf("master key encryption with pass error: %w", err)
	}

	fmt.Println("MK with PK: ", string(encryptedWithPK))
	fmt.Println("MK with PK: ", string(encryptedWithMP))

	return nil
}
