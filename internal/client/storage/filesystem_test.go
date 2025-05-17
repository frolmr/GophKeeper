package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestConfigDir(t *testing.T) (string, error) {
	testDir := t.TempDir()
	appDir := filepath.Join(testDir, domain.AppName)
	if err := os.MkdirAll(appDir, appDirPermissions); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}
	return appDir, nil
}

func newTestLocalStorage(t *testing.T) (*FileSystem, error) {
	appConfigDir, err := getTestConfigDir(t)
	if err != nil {
		return nil, fmt.Errorf("failed to get app directory: %w", err)
	}

	return &FileSystem{
		appDir: appConfigDir,
	}, nil
}

func TestLocalStorage(t *testing.T) {
	t.Run("SavePrivateKey and ReadPrivateKey", func(t *testing.T) {
		local, err := newTestLocalStorage(t)
		require.NoError(t, err)
		testData := []byte("test private key data")

		err = local.SavePrivateKey(testData)
		require.NoError(t, err)

		data, err := local.ReadPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, testData, data)
	})

	t.Run("SavePrivateKey to existing file", func(t *testing.T) {
		local, err := newTestLocalStorage(t)
		require.NoError(t, err)
		testData1 := []byte("initial data")
		testData2 := []byte("updated data")

		err = local.SavePrivateKey(testData1)
		require.NoError(t, err)

		err = local.SavePrivateKey(testData2)
		require.NoError(t, err)

		data, err := local.ReadPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, testData2, data)
	})

	t.Run("ReadPrivateKey when file doesn't exist", func(t *testing.T) {
		local, err := newTestLocalStorage(t)
		require.NoError(t, err)

		_, err = local.ReadPrivateKey()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "private key file does not exist")
	})

	t.Run("Directory has correct permissions", func(t *testing.T) {
		testDir := t.TempDir()
		appDir := filepath.Join(testDir, domain.AppName)

		os.RemoveAll(appDir)

		err := os.MkdirAll(appDir, appDirPermissions)
		require.NoError(t, err)

		info, err := os.Stat(appDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(appDirPermissions), info.Mode().Perm())
	})

	t.Run("File has correct permissions", func(t *testing.T) {
		local, err := newTestLocalStorage(t)
		require.NoError(t, err)
		testData := []byte("test data")

		err = local.SavePrivateKey(testData)
		require.NoError(t, err)

		filePath := filepath.Join(local.appDir, domain.PrivKeyFileName)

		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(keyFilePermissions), info.Mode().Perm())
	})
}
