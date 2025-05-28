package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserAndDevice atomically creates a new user with their first confirmed device.
//
// Parameters:
//
//	email      - User login email
//	password   - Plaintext password (will be hashed)
//	deviceName - Initial device name
//	deviceID   - Initial device identifier
//	masterKey  - Encrypted master key
//
// Returns:
//
//	error - Database or hashing errors
func (s *Storage) CreateUserAndDevice(ctx context.Context, email, password, deviceName, deviceID string, masterKey []byte) error {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Errorf("Transaction for user (%s) and device create error, err: %s", email, err.Error())
		return fmt.Errorf("error creating user transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Errorf("failed to rollback transaction: %w", rbErr)
			}
		}
	}()

	user, err := s.createAndReturnUser(ctx, tx, email, password)
	if err != nil {
		return err
	}

	if err := s.createConfirmedDevice(ctx, tx, deviceName, deviceID, masterKey, user); err != nil {
		return err
	}

	return tx.Commit()
}

// GetUserByEmail retrieves a user by email address.
func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := "SELECT uuid, email, password_hash, email_confirmed FROM users WHERE email = $1"
	return s.getUserBy(ctx, query, email)
}

// GetUserByUUID retrieves a user by their UUID.
func (s *Storage) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*domain.User, error) {
	query := "SELECT uuid, email, password_hash, email_confirmed FROM users WHERE uuid = $1"
	return s.getUserBy(ctx, query, userUUID)
}

func (s *Storage) getUserBy(ctx context.Context, query string, by any) (*domain.User, error) {
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", by, err.Error())
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	defer stmt.Close()

	var user domain.User
	err = stmt.QueryRowContext(ctx, by).Scan(&user.UUID, &user.Email, &user.PasswordHash, &user.EmailConfirmed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", by, err.Error())
			return nil, fmt.Errorf("error getting user: %w", err)
		}
	}

	return &user, nil
}

func (s *Storage) createAndReturnUser(ctx context.Context, tx *sql.Tx, email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", email, err.Error())
		return nil, fmt.Errorf("error user password encryption: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING uuid, email")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user insert: %s, err: %s", email, err.Error())
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	defer stmt.Close()

	var user domain.User
	err = stmt.QueryRow(email, hashedPassword).Scan(&user.UUID, &user.Email)
	if err != nil {
		s.logger.Errorf("User query fails for user: %s, err: %s", email, err.Error())
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &user, nil
}
