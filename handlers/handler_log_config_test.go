package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoggerService for testing
type MockLoggerService struct {
	mock.Mock
}

func (m *MockLoggerService) WithError(err error) *logrus.Entry {
	args := m.Called(err)
	return args.Get(0).(*logrus.Entry)
}

func (m *MockLoggerService) Error(args ...any) {
	m.Called(args...)
}

func (m *MockLoggerService) Info(args ...any) {
	m.Called(args...)
}

func (m *MockLoggerService) Debug(args ...any) {
	m.Called(args...)
}

func (m *MockLoggerService) Warn(args ...any) {
	m.Called(args...)
}

// MockRequestMetadataService for testing
type MockRequestMetadataService struct {
	mock.Mock
}

func (m *MockRequestMetadataService) GetIPAddress(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func (m *MockRequestMetadataService) GetUserAgent(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

func TestErrMsgOrNil(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "non-nil error",
			err:      assert.AnError,
			expected: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrMsgOrNil(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandlerConfig_GetRequestMetadata(t *testing.T) {
	mockRequestMetadata := &MockRequestMetadataService{}
	cfg := &HandlerConfig{
		RequestMetadataService: mockRequestMetadata,
	}

	req, _ := http.NewRequest("GET", "/test", nil)
	expectedIP := "192.168.1.1"
	expectedUA := "test-user-agent"

	mockRequestMetadata.On("GetIPAddress", req).Return(expectedIP)
	mockRequestMetadata.On("GetUserAgent", req).Return(expectedUA)

	ip, userAgent := cfg.GetRequestMetadata(req)

	assert.Equal(t, expectedIP, ip)
	assert.Equal(t, expectedUA, userAgent)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerConfig_LogHandlerError(t *testing.T) {
	mockLogger := &MockLoggerService{}

	cfg := &HandlerConfig{
		LoggerService: mockLogger,
	}

	ctx := context.Background()
	action := "test_action"
	details := "test_details"
	logMsg := "test_log_msg"
	ip := "192.168.1.1"
	ua := "test-user-agent"
	err := assert.AnError

	// Create a proper logrus entry
	logger := logrus.New()
	entry := logger.WithError(err)

	mockLogger.On("WithError", err).Return(entry)

	cfg.LogHandlerError(ctx, action, details, logMsg, ip, ua, err)

	mockLogger.AssertExpectations(t)
}

func TestHandlerConfig_LogHandlerError_NilError(t *testing.T) {
	mockLogger := &MockLoggerService{}

	cfg := &HandlerConfig{
		LoggerService: mockLogger,
	}

	ctx := context.Background()
	action := "test_action"
	details := "test_details"
	logMsg := "test_log_msg"
	ip := "192.168.1.1"
	ua := "test-user-agent"

	mockLogger.On("Error", logMsg).Return()

	cfg.LogHandlerError(ctx, action, details, logMsg, ip, ua, nil)

	mockLogger.AssertExpectations(t)
}

func TestHandlerConfig_LogHandlerSuccess(t *testing.T) {
	mockLogger := &MockLoggerService{}

	cfg := &HandlerConfig{
		LoggerService: mockLogger,
	}

	ctx := context.Background()
	action := "test_action"
	details := "test_details"
	ip := "192.168.1.1"
	ua := "test-user-agent"

	// LogHandlerSuccess currently has a TODO comment, so we'll just test that it doesn't panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerSuccess(ctx, action, details, ip, ua)
	})
}

func TestGetRequestMetadata_Legacy(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-user-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	ip, userAgent := GetRequestMetadata(req)

	assert.NotEmpty(t, ip)
	assert.Equal(t, "test-user-agent", userAgent)
}

func TestHandlersConfig_LogHandlerError_Legacy(t *testing.T) {
	logger := logrus.New()
	cfg := &HandlersConfig{
		Logger: logger,
	}

	ctx := context.Background()
	action := "test_action"
	details := "test_details"
	logMsg := "test_log_msg"
	ip := "192.168.1.1"
	ua := "test-user-agent"
	err := assert.AnError

	// This should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerError(ctx, action, details, logMsg, ip, ua, err)
	})
}

func TestHandlersConfig_LogHandlerSuccess_Legacy(t *testing.T) {
	logger := logrus.New()
	cfg := &HandlersConfig{
		Logger: logger,
	}

	ctx := context.Background()
	action := "test_action"
	details := "test_details"
	ip := "192.168.1.1"
	ua := "test-user-agent"

	// This should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerSuccess(ctx, action, details, ip, ua)
	})
}

func TestLogHandlerError_NilLoggerService(t *testing.T) {
	cfg := &HandlerConfig{
		LoggerService: nil,
	}
	// Should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerError(context.Background(), "action", "details", "logMsg", "ip", "ua", assert.AnError)
	})
}

func TestLogHandlerSuccess_NilLoggerService(t *testing.T) {
	cfg := &HandlerConfig{
		LoggerService: nil,
	}
	// Should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerSuccess(context.Background(), "action", "details", "ip", "ua")
	})
}

func TestGetRequestMetadata_NilRequest(t *testing.T) {
	cfg := &HandlerConfig{
		RequestMetadataService: nil,
	}
	// Should not panic, should return empty strings
	ip, ua := cfg.GetRequestMetadata(nil)
	assert.Equal(t, "", ip)
	assert.Equal(t, "", ua)
}
