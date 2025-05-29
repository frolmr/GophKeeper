package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/frolmr/GophKeeper/internal/server/api/auth"
	"github.com/frolmr/GophKeeper/internal/server/domain"
	mocks "github.com/frolmr/GophKeeper/internal/server/mocks/interceptors/auth"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor_Unary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validToken := "valid_token"
	userUUID := uuid.New()
	validClaims := &auth.Claims{UserUUID: userUUID}
	testUser := &domain.User{UUID: userUUID}

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockRepo := mocks.NewMockUsersRepository(ctrl)

	successHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	tests := []struct {
		name        string
		fullMethod  string
		metadata    metadata.MD
		setupMocks  func()
		expectedRes interface{}
		expectedErr error
	}{
		{
			name:        "Public RegisterUser bypasses auth",
			fullMethod:  "/users.Users/RegisterUser",
			setupMocks:  func() {},
			expectedRes: "success",
		},
		{
			name:        "Public LoginUser bypasses auth",
			fullMethod:  "/users.Users/LoginUser",
			setupMocks:  func() {},
			expectedRes: "success",
		},
		{
			name:        "Missing metadata",
			fullMethod:  "/secure.Method",
			metadata:    nil,
			setupMocks:  func() {},
			expectedErr: status.Error(codes.Unauthenticated, "metadata is not provided"),
		},
		{
			name:        "Missing authorization header",
			fullMethod:  "/secure.Method",
			metadata:    metadata.MD{},
			setupMocks:  func() {},
			expectedErr: status.Error(codes.Unauthenticated, "authorization token is not provided"),
		},
		{
			name:       "Invalid token",
			fullMethod: "/secure.Method",
			metadata:   metadata.MD{"authorization": []string{"Bearer bad_token"}},
			setupMocks: func() {
				mockJWT.EXPECT().
					VerifyAccessToken("bad_token").
					Return(nil, errors.New("invalid token")).
					Times(1)
			},
			expectedErr: status.Errorf(codes.Unauthenticated, "access token is invalid: invalid token"),
		},
		{
			name:       "Valid token but user not found",
			fullMethod: "/secure.Method",
			metadata:   metadata.MD{"authorization": []string{"Bearer " + validToken}},
			setupMocks: func() {
				mockJWT.EXPECT().
					VerifyAccessToken(validToken).
					Return(validClaims, nil).
					Times(1)
				mockRepo.EXPECT().
					GetUserByUUID(gomock.Any(), userUUID).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedErr: status.Error(codes.NotFound, "user not found"),
		},
		{
			name:       "Successful authentication",
			fullMethod: "/secure.Method",
			metadata:   metadata.MD{"authorization": []string{"Bearer " + validToken}},
			setupMocks: func() {
				mockJWT.EXPECT().
					VerifyAccessToken(validToken).
					Return(validClaims, nil).
					Times(1)
				mockRepo.EXPECT().
					GetUserByUUID(gomock.Any(), userUUID).
					Return(testUser, nil).
					Times(1)
			},
			expectedRes: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			interceptor := NewAuthInterceptor(mockJWT, mockRepo)
			unaryInterceptor := interceptor.Unary()

			ctx := context.Background()
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, tt.metadata)
			}

			res, err := unaryInterceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: tt.fullMethod},
				successHandler,
			)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				statusErr, _ := status.FromError(err)
				expectedStatus, _ := status.FromError(tt.expectedErr)
				assert.Equal(t, expectedStatus.Code(), statusErr.Code())
				assert.Contains(t, statusErr.Message(), expectedStatus.Message())
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestAuthInterceptor_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validToken := "valid_token"
	userUUID := uuid.New()
	validClaims := &auth.Claims{UserUUID: userUUID}
	testUser := &domain.User{UUID: userUUID}

	mockJWT := mocks.NewMockJWTManager(ctrl)
	mockRepo := mocks.NewMockUsersRepository(ctrl)

	mockJWT.EXPECT().
		VerifyAccessToken(validToken).
		Return(validClaims, nil).
		Times(1)
	mockRepo.EXPECT().
		GetUserByUUID(gomock.Any(), userUUID).
		Return(testUser, nil).
		Times(1)

	interceptor := NewAuthInterceptor(mockJWT, mockRepo)
	unaryInterceptor := interceptor.Unary()

	md := metadata.MD{"authorization": []string{"Bearer " + validToken}}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
		if !ok || user == nil {
			return nil, status.Error(codes.Internal, "user not in context")
		}
		assert.Equal(t, testUser.UUID, user.UUID)
		return "ok", nil
	}

	res, err := unaryInterceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/secure.Method"},
		testHandler,
	)

	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}
