package auth_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/stretchr/testify/assert"
)

// Mocking rand.Reader to simulate an error scenario in random state generation
type mockRandError struct{}

// Implementing the Read method for mockRandError to return an error when reading
func (m *mockRandError) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock error") // Simulating error when trying to read from rand.Reader
}

func TestGenerateState(t *testing.T) {
	tests := []struct {
		name           string
		overrideRand   func() io.Reader
		expectedState  string
		expectError    bool
		expectedLength int
	}{
		{
			name: "Valid State Generation",
			overrideRand: func() io.Reader {
				return rand.Reader
			},
			expectedState:  "valid_state",
			expectError:    false,
			expectedLength: 24,
		},
		{
			name: "Default State On Rand Error", // Scenario where rand.Reader fails and returns a default state
			overrideRand: func() io.Reader {
				return &mockRandError{}
			},
			expectedState:  "default_state",
			expectError:    true,
			expectedLength: 13,
		},
		{
			name: "State Length",
			overrideRand: func() io.Reader {
				return rand.Reader
			},
			expectedState:  "valid_state",
			expectError:    false,
			expectedLength: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup the original rand.Reader and replace it with the mock or real rand.Reader
			originalRand := rand.Reader
			rand.Reader = tt.overrideRand()
			defer func() { rand.Reader = originalRand }() // Restore the original rand.Reader after test

			state := auth.GenerateState()

			if tt.expectError {
				assert.Equal(t, tt.expectedState, state, "expected state %v but got %v", tt.expectedState, state)
			} else {
				assert.Len(t, state, tt.expectedLength, "expected length of state to be %d but got %d", tt.expectedLength, len(state))
			}
		})
	}
}
