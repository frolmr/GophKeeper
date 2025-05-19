package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/server/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *Storage) CreateUser(email, password string, masterKey []byte) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Password encryption failed for user: %s, err: %s", email, err.Error())
		return fmt.Errorf("error creating user: %w", err)
	}

	query := "INSERT INTO users (email, password_hash, master_key) VALUES ($1, $2, $3)"
	if _, err := s.db.Exec(query, email, string(hashedPassword), masterKey); err != nil {
		s.logger.Errorf("New user insertion failed, user: %s, err: %s", email, err.Error())
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (s *Storage) GetUserByEmail(email string) (*domain.User, error) {
	stmt, err := s.db.Prepare("SELECT id, email, password_hash FROM users WHERE email = $1")
	if err != nil {
		s.logger.Errorf("Can't prepare statement for user: %s, err: %s", email, err.Error())
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	defer stmt.Close()

	var user domain.User
	err = stmt.QueryRow(email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			s.logger.Errorf("User query fails for user: %s, err: %s", email, err.Error())
			return nil, fmt.Errorf("error getting user: %w", err)
		}
	}

	return &user, nil
}
