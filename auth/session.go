package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
)

// Allow override from test
var RandomReader io.Reader = rand.Reader

func GenerateState() string {
	b := make([]byte, 16)

	if _, err := io.ReadFull(RandomReader, b); err != nil {
		log.Println("Error generating random state:", err)
		return "default_state"
	}

	return base64.URLEncoding.EncodeToString(b)
}
