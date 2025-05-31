package adapter

import (
	"context"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	pb "github.com/frolmr/GophKeeper/pkg/proto/records"
)

// SendAddRecordRequest stores a new encrypted record on the server.
//
// Parameters:
//
//	encRec - Encrypted record data including metadata
//
// Returns:
//
//	error - Wrapped gRPC error if request fails
//
//nolint:dupl // these are different gRPC requests no way to parametrize
func (a *GRPCAdapter) SendAddRecordRequest(encRec domain.EncryptedRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	rec := pb.Record_builder{
		Uuid:     &encRec.UUID,
		Name:     &encRec.Name,
		Kind:     &encRec.Kind,
		Payload:  encRec.Payload,
		Metadata: encRec.Metadata,
	}.Build()

	req := pb.AddRecordRequest_builder{
		Record: rec,
	}.Build()

	_, err := a.recordClient.AddRecord(ctx, req)
	if err != nil {
		return fmt.Errorf("add record request failed: %w", err)
	}

	return nil
}

// SendUpdateRecordRequest modifies an existing encrypted record.
//
// Parameters:
//
//	encRec - Updated record data (must include UUID)
//
// Returns:
//
//	error - Wrapped gRPC error if request fails
//
//nolint:dupl // these are different gRPC requests no way to parametrize
func (a *GRPCAdapter) SendUpdateRecordRequest(encRec domain.EncryptedRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	rec := pb.Record_builder{
		Uuid:     &encRec.UUID,
		Name:     &encRec.Name,
		Kind:     &encRec.Kind,
		Payload:  encRec.Payload,
		Metadata: encRec.Metadata,
	}.Build()

	req := pb.UpdateRecordRequest_builder{
		Record: rec,
	}.Build()

	_, err := a.recordClient.UpdateRecord(ctx, req)
	if err != nil {
		return fmt.Errorf("update record request failed: %w", err)
	}

	return nil
}

// SendDeleteRecordRequest removes a record from the server.
//
// Parameters:
//
//	recUUID - UUID of the record to delete
//
// Returns:
//
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendDeleteRecordRequest(recUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := pb.DeleteRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	_, err := a.recordClient.DeleteRecord(ctx, req)
	if err != nil {
		return fmt.Errorf("delete record request failed: %w", err)
	}

	return nil
}

// SendGetRecordRequest retrieves a specific encrypted record.
//
// Parameters:
//
//	recUUID - UUID of the record to fetch
//
// Returns:
//
//	*domain.EncryptedRecord - The requested record
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendGetRecordRequest(recUUID string) (*domain.EncryptedRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := pb.GetRecordRequest_builder{
		Uuid: &recUUID,
	}.Build()

	resp, err := a.recordClient.GetRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update record request failed: %w", err)
	}

	encRec := domain.EncryptedRecord{
		UUID:     resp.GetRecord().GetUuid(),
		Name:     resp.GetRecord().GetName(),
		Kind:     resp.GetRecord().GetKind(),
		Payload:  resp.GetRecord().GetPayload(),
		Metadata: resp.GetRecord().GetMetadata(),
	}

	return &encRec, nil
}

// SendListRecordsRequest retrieves all records for the current user.
//
// Returns:
//
//	[]domain.EncryptedRecord - List of all user's records
//	error - Wrapped gRPC error if request fails
func (a *GRPCAdapter) SendListRecordsRequest() ([]domain.EncryptedRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req := pb.ListRecordsRequest_builder{}.Build()

	resp, err := a.recordClient.ListRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update record request failed: %w", err)
	}

	var encRecords []domain.EncryptedRecord
	for _, respRec := range resp.GetRecords() {
		encRec := domain.EncryptedRecord{
			UUID:     respRec.GetUuid(),
			Name:     respRec.GetName(),
			Kind:     respRec.GetKind(),
			Payload:  respRec.GetPayload(),
			Metadata: respRec.GetMetadata(),
		}
		encRecords = append(encRecords, encRec)
	}

	return encRecords, nil
}
