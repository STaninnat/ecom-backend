package userhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/mock"
)

// responseRecorder is a custom response writer that captures the status code
// and response body for testing purposes
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   string
}

func (r *responseRecorder) WriteHeader(status int)      { r.status = status }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body += string(b); return len(b), nil }

// --- MockUserService for GetUserService test ---
// MockUserService is a mock implementation of UserService for testing
type MockUserService struct{ UserService }

// mockHandlerLogger is a mock implementation of HandlerLogger interface
// for testing logging functionality
type mockHandlerLogger struct {
	mock.Mock
}

func (m *mockHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// MockAuth is a mock implementation of authentication functionality
// for testing auth-related operations
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

// MockAuthConfig is a mock implementation of AuthConfig for testing
// middleware functions that require authentication
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

// MockUserServiceForMiddleware is a mock implementation of UserService
// specifically designed for testing middleware functionality
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

// Add stub for PromoteUserToAdmin to satisfy UserService interface
func (m *MockUserServiceForMiddleware) PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error {
	args := m.Called(ctx, adminUser, targetUserID)
	return args.Error(0)
}

// TestHandlersConfig is a test-specific configuration that embeds HandlersConfig
// and provides a mock AuthConfig for testing
type TestHandlersConfig struct {
	*handlers.HandlersConfig
	TestAuth *MockAuthConfig
}

func (t *TestHandlersConfig) GetAuth() *MockAuthConfig {
	return t.TestAuth
}

// testUserConfig is a test double that tracks whether handler methods are called
// and captures the user passed to them for testing purposes
type testUserConfig struct {
	calledGetUser      bool
	calledUpdateUser   bool
	calledPromoteAdmin bool
	gotUser            database.User
}

func (cfg *testUserConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(database.User)
	cfg.calledGetUser = true
	if ok {
		cfg.gotUser = user
	}
}

func (cfg *testUserConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(database.User)
	cfg.calledUpdateUser = true
	if ok {
		cfg.gotUser = user
	}
}

func (cfg *testUserConfig) HandlerPromoteUserToAdmin(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(contextKeyUser).(database.User)
	cfg.calledPromoteAdmin = true
	if ok {
		cfg.gotUser = user
	}
}

// Provide AuthHandler wrappers that call the test double methods
func (cfg *testUserConfig) AuthHandlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerGetUser(w, r.WithContext(ctx))
}

func (cfg *testUserConfig) AuthHandlerUpdateUser(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerUpdateUser(w, r.WithContext(ctx))
}

func (cfg *testUserConfig) AuthHandlerPromoteUserToAdmin(w http.ResponseWriter, r *http.Request, user database.User) {
	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	cfg.HandlerPromoteUserToAdmin(w, r.WithContext(ctx))
}

// testUserExtractionConfig is a test double for testing user extraction middleware
// that embeds HandlersUserConfig to provide the base functionality
type testUserExtractionConfig struct {
	HandlersUserConfig
}

// dummyHandlerLogger is a no-op implementation of HandlerLogger for tests
// that don't need to verify logging behavior
type dummyHandlerLogger struct{}

func (d *dummyHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
}
func (d *dummyHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {}

// --- Mock for Update User ---
// mockUpdateUserService is a mock implementation of UserService
// specifically for testing user update functionality
type mockUpdateUserService struct {
	mock.Mock
}

func (m *mockUpdateUserService) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	return nil, nil // not used in update tests
}

func (m *mockUpdateUserService) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	args := m.Called(ctx, user, params)
	return args.Error(0)
}

func (m *mockUpdateUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil // not used in these tests
}

// Add stub for PromoteUserToAdmin to satisfy UserService interface
func (m *mockUpdateUserService) PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error {
	return nil // not used in these tests
}

// mockUpdateHandlerLogger is a mock implementation of HandlerLogger
// specifically for testing update user handler logging
type mockUpdateHandlerLogger struct {
	mock.Mock
}

func (m *mockUpdateHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockUpdateHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// --- Mock for Promote Admin ---
// mockPromoteUserService is a mock implementation of UserService
// specifically for testing user promotion to admin functionality
type mockPromoteUserService struct{ mock.Mock }

func (m *mockPromoteUserService) PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error {
	args := m.Called(ctx, adminUser, targetUserID)
	return args.Error(0)
}
func (m *mockPromoteUserService) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	return nil, nil
}
func (m *mockPromoteUserService) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	return nil
}
func (m *mockPromoteUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil
}

// mockPromoteLogger is a mock implementation of HandlerLogger
// specifically for testing promote admin handler logging
type mockPromoteLogger struct{ mock.Mock }

func (m *mockPromoteLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
func (m *mockPromoteLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// --- Mock for Get User ---
// mockGetUserService is a mock implementation of UserService
// specifically for testing user retrieval functionality
type mockGetUserService struct {
	mock.Mock
}

func (m *mockGetUserService) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *mockGetUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return database.User{}, nil // not used in these tests
}

func (m *mockGetUserService) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	return nil // not used in these tests
}

// Add stub for PromoteUserToAdmin to satisfy UserService interface
func (m *mockGetUserService) PromoteUserToAdmin(ctx context.Context, adminUser database.User, targetUserID string) error {
	return nil // not used in these tests
}

// mockGetHandlerLogger is a mock implementation of HandlerLogger
// specifically for testing get user handler logging
type mockGetHandlerLogger struct {
	mock.Mock
}

func (m *mockGetHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockGetHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}
