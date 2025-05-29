package domain

import "github.com/google/uuid"

// Record represents an encrypted data item stored by a user.
//
// Fields:
//
//	UUID     - Unique record identifier
//	UserUUID - Owner of the record
//	Name     - Human-readable record name
//	Kind     - Data type (password/card/text/bytes)
//	Payload  - Encrypted record data
//	Metadata - Optional unencrypted metadata
//
// Security:
// - Payload is encrypted with the user's master key
// - Metadata is stored in plaintext
type Record struct {
	UUID     uuid.UUID
	UserUUID uuid.UUID
	Name     string
	Kind     string
	Payload  []byte
	Metadata []byte
}
