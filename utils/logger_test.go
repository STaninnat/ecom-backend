// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// logger_test.go: Tests for logrus-based logger setup, hooks, and error handling.

// mockWriter is a mock implementation of io.Writer for testing error scenarios in WriterHook.
type mockWriter struct {
	buf        *bytes.Buffer
	errOnWrite bool
}

// Write writes to the buffer or returns an error if errOnWrite is true.
func (m *mockWriter) Write(p []byte) (int, error) {
	if m.errOnWrite {
		return 0, errors.New("write error")
	}
	return m.buf.Write(p)
}

// TestNewWriterHookAndLevels tests NewWriterHook and its Levels method for correct level mapping.
func TestNewWriterHookAndLevels(t *testing.T) {
	levels := []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel}
	hook := NewWriterHook(io.Discard, levels)
	got := hook.Levels()
	if len(got) != len(levels) {
		t.Errorf("expected %d levels, got %d", len(levels), len(got))
	}
	for _, lvl := range levels {
		found := false
		for _, g := range got {
			if g == lvl {
				found = true
			}
		}
		if !found {
			t.Errorf("level %v not found in hook.Levels()", lvl)
		}
	}
}

// TestWriterHookFire tests the Fire method of WriterHook for:
// - Writing log entries for matching levels
// - Skipping non-matching levels
// - Handling entry.String() errors
// - Handling writer errors
func TestWriterHookFire(t *testing.T) {
	buf := &bytes.Buffer{}
	hook := NewWriterHook(buf, []logrus.Level{logrus.InfoLevel})
	entry := &logrus.Entry{Logger: logrus.New(), Level: logrus.InfoLevel, Message: "test message"}
	err := hook.Fire(entry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected log message in buffer, got %q", buf.String())
	}

	// Should not write for levels not in LogLevels
	buf.Reset()
	entry.Level = logrus.ErrorLevel
	err = hook.Fire(entry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected buffer to be empty for non-matching level")
	}

	// Simulate error on entry.String()
	badEntry := &logrus.Entry{Logger: logrus.New(), Level: logrus.InfoLevel}
	badEntry.Logger.SetFormatter(&badFormatter{})
	err = hook.Fire(badEntry)
	if err == nil {
		t.Errorf("expected error from bad formatter, got nil")
	}

	// Simulate error on Write
	mw := &mockWriter{buf: &bytes.Buffer{}, errOnWrite: true}
	hook = NewWriterHook(mw, []logrus.Level{logrus.InfoLevel})
	entry = &logrus.Entry{Logger: logrus.New(), Level: logrus.InfoLevel, Message: "test"}
	err = hook.Fire(entry)
	if err == nil {
		t.Errorf("expected error from writer, got nil")
	}
}

// badFormatter is a logrus formatter that always returns an error, used to test error handling in WriterHook.
type badFormatter struct{}

// Format always returns an error for testing.
func (b *badFormatter) Format(*logrus.Entry) ([]byte, error) {
	return nil, errors.New("format error")
}

// TestInitLoggerBasic tests InitLogger for basic logger setup and configuration in dev mode.
func TestInitLoggerBasic(t *testing.T) {
	if err := os.Setenv("APP_MODE", "dev"); err != nil {
		t.Fatalf("os.Setenv failed: %v", err)
	}
	logger := InitLogger()
	if logger == nil {
		t.Fatal("InitLogger returned nil")
	}
	if logger.Level != logrus.DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.Level)
	}
	// We can't easily test file outputs or hooks without more advanced patching/mocking
}

// TestInitLoggerWithCreators_PanicOnInfoWriterError tests that InitLoggerWithCreators panics if info writer creation fails.
func TestInitLoggerWithCreators_PanicOnInfoWriterError(t *testing.T) {
	mockErr := errors.New("info rotator fail")
	panicMsg := "failed to create info log rotator: info rotator fail"
	//nolint:unparam // required to match InitLoggerWithCreators signature
	infoFail := func(string, ...rotatelogs.Option) (*rotatelogs.RotateLogs, error) {
		return nil, mockErr
	}
	errorOK := func(string, ...rotatelogs.Option) (*rotatelogs.RotateLogs, error) {
		return &rotatelogs.RotateLogs{}, nil
	}
	require.PanicsWithValue(t, panicMsg, func() {
		InitLoggerWithCreators(infoFail, errorOK)
	})
}

// TestInitLoggerWithCreators_PanicOnErrorWriterError tests that InitLoggerWithCreators panics if error writer creation fails.
func TestInitLoggerWithCreators_PanicOnErrorWriterError(t *testing.T) {
	mockErr := errors.New("error rotator fail")
	panicMsg := "failed to create error log rotator: error rotator fail"
	infoOK := func(string, ...rotatelogs.Option) (*rotatelogs.RotateLogs, error) {
		return &rotatelogs.RotateLogs{}, nil
	}
	//nolint:unparam // required to match InitLoggerWithCreators signature
	errorFail := func(string, ...rotatelogs.Option) (*rotatelogs.RotateLogs, error) {
		return nil, mockErr
	}
	require.PanicsWithValue(t, panicMsg, func() {
		InitLoggerWithCreators(infoOK, errorFail)
	})
}
