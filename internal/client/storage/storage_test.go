package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func getTestConfigDir(t *testing.T) (string, error) {
	testDir := t.TempDir()
	appDir := filepath.Join(testDir, domain.AppName)
	if err := os.MkdirAll(appDir, appDirPermissionsMask); err != nil {
		return "", fmt.Errorf("failed to create app directory: %w", err)
	}
	return appDir, nil
}

func newTestLocalStorage(t *testing.T) (*Storage, error) {
	appConfigDir, err := getTestConfigDir(t)
	if err != nil {
		return nil, fmt.Errorf("failed to get app directory: %w", err)
	}

	return &Storage{
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
		assert.Contains(t, err.Error(), "failed to read key file")
	})

	t.Run("Directory has correct permissions", func(t *testing.T) {
		testDir := t.TempDir()
		appDir := filepath.Join(testDir, domain.AppName)

		os.RemoveAll(appDir)

		err := os.MkdirAll(appDir, appDirPermissionsMask)
		require.NoError(t, err)

		info, err := os.Stat(appDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(appDirPermissionsMask), info.Mode().Perm())
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
		assert.Equal(t, os.FileMode(keyFilePermissionsMask), info.Mode().Perm())
	})
}

func TestTokenOperations(t *testing.T) {
	setupStorage := func(t *testing.T) *Storage {
		testDir := t.TempDir()
		appDir := filepath.Join(testDir, domain.AppName)
		require.NoError(t, os.MkdirAll(appDir, appDirPermissionsMask))

		dbPath := filepath.Join(appDir, dbFileName)
		db, err := bbolt.Open(dbPath, dbPersmissionsMask, nil)
		require.NoError(t, err)

		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(authBucket))
			return err
		})
		require.NoError(t, err)

		return &Storage{
			appDir: appDir,
			db:     db,
		}
	}

	t.Run("save and get token", func(t *testing.T) {
		storage := setupStorage(t)
		defer storage.Close()

		//nolint:gosec // it is ok in tests
		testToken := "test.jwt.token"
		err := storage.SaveToken(testToken)
		require.NoError(t, err)

		token, err := storage.GetToken()
		require.NoError(t, err)
		assert.Equal(t, testToken, token)
	})

	t.Run("save empty token", func(t *testing.T) {
		storage := setupStorage(t)
		defer storage.Close()

		err := storage.SaveToken("")
		require.Error(t, err)
		assert.Equal(t, "token cannot be empty", err.Error())
	})

	t.Run("get non-existent token", func(t *testing.T) {
		storage := setupStorage(t)
		defer storage.Close()

		_, err := storage.GetToken()
		require.Error(t, err)
		assert.Equal(t, "token not found", err.Error())
	})

	t.Run("save token with db error", func(t *testing.T) {
		storage := setupStorage(t)
		storage.Close()

		err := storage.SaveToken("test.token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database not open")
	})

	t.Run("get token with db error", func(t *testing.T) {
		storage := setupStorage(t)
		storage.Close()

		_, err := storage.GetToken()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database not open")
	})
}

func TestClose(t *testing.T) {
	t.Run("close nil db", func(t *testing.T) {
		storage := &Storage{db: nil}
		err := storage.Close()
		assert.NoError(t, err)
	})

	t.Run("close with db", func(t *testing.T) {
		storage, err := NewStorage()
		require.NoError(t, err)

		err = storage.Close()
		assert.NoError(t, err)

		err = storage.db.Update(func(tx *bbolt.Tx) error { return nil })
		assert.Error(t, err)
	})
}

func TestDBInitialization(t *testing.T) {
	t.Run("bucket creation on init", func(t *testing.T) {
		storage, err := NewStorage()
		require.NoError(t, err)
		defer storage.Close()

		err = storage.db.View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(authBucket))
			if bucket == nil {
				return errors.New("bucket not created")
			}
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("db file permissions", func(t *testing.T) {
		storage, err := NewStorage()
		require.NoError(t, err)
		defer storage.Close()

		dbPath := filepath.Join(storage.appDir, dbFileName)
		info, err := os.Stat(dbPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(dbPersmissionsMask), info.Mode().Perm())
	})
}
