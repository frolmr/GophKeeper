package domain

import "github.com/google/uuid"

// Device represents a registered client device that can access the system.
//
// Fields:
//
//	UUID      - Unique identifier for the device (assigned by server)
//	UserUUID  - Owner of the device
//	Name      - Human-readable device name (e.g., "John's MacBook")
//	ID        - Unique client-generated device identifier
//	MK        - Encrypted master key for this device (nil if unconfirmed)
//	PK        - Device public key in PEM format
//	Confirmed - Whether the device has been approved by user
//
// Security:
// - MK contains sensitive encryption material
// - PK should be verified before use
type Device struct {
	UUID      uuid.UUID
	UserUUID  uuid.UUID
	Name      string
	ID        string
	MK        []byte
	PK        []byte
	Confirmed bool
}
