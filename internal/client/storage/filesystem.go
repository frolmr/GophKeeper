package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/frolmr/GophKeeper/internal/client/domain"
)

const (
	keyFilePermissions = 0600
	appDirPermissions  = 0700
)

type FileSystem struct {
	appDir string
}

func NewLocalStorage() (*FileSystem, error) {
	appConfigDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get app directory: %w", err)
	}

	return &FileSystem{
		appDir: appConfigDir,
	}, nil
}

func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, domain.AppName)
	if err := os.MkdirAll(appDir, appDirPermissions); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}

	return appDir, nil
}

func (fs *FileSystem) SavePrivateKey(data []byte) error {
	filePath := filepath.Join(fs.appDir, domain.PrivKeyFileName)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, keyFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

func (fs *FileSystem) ReadPrivateKey() ([]byte, error) {
	filePath := filepath.Join(fs.appDir, domain.PrivKeyFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("private key file does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	return data, nil
}
