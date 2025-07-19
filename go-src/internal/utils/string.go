package utils

import (
	"crypto/rand"
	"encoding/base64"
	"regexp"
	"strings"
)

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to a simple random string if crypto/rand fails
		return generateFallbackRandomString(length)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// generateFallbackRandomString creates a simple random string as fallback
func generateFallbackRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// Slugify converts a string to a URL-friendly slug
func Slugify(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Replace multiple consecutive hyphens with single hyphen
	reg2 := regexp.MustCompile(`-+`)
	slug = reg2.ReplaceAllString(slug, "-")

	// Ensure slug is not empty
	if slug == "" {
		slug = "untitled"
	}

	return slug
}

// TruncateString truncates a string to the specified length
func TruncateString(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
