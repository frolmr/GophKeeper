package adapter

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	pb "github.com/frolmr/GophKeeper/pkg/proto/records"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockRecordsClient struct {
	mock.Mock
}

func (m *MockRecordsClient) AddRecord(ctx context.Context, in *pb.AddRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockRecordsClient) UpdateRecord(ctx context.Context, in *pb.UpdateRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockRecordsClient) DeleteRecord(ctx context.Context, in *pb.DeleteRecordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockRecordsClient) GetRecord(ctx context.Context, in *pb.GetRecordRequest, opts ...grpc.CallOption) (*pb.GetRecordResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetRecordResponse), args.Error(1)
}

func (m *MockRecordsClient) ListRecords(ctx context.Context, in *pb.ListRecordsRequest, opts ...grpc.CallOption) (*pb.ListRecordsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ListRecordsResponse), args.Error(1)
}

func TestSendAddRecordRequest_Success(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	uuid := "test-uuid"
	name := "test-name"
	kind := "test-kind"
	payload := []byte("test-payload")
	metadata := []byte("test-metadata")

	encRec := domain.EncryptedRecord{
		UUID:     uuid,
		Name:     name,
		Kind:     kind,
		Payload:  payload,
		Metadata: metadata,
	}

	expectedReq := pb.AddRecordRequest_builder{
		Record: pb.Record_builder{
			Uuid:     &uuid,
			Name:     &name,
			Kind:     &kind,
			Payload:  payload,
			Metadata: metadata,
		}.Build(),
	}.Build()

	mockClient.On("AddRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendAddRecordRequest(encRec)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendAddRecordRequest_Failure(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	uuid := "test-uuid"
	name := "test-name"
	kind := "test-kind"
	payload := []byte("test-payload")
	metadata := []byte("test-metadata")

	encRec := domain.EncryptedRecord{
		UUID:     uuid,
		Name:     name,
		Kind:     kind,
		Payload:  payload,
		Metadata: metadata,
	}

	expectedReq := pb.AddRecordRequest_builder{
		Record: pb.Record_builder{
			Uuid:     &uuid,
			Name:     &name,
			Kind:     &kind,
			Payload:  payload,
			Metadata: metadata,
		}.Build(),
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("AddRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendAddRecordRequest(encRec)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("add record request failed: %w", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendUpdateRecordRequest_Success(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	uuid := "test-uuid"
	name := "test-name"
	kind := "test-kind"
	payload := []byte("test-payload")
	metadata := []byte("test-metadata")

	encRec := domain.EncryptedRecord{
		UUID:     uuid,
		Name:     name,
		Kind:     kind,
		Payload:  payload,
		Metadata: metadata,
	}

	expectedReq := pb.UpdateRecordRequest_builder{
		Record: pb.Record_builder{
			Uuid:     &uuid,
			Name:     &name,
			Kind:     &kind,
			Payload:  payload,
			Metadata: metadata,
		}.Build(),
	}.Build()

	mockClient.On("UpdateRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendUpdateRecordRequest(encRec)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendUpdateRecordRequest_Failure(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	uuid := "test-uuid"
	name := "test-name"
	kind := "test-kind"
	payload := []byte("test-payload")
	metadata := []byte("test-metadata")

	encRec := domain.EncryptedRecord{
		UUID:     uuid,
		Name:     name,
		Kind:     kind,
		Payload:  payload,
		Metadata: metadata,
	}

	expectedReq := pb.UpdateRecordRequest_builder{
		Record: pb.Record_builder{
			Uuid:     &uuid,
			Name:     &name,
			Kind:     &kind,
			Payload:  payload,
			Metadata: metadata,
		}.Build(),
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("UpdateRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendUpdateRecordRequest(encRec)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("update record request failed: %w", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendDeleteRecordRequest_Success(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	recUUID := "test-uuid"

	expectedReq := pb.DeleteRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	mockClient.On("DeleteRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, nil)

	err := adapter.SendDeleteRecordRequest(recUUID)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSendDeleteRecordRequest_Failure(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	recUUID := "test-uuid"

	expectedReq := pb.DeleteRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("DeleteRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&emptypb.Empty{}, expectedErr)

	err := adapter.SendDeleteRecordRequest(recUUID)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("delete record request failed: %w", expectedErr).Error(), err.Error())
	mockClient.AssertExpectations(t)
}

func TestSendGetRecordRequest_Success(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	recUUID := "test-uuid"
	name := "test-name"
	kind := "test-kind"
	payload := []byte("test-payload")
	metadata := []byte("test-metadata")

	expectedReq := pb.GetRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	expectedResp := pb.GetRecordResponse_builder{
		Record: pb.Record_builder{
			Uuid:     &recUUID,
			Name:     &name,
			Kind:     &kind,
			Payload:  payload,
			Metadata: metadata,
		}.Build(),
	}.Build()

	mockClient.On("GetRecord", mock.Anything, expectedReq, mock.Anything).
		Return(expectedResp, nil)

	result, err := adapter.SendGetRecordRequest(recUUID)
	assert.NoError(t, err)
	assert.Equal(t, recUUID, result.UUID)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, kind, result.Kind)
	assert.Equal(t, payload, result.Payload)
	assert.Equal(t, metadata, result.Metadata)
	mockClient.AssertExpectations(t)
}

func TestSendGetRecordRequest_Failure(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	recUUID := "test-uuid"

	expectedReq := pb.GetRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	expectedErr := errors.New("connection error")
	mockClient.On("GetRecord", mock.Anything, expectedReq, mock.Anything).
		Return(&pb.GetRecordResponse{}, expectedErr)

	result, err := adapter.SendGetRecordRequest(recUUID)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("update record request failed: %w", expectedErr).Error(), err.Error())
	assert.Nil(t, result)
	mockClient.AssertExpectations(t)
}

func TestSendListRecordsRequest_Success(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	uuid1 := "uuid1"
	name1 := "name1"
	kind1 := "kind1"
	payload1 := []byte("payload1")
	metadata1 := []byte("metadata1")

	uuid2 := "uuid2"
	name2 := "name2"
	kind2 := "kind2"
	payload2 := []byte("payload2")
	metadata2 := []byte("metadata2")

	expectedResp := pb.ListRecordsResponse_builder{
		Records: []*pb.Record{
			pb.Record_builder{
				Uuid:     &uuid1,
				Name:     &name1,
				Kind:     &kind1,
				Payload:  payload1,
				Metadata: metadata1,
			}.Build(),
			pb.Record_builder{
				Uuid:     &uuid2,
				Name:     &name2,
				Kind:     &kind2,
				Payload:  payload2,
				Metadata: metadata2,
			}.Build(),
		},
	}.Build()

	mockClient.On("ListRecords", mock.Anything, &pb.ListRecordsRequest{}, mock.Anything).
		Return(expectedResp, nil)

	results, err := adapter.SendListRecordsRequest()
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, uuid1, results[0].UUID)
	assert.Equal(t, name1, results[0].Name)
	assert.Equal(t, kind1, results[0].Kind)
	assert.Equal(t, payload1, results[0].Payload)
	assert.Equal(t, metadata1, results[0].Metadata)

	assert.Equal(t, uuid2, results[1].UUID)
	assert.Equal(t, name2, results[1].Name)
	assert.Equal(t, kind2, results[1].Kind)
	assert.Equal(t, payload2, results[1].Payload)
	assert.Equal(t, metadata2, results[1].Metadata)

	mockClient.AssertExpectations(t)
}

func TestSendListRecordsRequest_Failure(t *testing.T) {
	mockClient := new(MockRecordsClient)
	adapter := &GRPCAdapter{recordClient: mockClient}

	expectedErr := errors.New("connection error")
	mockClient.On("ListRecords", mock.Anything, &pb.ListRecordsRequest{}, mock.Anything).
		Return(&pb.ListRecordsResponse{}, expectedErr)

	results, err := adapter.SendListRecordsRequest()
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("update record request failed: %w", expectedErr).Error(), err.Error())
	assert.Nil(t, results)
	mockClient.AssertExpectations(t)
}
