package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Add custom type for context key

type testKeyType string

const testKey testKeyType = "test-key"

func TestHandlerOptionalMiddleware_Success(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "test-token"})
	w := httptest.NewRecorder()

	expectedUser := database.User{ID: "user123", Role: "user"}
	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(expectedUser, nil)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, *user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerOptionalMiddleware_MissingToken(t *testing.T) {
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerOptionalMiddleware_InvalidToken(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "invalid-token", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerOptionalMiddleware_UserNotFound(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockUser := &MockUserService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}

	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		UserService:            mockUser,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "test-token"})
	w := httptest.NewRecorder()

	expectedClaims := &Claims{UserID: "user123"}

	mockRequestMetadata.On("GetIPAddress", req).Return("192.168.1.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-user-agent")
	mockAuth.On("ValidateAccessToken", "test-token", "test-secret").Return(expectedClaims, nil)
	mockUser.On("GetUserByID", req.Context(), "user123").Return(database.User{}, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockUser.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

// Optional: Legacy method structure test
func TestHandlersConfig_HandlerOptionalMiddleware_Structure(t *testing.T) {
	adapter := &HandlersConfig{}
	assert.NotNil(t, adapter)
}

func TestHandlerOptionalMiddleware_MalformedCookie(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}
	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""}) // Malformed/empty
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")
	mockAuth.On("ValidateAccessToken", "", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerOptionalMiddleware_MultipleCookies(t *testing.T) {
	mockAuth := &MockAuthService{}
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}
	cfg := &HandlerConfig{
		AuthService:            mockAuth,
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
		JWTSecret:              "test-secret",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")
	mockAuth.On("ValidateAccessToken", "", "test-secret").Return(nil, assert.AnError)

	// Set up logger expectations for error logging
	logger := logrus.New()
	entry := logger.WithError(assert.AnError)
	mockLogger.On("WithError", assert.AnError).Return(entry)

	handlerCalled := false
	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		handlerCalled = true
		assert.Nil(t, user)
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuth.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockRequestMetadata.AssertExpectations(t)
}

func TestHandlerOptionalMiddleware_ContextPropagation(t *testing.T) {
	mockLogger := &MockLoggerService{}
	mockRequestMetadata := &MockRequestMetadataService{}
	cfg := &HandlerConfig{
		LoggerService:          mockLogger,
		RequestMetadataService: mockRequestMetadata,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), testKey, "test-value")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockRequestMetadata.On("GetIPAddress", req).Return("127.0.0.1")
	mockRequestMetadata.On("GetUserAgent", req).Return("test-agent")

	testHandler := OptionalHandler(func(w http.ResponseWriter, r *http.Request, user *database.User) {
		assert.Equal(t, "test-value", r.Context().Value(testKey))
	})

	middleware := cfg.HandlerOptionalMiddleware(testHandler)
	middleware.ServeHTTP(w, req)
	mockRequestMetadata.AssertExpectations(t)
}
