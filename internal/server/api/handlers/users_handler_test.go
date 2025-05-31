package handlers

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	mocks "github.com/frolmr/GophKeeper/internal/server/mocks/handlers/users"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type mockStream struct {
	ctx context.Context
}

func (m *mockStream) SetHeader(md metadata.MD) error {
	m.ctx = metadata.NewOutgoingContext(m.ctx, md)
	return nil
}

func (m *mockStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockStream) Method() string                  { return "method" }
func (m *mockStream) SetTrailer(md metadata.MD) error { return nil }
func (m *mockStream) Context() context.Context        { return m.ctx }
func (m *mockStream) SendMsg(v interface{}) error     { return nil }
func (m *mockStream) RecvMsg(v interface{}) error     { return nil }

func createTestContext() context.Context {
	p := &peer.Peer{
		Addr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 12345,
		},
	}
	ctx := peer.NewContext(context.Background(), p)

	stream := &mockStream{ctx: ctx}

	return grpc.NewContextWithServerTransportStream(ctx, stream)
}

func TestUsersService_RegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockUsersRepository)
		ctx           context.Context
		input         *pb.RegisterUserRequest
		expectedError error
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, nil)
				repo.EXPECT().CreateUserAndDevice(
					gomock.Any(),
					"test@example.com",
					"password",
					"test-device",
					"device-id",
					[]byte("master-key"),
				).Return(nil)
			},
			ctx: context.WithValue(
				context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
				contextkeys.DeviceIDKey, "device-id",
			),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("password"),
				Mk:       []byte("master-key"),
			}.Build(),
			expectedError: nil,
		},
		{
			name:       "EmptyEmailOrPassword",
			setupMocks: func(repo *mocks.MockUsersRepository) {},
			ctx:        context.Background(),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr(""),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Error(codes.Unauthenticated, "invalid login or password"),
		},
		{
			name: "UserAlreadyExists",
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "existing@example.com").Return(&domain.User{}, nil)
			},
			ctx: context.Background(),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr("existing@example.com"),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Error(codes.AlreadyExists, "user already registered"),
		},
		{
			name: "MissingDeviceData",
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, nil)
			},
			ctx: context.Background(),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Error(codes.NotFound, "device data doesn't provided"),
		},
		{
			name: "DatabaseErrorOnGetUser",
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("db error"))
			},
			ctx: context.Background(),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Errorf(codes.Internal, "database error: %v", errors.New("db error")),
		},
		{
			name: "DatabaseErrorOnCreateUser",
			setupMocks: func(repo *mocks.MockUsersRepository) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, nil)
				repo.EXPECT().CreateUserAndDevice(
					gomock.Any(),
					"test@example.com",
					"password",
					"test-device",
					"device-id",
					[]byte("master-key"),
				).Return(errors.New("create error"))
			},
			ctx: context.WithValue(
				context.WithValue(context.Background(), contextkeys.DeviceNameKey, "test-device"),
				contextkeys.DeviceIDKey, "device-id",
			),
			input: pb.RegisterUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("password"),
				Mk:       []byte("master-key"),
			}.Build(),
			expectedError: status.Errorf(codes.Internal, "Can't create user: %v", errors.New("create error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUsersRepository(ctrl)
			jwt := mocks.NewMockJWTManager(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			service := NewUsersService(repo, logger, jwt)
			_, err := service.RegisterUser(tt.ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUsersService_LoginUser(t *testing.T) {
	userUUID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		setupMocks     func(*mocks.MockUsersRepository, *mocks.MockJWTManager)
		input          *pb.LoginUserRequest
		expectedError  error
		expectTokenSet bool
	}{
		{
			name: "Success",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
					UUID:           userUUID,
					PasswordHash:   string(hashedPassword),
					EmailConfirmed: true,
				}, nil)
				jwt.EXPECT().GenerateAccessToken(userUUID).Return("test-token", nil)
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("correct-password"),
			}.Build(),
			expectTokenSet: true,
		},
		{
			name: "TokenGenerationError",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
					UUID:           userUUID,
					PasswordHash:   string(hashedPassword),
					EmailConfirmed: true,
				}, nil)
				jwt.EXPECT().GenerateAccessToken(userUUID).Return("", errors.New("token error"))
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("correct-password"),
			}.Build(),
			expectedError: status.Errorf(codes.Internal, "can't generate auth token: %v", errors.New("token error")),
		},
		{
			name: "EmptyEmailOrPassword",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				// No expectations as we fail before repository calls
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr(""),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Error(codes.Unauthenticated, "invalid login or password"),
		},
		{
			name: "UserNotFound",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "nonexistent@example.com").Return(nil, nil)
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("nonexistent@example.com"),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Error(codes.NotFound, "User not found"),
		},
		{
			name: "DatabaseError",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(nil, errors.New("db error"))
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("password"),
			}.Build(),
			expectedError: status.Errorf(codes.Internal, "database error: %v", errors.New("db error")),
		},
		{
			name: "EmailNotConfirmed",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "unconfirmed@example.com").Return(&domain.User{
					UUID:           userUUID,
					PasswordHash:   string(hashedPassword),
					EmailConfirmed: false,
				}, nil)
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("unconfirmed@example.com"),
				Password: ptr("correct-password"),
			}.Build(),
			expectedError: status.Error(codes.Unauthenticated, "User email is unconfirmed! Please confirm email first!"),
		},
		{
			name: "InvalidPassword",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
					UUID:           userUUID,
					PasswordHash:   string(hashedPassword),
					EmailConfirmed: true,
				}, nil)
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("wrong-password"),
			}.Build(),
			expectedError: status.Error(codes.Unauthenticated, "invalid login or password"),
		},
		{
			name: "TokenGenerationError",
			setupMocks: func(repo *mocks.MockUsersRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
					UUID:           userUUID,
					PasswordHash:   string(hashedPassword),
					EmailConfirmed: true,
				}, nil)
				jwt.EXPECT().GenerateAccessToken(userUUID).Return("", errors.New("token error"))
			},
			input: pb.LoginUserRequest_builder{
				Email:    ptr("test@example.com"),
				Password: ptr("correct-password"),
			}.Build(),
			expectedError: status.Errorf(codes.Internal, "can't generate auth token: %v", errors.New("token error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUsersRepository(ctrl)
			jwt := mocks.NewMockJWTManager(ctrl)
			logger := zap.NewNop().Sugar()

			if tt.setupMocks != nil {
				tt.setupMocks(repo, jwt)
			}

			service := NewUsersService(repo, logger, jwt)
			ctx := createTestContext()
			_, err := service.LoginUser(ctx, tt.input)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				if tt.expectTokenSet {
					// Verify the token was set in headers
					stream := grpc.ServerTransportStreamFromContext(ctx).(*mockStream)
					md, ok := metadata.FromOutgoingContext(stream.ctx)
					assert.True(t, ok)
					assert.Equal(t, []string{"Bearer test-token"}, md.Get("authorization"))
				}
			}
		})
	}
}

func ptr(s string) *string {
	return &s
}
