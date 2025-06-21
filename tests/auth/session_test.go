package auth_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/stretchr/testify/assert"
)

// Mocking rand.Reader to simulate an error scenario in random state generation
type mockReaderError struct{}

// Implementing the Read method for mockRandError to return an error when reading
func (m *mockReaderError) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("mock error") // Simulating error when trying to read from rand.Reader
}

type mockReaderFixed struct{}

func (m *mockReaderFixed) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(i % 256)
	}
	return len(p), nil
}

func TestGenerateState(t *testing.T) {
	t.Run("returns 24-character base64 string", func(t *testing.T) {
		original := auth.RandomReader
		defer func() { auth.RandomReader = original }()

		auth.RandomReader = &mockReaderFixed{}
		state := auth.GenerateState()

		assert.Len(t, state, 24, "base64 string length should be 24")
		_, err := base64.URLEncoding.DecodeString(state)
		assert.NoError(t, err, "should be valid base64")
	})

	t.Run("returns different values each call", func(t *testing.T) {
		state1 := auth.GenerateState()
		state2 := auth.GenerateState()
		assert.NotEqual(t, state1, state2, "each call should generate unique state")
	})

	t.Run("returns default_state on error", func(t *testing.T) {
		original := auth.RandomReader
		defer func() { auth.RandomReader = original }()

		auth.RandomReader = &mockReaderError{}
		state := auth.GenerateState()

		assert.Equal(t, "default_state", state, "fallback state should be returned on error")
	})
}
