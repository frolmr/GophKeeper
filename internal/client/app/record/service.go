package record

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/frolmr/GophKeeper/internal/client/domain"
)

// Connector defines record-related network operations.
type Connector interface {
	SendAddRecordRequest(encRec domain.EncryptedRecord) error
	SendGetDeviceMKRequest() ([]byte, error)
	SendGetRecordRequest(recUUID string) (*domain.EncryptedRecord, error)
	SendDeleteRecordRequest(recUUID string) error
	SendListRecordsRequest() ([]domain.EncryptedRecord, error)
	SendUpdateRecordRequest(encRec domain.EncryptedRecord) error
}

// Repository defines persistence operations for record-related data.
type Repository interface {
	ReadPrivateKey() ([]byte, error)
}

// CryptoProcessor defines cryptographic operations for record handling.
type CryptoProcessor interface {
	BytesToPrivateKey([]byte) (*rsa.PrivateKey, error)
	DecryptWithPrivateKey([]byte, *rsa.PrivateKey) ([]byte, error)
	EncryptWithMasterKey(plaintext []byte, mk []byte) ([]byte, error)
	DecryptWithMasterKey(ciphertext []byte, mk []byte) ([]byte, error)
}

// RecordService manages secure record operations including:
// - Encryption/decryption
// - Synchronization with server
//
// Security Note:
// All record payloads are encrypted with master key before transmission.
type RecordService struct {
	repository    Repository
	cryptoService CryptoProcessor
	connector     Connector
}

// NewRecordService creates a new record service instance.
//
// Parameters:
//
//	repo - Storage repository for device keys
//	cp   - Cryptographic processor
//	rc   - Network connector
func NewRecordService(repo Repository, cp CryptoProcessor, rc Connector) *RecordService {
	return &RecordService{
		repository:    repo,
		cryptoService: cp,
		connector:     rc,
	}
}

// AddRecord encrypts and stores a new record on the server.
//
// Parameters:
//
//	rec - Record to add (will be encrypted)
//
// Returns:
//
//	error - Encryption or network errors
func (s *RecordService) AddRecord(rec *domain.RawRecord) error {
	privKeyBytes, err := s.repository.ReadPrivateKey()
	if err != nil {
		return fmt.Errorf("error reading PK: %w", err)
	}

	privKey, err := s.cryptoService.BytesToPrivateKey(privKeyBytes)
	if err != nil {
		return fmt.Errorf("error private key convertation: %w", err)
	}

	mk, err := s.connector.SendGetDeviceMKRequest()
	if err != nil {
		return fmt.Errorf("error getting users MK: %w", err)
	}

	decryptedMK, err := s.cryptoService.DecryptWithPrivateKey(mk, privKey)
	if err != nil {
		return fmt.Errorf("MK decryption failed: %w", err)
	}

	payload, err := json.Marshal(rec.Payload)
	if err != nil {
		return fmt.Errorf("error payload convertation: %w", err)
	}

	encPayload, err := s.cryptoService.EncryptWithMasterKey(payload, decryptedMK)
	if err != nil {
		return fmt.Errorf("failed to encrypt payload: %w", err)
	}

	encRec := domain.EncryptedRecord{
		Name:     rec.Name,
		Kind:     rec.Kind,
		Payload:  encPayload,
		Metadata: rec.Metadata,
	}

	return s.connector.SendAddRecordRequest(encRec)
}

// DeleteRecord deletes record on the server.
//
// Parameters:
//
//	recUUID - Record uuid to delete
//
// Returns:
//
//	error - Encryption or network errors
func (s *RecordService) DeleteRecord(recUUID string) error {
	if err := s.connector.SendDeleteRecordRequest(recUUID); err != nil {
		return fmt.Errorf("delete record request failed: %w", err)
	}

	return nil
}

// ListRecords retrieves all records registered for the current user.
//
// Returns:
//
//	[]domain.EncryptedRecord - List of registered encrypted records
//	error - Network errors
func (s *RecordService) ListRecords() ([]domain.EncryptedRecord, error) {
	records, err := s.connector.SendListRecordsRequest()
	if err != nil {
		return nil, fmt.Errorf("list devices request failed: %w", err)
	}

	return records, nil
}

// ReadRecord retrieves records registered for the current user and decrypts it.
//
// Returns:
//
//	domain.RawRecord - Decrypted record
//	error - Network errors
func (s *RecordService) ReadRecord(recID string) (*domain.RawRecord, error) {
	privKeyBytes, err := s.repository.ReadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error reading PK: %w", err)
	}

	privKey, err := s.cryptoService.BytesToPrivateKey(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("error private key convertation: %w", err)
	}

	mk, err := s.connector.SendGetDeviceMKRequest()
	if err != nil {
		return nil, fmt.Errorf("error getting users MK: %w", err)
	}

	decryptedMK, err := s.cryptoService.DecryptWithPrivateKey(mk, privKey)
	if err != nil {
		return nil, fmt.Errorf("MK decryption failed: %w", err)
	}

	encRec, err := s.connector.SendGetRecordRequest(recID)
	if err != nil {
		return nil, fmt.Errorf("error getting record: %w", err)
	}

	decPayload, err := s.cryptoService.DecryptWithMasterKey(encRec.Payload, decryptedMK)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var payload any
	switch encRec.Kind {
	case domain.PasswordPayloadKind:
		payload = domain.PasswordPayload{}
	case domain.CardPayloadKind:
		payload = domain.CardPayload{}
	case domain.TextPayloadKind:
		payload = domain.TextPayload{}
	case domain.BytesPayloadKind:
		payload = domain.BytePayload{}
	default:
		return nil, errors.New("unknown payload kind")
	}

	if err := json.Unmarshal(decPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	rawRec := domain.RawRecord{
		Name:     encRec.Name,
		Kind:     encRec.Kind,
		Payload:  payload,
		Metadata: encRec.Metadata,
	}

	return &rawRec, nil
}

// UpdateRecord encrypts and stores a updated record on the server.
//
// Parameters:
//
//	rec - Record to update (will be encrypted)
//
// Returns:
//
//	error - Encryption or network errors
func (s *RecordService) UpdateRecord(rec *domain.RawRecord, id string) error {
	privKeyBytes, err := s.repository.ReadPrivateKey()
	if err != nil {
		return fmt.Errorf("error reading PK: %w", err)
	}

	privKey, err := s.cryptoService.BytesToPrivateKey(privKeyBytes)
	if err != nil {
		return fmt.Errorf("error private key convertation: %w", err)
	}

	mk, err := s.connector.SendGetDeviceMKRequest()
	if err != nil {
		return fmt.Errorf("error getting users MK: %w", err)
	}

	decryptedMK, err := s.cryptoService.DecryptWithPrivateKey(mk, privKey)
	if err != nil {
		return fmt.Errorf("MK decryption failed: %w", err)
	}

	payload, err := json.Marshal(rec.Payload)
	if err != nil {
		return fmt.Errorf("error payload convertation: %w", err)
	}

	encPayload, err := s.cryptoService.EncryptWithMasterKey(payload, decryptedMK)
	if err != nil {
		return fmt.Errorf("failed to encrypt payload: %w", err)
	}

	encRec := domain.EncryptedRecord{
		UUID:     id,
		Name:     rec.Name,
		Kind:     rec.Kind,
		Payload:  encPayload,
		Metadata: rec.Metadata,
	}

	return s.connector.SendUpdateRecordRequest(encRec)
}
