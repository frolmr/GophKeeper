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

func TestCreateAndReturnUser(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	tests := []struct {
		name        string
		email       string
		password    string
		mockClosure func(mock sqlmock.Sqlmock)
		wantUser    *domain.User
		wantErr     bool
	}{
		{
			name:     "successful user creation",
			email:    "test@example.com",
			password: "password123",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email")
				mock.ExpectQuery("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email").
					WithArgs("test@example.com", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "email"}).
						AddRow(uuid.New(), "test@example.com"))
			},
			wantUser: &domain.User{Email: "test@example.com"},
			wantErr:  false,
		},
		{
			name:     "prepare statement error",
			email:    "test@example.com",
			password: "password123",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users").
					WillReturnError(errors.New("prepare error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
		{
			name:     "query execution error",
			email:    "test@example.com",
			password: "password123",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email")
				mock.ExpectQuery("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email").
					WithArgs("test@example.com", sqlmock.AnyArg()).
					WillReturnError(errors.New("execution error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			user, err := storage.createAndReturnUser(context.Background(), tx, tt.email, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser.Email, user.Email)
				assert.NotNil(t, user.UUID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUUID := uuid.New()
	tests := []struct {
		name        string
		email       string
		mockClosure func(mock sqlmock.Sqlmock)
		wantUser    *domain.User
		wantErr     bool
	}{
		{
			name:  "successful get user",
			email: "test@example.com",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE email = \\$1").
					ExpectQuery().
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "email", "password_hash", "email_confirmed"}).
						AddRow(testUUID, "test@example.com", "hashed_password", true))
			},
			wantUser: &domain.User{
				UUID:           testUUID,
				Email:          "test@example.com",
				PasswordHash:   "hashed_password",
				EmailConfirmed: true,
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE email = \\$1").
					ExpectQuery().
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantUser: nil,
			wantErr:  false,
		},
		{
			name:  "prepare statement error",
			email: "test@example.com",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE email = \\$1").
					WillReturnError(errors.New("prepare error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
		{
			name:  "query execution error",
			email: "test@example.com",
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE email = \\$1").
					ExpectQuery().
					WithArgs("test@example.com").
					WillReturnError(errors.New("execution error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			user, err := storage.GetUserByEmail(context.Background(), tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserByUUID(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUUID := uuid.New()
	tests := []struct {
		name        string
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantUser    *domain.User
		wantErr     bool
	}{
		{
			name:     "successful get user by UUID",
			userUUID: testUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE uuid = \\$1").
					ExpectQuery().
					WithArgs(testUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "email", "password_hash", "email_confirmed"}).
						AddRow(testUUID, "test@example.com", "hashed_password", true))
			},
			wantUser: &domain.User{
				UUID:           testUUID,
				Email:          "test@example.com",
				PasswordHash:   "hashed_password",
				EmailConfirmed: true,
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			userUUID: uuid.New(),
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, email, password_hash, email_confirmed FROM users WHERE uuid = \\$1").
					ExpectQuery().
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			wantUser: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			user, err := storage.GetUserByUUID(context.Background(), tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, user)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateUserAndDevice(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUUID := uuid.New()
	testDeviceID := "device123"
	testMasterKey := []byte("master-key")

	tests := []struct {
		name        string
		email       string
		password    string
		deviceName  string
		deviceID    string
		masterKey   []byte
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful user and device creation",
			email:      "test@example.com",
			password:   "password123",
			deviceName: "test-device",
			deviceID:   testDeviceID,
			masterKey:  testMasterKey,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email")
				mock.ExpectQuery("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email").
					WithArgs("test@example.com", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "email"}).
						AddRow(testUUID, "test@example.com"))
				mock.ExpectPrepare("INSERT INTO devices \\(name, id, master_key, user_uuid, confirmed\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, true\\)")
				mock.ExpectExec("INSERT INTO devices \\(name, id, master_key, user_uuid, confirmed\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, true\\)").
					WithArgs("test-device", testDeviceID, testMasterKey, testUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:       "user creation fails",
			email:      "test@example.com",
			password:   "password123",
			deviceName: "test-device",
			deviceID:   testDeviceID,
			masterKey:  testMasterKey,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email")
				mock.ExpectQuery("INSERT INTO users \\(email, password_hash\\) VALUES \\(\\$1, \\$2\\) RETURNING uuid, email").
					WithArgs("test@example.com", sqlmock.AnyArg()).
					WillReturnError(errors.New("user creation error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.CreateUserAndDevice(
				context.Background(),
				tt.email,
				tt.password,
				tt.deviceName,
				tt.deviceID,
				tt.masterKey,
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
