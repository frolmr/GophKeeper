// Package domain defines the core data structures and constants
// for the GophKeeper client application.
//
// The package contains:
// - Application-wide constants
// - Device identity structures
// - Record data models (both raw and encrypted)
// - Payload type definitions
//
// Security Note:
// Contains sensitive data structures - handle with care
package domain

// Application constants
const (
	// AppName identifies the application for device ID generation
	AppName = "GophKeeper"

	// PrivKeyFileName defines the default filename for storing private keys
	PrivKeyFileName = "pk"

	// Payload kind constants
	PasswordPayloadKind = "password" // For login credentials
	CardPayloadKind     = "card"     // For payment card information
	TextPayloadKind     = "text"     // For arbitrary text data
	BytesPayloadKind    = "bytes"    // For binary data
)
