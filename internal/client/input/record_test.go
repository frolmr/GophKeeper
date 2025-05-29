package input

import (
	"errors"
	"strings"
	"testing"

	"github.com/frolmr/GophKeeper/internal/client/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestNewRecordInputReader(t *testing.T) {
	t.Run("with nil reader uses stdin", func(t *testing.T) {
		reader := NewRecordInputReader(nil)
		assert.NotNil(t, reader)
	})

	t.Run("with custom reader", func(t *testing.T) {
		customReader := strings.NewReader("")
		reader := NewRecordInputReader(customReader)
		assert.NotNil(t, reader)
	})
}

func TestReadPasswordInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := "test-user\ntest-password\n"
		expected := &domain.PasswordPayload{
			Login:    "test-user",
			Password: "test-password",
		}

		reader := NewRecordInputReader(strings.NewReader(input))
		result, err := reader.readPasswordInput()

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("login read error", func(t *testing.T) {
		reader := NewRecordInputReader(&errorReader{})
		_, err := reader.readPasswordInput()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read login")
	})

	t.Run("password read error", func(t *testing.T) {
		input := "test-user\n" // Only provide login, not password
		reader := NewRecordInputReader(strings.NewReader(input))
		_, err := reader.readPasswordInput()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read password")
	})
}

func TestReadCardInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := "1234567890\nJohn Doe\n123\n12/25\n"
		expected := &domain.CardPayload{
			Number:     "1234567890",
			Owner:      "John Doe",
			CVC:        "123",
			ExpiryDate: "12/25",
		}

		reader := NewRecordInputReader(strings.NewReader(input))
		result, err := reader.readCardInput()

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("read error", func(t *testing.T) {
		tests := []struct {
			name   string
			input  string
			errMsg string
		}{
			{"number error", "", "can't read card number"},
			{"owner error", "1234\n", "can't read owner"},
			{"cvc error", "1234\nJohn Doe\n", "can't read cvc"},
			{"expiry error", "1234\nJohn Doe\n123\n", "can't read expiry date"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reader := NewRecordInputReader(strings.NewReader(tt.input))
				_, err := reader.readCardInput()
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			})
		}
	})
}

func TestReadTextInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := "sample text content\n"
		expected := &domain.TextPayload{
			Text: "sample text content",
		}

		reader := NewRecordInputReader(strings.NewReader(input))
		result, err := reader.readTextInput()

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("read error", func(t *testing.T) {
		reader := NewRecordInputReader(&errorReader{})
		_, err := reader.readTextInput()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read text")
	})
}

func TestReadByteInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := "binary data\n"
		expected := &domain.BytePayload{
			Bytes: []byte("binary data\n"),
		}

		reader := NewRecordInputReader(strings.NewReader(input))
		result, err := reader.readByteInput()

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("read error", func(t *testing.T) {
		reader := NewRecordInputReader(&errorReader{})
		_, err := reader.readByteInput()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read text")
	})
}

func TestReadMetadataInput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := "metadata content\n"
		expected := []byte("metadata content\n")

		reader := NewRecordInputReader(strings.NewReader(input))
		result, err := reader.readMetadataInput()

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("read error", func(t *testing.T) {
		reader := NewRecordInputReader(&errorReader{})
		_, err := reader.readMetadataInput()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't read metadata")
	})
}

func TestReadInput_ErrorCases(t *testing.T) {
	t.Run("payload read error", func(t *testing.T) {
		reader := NewRecordInputReader(&errorReader{})

		_, err := reader.ReadInput("test", domain.PasswordPayloadKind)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read payload")
	})

	t.Run("metadata read error", func(t *testing.T) {
		// First provide valid payload input, then error for metadata
		input := "test-login\ntest-pwd\n" // Valid password input
		reader := NewRecordInputReader(&partialErrorReader{
			data:   []byte(input),
			errPos: len(input),
		})

		_, err := reader.ReadInput("test", domain.PasswordPayloadKind)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read metadata")
	})
}

// partialErrorReader returns data until errPos, then returns error
type partialErrorReader struct {
	data   []byte
	pos    int
	errPos int
}

func (r *partialErrorReader) Read(p []byte) (n int, err error) {
	if r.pos >= r.errPos {
		return 0, errors.New("simulated read error")
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
