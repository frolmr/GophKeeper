package domain

// RawRecord represents unencrypted record data before secure storage
type RawRecord struct {
	Name     string // Record name/title
	Kind     string // One of the PayloadKind constants
	Payload  any    // Typed payload (see below)
	Metadata []byte // Optional metadata (non-encrypted)
}

// PasswordPayload stores login credential information
type PasswordPayload struct {
	Login    string `json:"login"`    // Username or email
	Password string `json:"password"` // Password (in plaintext - will be encrypted)
}

// CardPayload stores payment card information
// TODO: replace with proper validation and formatting
type CardPayload struct {
	Number     string `json:"number"` // Card number
	Owner      string `json:"owner"`  // Cardholder name
	CVC        string `json:"cvc"`    // Security code
	ExpiryDate string `json:"exp"`    // Expiration date (MM/YY)
}

// TextPayload stores arbitrary text content
type TextPayload struct {
	Text string `json:"text"` // Content data
}

// BytePayload stores binary data
type BytePayload struct {
	Bytes []byte `json:"bytes"` // Raw binary content
}

// EncryptedRecord represents data securely stored on server
type EncryptedRecord struct {
	UUID     string // Unique record identifier
	Name     string // Record name/title
	Kind     string // Payload type identifier
	Payload  []byte // Encrypted payload data
	Metadata []byte // Optional metadata (non-encrypted)
}
