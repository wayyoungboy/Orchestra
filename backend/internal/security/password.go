package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrHashFailed       = errors.New("failed to hash password")
)

const bcryptCost = 12

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) < 6 {
		return "", ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", ErrHashFailed
	}

	return string(hash), nil
}

// VerifyPassword checks if the password matches the hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}