package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetDeviceByUserUUIDAndID(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	testDeviceID := "device123"
	testDeviceUUID := uuid.New()

	tests := []struct {
		name        string
		deviceID    string
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantDevice  *domain.Device
		wantErr     bool
	}{
		{
			name:     "successful get device by ID",
			deviceID: testDeviceID,
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND id = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, testDeviceID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "id", "master_key", "public_key", "confirmed"}).
						AddRow(testDeviceUUID, "test-device", testDeviceID, []byte("master-key"), []byte("public-key"), true))
			},
			wantDevice: &domain.Device{
				UUID:      testDeviceUUID,
				Name:      "test-device",
				ID:        testDeviceID,
				MK:        []byte("master-key"),
				PK:        []byte("public-key"),
				Confirmed: true,
			},
			wantErr: false,
		},
		{
			name:     "device not found",
			deviceID: "nonexistent",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND id = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, "nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantDevice: nil,
			wantErr:    false,
		},
		{
			name:     "prepare statement error",
			deviceID: testDeviceID,
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND id = \\$2").
					WillReturnError(errors.New("prepare error"))
			},
			wantDevice: nil,
			wantErr:    true,
		},
		{
			name:     "query execution error",
			deviceID: testDeviceID,
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND id = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, testDeviceID).
					WillReturnError(errors.New("execution error"))
			},
			wantDevice: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			device, err := storage.GetDeviceByUserUUIDAndID(context.Background(), tt.deviceID, tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDevice, device)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetDeviceByUserUUIDAndUUID(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	testDeviceUUID := uuid.New()

	tests := []struct {
		name        string
		deviceUUID  uuid.UUID
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantDevice  *domain.Device
		wantErr     bool
	}{
		{
			name:       "successful get device by UUID",
			deviceUUID: testDeviceUUID,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND uuid = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, testDeviceUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "id", "master_key", "public_key", "confirmed"}).
						AddRow(testDeviceUUID, "test-device", "device123", []byte("master-key"), []byte("public-key"), true))
			},
			wantDevice: &domain.Device{
				UUID:      testDeviceUUID,
				Name:      "test-device",
				ID:        "device123",
				MK:        []byte("master-key"),
				PK:        []byte("public-key"),
				Confirmed: true,
			},
			wantErr: false,
		},
		{
			name:       "device not found",
			deviceUUID: uuid.New(),
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = \\$1 AND uuid = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			wantDevice: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			device, err := storage.GetDeviceByUserUUIDAndUUID(context.Background(), tt.deviceUUID, tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDevice, device)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAllUserDevices(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	device1UUID := uuid.New()
	device2UUID := uuid.New()

	tests := []struct {
		name        string
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantDevices []*domain.Device
		wantErr     bool
	}{
		{
			name:     "successful get all devices",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, confirmed FROM devices WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "confirmed"}).
						AddRow(device1UUID, "device1", true).
						AddRow(device2UUID, "device2", false))
			},
			wantDevices: []*domain.Device{
				{UUID: device1UUID, Name: "device1", Confirmed: true},
				{UUID: device2UUID, Name: "device2", Confirmed: false},
			},
			wantErr: false,
		},
		{
			name:     "no devices found",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, confirmed FROM devices WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "confirmed"}))
			},
			wantDevices: []*domain.Device(nil),
			wantErr:     false,
		},
		{
			name:     "prepare statement error",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, confirmed FROM devices WHERE user_uuid = \\$1").
					WillReturnError(errors.New("prepare error"))
			},
			wantDevices: nil,
			wantErr:     true,
		},
		{
			name:     "query execution error",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, confirmed FROM devices WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantDevices: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			devices, err := storage.GetAllUserDevices(context.Background(), tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDevices, devices)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddDevice(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUser := &domain.User{
		UUID:  uuid.New(),
		Email: "test@example.com",
	}
	testDeviceName := "test-device"
	testDeviceID := "device123"
	testPublicKey := []byte("public-key")

	tests := []struct {
		name        string
		deviceName  string
		deviceID    string
		publicKey   []byte
		user        *domain.User
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful add device",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			publicKey:  testPublicKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, public_key, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					ExpectExec().
					WithArgs(testDeviceName, testDeviceID, testPublicKey, testUser.UUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "prepare statement error",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			publicKey:  testPublicKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, public_key, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:       "execution error",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			publicKey:  testPublicKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, public_key, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					ExpectExec().
					WithArgs(testDeviceName, testDeviceID, testPublicKey, testUser.UUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.AddDevice(context.Background(), tt.deviceName, tt.deviceID, tt.publicKey, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestConfirmDevice(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	testDeviceUUID := uuid.New()
	testMasterKey := []byte("master-key")

	tests := []struct {
		name        string
		deviceUUID  uuid.UUID
		userUUID    uuid.UUID
		masterKey   []byte
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful confirm device",
			deviceUUID: testDeviceUUID,
			userUUID:   testUserUUID,
			masterKey:  testMasterKey,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE devices SET confirmed = TRUE, master_key = \\$1 WHERE uuid = \\$2 AND user_uuid = \\$3").
					ExpectExec().
					WithArgs(testMasterKey, testDeviceUUID, testUserUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "prepare statement error",
			deviceUUID: testDeviceUUID,
			userUUID:   testUserUUID,
			masterKey:  testMasterKey,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE devices SET confirmed = TRUE, master_key = \\$1 WHERE uuid = \\$2 AND user_uuid = \\$3").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:       "execution error",
			deviceUUID: testDeviceUUID,
			userUUID:   testUserUUID,
			masterKey:  testMasterKey,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE devices SET confirmed = TRUE, master_key = \\$1 WHERE uuid = \\$2 AND user_uuid = \\$3").
					ExpectExec().
					WithArgs(testMasterKey, testDeviceUUID, testUserUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.ConfirmDevice(context.Background(), tt.deviceUUID, tt.userUUID, tt.masterKey)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateConfirmedDevice(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUser := &domain.User{
		UUID:  uuid.New(),
		Email: "test@example.com",
	}
	testDeviceName := "test-device"
	testDeviceID := "device123"
	testMasterKey := []byte("master-key")

	tests := []struct {
		name        string
		deviceName  string
		deviceID    string
		masterKey   []byte
		user        *domain.User
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful create confirmed device",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			masterKey:  testMasterKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, master_key, user_uuid, confirmed\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, true\\)").
					ExpectExec().
					WithArgs(testDeviceName, testDeviceID, testMasterKey, testUser.UUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "prepare statement error",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			masterKey:  testMasterKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, master_key, user_uuid, confirmed\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, true\\)").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:       "execution error",
			deviceName: testDeviceName,
			deviceID:   testDeviceID,
			masterKey:  testMasterKey,
			user:       testUser,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, master_key, user_uuid, confirmed\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, true\\)").
					ExpectExec().
					WithArgs(testDeviceName, testDeviceID, testMasterKey, testUser.UUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			err = storage.createConfirmedDevice(context.Background(), tx, tt.deviceName, tt.deviceID, tt.masterKey, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
