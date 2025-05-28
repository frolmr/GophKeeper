package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

const (
	emailMaxLength    = 254
	passwordMinLength = 8
)

// IsValidEmail checks if the provided string is a valid email address.
// It verifies:
// - Length doesn't exceed 254 characters
// - Matches basic email format pattern
// - Can be parsed by Go's mail.ParseAddress
// Returns true if the email is valid, false otherwise.
func IsValidEmail(email string) bool {
	if len(email) > emailMaxLength || email == "" {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart, domain := parts[0], parts[1]

	localPartRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+$`)
	if !localPartRegex.MatchString(localPart) {
		return false
	}

	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return false // Must have at least domain and TLD
	}

	for _, part := range domainParts {
		domainRegex := regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
		if !domainRegex.MatchString(part) {
			return false
		}
	}

	tld := domainParts[len(domainParts)-1]
	tldRegex := regexp.MustCompile(`^[a-zA-Z]{2,}$`)
	if !tldRegex.MatchString(tld) {
		return false
	}

	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsPasswordStrong checks if the password meets strength requirements.
// It verifies:
// - Minimum length of 8 characters
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one number
// - Contains at least one special character
// Returns true if the password meets all requirements, false otherwise.
func IsPasswordStrong(password string) bool {
	if len(password) < passwordMinLength {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)

	hasNumber := regexp.MustCompile(`\d`).MatchString(password)

	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
