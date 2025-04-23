package utils_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type brokenWriter struct{}

func (b brokenWriter) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

type mockWriter struct {
	buf bytes.Buffer
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}

func TestWriterHookFireWritesLog(t *testing.T) {
	writer := &mockWriter{}
	hook := &utils.WriterHook{
		Writer:    writer,
		LogLevels: []logrus.Level{logrus.InfoLevel},
	}

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel
	entry.Message = "test message"

	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(writer.buf.String(), "test message") {
		t.Errorf("log output does not contain expected message")
	}
}

func TestWriterHookFireHandlesError(t *testing.T) {
	hook := &utils.WriterHook{
		Writer: brokenWriter{},
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
		},
	}

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel
	entry.Message = "should fail"

	err := hook.Fire(entry)
	if err == nil {
		t.Errorf("expected error but got nil")
	}
}

func TestWriterHookLevelsMethod(t *testing.T) {
	hook := &utils.WriterHook{
		LogLevels: []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel},
	}

	levels := hook.Levels()
	if len(levels) != 2 || levels[0] != logrus.InfoLevel || levels[1] != logrus.ErrorLevel {
		t.Errorf("unexpected log levels returned")
	}
}

func TestWriterHookFireIgnoresUnlistedLevels(t *testing.T) {
	writer := &mockWriter{}
	hook := &utils.WriterHook{
		Writer:    writer,
		LogLevels: []logrus.Level{logrus.ErrorLevel},
	}

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel // not in hook.LogLevels
	entry.Message = "should not be logged"

	err := hook.Fire(entry)
	require.NoError(t, err)
	require.Equal(t, "", writer.buf.String(), "unexpected log written")
}

func TestLoggerBehavior_UnknownModeDefaults(t *testing.T) {
	os.Setenv("APP_MODE", "staging")
	buf := new(bytes.Buffer)

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(io.Discard)
	logger.AddHook(&utils.WriterHook{
		Writer:    buf,
		LogLevels: logrus.AllLevels,
	})

	logger.Info("test")

	require.NotEmpty(t, buf.String(), "expected log output even in unknown APP_MODE")
}

func TestLoggerBehavior(t *testing.T) {
	testCases := []struct {
		name          string
		appModeEnv    string
		logLevel      logrus.Level
		logMessage    string
		expectedLevel string
		expectedMsg   string
	}{
		{
			name:          "info_log_in_prod_mode",
			appModeEnv:    "prod",
			logLevel:      logrus.InfoLevel,
			logMessage:    "info level test",
			expectedLevel: "info",
			expectedMsg:   "info level test",
		},
		{
			name:          "error_log_in_prod_mode",
			appModeEnv:    "prod",
			logLevel:      logrus.ErrorLevel,
			logMessage:    "error level test",
			expectedLevel: "error",
			expectedMsg:   "error level test",
		},
		{
			name:          "info_log_in_dev_mode",
			appModeEnv:    "dev",
			logLevel:      logrus.InfoLevel,
			logMessage:    "dev info log",
			expectedLevel: "info",
			expectedMsg:   "dev info log",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("APP_MODE", tc.appModeEnv)

			buf := new(bytes.Buffer)
			logger := logrus.New()
			logger.SetFormatter(&logrus.JSONFormatter{
				PrettyPrint:     true,
				TimestampFormat: time.RFC3339,
			})
			logger.SetOutput(io.Discard)

			logger.AddHook(&utils.WriterHook{
				Writer: buf,
				LogLevels: []logrus.Level{
					logrus.InfoLevel,
					logrus.WarnLevel,
					logrus.DebugLevel,
					logrus.ErrorLevel,
					logrus.FatalLevel,
					logrus.PanicLevel,
				},
			})

			logger.SetLevel(logrus.DebugLevel)

			switch tc.logLevel {
			case logrus.InfoLevel:
				logger.Info(tc.logMessage)
			case logrus.ErrorLevel:
				logger.Error(tc.logMessage)
			}

			logStr := buf.String()
			require.NotEmpty(t, logStr)

			var logData map[string]any
			err := json.Unmarshal([]byte(logStr), &logData)
			require.NoError(t, err, "failed to parse log JSON")

			if logData["level"] != tc.expectedLevel {
				t.Errorf("expected level to be %s, got %v", tc.expectedLevel, logData["level"])
			}
			if logData["msg"] != tc.expectedMsg {
				t.Errorf("expected message to be %s, got %v", tc.expectedMsg, logData["msg"])
			}
		})
	}
}
