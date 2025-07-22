package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the given password using bcrypt and returns the resulting hash string.
// Returns an error if the password is too short or hashing fails.
func HashPassword(password string) (string, error) {
	if len(password) <= 8 {
		return "", errors.New("password is too short, it must have at least 8 characters")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// log.Printf("Couldn't hash password")
		return "", err
	}

	return string(bytes), nil
}

// CheckPasswordHash compares a plaintext password with its bcrypt hash and returns an error if they do not match.
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("password mismatch")
		}

		return fmt.Errorf("error comparing password: %w", err)
	}

	return nil
}
