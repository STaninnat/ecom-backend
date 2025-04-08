package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GenerateState() string {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		log.Println("Error generating random state:", err)
		return "default_state"
	}

	return base64.URLEncoding.EncodeToString(b)
}
