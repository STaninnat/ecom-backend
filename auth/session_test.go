package auth

import (
	"errors"
	"strings"
	"testing"
)

type failReader struct{}

func (f *failReader) Read(p []byte) (int, error) {
	return 0, errors.New("fail")
}

func TestGenerateState_Success(t *testing.T) {
	old := RandomReader
	defer func() { RandomReader = old }()
	RandomReader = strings.NewReader("abcdefghijklmnopABCDEFGHIJKLMNOP")

	state := GenerateState()
	if state == "" || state == "default_state" {
		t.Errorf("expected non-empty, non-default state, got %q", state)
	}
}

func TestGenerateState_Fail(t *testing.T) {
	old := RandomReader
	defer func() { RandomReader = old }()
	RandomReader = &failReader{}

	state := GenerateState()
	if state != "default_state" {
		t.Errorf("expected default_state, got %q", state)
	}
}
