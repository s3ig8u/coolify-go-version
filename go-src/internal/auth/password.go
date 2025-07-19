package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Password utilities for secure password management

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	// Use cost 12 for a good balance of security and performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength checks if a password meets minimum requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	// Add more validation as needed
	// - Check for uppercase, lowercase, numbers, special characters
	// - Check against common password lists
	// - Check for patterns

	return nil
}

// Common password validation errors
var (
	ErrPasswordTooShort = &PasswordError{Message: "Password must be at least 8 characters long"}
)

// PasswordError represents a password validation error
type PasswordError struct {
	Message string
}

func (e *PasswordError) Error() string {
	return e.Message
}
