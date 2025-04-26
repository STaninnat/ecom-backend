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
		name              string
		ctx               context.Context
		status            string
		action            string
		details           string
		errMsg            string
		ua                string
		ip                string
		expectedLevel     string
		expectedRequestID string
	}{
		{
			name:              "success_with_userid_and_requestid",
			ctx:               context.WithValue(context.WithValue(context.Background(), utils.ContextKeyUserID, "u-123"), utils.ContextKeyRequestID, "req-123"),
			status:            "success",
			action:            "signup",
			details:           "local",
			ua:                "agent-a",
			ip:                "1.1.1.1",
			expectedLevel:     "info",
			expectedRequestID: "req-123",
		},
		{
			name:              "fail_with_error_and_requestid",
			ctx:               context.WithValue(context.WithValue(context.Background(), utils.ContextKeyUserID, "u-999"), utils.ContextKeyRequestID, "req-999"),
			status:            "fail",
			action:            "sign",
			details:           "pwd mismatch",
			errMsg:            "invalid password",
			ua:                "agent-b",
			ip:                "1.2.3.4",
			expectedLevel:     "error",
			expectedRequestID: "req-999",
		},
		{
			name:              "no_userid_or_requestid_ua_ip",
			ctx:               context.Background(),
			status:            "success",
			action:            "health",
			ua:                "",
			ip:                "",
			expectedLevel:     "info",
			expectedRequestID: "",
		},
		{
			name:              "no_userid_with_requestid",
			ctx:               context.WithValue(context.WithValue(context.Background(), utils.ContextKeyRequestID, "req-123"), utils.ContextKeyUserID, nil),
			status:            "success",
			action:            "login",
			details:           "social login",
			ua:                "agent-c",
			ip:                "1.3.3.3",
			expectedLevel:     "info",
			expectedRequestID: "req-123",
		},
		{
			name:              "no_requestid_no_userid",
			ctx:               context.Background(),
			status:            "success",
			action:            "logout",
			details:           "normal logout",
			ua:                "agent-d",
			ip:                "1.4.4.4",
			expectedLevel:     "info",
			expectedRequestID: "",
		},
		{
			name:              "status_pending",
			ctx:               context.WithValue(context.WithValue(context.Background(), utils.ContextKeyUserID, "u-456"), utils.ContextKeyRequestID, "req-456"),
			status:            "pending",
			action:            "verify",
			details:           "verification in process",
			ua:                "agent-e",
			ip:                "1.5.5.5",
			expectedLevel:     "info",
			expectedRequestID: "req-456",
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
