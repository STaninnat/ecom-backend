package utils_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type captureWriter struct{ bytes.Buffer }

func (w *captureWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p) }

func TestLogUserAction(t *testing.T) {
	testCases := []struct {
		name          string
		ctx           context.Context
		status        string
		action        string
		details       string
		errMsg        string
		ua            string
		ip            string
		expectedLevel string
	}{
		{
			name:          "success_with_userid",
			ctx:           context.WithValue(context.Background(), utils.ContextKeyUserID, "u-123"),
			status:        "success",
			action:        "signup",
			details:       "local",
			ua:            "agent-a",
			ip:            "1.1.1.1",
			expectedLevel: "info",
		},
		{
			name:          "fail_with_error",
			ctx:           context.WithValue(context.Background(), utils.ContextKeyUserID, "u-999"),
			status:        "fail",
			action:        "sign",
			details:       "pwd mismatch",
			errMsg:        "invalid password",
			ua:            "agent-b",
			ip:            "1.2.3.4",
			expectedLevel: "error",
		},
		{
			name:          "no_userid_ua_ip",
			ctx:           context.Background(),
			status:        "success",
			action:        "health",
			ua:            "",
			ip:            "",
			expectedLevel: "info",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := &captureWriter{}
			logger := logrus.New()
			logger.SetFormatter(&logrus.JSONFormatter{})
			logger.SetOutput(w)

			utils.LogUserAction(utils.ActionLogParams{
				Logger:    logger,
				Ctx:       tc.ctx,
				Action:    tc.action,
				Status:    tc.status,
				Details:   tc.details,
				ErrorMsg:  tc.errMsg,
				UserAgent: tc.ua,
				IP:        tc.ip,
			})

			var m map[string]any
			err := json.Unmarshal(w.Bytes(), &m)

			require.NoError(t, err)
			require.Equal(t, tc.expectedLevel, m["level"])
			require.Equal(t, tc.action, m["action"])

			if uid := tc.ctx.Value(utils.ContextKeyUserID); uid != nil {
				require.Equal(t, uid, m["userID"])
			}
			if tc.errMsg != "" {
				require.Equal(t, tc.errMsg, m["error"])
			}

			require.Equal(t, tc.ua, m["userAgent"])
			require.Equal(t, tc.ip, m["ip"])
		})

	}
}
