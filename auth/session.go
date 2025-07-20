package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// RandomReader is the source of randomness for session state generation. It can be overridden for testing.
var RandomReader io.Reader = rand.Reader

// GenerateState generates a random URL-safe state string for session management or OAuth flows.
// Returns a default value if random generation fails.
func GenerateState() string {
	b := make([]byte, 16)

	if _, err := io.ReadFull(RandomReader, b); err != nil {
		// Only log unexpected errors if needed; otherwise, just return a default value.
		return "default_state"
	}

	return base64.URLEncoding.EncodeToString(b)
}
