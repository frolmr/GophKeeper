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

func TestAddRecord(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	testName := "test-record"
	testKind := "password"
	testPayload := []byte("test-payload")
	testMetadata := []byte("test-metadata")

	tests := []struct {
		name        string
		recordName  string
		kind        string
		payload     []byte
		metadata    []byte
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful add record",
			recordName: testName,
			kind:       testKind,
			payload:    testPayload,
			metadata:   testMetadata,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO records \\(name, kind, payload, metadata, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
					ExpectExec().
					WithArgs(testName, testKind, testPayload, testMetadata, testUserUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "prepare statement error",
			recordName: testName,
			kind:       testKind,
			payload:    testPayload,
			metadata:   testMetadata,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO records \\(name, kind, payload, metadata, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:       "execution error",
			recordName: testName,
			kind:       testKind,
			payload:    testPayload,
			metadata:   testMetadata,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO records \\(name, kind, payload, metadata, user_uuid\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\)").
					ExpectExec().
					WithArgs(testName, testKind, testPayload, testMetadata, testUserUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.AddRecord(context.Background(), tt.recordName, tt.kind, tt.payload, tt.metadata, tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testRecord := &domain.Record{
		UUID:     uuid.New(),
		Name:     "test-record",
		Kind:     "password",
		Payload:  []byte("test-payload"),
		Metadata: []byte("test-metadata"),
		UserUUID: uuid.New(),
	}

	tests := []struct {
		name        string
		record      *domain.Record
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:   "successful update record",
			record: testRecord,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE records SET name = \\$1, kind = \\$2, payload = \\$3, metadata = \\$4 WHERE uuid = \\$5").
					ExpectExec().
					WithArgs(testRecord.Name, testRecord.Kind, testRecord.Payload, testRecord.Metadata, testRecord.UUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:   "prepare statement error",
			record: testRecord,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE records SET name = \\$1, kind = \\$2, payload = \\$3, metadata = \\$4 WHERE uuid = \\$5").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:   "execution error",
			record: testRecord,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE records SET name = \\$1, kind = \\$2, payload = \\$3, metadata = \\$4 WHERE uuid = \\$5").
					ExpectExec().
					WithArgs(testRecord.Name, testRecord.Kind, testRecord.Payload, testRecord.Metadata, testRecord.UUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.UpdateRecord(context.Background(), tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteRecord(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testRecordUUID := uuid.New()

	tests := []struct {
		name        string
		recordUUID  uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:       "successful delete record",
			recordUUID: testRecordUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("DELETE FROM records WHERE uuid = \\$1").
					ExpectExec().
					WithArgs(testRecordUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "prepare statement error",
			recordUUID: testRecordUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("DELETE FROM records WHERE uuid = \\$1").
					WillReturnError(errors.New("prepare error"))
			},
			wantErr: true,
		},
		{
			name:       "execution error",
			recordUUID: testRecordUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("DELETE FROM records WHERE uuid = \\$1").
					ExpectExec().
					WithArgs(testRecordUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			err := storage.DeleteRecord(context.Background(), tt.recordUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetRecordByUserUUIDAndUUID(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	testRecordUUID := uuid.New()

	tests := []struct {
		name        string
		recordUUID  uuid.UUID
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantRecord  *domain.Record
		wantErr     bool
	}{
		{
			name:       "successful get record by UUID",
			recordUUID: testRecordUUID,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1 AND uuid = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, testRecordUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "kind", "payload", "metadata"}).
						AddRow(testRecordUUID, "test-record", "password", []byte("test-payload"), []byte("test-metadata")))
			},
			wantRecord: &domain.Record{
				UUID:     testRecordUUID,
				Name:     "test-record",
				Kind:     "password",
				Payload:  []byte("test-payload"),
				Metadata: []byte("test-metadata"),
			},
			wantErr: false,
		},
		{
			name:       "record not found",
			recordUUID: uuid.New(),
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1 AND uuid = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			wantRecord: nil,
			wantErr:    false,
		},
		{
			name:       "prepare statement error",
			recordUUID: testRecordUUID,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1 AND uuid = \\$2").
					WillReturnError(errors.New("prepare error"))
			},
			wantRecord: nil,
			wantErr:    true,
		},
		{
			name:       "query execution error",
			recordUUID: testRecordUUID,
			userUUID:   testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1 AND uuid = \\$2").
					ExpectQuery().
					WithArgs(testUserUUID, testRecordUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantRecord: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			record, err := storage.GetRecordByUserUUIDAndUUID(context.Background(), tt.recordUUID, tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRecord, record)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAllUserRecords(t *testing.T) {
	db, mock, logger := NewMock()
	defer db.Close()

	storage := NewStorage(db, logger)

	testUserUUID := uuid.New()
	record1UUID := uuid.New()
	record2UUID := uuid.New()

	tests := []struct {
		name        string
		userUUID    uuid.UUID
		mockClosure func(mock sqlmock.Sqlmock)
		wantRecords []*domain.Record
		wantErr     bool
	}{
		{
			name:     "successful get all records",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "kind", "payload", "metadata"}).
						AddRow(record1UUID, "record1", "password", []byte("payload1"), []byte("metadata1")).
						AddRow(record2UUID, "record2", "note", []byte("payload2"), []byte("metadata2")))
			},
			wantRecords: []*domain.Record{
				{
					UUID:     record1UUID,
					Name:     "record1",
					Kind:     "password",
					Payload:  []byte("payload1"),
					Metadata: []byte("metadata1"),
				},
				{
					UUID:     record2UUID,
					Name:     "record2",
					Kind:     "note",
					Payload:  []byte("payload2"),
					Metadata: []byte("metadata2"),
				},
			},
			wantErr: false,
		},
		{
			name:     "no records found",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnRows(sqlmock.NewRows([]string{"uuid", "name", "kind", "payload", "metadata"}))
			},
			wantRecords: []*domain.Record(nil),
			wantErr:     false,
		},
		{
			name:     "prepare statement error",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1").
					WillReturnError(errors.New("prepare error"))
			},
			wantRecords: nil,
			wantErr:     true,
		},
		{
			name:     "query execution error",
			userUUID: testUserUUID,
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = \\$1").
					ExpectQuery().
					WithArgs(testUserUUID).
					WillReturnError(errors.New("execution error"))
			},
			wantRecords: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure(mock)

			records, err := storage.GetAllUserRecords(context.Background(), tt.userUUID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRecords, records)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
