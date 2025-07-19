package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomString generates a random string of given length
func RandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// TODO: Add helpers for secure token generation, etc.
