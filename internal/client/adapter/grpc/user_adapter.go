package adapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/frolmr/GophKeeper/pkg/proto/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	requestTimeout = 5 * time.Second
)

type UserGRPCAdapter struct {
	client pb.UsersClient
}

func NewUserGRPCAdapter(clientConn *grpc.ClientConn) *UserGRPCAdapter {
	client := pb.NewUsersClient(clientConn)

	return &UserGRPCAdapter{
		client: client,
	}
}

func (ua *UserGRPCAdapter) SendRegisterRequest(email, password string, mk []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	var header metadata.MD
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs())

	req := &pb.RegisterUserRequest{
		Email:    &email,
		Password: &password,
		Mk:       mk,
	}

	resp, err := ua.client.RegisterUser(ctx, req, grpc.Header((&header)))
	if err != nil {
		return err
	}

	if !*resp.Received {
		if resp.Error != nil {
			return errors.New(*resp.Error)
		}
		return errors.New("server did not acknowledge user registration")
	}

	if authHeaders := header.Get("authorization"); len(authHeaders) > 0 {
		fmt.Println("TOKEN: ", authHeaders)
	} else {
		return errors.New("server did not return authorization token")
	}

	return nil
}

func (ua *UserGRPCAdapter) SendLoginRequest() {}
