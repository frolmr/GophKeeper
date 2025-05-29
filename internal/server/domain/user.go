package domain

import "github.com/google/uuid"

// User represents an authenticated system account.
//
// Fields:
//
//	UUID           - Unique user identifier
//	Email          - Login email address
//	PasswordHash   - BCrypt hash of user password
//	EmailConfirmed - Whether email address has been verified
//
// Security:
// - PasswordHash should never be logged
// - EmailConfirmed must be checked for sensitive operations
type User struct {
	UUID           uuid.UUID
	Email          string
	PasswordHash   string
	EmailConfirmed bool
}
