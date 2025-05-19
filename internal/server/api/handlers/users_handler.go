package handlers

import (
	"context"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersRepository interface {
	CreateUser(email, password string, maskterKey []byte) error
	GetUserByEmail(email string) (*domain.User, error)
}

type UsersService struct {
	pb.UnimplementedUsersServer
	repo   UsersRepository
	logger *zap.SugaredLogger
}

func NewUserService(repo UsersRepository, lgr *zap.SugaredLogger) *UsersService {
	return &UsersService{
		repo:   repo,
		logger: lgr,
	}
}

func (us *UsersService) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*pb.Ack, error) {
	received := true
	errorMsg := ""

	email := *in.Email
	password := *in.Password

	if email == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	existingUser, err := us.repo.GetUserByEmail(email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "user already registered")
	}

	err = us.repo.CreateUser(email, password, in.GetMk())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't create user: %v", err)
	}
	return &pb.Ack{Received: &received, Error: &errorMsg}, nil
}

func (us *UsersService) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.Ack, error) {
	received := true
	errorMsg := ""

	email := *in.Email
	password := *in.Password

	if email == "" || password == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	existingUser, err := us.repo.GetUserByEmail(email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}

	return &pb.Ack{Received: &received, Error: &errorMsg}, nil
}

func (us *UsersService) AddDevice(ctx context.Context, in *pb.AddDeviceRequest) (*pb.Ack, error) {
	return nil, nil
}
