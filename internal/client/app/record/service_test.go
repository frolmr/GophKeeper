package record

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	mocks "github.com/frolmr/GophKeeper/internal/client/mocks/app/record"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRecordService_AddRecord(t *testing.T) {
	tests := []struct {
		name         string
		record       *domain.RawRecord
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful record addition",
			record: &domain.RawRecord{
				Name: "test",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "user",
					Password: "pass",
				},
			},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				payloadBytes, _ := json.Marshal(domain.PasswordPayload{Login: "user", Password: "pass"})
				cp.EXPECT().EncryptWithMasterKey(payloadBytes, []byte("decrypted-mk")).Return([]byte("encrypted-payload"), nil)

				rc.EXPECT().SendAddRecordRequest(gomock.Any()).DoAndReturn(func(rec domain.EncryptedRecord) error {
					assert.Equal(t, "test", rec.Name)
					assert.Equal(t, domain.PasswordPayloadKind, rec.Kind)
					assert.NotEmpty(t, rec.Payload)
					return nil
				})
			},
			expectError: false,
		},
		{
			name:   "fails when reading private key fails",
			record: &domain.RawRecord{},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return(nil, errors.New("read error"))
				rc.EXPECT().SendGetDeviceMKRequest().Times(0)
			},
			expectError:  true,
			errorMessage: "error reading PK: read error",
		},
		{
			name:   "fails when getting MK fails",
			record: &domain.RawRecord{},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return(nil, errors.New("mk error"))
			},
			expectError:  true,
			errorMessage: "error getting users MK: mk error",
		},
		{
			name: "fails when payload encryption fails",
			record: &domain.RawRecord{
				Name: "test",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "user",
					Password: "pass",
				},
			},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				payloadBytes, _ := json.Marshal(domain.PasswordPayload{Login: "user", Password: "pass"})
				cp.EXPECT().EncryptWithMasterKey(payloadBytes, []byte("decrypted-mk")).Return(nil, errors.New("encryption error"))
			},
			expectError:  true,
			errorMessage: "failed to encrypt payload: encryption error",
		},
		{
			name: "fails when payload marshaling fails",
			record: &domain.RawRecord{
				Name:    "test",
				Kind:    domain.PasswordPayloadKind,
				Payload: make(chan int), // Unmarshalable payload
			},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)
				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)
			},
			expectError:  true,
			errorMessage: "error payload convertation: json: unsupported type: chan int",
		},
		{
			name: "fails when private key conversion fails",
			record: &domain.RawRecord{
				Name: "test",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "user",
					Password: "pass",
				},
			},
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(nil, errors.New("conversion error"))
				rc.EXPECT().SendGetDeviceMKRequest().Times(0)
			},
			expectError:  true,
			errorMessage: "error private key convertation: conversion error",
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

			service := NewRecordService(mockRepo, mockCrypto, mockConnector)
			err := service.AddRecord(tt.record)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordService_DeleteRecord(t *testing.T) {
	tests := []struct {
		name         string
		recUUID      string
		setupMocks   func(*mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name:    "successful record deletion",
			recUUID: "record-123",
			setupMocks: func(rc *mocks.MockConnector) {
				rc.EXPECT().SendDeleteRecordRequest("record-123").Return(nil)
			},
			expectError: false,
		},
		{
			name:    "fails when delete request fails",
			recUUID: "record-123",
			setupMocks: func(rc *mocks.MockConnector) {
				rc.EXPECT().SendDeleteRecordRequest("record-123").Return(errors.New("delete error"))
			},
			expectError:  true,
			errorMessage: "delete record request failed: delete error",
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

			service := NewRecordService(mockRepo, mockCrypto, mockConnector)
			err := service.DeleteRecord(tt.recUUID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordService_ListRecords(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockConnector)
		expectError  bool
		expectedList []domain.EncryptedRecord
		errorMessage string
	}{
		{
			name: "successful record listing",
			setupMocks: func(rc *mocks.MockConnector) {
				expected := []domain.EncryptedRecord{
					{UUID: "record-1", Name: "Test 1"},
					{UUID: "record-2", Name: "Test 2"},
				}
				rc.EXPECT().SendListRecordsRequest().Return(expected, nil)
			},
			expectError: false,
			expectedList: []domain.EncryptedRecord{
				{UUID: "record-1", Name: "Test 1"},
				{UUID: "record-2", Name: "Test 2"},
			},
		},
		{
			name: "fails when list request fails",
			setupMocks: func(rc *mocks.MockConnector) {
				rc.EXPECT().SendListRecordsRequest().Return(nil, errors.New("list error"))
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

			service := NewRecordService(mockRepo, mockCrypto, mockConnector)
			records, err := service.ListRecords()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, records)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedList, records)
			}
		})
	}
}

func TestRecordService_ReadRecord(t *testing.T) {
	tests := []struct {
		name         string
		recID        string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		expectedRec  *domain.RawRecord
		errorMessage string
	}{
		{
			name:  "successful record reading (password kind)",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				encRec := &domain.EncryptedRecord{
					UUID:    "record-123",
					Name:    "Test Password",
					Kind:    domain.PasswordPayloadKind,
					Payload: []byte("encrypted-payload"),
				}
				rc.EXPECT().SendGetRecordRequest("record-123").Return(encRec, nil)

				payloadBytes, _ := json.Marshal(domain.PasswordPayload{Login: "user", Password: "pass"})
				cp.EXPECT().DecryptWithMasterKey([]byte("encrypted-payload"), []byte("decrypted-mk")).Return(payloadBytes, nil)
			},
			expectError: false,
			expectedRec: &domain.RawRecord{
				Name: "Test Password",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "user",
					Password: "pass",
				},
			},
		},
		{
			name:  "fails when getting record fails",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				rc.EXPECT().SendGetRecordRequest("record-123").Return(nil, errors.New("get error"))
			},
			expectError:  true,
			errorMessage: "error getting record: get error",
		},
		{
			name:  "fails when payload decryption fails",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				encRec := &domain.EncryptedRecord{
					UUID:    "record-123",
					Name:    "Test Password",
					Kind:    domain.PasswordPayloadKind,
					Payload: []byte("encrypted-payload"),
				}
				rc.EXPECT().SendGetRecordRequest("record-123").Return(encRec, nil)

				cp.EXPECT().DecryptWithMasterKey([]byte("encrypted-payload"), []byte("decrypted-mk")).Return(nil, errors.New("decryption error"))
			},
			expectError:  true,
			errorMessage: "failed to decrypt payload: decryption error",
		},
		{
			name:  "fails for unknown payload kind",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				encRec := &domain.EncryptedRecord{
					UUID:    "record-123",
					Name:    "Test Unknown",
					Kind:    "unknown-kind",
					Payload: []byte("encrypted-payload"),
				}
				rc.EXPECT().SendGetRecordRequest("record-123").Return(encRec, nil)

				cp.EXPECT().DecryptWithMasterKey([]byte("encrypted-payload"), []byte("decrypted-mk")).Return([]byte("{}"), nil)
			},
			expectError:  true,
			errorMessage: "unknown payload kind",
		},
		{
			name:  "fails when payload unmarshaling fails",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)
				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)
				encRec := &domain.EncryptedRecord{
					UUID:    "record-123",
					Name:    "Test Password",
					Kind:    domain.PasswordPayloadKind,
					Payload: []byte("encrypted-payload"),
				}
				rc.EXPECT().SendGetRecordRequest("record-123").Return(encRec, nil)
				cp.EXPECT().DecryptWithMasterKey([]byte("encrypted-payload"), []byte("decrypted-mk")).Return([]byte("invalid-json"), nil)
			},
			expectError:  true,
			errorMessage: "failed to unmarshal payload: invalid character 'i' looking for beginning of value",
		},
		{
			name:  "fails when private key conversion fails",
			recID: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(nil, errors.New("conversion error"))
				rc.EXPECT().SendGetDeviceMKRequest().Times(0)
			},
			expectError:  true,
			errorMessage: "error private key convertation: conversion error",
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

			service := NewRecordService(mockRepo, mockCrypto, mockConnector)
			record, err := service.ReadRecord(tt.recID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRec.Name, record.Name)
				assert.Equal(t, tt.expectedRec.Kind, record.Kind)

				switch v := record.Payload.(type) {
				case domain.PasswordPayload:
					assert.Equal(t, tt.expectedRec.Payload.(domain.PasswordPayload), v)
				case map[string]interface{}:
					expected := tt.expectedRec.Payload.(domain.PasswordPayload)
					assert.Equal(t, expected.Login, v["login"])
					assert.Equal(t, expected.Password, v["password"])
				default:
					t.Errorf("unexpected payload type: %T", v)
				}
			}
		})
	}
}

func TestRecordService_UpdateRecord(t *testing.T) {
	tests := []struct {
		name         string
		record       *domain.RawRecord
		id           string
		setupMocks   func(*mocks.MockRepository, *mocks.MockCryptoProcessor, *mocks.MockConnector)
		expectError  bool
		errorMessage string
	}{
		{
			name: "successful record update",
			record: &domain.RawRecord{
				Name: "updated",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "new-user",
					Password: "new-pass",
				},
			},
			id: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)

				payloadBytes, _ := json.Marshal(domain.PasswordPayload{Login: "new-user", Password: "new-pass"})
				cp.EXPECT().EncryptWithMasterKey(payloadBytes, []byte("decrypted-mk")).Return([]byte("encrypted-payload"), nil)

				rc.EXPECT().SendUpdateRecordRequest(gomock.Any()).DoAndReturn(func(rec domain.EncryptedRecord) error {
					assert.Equal(t, "record-123", rec.UUID)
					assert.Equal(t, "updated", rec.Name)
					assert.Equal(t, domain.PasswordPayloadKind, rec.Kind)
					assert.NotEmpty(t, rec.Payload)
					return nil
				})
			},
			expectError: false,
		},
		{
			name:   "fails when MK decryption fails",
			record: &domain.RawRecord{},
			id:     "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)

				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return(nil, errors.New("decryption error"))
			},
			expectError:  true,
			errorMessage: "MK decryption failed: decryption error",
		},
		{
			name: "fails when payload marshaling fails",
			record: &domain.RawRecord{
				Name:    "test",
				Kind:    domain.PasswordPayloadKind,
				Payload: make(chan int), // Unmarshalable payload
			},
			id: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				privKey := &rsa.PrivateKey{}
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(privKey, nil)
				rc.EXPECT().SendGetDeviceMKRequest().Return([]byte("encrypted-mk"), nil)
				cp.EXPECT().DecryptWithPrivateKey([]byte("encrypted-mk"), privKey).Return([]byte("decrypted-mk"), nil)
			},
			expectError:  true,
			errorMessage: "error payload convertation: json: unsupported type: chan int",
		},
		{
			name: "fails when private key conversion fails",
			record: &domain.RawRecord{
				Name: "test",
				Kind: domain.PasswordPayloadKind,
				Payload: domain.PasswordPayload{
					Login:    "user",
					Password: "pass",
				},
			},
			id: "record-123",
			setupMocks: func(r *mocks.MockRepository, cp *mocks.MockCryptoProcessor, rc *mocks.MockConnector) {
				r.EXPECT().ReadPrivateKey().Return([]byte("private-key-bytes"), nil)
				cp.EXPECT().BytesToPrivateKey([]byte("private-key-bytes")).Return(nil, errors.New("conversion error"))
				rc.EXPECT().SendGetDeviceMKRequest().Times(0)
			},
			expectError:  true,
			errorMessage: "error private key convertation: conversion error",
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

			service := NewRecordService(mockRepo, mockCrypto, mockConnector)
			err := service.UpdateRecord(tt.record, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
