package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Allow override from test
var RandomReader io.Reader = rand.Reader

func GenerateState() string {
	b := make([]byte, 16)

	if _, err := io.ReadFull(RandomReader, b); err != nil {
		// Only log unexpected errors if needed; otherwise, just return a default value.
		return "default_state"
	}

	return base64.URLEncoding.EncodeToString(b)
}
