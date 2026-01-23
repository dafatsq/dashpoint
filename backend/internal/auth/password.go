package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptCost is the default cost for bcrypt hashing
	DefaultBcryptCost = 12
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPIN hashes a PIN using bcrypt (same as password but conceptually separate)
func HashPIN(pin string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pin), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash PIN: %w", err)
	}
	return string(bytes), nil
}

// CheckPIN compares a PIN with its hash
func CheckPIN(pin, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin))
	return err == nil
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken creates a SHA-256 hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
