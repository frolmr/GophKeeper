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

	rec := pb.Record{
		Uuid:     &encRec.UUID,
		Name:     &encRec.Name,
		Kind:     &encRec.Kind,
		Payload:  encRec.Payload,
		Metadata: encRec.Metadata,
	}

	req := &pb.AddRecordRequest{
		Record: &rec,
	}

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

	rec := pb.Record{
		Uuid:     &encRec.UUID,
		Name:     &encRec.Name,
		Kind:     &encRec.Kind,
		Payload:  encRec.Payload,
		Metadata: encRec.Metadata,
	}

	req := &pb.UpdateRecordRequest{
		Record: &rec,
	}

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

	req := &pb.DeleteRecordRequest{
		Uuid: &recUUID,
	}

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

	req := &pb.GetRecordRequest{
		Uuid: &recUUID,
	}

	resp, err := a.recordClient.GetRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update record request failed: %w", err)
	}

	encRec := domain.EncryptedRecord{
		UUID:     *resp.Record.Uuid,
		Name:     *resp.Record.Name,
		Kind:     *resp.Record.Kind,
		Payload:  resp.Record.Payload,
		Metadata: resp.Record.Metadata,
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

	req := &pb.ListRecordsRequest{}

	resp, err := a.recordClient.ListRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update record request failed: %w", err)
	}

	var encRecords []domain.EncryptedRecord
	for _, respRec := range resp.Records {
		encRec := domain.EncryptedRecord{
			UUID:     *respRec.Uuid,
			Name:     *respRec.Name,
			Kind:     *respRec.Kind,
			Payload:  respRec.Payload,
			Metadata: respRec.Metadata,
		}
		encRecords = append(encRecords, encRec)
	}

	return encRecords, nil
}
