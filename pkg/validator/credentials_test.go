package validator

import "testing"

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "test@example.com", true},
		{"Valid with subdomain", "user.name@mail.example.com", true},
		{"Valid with +", "user+tag@example.com", true},
		{"Valid with numbers", "user123@example.com", true},
		{"Valid with hyphen", "user-name@example.com", true},
		{"Valid with multiple subdomains", "user@mail.sub.example.com", true},
		{"Too long", "a@" + string(make([]byte, 253)), false},
		{"Exceeds max length", "a@" + string(make([]byte, 254)), false},
		{"Missing @", "invalid-email.com", false},
		{"Missing local part", "@example.com", false},
		{"Missing domain", "user@", false},
		{"Invalid chars in domain", "user@exa mple.com", false},
		{"Consecutive dots", "user@example..com", false},
		{"Leading dot", ".user@example.com", false},
		{"Trailing dot", "user.@example.com", false},
		{"Invalid special char", "user@exa<mple.com", false},
		{"Empty string", "", false},
		{"Only whitespace", "   ", false},
		{"Invalid TLD", "user@example.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.expected)
			}
		})
	}
}

func TestIsPasswordStrong(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"Strong password", "P@ssw0rd", true},
		{"Too short", "Short1!", false},
		{"Missing uppercase", "password1!", false},
		{"Missing lowercase", "PASSWORD1!", false},
		{"Missing number", "Password!", false},
		{"Missing special", "Password1", false},
		{"Valid complex", "V3ry$tr0ngP@ss", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPasswordStrong(tt.password); got != tt.expected {
				t.Errorf("IsPasswordStrong(%q) = %v, want %v", tt.password, got, tt.expected)
			}
		})
	}
}
