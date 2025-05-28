package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/google/uuid"
)

// AddRecord stores a new encrypted record for a user.
//
// Parameters:
//
//	name     - Record identifier
//	kind     - Data type (password/card/text/bytes)
//	payload  - Encrypted data
//	metadata - Optional plaintext metadata
//	userUUID - Owner reference
//
// Returns:
//
//	error - Database errors
func (s *Storage) AddRecord(ctx context.Context, name, kind string, payload, metadata []byte, userUUID uuid.UUID) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO records (name, kind, payload, metadata, user_uuid) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) record insert, err: %s", userUUID, err.Error())
		return fmt.Errorf("error creating record: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, kind, payload, metadata, userUUID)
	if err != nil {
		s.logger.Errorf("Record query fails for user: %s, err: %s", userUUID, err.Error())
		return fmt.Errorf("error creating record: %w", err)
	}

	return nil
}

// UpdateRecord modifies an existing record.
//
// Parameters:
//
//	record - Updated record data
//
// Returns:
//
//	error - Database errors
func (s *Storage) UpdateRecord(ctx context.Context, record *domain.Record) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE records SET name = $1, kind = $2, payload = $3, metadata = $4 WHERE uuid = $5")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) record insert, err: %s", record.UserUUID, err.Error())
		return fmt.Errorf("error creating record: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(record.Name, record.Kind, record.Payload, record.Metadata, record.UUID)
	if err != nil {
		s.logger.Errorf("Record query fails for user: %s, err: %s", record.UserUUID, err.Error())
		return fmt.Errorf("error creating record: %w", err)
	}

	return nil
}

// DeleteRecord removes a record by its UUID.
func (s *Storage) DeleteRecord(ctx context.Context, recordUUID uuid.UUID) error {
	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM records WHERE uuid = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for record (%s) deletion, err: %s", recordUUID, err.Error())
		return fmt.Errorf("error deleting record: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(recordUUID)
	if err != nil {
		s.logger.Errorf("Record delete query fails %s, err: %s", recordUUID, err.Error())
		return fmt.Errorf("error deleting record: %w", err)
	}

	return nil
}

// GetRecordByUserUUIDAndUUID retrieves a specific record with ownership check.
func (s *Storage) GetRecordByUserUUIDAndUUID(ctx context.Context, recordUUID, userUUID uuid.UUID) (*domain.Record, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = $1 AND uuid = $2")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s to get record, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error getting record: %w", err)
	}
	defer stmt.Close()

	var record domain.Record
	err = stmt.QueryRowContext(ctx, userUUID, recordUUID).Scan(&record.UUID, &record.Name, &record.Kind, &record.Payload, &record.Metadata)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("Record query fails for user: %s, err: %s", userUUID, err.Error())
			return nil, fmt.Errorf("error getting record: %w", err)
		}
	}

	return &record, nil
}

// GetAllUserRecords lists all records belonging to a user.
func (s *Storage) GetAllUserRecords(ctx context.Context, userUUID uuid.UUID) ([]*domain.Record, error) {
	var records []*domain.Record
	stmt, err := s.db.PrepareContext(ctx, "SELECT uuid, name, kind, payload, metadata FROM records WHERE user_uuid = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) records list, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error adding device: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userUUID)
	if err != nil {
		s.logger.Errorf("Can't query records for user: %s, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error getting all users records: %w", err)
	}

	for rows.Next() {
		var record domain.Record
		err := rows.Scan(&record.UUID, &record.Name, &record.Kind, &record.Payload, &record.Metadata)
		if err != nil {
			s.logger.Errorf("Can't scan records to struct for user: %s, err: %s", userUUID, err.Error())
			return nil, fmt.Errorf("error getting all users records: %w", err)
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorf("Got rows.Err() for user: %s, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error getting all users records: %w", err)
	}

	return records, nil
}
