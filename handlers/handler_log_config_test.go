package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoggerService is a mock implementation of the LoggerService interface for testing.
type MockLoggerService struct {
	mock.Mock
}

// WithError mocks the WithError method for LoggerService.
func (m *MockLoggerService) WithError(err error) *logrus.Entry {
	args := m.Called(err)
	return args.Get(0).(*logrus.Entry)
}

// Error mocks the Error method for LoggerService.
func (m *MockLoggerService) Error(args ...any) {
	m.Called(args...)
}

// Info mocks the Info method for LoggerService.
func (m *MockLoggerService) Info(args ...any) {
	m.Called(args...)
}

// Debug mocks the Debug method for LoggerService.
func (m *MockLoggerService) Debug(args ...any) {
	m.Called(args...)
}

// Warn mocks the Warn method for LoggerService.
func (m *MockLoggerService) Warn(args ...any) {
	m.Called(args...)
}

// MockRequestMetadataService is a mock implementation of the RequestMetadataService interface for testing.
type MockRequestMetadataService struct {
	mock.Mock
}

// GetIPAddress mocks the GetIPAddress method for RequestMetadataService.
func (m *MockRequestMetadataService) GetIPAddress(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

// GetUserAgent mocks the GetUserAgent method for RequestMetadataService.
func (m *MockRequestMetadataService) GetUserAgent(r *http.Request) string {
	args := m.Called(r)
	return args.String(0)
}

// TestErrMsgOrNil tests the ErrMsgOrNil utility function.
// It checks that the correct error message or empty string is returned.
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

// TestHandlerConfig_GetRequestMetadata tests the GetRequestMetadata method of HandlerConfig.
// It checks that the correct IP address and user agent are returned from the mock service.
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

// TestHandlerConfig_LogHandlerError tests the LogHandlerError method of HandlerConfig.
// It checks that the logger's WithError method is called and expectations are met.
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

// TestHandlerConfig_LogHandlerError_NilError tests LogHandlerError with a nil error.
// It checks that the logger's Error method is called with the log message.
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

// TestHandlerConfig_LogHandlerSuccess tests the LogHandlerSuccess method of HandlerConfig.
// It checks that the method does not panic (currently a placeholder).
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

// TestGetRequestMetadata_Legacy tests the legacy GetRequestMetadata function.
// It checks that the correct IP address and user agent are extracted from the request headers.
func TestGetRequestMetadata_Legacy(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-user-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	ip, userAgent := GetRequestMetadata(req)

	assert.NotEmpty(t, ip)
	assert.Equal(t, "test-user-agent", userAgent)
}

// TestHandlersConfig_LogHandlerError_Legacy tests the legacy LogHandlerError method of HandlersConfig.
// It checks that the method does not panic when called with an error.
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

// TestHandlersConfig_LogHandlerSuccess_Legacy tests the legacy LogHandlerSuccess method of HandlersConfig.
// It checks that the method does not panic when called.
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

// TestLogHandlerError_NilLoggerService tests LogHandlerError with a nil LoggerService.
// It checks that the method does not panic when called.
func TestLogHandlerError_NilLoggerService(t *testing.T) {
	cfg := &HandlerConfig{
		LoggerService: nil,
	}
	// Should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerError(context.Background(), "action", "details", "logMsg", "ip", "ua", assert.AnError)
	})
}

// TestLogHandlerSuccess_NilLoggerService tests LogHandlerSuccess with a nil LoggerService.
// It checks that the method does not panic when called.
func TestLogHandlerSuccess_NilLoggerService(t *testing.T) {
	cfg := &HandlerConfig{
		LoggerService: nil,
	}
	// Should not panic
	assert.NotPanics(t, func() {
		cfg.LogHandlerSuccess(context.Background(), "action", "details", "ip", "ua")
	})
}

// TestGetRequestMetadata_NilRequest tests GetRequestMetadata with a nil request and service.
// It checks that the method returns empty strings and does not panic.
func TestGetRequestMetadata_NilRequest(t *testing.T) {
	cfg := &HandlerConfig{
		RequestMetadataService: nil,
	}
	// Should not panic, should return empty strings
	ip, ua := cfg.GetRequestMetadata(nil)
	assert.Equal(t, "", ip)
	assert.Equal(t, "", ua)
}
