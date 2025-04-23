package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type bufLogger struct{ bytes.Buffer }

func (b *bufLogger) Write(p []byte) (int, error) { return b.Buffer.Write(p) }

func newMockConfig(w *bufLogger) *handlers.HandlersConfig {
	lg := logrus.New()
	lg.SetFormatter(&logrus.JSONFormatter{})
	lg.SetOutput(w)
	return &handlers.HandlersConfig{Logger: lg}
}

func TestLogHandlerFunctions(t *testing.T) {
	baseCtx := context.Background()
	ctxWithID := context.WithValue(baseCtx, utils.ContextKeyUserID, "uid-1")

	testCases := []struct {
		name    string
		ok      bool // true = success path
		withUID bool
		withErr bool
	}{
		{"success_with_uid", true, true, false},
		{"success_no_uid", true, false, false},
		{"error_with_err_uid", false, true, true},
		{"error_no_err_no_uid", false, false, false},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			buf := &bufLogger{}
			cfg := newMockConfig(buf)

			ctx := baseCtx
			if c.withUID {
				ctx = ctxWithID
			}

			if c.ok {
				cfg.LogHandlerSuccess(ctx, "signup", "done", "1.1.1.1", "ua-test")
			} else {
				var e error
				if c.withErr {
					e = errors.New("boom")
				}
				cfg.LogHandlerError(ctx, "sign", "fail", "log msg", "1.1.1.1", "ua-x", e)
			}

			// pick last line
			lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
			last := lines[len(lines)-1]

			var entry map[string]any
			require.NoError(t, json.Unmarshal(last, &entry))

			if c.ok {
				require.Equal(t, "success", entry["status"])
				require.Equal(t, "signup", entry["action"])
			} else {
				require.Equal(t, "fail", entry["status"])
				require.Equal(t, "sign", entry["action"])
				if c.withErr {
					require.Contains(t, entry, "error")
				} else {
					require.NotContains(t, entry, "error")
				}
			}

			if c.withUID {
				require.Equal(t, "uid-1", entry["userID"])
			} else {
				require.Nil(t, entry["userID"])
			}
			require.Equal(t, "1.1.1.1", entry["ip"])
		})
	}
}
