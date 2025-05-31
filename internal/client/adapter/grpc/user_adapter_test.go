package adapter

import (
	"context"
	"errors"
	"fmt"
	"testing"

	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockUsersClient struct {
	mock.Mock
	setHeaders bool
}

func (m *MockUsersClient) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUsersClient) LoginUser(ctx context.Context, in *pb.LoginUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)

	if m.setHeaders {
		if len(opts) > 0 {
			if header, ok := opts[0].(grpc.HeaderCallOption); ok {
				*header.HeaderAddr = metadata.Pairs("authorization", "test-token")
			}
		}
	}

	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestSendRegisterRequest_Success(t *testing.T) {
	mockClient := new(MockUsersClient)
	adapter := &GRPCAdapter{userClient: mockClient}

	email := "test@example.com"
	password := "password123"
	mk := []byte("masterkey")

	expectedReq := pb.RegisterUserRequest_builder{
		Email:    &email,
		Password: &password,
		Mk:       mk,
	}.Build()

	mockClient.On("RegisterUser", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendRegisterRequest(email, password, mk)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendRegisterRequest_Failure(t *testing.T) {
	mockClient := new(MockUsersClient)
	adapter := &GRPCAdapter{userClient: mockClient}

	email := "test@example.com"
	password := "password123"
	mk := []byte("masterkey")

	expectedReq := pb.RegisterUserRequest_builder{
		Email:    &email,
		Password: &password,
		Mk:       mk,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("RegisterUser", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendRegisterRequest(email, password, mk)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("register request failed: %w", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendLoginRequest_Success(t *testing.T) {
	mockClient := new(MockUsersClient)
	mockClient.setHeaders = true // Enable header setting for this test
	adapter := &GRPCAdapter{userClient: mockClient}

	email := "test@example.com"
	password := "password123"

	expectedReq := pb.LoginUserRequest_builder{
		Email:    &email,
		Password: &password,
	}.Build()

	mockClient.On("LoginUser", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	token, err := adapter.SendLoginRequest(email, password)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
	mockClient.AssertExpectations(t)
}

func TestSendLoginRequest_Failure(t *testing.T) {
	t.Run("gRPC error", func(t *testing.T) {
		mockClient := new(MockUsersClient)
		adapter := &GRPCAdapter{userClient: mockClient}

		email := "test@example.com"
		password := "password123"

		expectedReq := pb.LoginUserRequest_builder{
			Email:    &email,
			Password: &password,
		}.Build()

		expectedErr := errors.New("connection error")
		mockClient.On("LoginUser", mock.Anything, expectedReq, mock.Anything).
			Return(&emptypb.Empty{}, expectedErr)

		token, err := adapter.SendLoginRequest(email, password)
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("login request failed: %w", expectedErr).Error(), err.Error())
		assert.Empty(t, token)
	})

	t.Run("missing auth token", func(t *testing.T) {
		mockClient := new(MockUsersClient)
		mockClient.setHeaders = false // Explicitly disable header setting
		adapter := &GRPCAdapter{userClient: mockClient}

		email := "test@example.com"
		password := "password123"

		expectedReq := pb.LoginUserRequest_builder{
			Email:    &email,
			Password: &password,
		}.Build()

		mockClient.On("LoginUser", mock.Anything, expectedReq, mock.Anything).
			Return(&emptypb.Empty{}, nil)

		token, err := adapter.SendLoginRequest(email, password)
		assert.Error(t, err)
		assert.Equal(t, "server did not return authorization token", err.Error())
		assert.Empty(t, token)
	})
}
