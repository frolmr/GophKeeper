package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	mocks "github.com/frolmr/GophKeeper/internal/server/mocks/handlers/records"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/records"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecordsService_AddRecord(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	recordName := "test-record"
	recordKind := "password"
	recordPayload := []byte("payload")
	recordMetadata := []byte("metadata")

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockRecordsRepository)
		ctx           context.Context
		input         *pb.AddRecordRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().AddRecord(
					gomock.Any(),
					recordName,
					recordKind,
					recordPayload,
					recordMetadata,
					userUUID,
				).Return(nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.AddRecordRequest{
				Record: &pb.Record{
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: nil,
		},
		{
			name:       "UserNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.Background(),
			input: &pb.AddRecordRequest{
				Record: &pb.Record{
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().AddRecord(
					gomock.Any(),
					recordName,
					recordKind,
					recordPayload,
					recordMetadata,
					userUUID,
				).Return(errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.AddRecordRequest{
				Record: &pb.Record{
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Errorf(codes.Internal, "can't add record: %v", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRecordsRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewRecordsService(repo, logger)
			_, err := service.AddRecord(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordsService_UpdateRecord(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	recordUUID := uuid.New()
	recordUUIDStr := recordUUID.String()
	invalidUUID := "invalid-uuid"
	recordName := "updated-record"
	recordKind := "password"
	recordPayload := []byte("updated-payload")
	recordMetadata := []byte("updated-metadata")

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockRecordsRepository)
		ctx           context.Context
		input         *pb.UpdateRecordRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(&domain.Record{UUID: recordUUID}, nil)
				repo.EXPECT().UpdateRecord(
					gomock.Any(),
					gomock.Any(),
				).Return(nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &recordUUIDStr,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: nil,
		},
		{
			name:       "UserNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.Background(),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &recordUUIDStr,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "InvalidUUID",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &invalidUUID,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Error(codes.InvalidArgument, "can't parse record uuid"),
		},
		{
			name: "RecordNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &recordUUIDStr,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Error(codes.NotFound, "no records to update"),
		},
		{
			name: "DatabaseErrorOnGet",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &recordUUIDStr,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Errorf(codes.Internal, "get record error: %v", errors.New("db error")),
		},
		{
			name: "DatabaseErrorOnUpdate",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(&domain.Record{UUID: recordUUID}, nil)
				repo.EXPECT().UpdateRecord(
					gomock.Any(),
					gomock.Any(),
				).Return(errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.UpdateRecordRequest{
				Record: &pb.Record{
					Uuid:     &recordUUIDStr,
					Name:     &recordName,
					Kind:     &recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				},
			},
			expectedError: status.Errorf(codes.Internal, "can't add record: %v", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRecordsRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewRecordsService(repo, logger)
			_, err := service.UpdateRecord(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordsService_DeleteRecord(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	recordUUID := uuid.New()
	recordUUIDStr := recordUUID.String()
	invalidUUID := "invalid-uuid"

	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockRecordsRepository)
		ctx           context.Context
		input         *pb.DeleteRecordRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(&domain.Record{UUID: recordUUID}, nil)
				repo.EXPECT().DeleteRecord(
					gomock.Any(),
					recordUUID,
				).Return(nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.DeleteRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: nil,
		},
		{
			name:       "UserNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.Background(),
			input: &pb.DeleteRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "InvalidUUID",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.DeleteRecordRequest{
				Uuid: &invalidUUID,
			},
			expectedError: status.Error(codes.InvalidArgument, "can't parse record uuid"),
		},
		{
			name: "RecordNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.DeleteRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Error(codes.NotFound, "no records to delete"),
		},
		{
			name: "DatabaseErrorOnGet",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.DeleteRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Errorf(codes.Internal, "get record error: %v", errors.New("db error")),
		},
		{
			name: "DatabaseErrorOnDelete",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(&domain.Record{UUID: recordUUID}, nil)
				repo.EXPECT().DeleteRecord(
					gomock.Any(),
					recordUUID,
				).Return(errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.DeleteRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Errorf(codes.Internal, "can't add record: %v", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRecordsRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewRecordsService(repo, logger)
			_, err := service.DeleteRecord(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecordsService_GetRecord(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	recordUUID := uuid.New()
	recordUUIDStr := recordUUID.String()
	invalidUUID := "invalid-uuid"
	recordName := "test-record"
	recordKind := "password"
	recordPayload := []byte("payload")
	recordMetadata := []byte("metadata")

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockRecordsRepository)
		ctx            context.Context
		input          *pb.GetRecordRequest
		expectedError  error
		expectedRecord *pb.Record
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(&domain.Record{
					UUID:     recordUUID,
					Name:     recordName,
					Kind:     recordKind,
					Payload:  recordPayload,
					Metadata: recordMetadata,
				}, nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.GetRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedRecord: &pb.Record{
				Uuid:     &recordUUIDStr,
				Name:     &recordName,
				Kind:     &recordKind,
				Payload:  recordPayload,
				Metadata: recordMetadata,
			},
		},
		{
			name:       "UserNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.Background(),
			input: &pb.GetRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "InvalidUUID",
			setupMocks: func(repo *mocks.MockRecordsRepository) {},
			ctx:        context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.GetRecordRequest{
				Uuid: &invalidUUID,
			},
			expectedError: status.Error(codes.InvalidArgument, "can't parse record uuid"),
		},
		{
			name: "RecordNotFound",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, nil)
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.GetRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Error(codes.NotFound, "no records to get"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetRecordByUserUUIDAndUUID(
					gomock.Any(),
					recordUUID,
					userUUID,
				).Return(nil, errors.New("db error"))
			},
			ctx: context.WithValue(context.Background(), contextkeys.UserKey, user),
			input: &pb.GetRecordRequest{
				Uuid: &recordUUIDStr,
			},
			expectedError: status.Errorf(codes.Internal, "get record error: %v", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRecordsRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewRecordsService(repo, logger)
			resp, err := service.GetRecord(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRecord, resp.Record)
			}
		})
	}
}

func TestRecordsService_ListRecords(t *testing.T) {
	userUUID := uuid.New()
	user := &domain.User{UUID: userUUID}
	recordUUID1 := uuid.New()
	recordUUID2 := uuid.New()
	recordName1 := "record-1"
	recordName2 := "record-2"
	recordKind := "password"
	recordPayload := []byte("payload")
	recordMetadata := []byte("metadata")

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockRecordsRepository)
		ctx            context.Context
		expectedError  error
		expectedCount  int
		expectedRecord *pb.Record
	}{
		{
			name: "SuccessWithRecords",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetAllUserRecords(gomock.Any(), userUUID).Return([]*domain.Record{
					{
						UUID:     recordUUID1,
						Name:     recordName1,
						Kind:     recordKind,
						Payload:  recordPayload,
						Metadata: recordMetadata,
					},
					{
						UUID:     recordUUID2,
						Name:     recordName2,
						Kind:     recordKind,
						Payload:  recordPayload,
						Metadata: recordMetadata,
					},
				}, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name:          "UserNotFound",
			setupMocks:    func(repo *mocks.MockRecordsRepository) {},
			ctx:           context.Background(),
			expectedError: status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetAllUserRecords(gomock.Any(), userUUID).Return(nil, errors.New("db error"))
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: status.Errorf(codes.Internal, "can't list records: %v", errors.New("db error")),
		},
		{
			name: "EmptyList",
			setupMocks: func(repo *mocks.MockRecordsRepository) {
				repo.EXPECT().GetAllUserRecords(gomock.Any(), userUUID).Return([]*domain.Record{}, nil)
			},
			ctx:           context.WithValue(context.Background(), contextkeys.UserKey, user),
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockRecordsRepository(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewRecordsService(repo, logger)
			resp, err := service.ListRecords(tt.ctx, &pb.ListRecordsRequest{})

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.GetRecords(), tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, recordUUID1.String(), *resp.Records[0].Uuid)
					assert.Equal(t, recordName1, *resp.Records[0].Name)
					assert.Equal(t, recordKind, *resp.Records[0].Kind)
				}
			}
		})
	}
}
