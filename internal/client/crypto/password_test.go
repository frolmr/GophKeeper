package crypto

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateStrongPassword(t *testing.T) {
	cs := NewCryptoService()
	const testRuns = 100

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "generates password of correct length",
			testFunc: func(t *testing.T) {
				password, err := cs.GenerateStrongPassword()
				require.NoError(t, err)
				assert.Len(t, password, masterPasswordLength)
			},
		},
		{
			name: "password contains characters from all categories",
			testFunc: func(t *testing.T) {
				for i := 0; i < testRuns; i++ {
					password, err := cs.GenerateStrongPassword()
					require.NoError(t, err)

					var hasLower, hasUpper, hasDigit, hasSpecial bool

					for _, c := range password {
						switch {
						case unicode.IsLower(c):
							hasLower = true
						case unicode.IsUpper(c):
							hasUpper = true
						case unicode.IsDigit(c):
							hasDigit = true
						case strings.ContainsRune(specialChars, c):
							hasSpecial = true
						}
					}

					assert.True(t, hasLower, "password should contain lowercase letters")
					assert.True(t, hasUpper, "password should contain uppercase letters")
					assert.True(t, hasDigit, "password should contain digits")
					assert.True(t, hasSpecial, "password should contain special characters")
				}
			},
		},
		{
			name: "passwords are unique",
			testFunc: func(t *testing.T) {
				seen := make(map[string]bool)
				for i := 0; i < testRuns; i++ {
					password, err := cs.GenerateStrongPassword()
					require.NoError(t, err)
					assert.False(t, seen[password], "duplicate password generated: %v", password)
					seen[password] = true
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestGetRandomChar(t *testing.T) {
	const testRuns = 100

	tests := []struct {
		name        string
		charSet     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "returns character from set",
			charSet: "abc",
		},
		{
			name:        "handles empty character set",
			charSet:     "",
			wantErr:     true,
			errContains: "empty character set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				_, err := getRandomChar(tt.charSet)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			for i := 0; i < testRuns; i++ {
				char, err := getRandomChar(tt.charSet)
				require.NoError(t, err)
				assert.Contains(t, tt.charSet, string(char))
			}
		})
	}
}

func TestGetRandomInt(t *testing.T) {
	const testRuns = 100

	tests := []struct {
		name        string
		min         int
		max         int
		wantErr     bool
		errContains string
	}{
		{
			name: "returns number in range",
			min:  5,
			max:  10,
		},
		{
			name:        "handles invalid range",
			min:         10,
			max:         5,
			wantErr:     true,
			errContains: "invalid range: min >= max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				_, err := getRandomInt(tt.min, tt.max)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			for i := 0; i < testRuns; i++ {
				n, err := getRandomInt(tt.min, tt.max)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, n, tt.min)
				assert.Less(t, n, tt.max)
			}
		})
	}
}
