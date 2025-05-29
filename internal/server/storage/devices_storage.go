package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/google/uuid"
)

// GetDeviceByUserUUIDAndID retrieves a device by its client-generated ID and owner.
//
// Parameters:
//
//	ctx      - Context for cancellation/timeout
//	id       - Client device identifier
//	userUUID - Owner's UUID
//
// Returns:
//
//	*domain.Device - Found device (nil if not found)
//	error          - Database or scanning errors
func (s *Storage) GetDeviceByUserUUIDAndID(ctx context.Context, id string, userUUID uuid.UUID) (*domain.Device, error) {
	query := "SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = $1 AND id = $2"
	return s.getDeviceByUserUUIDAnd(ctx, query, userUUID, id)
}

// GetDeviceByUserUUIDAndUUID retrieves a device by its server-generated UUID and owner.
//
// Parameters:
//
//	ctx        - Context for cancellation/timeout
//	deviceUUID - Client device UUID
//	userUUID   - Owner's UUID
//
// Returns:
//
//	*domain.Device - Found device (nil if not found)
//	error          - Database or scanning errors
func (s *Storage) GetDeviceByUserUUIDAndUUID(ctx context.Context, deviceUUID, userUUID uuid.UUID) (*domain.Device, error) {
	query := "SELECT uuid, name, id, master_key, public_key, confirmed FROM devices WHERE user_uuid = $1 AND uuid = $2"
	return s.getDeviceByUserUUIDAnd(ctx, query, userUUID, deviceUUID)
}

func (s *Storage) getDeviceByUserUUIDAnd(ctx context.Context, query string, userUUID uuid.UUID, by any) (*domain.Device, error) {
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		s.logger.Errorf("Can't prepare statement for device: %s, err: %s", by, err.Error())
		return nil, fmt.Errorf("error getting device: %w", err)
	}
	defer stmt.Close()

	var device domain.Device
	err = stmt.QueryRowContext(ctx, userUUID, by).Scan(&device.UUID, &device.Name, &device.ID, &device.MK, &device.PK, &device.Confirmed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("Device query fails for user: %s, err: %s", userUUID, err.Error())
			return nil, fmt.Errorf("error getting user: %w", err)
		}
	}

	return &device, nil
}

// GetAllUserDevices lists all devices registered to a user.
//
// Returns:
//
//	[]*domain.Device - List of devices (empty slice if none)
//	error            - Database errors
func (s *Storage) GetAllUserDevices(ctx context.Context, userUUID uuid.UUID) ([]*domain.Device, error) {
	var devices []*domain.Device
	stmt, err := s.db.PrepareContext(ctx, "SELECT uuid, name, confirmed FROM devices WHERE user_uuid = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) device list, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error adding device: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userUUID)
	if err != nil {
		s.logger.Errorf("Can't query devices for user: %s, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error getting all users devices: %w", err)
	}

	for rows.Next() {
		var device domain.Device
		err := rows.Scan(&device.UUID, &device.Name, &device.Confirmed)
		if err != nil {
			s.logger.Errorf("Can't scan devices to struct for user: %s, err: %s", userUUID, err.Error())
			return nil, fmt.Errorf("error getting all users devices: %w", err)
		}
		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		s.logger.Errorf("Got rows.Err() for user: %s, err: %s", userUUID, err.Error())
		return nil, fmt.Errorf("error getting all users devices: %w", err)
	}

	return devices, nil
}

// AddDevice registers a new unconfirmed device for a user.
//
// Parameters:
//
//	name      - Human-readable device name
//	id        - Client-generated device ID
//	publicKey - Device public key in PEM format
//	user      - Owner reference
//
// Returns:
//
//	error - Database errors
func (s *Storage) AddDevice(ctx context.Context, name, id string, publicKey []byte, user *domain.User) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO devices (name, id, public_key, user_uuid) VALUES ($1, $2, $3, $4)")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) device insert, err: %s", user.Email, err.Error())
		return fmt.Errorf("error adding device: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, id, publicKey, user.UUID)
	if err != nil {
		s.logger.Errorf("Device query fails for user: %s, err: %s", user.Email, err.Error())
		return fmt.Errorf("error adding device: %w", err)
	}

	return nil
}

// ConfirmDevice approves a device and stores its master key.
//
// Parameters:
//
//	deviceUUID - Device to confirm
//	userUUID   - Verifying owner
//	masterKey  - Encrypted master key
//
// Returns:
//
//	error - Database errors
func (s *Storage) ConfirmDevice(ctx context.Context, deviceUUID, userUUID uuid.UUID, masterKey []byte) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE devices SET confirmed = TRUE, master_key = $1 WHERE uuid = $2 AND user_uuid = $3")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for device (%s) update, err: %s", deviceUUID, err.Error())
		return fmt.Errorf("error updating device: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(masterKey, deviceUUID, userUUID)
	if err != nil {
		s.logger.Errorf("Device update query fails for device: %s, err: %s", deviceUUID, err.Error())
		return fmt.Errorf("error updating device: %w", err)
	}

	return nil
}

func (s *Storage) createConfirmedDevice(ctx context.Context, tx *sql.Tx, name, id string, masterKey []byte, user *domain.User) error {
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO devices (name, id, master_key, user_uuid, confirmed) VALUES ($1, $2, $3, $4, true)")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user (%s) device insert, err: %s", user.Email, err.Error())
		return fmt.Errorf("error creating device: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, id, masterKey, user.UUID)
	if err != nil {
		s.logger.Errorf("Device query fails for user: %s, err: %s", user.Email, err.Error())
		return fmt.Errorf("error creating device: %w", err)
	}

	return nil
}
