// Package input handles user input collection for record creation in GophKeeper.
//
// The package provides:
// - Structured input collection for different record types
// - Standardized input validation
// - Flexible input source (stdin or any io.Reader)
//
// Usage:
//
//	reader := input.NewRecordInputReader(os.Stdin)
//	record, err := reader.ReadInput("MyPassword", "password")
package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/frolmr/GophKeeper/internal/client/domain"
)

// RecordInputReader handles collection of user input for record creation.
// Supports multiple input sources through io.Reader interface.
type RecordInputReader struct {
	inputReader io.Reader
}

// NewRecordInputReader creates a new input reader instance.
//
// Parameters:
//
//	inputReader - io.Reader to use for input (nil defaults to os.Stdin)
func NewRecordInputReader(inputReader io.Reader) *RecordInputReader {
	if inputReader == nil {
		inputReader = os.Stdin
	}
	return &RecordInputReader{
		inputReader: inputReader,
	}
}

// ReadInput collects user input for a new record based on its kind.
//
// Parameters:
//
//	name - Record name/title
//	kind - Record type (must match domain.PayloadKind constants)
//
// Returns:
//
//	*domain.RawRecord - Populated record structure
//	error             - Input or validation errors
//
// Supported kinds:
//   - password: Collects login credentials
//   - card:     Collects payment card details
//   - text:     Collects arbitrary text
//   - bytes:    Collects binary data
func (ir *RecordInputReader) ReadInput(name, kind string) (*domain.RawRecord, error) {
	var payload any
	var err error
	switch kind {
	case domain.PasswordPayloadKind:
		payload, err = ir.readPasswordInput()
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	case domain.CardPayloadKind:
		payload, err = ir.readCardInput()
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	case domain.TextPayloadKind:
		payload, err = ir.readTextInput()
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	case domain.BytesPayloadKind:
		payload, err = ir.readByteInput()
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	default:
		return nil, errors.New("unknown record kind")
	}

	metadata, err := ir.readMetadataInput()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	return &domain.RawRecord{
		Name:     name,
		Kind:     kind,
		Payload:  payload,
		Metadata: metadata,
	}, nil
}

// readPasswordInput collects credentials for a password record.
//
// Prompts for:
// - Login (username/email)
// - Password
//
// Returns:
//
//	*domain.PasswordPayload - Structured credentials
//	error                   - Input errors
func (ir *RecordInputReader) readPasswordInput() (*domain.PasswordPayload, error) {
	reader := bufio.NewReader(ir.inputReader)

	fmt.Print("Enter Login: ")
	login, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read login: %w", err)
	}
	login = strings.TrimSpace(login)

	fmt.Print("Enter Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read password: %w", err)
	}
	password = strings.TrimSpace(password)

	return &domain.PasswordPayload{Login: login, Password: password}, nil
}

// readCardInput collects payment card details.
//
// Prompts for:
// - Card number
// - Cardholder name
// - CVC code
// - Expiration date
//
// Returns:
//
//	*domain.CardPayload - Structured card details
//	error               - Input errors
//
// TODO: Add validation for card number format and expiration date
func (ir *RecordInputReader) readCardInput() (*domain.CardPayload, error) {
	reader := bufio.NewReader(ir.inputReader)

	fmt.Print("Enter Card Number: ")
	number, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read card number: %w", err)
	}
	number = strings.TrimSpace(number)

	fmt.Print("Enter Owner: ")
	owner, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read owner: %w", err)
	}
	owner = strings.TrimSpace(owner)

	fmt.Print("Enter CVC: ")
	cvc, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read cvc: %w", err)
	}
	cvc = strings.TrimSpace(cvc)

	fmt.Print("Enter Expiry Date: ")
	exp, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read expiry date: %w", err)
	}
	exp = strings.TrimSpace(exp)

	payload := domain.CardPayload{
		Number:     number,
		Owner:      owner,
		CVC:        cvc,
		ExpiryDate: exp,
	}
	return &payload, nil
}

// readTextInput collects arbitrary text content.
//
// Returns:
//
//	*domain.TextPayload - Text content
//	error               - Input errors
func (ir *RecordInputReader) readTextInput() (*domain.TextPayload, error) {
	reader := bufio.NewReader(ir.inputReader)

	fmt.Print("Enter Text: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read text: %w", err)
	}
	text = strings.TrimSpace(text)

	return &domain.TextPayload{Text: text}, nil
}

// readByteInput collects binary data input.
// Currently treats input as text bytes.
//
// Returns:
//
//	*domain.BytePayload - Byte content
//	error               - Input errors
func (ir *RecordInputReader) readByteInput() (*domain.BytePayload, error) {
	reader := bufio.NewReader(ir.inputReader)

	fmt.Print("Enter Text: ")
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read text: %w", err)
	}

	return &domain.BytePayload{Bytes: bytes}, nil
}

// readMetadataInput collects optional metadata for records.
//
// Returns:
//
//	[]byte - Raw metadata bytes
//	error  - Input errors
func (ir *RecordInputReader) readMetadataInput() ([]byte, error) {
	reader := bufio.NewReader(ir.inputReader)

	fmt.Print("Enter Metadata: ")
	meta, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("can't read metadata: %w", err)
	}

	return meta, nil
}
