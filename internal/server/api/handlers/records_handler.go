package handlers

import (
	"context"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/frolmr/GophKeeper/pkg/contextkeys"
	pb "github.com/frolmr/GophKeeper/pkg/proto/records"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RecordsRepository interface {
	AddRecord(ctx context.Context, name, kind string, payload, metadata []byte, userUUID uuid.UUID) error
	UpdateRecord(ctx context.Context, record *domain.Record) error
	DeleteRecord(ctx context.Context, recordUUID uuid.UUID) error
	GetRecordByUserUUIDAndUUID(ctx context.Context, recordUUID, userUUID uuid.UUID) (*domain.Record, error)
	GetAllUserRecords(ctx context.Context, userUUID uuid.UUID) ([]*domain.Record, error)
}

// RecordsService implements the gRPC Records service for:
// - Record addition
// - Record deleteion
// - Record udpate
// - Record read
// - Records list
type RecordsService struct {
	pb.UnimplementedRecordsServer
	repo   RecordsRepository
	logger *zap.SugaredLogger
}

// NewRecordsService creates a new records service handler.
func NewRecordsService(repo RecordsRepository, lgr *zap.SugaredLogger) *RecordsService {
	return &RecordsService{
		repo:   repo,
		logger: lgr,
	}
}

// AddRecord adds a new record for the authenticated user.
func (rs *RecordsService) AddRecord(ctx context.Context, in *pb.AddRecordRequest) (*emptypb.Empty, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	inRec := in.GetRecord()

	if err := rs.repo.AddRecord(ctx, inRec.GetName(), inRec.GetKind(), inRec.GetPayload(), inRec.GetMetadata(), user.UUID); err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't add record: %v", err)
	}

	return nil, nil
}

// UpdateRecord update the existing record for the authenticated user.
func (rs *RecordsService) UpdateRecord(ctx context.Context, in *pb.UpdateRecordRequest) (*emptypb.Empty, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	inRec := in.GetRecord()
	recUUID, err := uuid.Parse(inRec.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "can't parse record uuid")
	}

	existingRecord, err := rs.repo.GetRecordByUserUUIDAndUUID(ctx, recUUID, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "get record error: %v", err)
	}

	if existingRecord == nil {
		return nil, status.Error(codes.NotFound, "no records to update")
	}

	rec := domain.Record{
		UUID:     recUUID,
		Name:     inRec.GetName(),
		Kind:     inRec.GetKind(),
		Payload:  inRec.GetPayload(),
		Metadata: inRec.GetMetadata(),
	}

	if err := rs.repo.UpdateRecord(ctx, &rec); err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't add record: %v", err)
	}

	return nil, nil
}

// DeleteRecord deletes the existing record for the authenticated user.
func (rs *RecordsService) DeleteRecord(ctx context.Context, in *pb.DeleteRecordRequest) (*emptypb.Empty, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	recUUID, err := uuid.Parse(in.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "can't parse record uuid")
	}

	existingRec, err := rs.repo.GetRecordByUserUUIDAndUUID(ctx, recUUID, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "get record error: %v", err)
	}

	if existingRec == nil {
		return nil, status.Error(codes.NotFound, "no records to delete")
	}

	if err := rs.repo.DeleteRecord(ctx, recUUID); err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't add record: %v", err)
	}

	return nil, nil
}

// GetRecord gets the existing record for the authenticated user.
func (rs *RecordsService) GetRecord(ctx context.Context, in *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	reqRecUUID, err := uuid.Parse(in.GetUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "can't parse record uuid")
	}

	existingRec, err := rs.repo.GetRecordByUserUUIDAndUUID(ctx, reqRecUUID, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "get record error: %v", err)
	}

	if existingRec == nil {
		return nil, status.Error(codes.NotFound, "no records to get")
	}

	respRecUUID := existingRec.UUID.String()
	respRec := pb.Record_builder{
		Uuid:     &respRecUUID,
		Name:     &existingRec.Name,
		Kind:     &existingRec.Kind,
		Payload:  existingRec.Payload,
		Metadata: existingRec.Metadata,
	}.Build()

	return pb.GetRecordResponse_builder{Record: respRec}.Build(), nil
}

// ListRecords retrievs all the records for the authenticated user.
func (rs *RecordsService) ListRecords(ctx context.Context, in *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(*domain.User)
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	records, err := rs.repo.GetAllUserRecords(ctx, user.UUID)
	if err != nil {
		// TODO: server errors shouldn't be passed to user, delete after debug
		return nil, status.Errorf(codes.Internal, "can't list records: %v", err)
	}

	var respRecords []*pb.Record
	for _, record := range records {
		recordUUID := record.UUID.String()
		respRec := pb.Record_builder{
			Uuid:     &recordUUID,
			Name:     &record.Name,
			Kind:     &record.Kind,
			Payload:  record.Payload,
			Metadata: record.Metadata,
		}.Build()
		respRecords = append(respRecords, respRec)
	}

	return pb.ListRecordsResponse_builder{Records: respRecords}.Build(), nil
}
