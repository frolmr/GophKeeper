package user

import (
	"crypto/rsa"
	"errors"
	"os"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/crypto"
	mocks "github.com/frolmr/GophKeeper/internal/client/mocks/app/user"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		password     string
		setupMocks   func(*mocks.MockConnector, *mocks.MockRepository)
		expectError  bool
		errorMessage string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(uc *mocks.MockConnector, r *mocks.MockRepository) {
				uc.EXPECT().SendLoginRequest("test@example.com", "password123").
					Return("valid-token", nil)
				r.EXPECT().SaveToken("valid-token").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "login fails on connector",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(uc *mocks.MockConnector, r *mocks.MockRepository) {
				uc.EXPECT().SendLoginRequest("test@example.com", "wrongpassword").
					Return("", errors.New("invalid credentials"))
				r.EXPECT().SaveToken(gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "response token missing: invalid credentials",
		},
		{
			name:     "login fails on token save",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(uc *mocks.MockConnector, r *mocks.MockRepository) {
				uc.EXPECT().SendLoginRequest("test@example.com", "password123").
					Return("valid-token", nil)
				r.EXPECT().SaveToken("valid-token").
					Return(errors.New("storage error"))
			},
			expectError:  true,
			errorMessage: "token save error: storage error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockConnector := mocks.NewMockConnector(ctrl)
			mockRepo := mocks.NewMockRepository(ctrl)
			mockCrypto := mocks.NewMockCryptoProcessor(ctrl)

			tt.setupMocks(mockConnector, mockRepo)

			service := NewUserService(mockRepo, mockCrypto, mockConnector)
			err := service.Login(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		password     string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name:     "successful registration",
			email:    "new@example.com",
			password: "newpassword",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, uc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return(nil, &os.PathError{Op: "open", Path: "nonexistent", Err: os.ErrNotExist})

				dk := &crypto.DeviceKey{PrivateKey: &rsa.PrivateKey{}}
				cp.EXPECT().GenerateDeviceKey().Return(dk, nil)
				cp.EXPECT().PrivateKeyToBytes(dk.PrivateKey).Return([]byte("private-key-bytes"), nil)
				r.EXPECT().SavePrivateKey([]byte("private-key-bytes")).Return(nil)

				mk := []byte("master-key")
				cp.EXPECT().GenerateMasterKey().Return(mk, nil)
				cp.EXPECT().EncryptWithPublicKey(mk, &dk.PrivateKey.PublicKey).Return([]byte("encrypted-mk"), nil)
				uc.EXPECT().SendRegisterRequest("new@example.com", "newpassword", []byte("encrypted-mk")).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "registration fails when private key exists but can't be read",
			email:    "new@example.com",
			password: "newpassword",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, uc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return(nil, errors.New("read error"))
				uc.EXPECT().SendRegisterRequest(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "failed to get pk: error private key reading: read error",
		},
		{
			name:     "registration fails when master key generation fails",
			email:    "new@example.com",
			password: "newpassword",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, uc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("existing-key"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("existing-key")).Return(&rsa.PrivateKey{}, nil)

				cp.EXPECT().GenerateMasterKey().Return(nil, errors.New("generation error"))
				uc.EXPECT().SendRegisterRequest(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectError:  true,
			errorMessage: "master key generation error: generation error",
		},
		{
			name:     "registration fails when sending request fails",
			email:    "new@example.com",
			password: "newpassword",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, uc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("existing-key"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("existing-key")).Return(&rsa.PrivateKey{}, nil)

				mk := []byte("master-key")
				cp.EXPECT().GenerateMasterKey().Return(mk, nil)
				cp.EXPECT().EncryptWithPublicKey(mk, gomock.Any()).Return([]byte("encrypted-mk"), nil)
				uc.EXPECT().SendRegisterRequest("new@example.com", "newpassword", []byte("encrypted-mk")).
					Return(errors.New("connection error"))
			},
			expectError:  true,
			errorMessage: "response token missing: connection error",
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

			service := NewUserService(mockRepo, mockCrypto, mockConnector)
			err := service.Register(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_getOrGenPrivateKey(t *testing.T) {
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

			service := NewUserService(mockRepo, mockCrypto, mockConnector)
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
