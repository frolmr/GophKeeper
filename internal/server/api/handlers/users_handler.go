package handlers

import (
	"context"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UsersRepository interface {
	CreateUserAndDevice(ctx context.Context, email, password, deviceName, deviceID string, maskterKey []byte) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetDeviceByUserUUIDAndID(ctx context.Context, name string, userUUID uuid.UUID) (*domain.Device, error)
}

type JWTManager interface {
	GenerateAccessToken(userUUID uuid.UUID) (string, error)
}

// UsersService implements the gRPC Users service for:
// - User registration
// - User login
type UsersService struct {
	pb.UnimplementedUsersServer
	repo       UsersRepository
	jwtManager JWTManager
	logger     *zap.SugaredLogger
}

// NewDevicesService creates a new devices service handler.
func NewUsersService(repo UsersRepository, lgr *zap.SugaredLogger, jwtManager JWTManager) *UsersService {
	return &UsersService{
		repo:       repo,
		logger:     lgr,
		jwtManager: jwtManager,
	}
}

// RegisterUser registers a new user.
func (us *UsersService) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*emptypb.Empty, error) {
	email := *in.Email
	password := *in.Password

	if email == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	existingUser, err := us.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "user already registered")
	}

	deviceName, ok := ctx.Value(contextkeys.DeviceNameKey).(string)
	if !ok {
		return nil, status.Error(codes.NotFound, "device data doesn't provided")
	}

	deviceID, ok := ctx.Value(contextkeys.DeviceIDKey).(string)
	if !ok {
		return nil, status.Error(codes.NotFound, "device data doesn't provided")
	}

	err = us.repo.CreateUserAndDevice(ctx, email, password, deviceName, deviceID, in.GetMk())
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "Can't create user: %v", err)
	}
	return nil, nil
}

// LoginUser logsin the existing user.
func (us *UsersService) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*emptypb.Empty, error) {
	email := *in.Email
	password := *in.Password

	if email == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	existingUser, err := us.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	if existingUser == nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	if !existingUser.EmailConfirmed {
		return nil, status.Error(codes.Unauthenticated, "User email is unconfirmed! Please confirm email first!")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	token, err := us.jwtManager.GenerateAccessToken(existingUser.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't generate auth token: %v", err)
	}
	header := metadata.Pairs("authorization", "Bearer "+token)
	if err := grpc.SetHeader(ctx, header); err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't set auth header: %v", err)
	}

	return nil, nil
}
