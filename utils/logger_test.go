package utils

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

type mockWriter struct {
	buf        *bytes.Buffer
	errOnWrite bool
}

func (m *mockWriter) Write(p []byte) (int, error) {
	if m.errOnWrite {
		return 0, errors.New("write error")
	}
	return m.buf.Write(p)
}

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

type badFormatter struct{}

func (b *badFormatter) Format(*logrus.Entry) ([]byte, error) {
	return nil, errors.New("format error")
}

func TestInitLoggerBasic(t *testing.T) {
	// Patch environment to force dev mode
	os.Setenv("APP_MODE", "dev")
	logger := InitLogger()
	if logger == nil {
		t.Fatal("InitLogger returned nil")
	}
	if logger.Level != logrus.DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.Level)
	}
	// We can't easily test file outputs or hooks without more advanced patching/mocking
}
