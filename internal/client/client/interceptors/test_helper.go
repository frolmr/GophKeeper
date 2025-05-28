package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

type mockInvoker struct {
	lastContext context.Context
	called      bool
	err         error
}

func (m *mockInvoker) invoker(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	opts ...grpc.CallOption) error {
	m.lastContext = ctx
	m.called = true
	return m.err
}
