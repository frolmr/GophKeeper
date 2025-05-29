package device

import (
	"crypto/rsa"
	"errors"
	"os"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/crypto"
	"github.com/frolmr/GophKeeper/internal/client/domain"
	mocks "github.com/frolmr/GophKeeper/internal/client/mocks/app/device"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeviceService_AddDevice(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful device addition",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				// Setup for getOrGenPrivateKey
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})
				dk := &crypto.DeviceKey{PrivateKey: &rsa.PrivateKey{}}
				cp.EXPECT().GenerateDeviceKey().Return(dk, nil)
				cp.EXPECT().PrivateKeyToBytes(dk.PrivateKey).Return([]byte("private-key-bytes"), nil)
				r.EXPECT().SavePrivateKey([]byte("private-key-bytes")).Return(nil)

				// Setup for AddDevice
				cp.EXPECT().PublicKeyToBytes(&dk.PrivateKey.PublicKey).Return([]byte("public-key-bytes"), nil)
				dc.EXPECT().SendAddDeviceRequest([]byte("public-key-bytes")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "fails when private key generation fails",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})
				cp.EXPECT().GenerateDeviceKey().Return(nil, errors.New("generation error"))
				dc.EXPECT().SendAddDeviceRequest(gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "failed to get pk: device key generation error: generation error",
		},
		{
			name: "fails when public key conversion fails",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				// Setup for getOrGenPrivateKey
				r.EXPECT().ReadPrivateKey().Return([]byte("existing-key"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("existing-key")).Return(privKey, nil)

				// Setup for AddDevice
				cp.EXPECT().PublicKeyToBytes(&privKey.PublicKey).Return(nil, errors.New("conversion error"))
				dc.EXPECT().SendAddDeviceRequest(gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "error getting public key from private: conversion error",
		},
		{
			name: "fails when add device request fails",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				// Setup for getOrGenPrivateKey
				r.EXPECT().ReadPrivateKey().Return([]byte("existing-key"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("existing-key")).Return(privKey, nil)

				// Setup for AddDevice
				cp.EXPECT().PublicKeyToBytes(&privKey.PublicKey).Return([]byte("public-key-bytes"), nil)
				dc.EXPECT().SendAddDeviceRequest([]byte("public-key-bytes")).Return(errors.New("request failed"))
			},
			expectError:  true,
			errorMessage: "failed to add device: request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockCrypto := mocks.NewMockCryptoProcessor(ctrl)
			mockConnector := mocks.NewMockConnector(ctrl)

			tt.setupMocks(mockRepo, mockCrypto, mockConnector)

			service := NewDeviceService(mockRepo, mockCrypto, mockConnector)
			err := service.AddDevice()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceService_ConfirmDevice(t *testing.T) {
	tests := []struct {
		name         string
		deviceID     string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name:     "successful device confirmation",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				// Setup private key
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				// Setup MK retrieval and decryption
				dc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				// Setup device PK retrieval
				dc.EXPECT().SendGetDevicePKRequest("device-123").Return([]byte("device-pk-bytes"), nil)
				pubKey := &rsa.PublicKey{}
				cp.EXPECT().BytesToPublicKey([]byte("device-pk-bytes")).Return(pubKey, nil)

				// Setup MK encryption and approval
				cp.EXPECT().EncryptWithPublicKey([]byte("decrypted-mk"), pubKey).Return([]byte("re-encrypted-mk"), nil)
				dc.EXPECT().SendApproveDeviceRequest("device-123", []byte("re-encrypted-mk")).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "fails when reading private key fails",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return(nil, errors.New("read error"))
				dc.EXPECT().SendGetDeviceMKRequest().Times(0)
			},
			expectError:  true,
			errorMessage: "error reading PK: read error",
		},
		{
			name:     "fails when getting MK fails",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				dc.EXPECT().SendGetDeviceMKRequest().Return(nil, errors.New("mk error"))
				dc.EXPECT().SendGetDevicePKRequest(gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "error getting users MK: mk error",
		},
		{
			name:     "fails when MK decryption fails",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				dc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return(nil, errors.New("decryption error"))

				dc.EXPECT().SendGetDevicePKRequest(gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "MK decryption failed: decryption error",
		},
		{
			name:     "fails when getting device PK fails",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				dc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				dc.EXPECT().SendGetDevicePKRequest("device-123").Return(nil, errors.New("pk error"))
				dc.EXPECT().SendApproveDeviceRequest(gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "error getting device by id: pk error",
		},
		{
			name:     "fails when approving device fails",
			deviceID: "device-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, dc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				dc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				dc.EXPECT().SendGetDevicePKRequest("device-123").Return([]byte("device-pk-bytes"), nil)
				pubKey := &rsa.PublicKey{}
				cp.EXPECT().BytesToPublicKey([]byte("device-pk-bytes")).Return(pubKey, nil)

				cp.EXPECT().EncryptWithPublicKey([]byte("decrypted-mk"), pubKey).Return([]byte("re-encrypted-mk"), nil)
				dc.EXPECT().SendApproveDeviceRequest("device-123", []byte("re-encrypted-mk")).Return(errors.New("approval error"))
			},
			expectError:  true,
			errorMessage: "approval error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockCrypto := mocks.NewMockCryptoProcessor(ctrl)
			mockConnector := mocks.NewMockConnector(ctrl)

			tt.setupMocks(mockRepo, mockCrypto, mockConnector)

			service := NewDeviceService(mockRepo, mockCrypto, mockConnector)
			err := service.ConfirmDevice(tt.deviceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeviceService_ListDevices(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockConnector)
		expectError  bool
		expectedList []domain.DeviceOnServer
		errorMessage string
	}{
		{
			name: "successful device listing",
			setupMocks: func(dc *mocks.MockConnector) {
				expected := []domain.DeviceOnServer{
					{ID: "device-1", Name: "Laptop"},
					{ID: "device-2", Name: "Phone"},
				}
				dc.EXPECT().SendListDevicesRequest().Return(expected, nil)
			},
			expectError: false,
			expectedList: []domain.DeviceOnServer{
				{ID: "device-1", Name: "Laptop"},
				{ID: "device-2", Name: "Phone"},
			},
		},
		{
			name: "fails when list request fails",
			setupMocks: func(dc *mocks.MockConnector) {
				dc.EXPECT().SendListDevicesRequest().Return(nil, errors.New("list error"))
			},
			expectError:  true,
			errorMessage: "list devices request failed: list error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockCrypto := mocks.NewMockCryptoProcessor(ctrl)
			mockConnector := mocks.NewMockConnector(ctrl)

			tt.setupMocks(mockConnector)

			service := NewDeviceService(mockRepo, mockCrypto, mockConnector)
			devices, err := service.ListDevices()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, devices)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedList, devices)
			}
		})
	}
}

func TestDeviceService_getOrGenPrivateKey(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor)
		expectError  bool
		errorMessage string
	}{
		{
			name: "existing private key",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor) {
				r.EXPECT().ReadPrivateKey().Return([]byte("existing-key"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("existing-key")).Return(&rsa.PrivateKey{}, nil)
			},
			expectError: false,
		},
		{
			name: "new private key generation",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor) {
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})

				dk := &crypto.DeviceKey{PrivateKey: &rsa.PrivateKey{}}
				cp.EXPECT().GenerateDeviceKey().Return(dk, nil)
				cp.EXPECT().PrivateKeyToBytes(dk.PrivateKey).Return([]byte("private-key-bytes"), nil)
				r.EXPECT().SavePrivateKey([]byte("private-key-bytes")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "error reading private key",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor) {
				r.EXPECT().ReadPrivateKey().Return(nil, errors.New("read error"))
				cp.EXPECT().GenerateDeviceKey().Times(0)
			},
			expectError:  true,
			errorMessage: "error private key reading: read error",
		},
		{
			name: "error generating device key",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor) {
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})
				cp.EXPECT().GenerateDeviceKey().Return(nil, errors.New("generation error"))
			},
			expectError:  true,
			errorMessage: "device key generation error: generation error",
		},
		{
			name: "error saving private key",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor) {
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})

				dk := &crypto.DeviceKey{PrivateKey: &rsa.PrivateKey{}}
				cp.EXPECT().GenerateDeviceKey().Return(dk, nil)
				cp.EXPECT().PrivateKeyToBytes(dk.PrivateKey).Return([]byte("private-key-bytes"), nil)
				r.EXPECT().SavePrivateKey([]byte("private-key-bytes")).Return(errors.New("save error"))
			},
			expectError:  true,
			errorMessage: "failed to save private key: save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockCrypto := mocks.NewMockCryptoProcessor(ctrl)
			mockConnector := mocks.NewMockConnector(ctrl)

			tt.setupMocks(mockRepo, mockCrypto)

			service := NewDeviceService(mockRepo, mockCrypto, mockConnector)
			_, err := service.getOrGenPrivateKey()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
