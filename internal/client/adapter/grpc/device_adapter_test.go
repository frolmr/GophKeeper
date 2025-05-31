package adapter

import (
	"context"
	"errors"
	"fmt"
	"testing"

	pb "github.com/frolmr/GophKeeper/pkg/proto/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockDevicesClient struct {
	mock.Mock
}

func (m *MockDevicesClient) ListDevices(ctx context.Context, in *pb.ListDevicesRequest, opts ...grpc.CallOption) (*pb.ListDevicesResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListDevicesResponse), args.Error(1)
}

func (m *MockDevicesClient) GetDevicePK(ctx context.Context, in *pb.GetDevicePKRequest, opts ...grpc.CallOption) (*pb.GetDevicePKResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetDevicePKResponse), args.Error(1)
}

func (m *MockDevicesClient) GetDeviceMK(ctx context.Context, in *pb.GetDeviceMKRequest, opts ...grpc.CallOption) (*pb.GetDeviceMKResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetDeviceMKResponse), args.Error(1)
}

func (m *MockDevicesClient) AddDevice(ctx context.Context, in *pb.AddDeviceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockDevicesClient) ApproveDevice(ctx context.Context, in *pb.ApproveDeviceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func TestSendAddDeviceRequest_Success(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	pk := []byte("public-key")

	expectedReq := pb.AddDeviceRequest_builder{
		Pk: pk,
	}.Build()

	mockClient.On("AddDevice", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendAddDeviceRequest(pk)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendAddDeviceRequest_AlreadyExists(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	pk := []byte("public-key")

	expectedReq := pb.AddDeviceRequest_builder{
		Pk: pk,
	}.Build()

	expectedErr := status.Error(codes.AlreadyExists, "device already exists")
	mockClient.On("AddDevice", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendAddDeviceRequest(pk)
	assert.Error(t, err)
	assert.Equal(t, "device already registered", err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendAddDeviceRequest_Failure(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	pk := []byte("public-key")

	expectedReq := pb.AddDeviceRequest_builder{
		Pk: pk,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("AddDevice", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendAddDeviceRequest(pk)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("add device request failed: %w", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendListDevicesRequest_Success(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	uuid1 := "uuid1"
	name1 := "device1"
	conf1 := true
	uuid2 := "uuid2"
	name2 := "device2"
	conf2 := false

	expectedDevices := []*pb.Device{
		pb.Device_builder{
			Uuid:      &uuid1,
			Name:      &name1,
			Confirmed: &conf1,
		}.Build(),
		pb.Device_builder{
			Uuid:      &uuid2,
			Name:      &name2,
			Confirmed: &conf2,
		}.Build(),
	}

	expectedResp := pb.ListDevicesResponse_builder{
		Devices: expectedDevices,
	}.Build()

	mockClient.On("ListDevices", mock.Anything, &pb.ListDevicesRequest{}, mock.Anything).
		Return(expectedResp, nil)

	devices, err := adapter.SendListDevicesRequest()
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, uuid1, devices[0].ID)
	assert.Equal(t, name1, devices[0].Name)
	assert.True(t, devices[0].Confirmed)
	assert.Equal(t, uuid2, devices[1].ID)
	assert.Equal(t, name2, devices[1].Name)
	assert.False(t, devices[1].Confirmed)
	mockClient.AssertExpectations(t)
}

func TestSendListDevicesRequest_Failure(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	expectedErr := errors.New("connection error")
	mockClient.On("ListDevices", mock.Anything, &pb.ListDevicesRequest{}, mock.Anything).
		Return(&pb.ListDevicesResponse{}, expectedErr)

	devices, err := adapter.SendListDevicesRequest()
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("list devices response error: %v", expectedErr).Error(), err.Error())
	assert.Nil(t, devices)
	mockClient.AssertExpectations(t)
}

func TestSendGetDevicePKRequest_Success(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	deviceID := "uuid1"
	pk := []byte("public-key")

	expectedReq := pb.GetDevicePKRequest_builder{
		Uuid: &deviceID,
	}.Build()

	expectedResp := pb.GetDevicePKResponse_builder{
		Pk: pk,
	}.Build()

	mockClient.On("GetDevicePK", mock.Anything, expectedReq, mock.Anything).
		Return(expectedResp, nil)

	result, err := adapter.SendGetDevicePKRequest(deviceID)
	assert.NoError(t, err)
	assert.Equal(t, pk, result)
	mockClient.AssertExpectations(t)
}

func TestSendGetDevicePKRequest_Failure(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	deviceID := "uuid1"

	expectedReq := pb.GetDevicePKRequest_builder{
		Uuid: &deviceID,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("GetDevicePK", mock.Anything, expectedReq, mock.Anything).
		Return(&pb.GetDevicePKResponse{}, expectedErr)

	result, err := adapter.SendGetDevicePKRequest(deviceID)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("get devices PK response error: %v", expectedErr).Error(), err.Error())
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}

func TestSendApproveDeviceRequest_Success(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	deviceID := "uuid1"
	mk := []byte("master-key")

	expectedReq := pb.ApproveDeviceRequest_builder{
		Uuid: &deviceID,
		Mk:   mk,
	}.Build()

	mockClient.On("ApproveDevice", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendApproveDeviceRequest(deviceID, mk)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendApproveDeviceRequest_Failure(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	deviceID := "uuid1"
	mk := []byte("master-key")

	expectedReq := pb.ApproveDeviceRequest_builder{
		Uuid: &deviceID,
		Mk:   mk,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("ApproveDevice", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendApproveDeviceRequest(deviceID, mk)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("approve devices response error: %v", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendGetDeviceMKRequest_Success(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	mk := []byte("master-key")

	expectedResp := pb.GetDeviceMKResponse_builder{
		Mk: mk,
	}.Build()

	mockClient.On("GetDeviceMK", mock.Anything, &pb.GetDeviceMKRequest{}, mock.Anything).
		Return(expectedResp, nil)

	result, err := adapter.SendGetDeviceMKRequest()
	assert.NoError(t, err)
	assert.Equal(t, mk, result)
	mockClient.AssertExpectations(t)
}

func TestSendGetDeviceMKRequest_Failure(t *testing.T) {
	mockClient := new(MockDevicesClient)
	adapter := &GRPCAdapter{deviceClient: mockClient}

	expectedErr := errors.New("connection error")
	mockClient.On("GetDeviceMK", mock.Anything, &pb.GetDeviceMKRequest{}, mock.Anything).
		Return(&pb.GetDeviceMKResponse{}, expectedErr)

	result, err := adapter.SendGetDeviceMKRequest()
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("login request failed: %w", expectedErr).Error(), err.Error())
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}
