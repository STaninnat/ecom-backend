package utils

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

type testHook struct {
	entries []*logrus.Entry
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(e *logrus.Entry) error {
	h.entries = append(h.entries, e)
	return nil
}

func newTestLogger() (*logrus.Logger, *testHook, *bytes.Buffer) {
	logger := logrus.New()
	buf := &bytes.Buffer{}
	logger.SetOutput(buf)
	hook := &testHook{}
	logger.AddHook(hook)
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	return logger, hook, buf
}

func TestLogUserAction(t *testing.T) {
	statuses := []struct {
		status  string
		wantMsg string
		wantLvl logrus.Level
	}{
		{"pending", "User action pending", logrus.InfoLevel},
		{"fail", "User action failed", logrus.ErrorLevel},
		{"success", "User action success", logrus.InfoLevel},
		{"", "User action success", logrus.InfoLevel},
	}

	for _, tc := range statuses {
		t.Run(tc.status, func(t *testing.T) {
			logger, hook, buf := newTestLogger()
			ctx := context.WithValue(context.Background(), ContextKeyUserID, "u123")
			ctx = context.WithValue(ctx, ContextKeyRequestID, "r456")
			params := ActionLogParams{
				Logger:    logger,
				Ctx:       ctx,
				Action:    "login",
				Status:    tc.status,
				Details:   "details",
				ErrorMsg:  "",
				UserAgent: "ua",
				IP:        "127.0.0.1",
			}
			LogUserAction(params)
			if len(hook.entries) == 0 {
				t.Fatalf("no log entries captured")
			}
			entry := hook.entries[len(hook.entries)-1]
			if entry.Level != tc.wantLvl {
				t.Errorf("expected level %v, got %v", tc.wantLvl, entry.Level)
			}
			if !strings.Contains(buf.String(), tc.wantMsg) {
				t.Errorf("expected log message %q in output, got %q", tc.wantMsg, buf.String())
			}
			if entry.Data["userID"] != "u123" || entry.Data["request_id"] != "r456" {
				t.Errorf("context values not logged correctly: %+v", entry.Data)
			}
		})
	}

	t.Run("with error message", func(t *testing.T) {
		logger, hook, buf := newTestLogger()
		ctx := context.WithValue(context.Background(), ContextKeyUserID, "u1")
		ctx = context.WithValue(ctx, ContextKeyRequestID, "r2")
		params := ActionLogParams{
			Logger:    logger,
			Ctx:       ctx,
			Action:    "update",
			Status:    "fail",
			Details:   "bad stuff",
			ErrorMsg:  "something went wrong",
			UserAgent: "ua",
			IP:        "ip",
		}
		LogUserAction(params)
		if len(hook.entries) == 0 {
			t.Fatalf("no log entries captured")
		}
		entry := hook.entries[len(hook.entries)-1]
		if entry.Data["error"] != "something went wrong" {
			t.Errorf("expected error field in log entry, got %+v", entry.Data)
		}
		if entry.Level != logrus.ErrorLevel {
			t.Errorf("expected ErrorLevel, got %v", entry.Level)
		}
		if !strings.Contains(buf.String(), "User action failed") {
			t.Errorf("expected fail message in log output")
		}
	})
}
