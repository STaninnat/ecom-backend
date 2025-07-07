package userhandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInitUserService_MissingHandlersConfig(t *testing.T) {
	cfg := &HandlersUserConfig{HandlersConfig: nil}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

func TestInitUserService_MissingDB(t *testing.T) {
	cfg := &HandlersUserConfig{HandlersConfig: &handlers.HandlersConfig{APIConfig: &config.APIConfig{DB: nil}}}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

func TestGetUserService_AlreadyInitialized(t *testing.T) {
	mockService := new(MockUserService)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		userService:    mockService,
	}
	service := cfg.GetUserService()
	assert.Equal(t, mockService, service)
}

func TestGetUserService_InitializesWithNilConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: nil,
		userService:    nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

func TestGetUserService_InitializesWithNilDB(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{APIConfig: &config.APIConfig{DB: nil}},
		userService:    nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

func TestGetUserService_InitializesWithValidConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{DB: &database.Queries{}},
		},
		userService: nil,
	}
	service := cfg.GetUserService()
	assert.NotNil(t, service)
}

// --- handleUserError ---

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   string
}

func (r *responseRecorder) WriteHeader(status int)      { r.status = status }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body += string(b); return len(b), nil }

func TestHandleUserError_KnownError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "update_failed", Message: "fail", Err: errors.New("db")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "update_failed", "fail", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_UnknownError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := errors.New("unknown error")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_CommitError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "commit_error", Message: "Commit failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "commit_error", "Commit failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_TransactionError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "transaction_error", Message: "Transaction failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "transaction_error", "Transaction failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

// --- MockUserService for GetUserService test ---
type MockUserService struct{ UserService }

func TestInitUserService_Success(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{DB: &database.Queries{}},
		},
		Logger: new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.NoError(t, err)
	assert.NotNil(t, cfg.userService)
}

func TestHandleUserError_DefaultCase(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "other_error", Message: "fail", Err: errors.New("db")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "internal_error", "fail", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_NonUserError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := errors.New("unexpected error")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

// Add mockHandlerLogger for HandlerLogger interface

type mockHandlerLogger struct {
	mock.Mock
}

func (m *mockHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockAuth for testing
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) ValidateAccessToken(token, secret string) (*handlers.Claims, error) {
	args := m.Called(token, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*handlers.Claims), args.Error(1)
}

func (m *MockAuth) GetUserByID(ctx context.Context, id string) (database.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return database.User{}, args.Error(1)
	}
	return args.Get(0).(database.User), args.Error(1)
}

func TestHandleUserError_UserNotFound(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "user_not_found", Message: "User not found", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "user_not_found", "User not found", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusNotFound, w.status)
	assert.Contains(t, w.body, "User not found")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_InvalidRequest(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid request", Err: errors.New("validation error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "invalid_request", "Invalid request", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusBadRequest, w.status)
	assert.Contains(t, w.body, "Invalid request")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_UpdateFailed(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "update_failed", Message: "Update failed", Err: errors.New("db error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "update_failed", "Update failed", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Something went wrong")
	mockLogger.AssertExpectations(t)
}

func TestHandleUserError_DefaultAppError(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger: mockLogger,
	}
	err := &handlers.AppError{Code: "custom_error", Message: "Custom error", Err: errors.New("custom error")}
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "internal_error", "Custom error", "ip", "ua", err.Err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
	mockLogger.AssertExpectations(t)
}

func TestInitUserService_NilHandlersConfig(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: nil,
		Logger:         new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

func TestInitUserService_NilDB(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{APIConfig: &config.APIConfig{DB: nil}},
		Logger:         new(mockHandlerLogger),
	}
	err := cfg.InitUserService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// MockAuthConfig for testing middleware functions
type MockAuthConfig struct {
	mock.Mock
}

func (m *MockAuthConfig) ValidateAccessToken(token, secret string) (*handlers.Claims, error) {
	args := m.Called(token, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*handlers.Claims), args.Error(1)
}

func (m *MockAuthConfig) GetUserByID(ctx context.Context, id string) (database.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return database.User{}, args.Error(1)
	}
	return args.Get(0).(database.User), args.Error(1)
}

// MockUserService for middleware tests
type MockUserServiceForMiddleware struct {
	mock.Mock
}

func (m *MockUserServiceForMiddleware) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *MockUserServiceForMiddleware) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	args := m.Called(ctx, user, params)
	return args.Error(0)
}

func (m *MockUserServiceForMiddleware) GetUserByID(ctx context.Context, id string) (database.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return database.User{}, args.Error(1)
	}
	return args.Get(0).(database.User), args.Error(1)
}

// Test-specific configuration for middleware testing
type TestHandlersConfig struct {
	*handlers.HandlersConfig
	TestAuth *MockAuthConfig
}

func (t *TestHandlersConfig) GetAuth() *MockAuthConfig {
	return t.TestAuth
}

func TestExtractUserFromRequest_NoAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request without Authorization header
	r := httptest.NewRequest("GET", "/user", nil)

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

func TestExtractUserFromRequest_EmptyAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with empty Authorization header
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

func TestExtractUserFromRequest_ShortAuthHeader(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with short Authorization header
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "Short")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

func TestExtractUserFromRequest_InvalidBearerPrefix(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Create a request with invalid Bearer prefix
	r := httptest.NewRequest("GET", "/user", nil)
	r.Header.Set("Authorization", "InvalidPrefix valid-token")

	user, err := cfg.extractUserFromRequest(r)
	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Contains(t, err.Error(), "missing or invalid Authorization header")
}

// Test UserExtractionMiddleware by testing its behavior indirectly
func TestUserExtractionMiddleware_Behavior(t *testing.T) {
	mockLogger := new(mockHandlerLogger)
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger:    logrus.New(),
			APIConfig: &config.APIConfig{},
		},
		Logger:      mockLogger,
		userService: mockUserService,
	}

	// Test that the middleware returns a handler function
	middleware := cfg.UserExtractionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	assert.NotNil(t, middleware)

	// Test that it's callable
	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, r)
	// Should not panic and should complete successfully
}

// Test extractUserFromRequest with simple scenarios that don't require complex mocking
func TestExtractUserFromRequest_SimpleScenarios(t *testing.T) {
	mockUserService := new(MockUserServiceForMiddleware)

	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{},
		},
		userService: mockUserService,
	}

	// Test cases that don't require auth mocking
	testCases := []struct {
		name          string
		authHeader    string
		expectedError string
	}{
		{
			name:          "No Authorization header",
			authHeader:    "",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Empty Authorization header",
			authHeader:    "",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Short Authorization header",
			authHeader:    "Short",
			expectedError: "missing or invalid Authorization header",
		},
		{
			name:          "Invalid Bearer prefix",
			authHeader:    "InvalidPrefix valid-token",
			expectedError: "missing or invalid Authorization header",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/user", nil)
			if tc.authHeader != "" {
				r.Header.Set("Authorization", tc.authHeader)
			}

			user, err := cfg.extractUserFromRequest(r)
			assert.Error(t, err)
			assert.Equal(t, database.User{}, user)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
