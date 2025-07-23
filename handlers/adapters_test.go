// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// adapters_test.go: Tests for adapter implementations that integrate services with handler middleware, including legacy support.

const (
	testTokenString       = "test-token"
	testSecret            = "test-secret"
	testIPAddress         = "192.168.1.1"
	testUserAgentAdapters = "test-user-agent"
)

// TestHandlerConfigAuthAdapter_ValidateAccessToken_Success tests successful validation of an access token.
// It checks that the adapter returns the expected claims and no error.
func TestHandlerConfigAuthAdapter_ValidateAccessToken_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	adapter := &handlerConfigAuthAdapter{authService: mockAuthService}

	tokenString := testTokenString
	secret := testSecret
	expectedClaims := &Claims{UserID: "user123"}

	mockAuthService.On("ValidateAccessToken", tokenString, secret).Return(expectedClaims, nil)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	mockAuthService.AssertExpectations(t)
}

// TestHandlerConfigAuthAdapter_ValidateAccessToken_Error tests error handling during access token validation.
// It checks that the adapter returns an error and nil claims for invalid tokens.
func TestHandlerConfigAuthAdapter_ValidateAccessToken_Error(t *testing.T) {
	mockAuthService := &MockAuthService{}
	adapter := &handlerConfigAuthAdapter{authService: mockAuthService}

	tokenString := "invalid-token"
	secret := testSecret
	expectedError := assert.AnError

	mockAuthService.On("ValidateAccessToken", tokenString, secret).Return(nil, expectedError)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, expectedError, err)
	mockAuthService.AssertExpectations(t)
}

// TestHandlerConfigUserAdapter_GetUserByID_Success tests successful retrieval of a user by ID.
// It checks that the adapter returns the expected user and no error.
func TestHandlerConfigUserAdapter_GetUserByID_Success(t *testing.T) {
	mockUserService := &MockUserService{}
	adapter := &handlerConfigUserAdapter{userService: mockUserService}

	ctx := context.Background()
	userID := "user123"
	expectedUser := database.User{ID: userID, Email: "test@example.com"}

	mockUserService.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

	user, err := adapter.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockUserService.AssertExpectations(t)
}

// TestHandlerConfigUserAdapter_GetUserByID_Error tests error handling when retrieving a user by ID.
// It checks that the adapter returns an error and an empty user for nonexistent IDs.
func TestHandlerConfigUserAdapter_GetUserByID_Error(t *testing.T) {
	mockUserService := &MockUserService{}
	adapter := &handlerConfigUserAdapter{userService: mockUserService}

	ctx := context.Background()
	userID := "nonexistent"
	expectedError := assert.AnError

	mockUserService.On("GetUserByID", ctx, userID).Return(database.User{}, expectedError)

	user, err := adapter.GetUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, database.User{}, user)
	assert.Equal(t, expectedError, err)
	mockUserService.AssertExpectations(t)
}

// TestHandlerConfigLoggerAdapter_WithError tests the WithError method of the logger adapter.
// It checks that the adapter returns a non-nil log entry with the error attached.
func TestHandlerConfigLoggerAdapter_WithError(t *testing.T) {
	mockLoggerService := &MockLoggerService{}
	adapter := &handlerConfigLoggerAdapter{loggerService: mockLoggerService}

	err := assert.AnError
	logger := logrus.New()
	entry := logger.WithError(err)

	mockLoggerService.On("WithError", err).Return(entry)

	result := adapter.WithError(err)

	assert.NotNil(t, result)
	mockLoggerService.AssertExpectations(t)
}

// TestHandlerConfigLoggerAdapter_Error tests the Error method of the logger adapter.
// It checks that the adapter calls the logger service's Error method as expected.
func TestHandlerConfigLoggerAdapter_Error(t *testing.T) {
	mockLoggerService := &MockLoggerService{}
	adapter := &handlerConfigLoggerAdapter{loggerService: mockLoggerService}

	mockLoggerService.On("Error", "test", "message").Return()

	adapter.Error("test", "message")

	mockLoggerService.AssertExpectations(t)
}

// TestHandlerConfigMetadataAdapter_GetIPAddress tests retrieval of the IP address from a request.
// It checks that the adapter returns the expected IP address from the metadata service.
func TestHandlerConfigMetadataAdapter_GetIPAddress(t *testing.T) {
	mockMetadataService := &MockRequestMetadataService{}
	adapter := &handlerConfigMetadataAdapter{metadataService: mockMetadataService}

	req, _ := http.NewRequest("GET", "/test", nil)
	expectedIP := testIPAddress

	mockMetadataService.On("GetIPAddress", req).Return(expectedIP)

	ip := adapter.GetIPAddress(req)

	assert.Equal(t, expectedIP, ip)
	mockMetadataService.AssertExpectations(t)
}

// TestHandlerConfigMetadataAdapter_GetUserAgent tests retrieval of the user agent from a request.
// It checks that the adapter returns the expected user agent from the metadata service.
func TestHandlerConfigMetadataAdapter_GetUserAgent(t *testing.T) {
	mockMetadataService := &MockRequestMetadataService{}
	adapter := &handlerConfigMetadataAdapter{metadataService: mockMetadataService}

	req, _ := http.NewRequest("GET", "/test", nil)
	expectedUA := testUserAgentAdapters

	mockMetadataService.On("GetUserAgent", req).Return(expectedUA)

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, expectedUA, ua)
	mockMetadataService.AssertExpectations(t)
}

// Test Legacy Adapters
// ===================

// MockLegacyAuth is a mock for testing the legacy auth adapter.
type MockLegacyAuth struct {
	mock.Mock
}

// ValidateAccessToken mocks the ValidateAccessToken method for legacy auth.
func (m *MockLegacyAuth) ValidateAccessToken(tokenString, secret string) (*auth.Claims, error) {
	args := m.Called(tokenString, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

// TestLegacyAuthService_ValidateAccessToken_Success tests successful validation of an access token using the legacy auth service.
// It checks that the adapter returns the expected claims and no error.
func TestLegacyAuthService_ValidateAccessToken_Success(t *testing.T) {
	mockAuth := &MockLegacyAuth{}
	adapter := &legacyAuthService{auth: mockAuth}

	tokenString := testTokenString
	secret := testSecret
	expectedAuthClaims := &auth.Claims{UserID: "user123"}

	mockAuth.On("ValidateAccessToken", tokenString, secret).Return(expectedAuthClaims, nil)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	mockAuth.AssertExpectations(t)
}

// TestLegacyAuthService_ValidateAccessToken_Error tests error handling during access token validation in the legacy auth service.
// It checks that the adapter returns an error and nil claims for invalid tokens.
func TestLegacyAuthService_ValidateAccessToken_Error(t *testing.T) {
	mockAuth := &MockLegacyAuth{}
	adapter := &legacyAuthService{auth: mockAuth}

	tokenString := "invalid-token"
	secret := testSecret
	expectedError := assert.AnError

	mockAuth.On("ValidateAccessToken", tokenString, secret).Return(nil, expectedError)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, expectedError, err)
	mockAuth.AssertExpectations(t)
}

// Note: Legacy user service tests are skipped because they require *database.Queries
// which is difficult to mock properly. In a real scenario, these would be tested
// with integration tests using a test database.

// TestLegacyUserService_GetUserByID tests retrieval of a user by ID using the legacy user service.
// It checks that the adapter returns the expected user and no error, and verifies SQL expectations.
func TestLegacyUserService_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// defer func() {
	// 	if err := db.Close(); err != nil {
	// 		t.Errorf("db.Close() failed: %v", err)
	// 	}
	// }()

	queries := database.New(db)
	adapter := &legacyUserService{db: queries}

	ctx := context.Background()
	userID := "user123"
	expectedUser := database.User{ID: userID, Name: "Test User", Email: "test@example.com"}

	// Set up expected query and result
	mock.ExpectQuery(`SELECT id, name, email, password, provider, provider_id, phone, address, role, created_at, updated_at FROM users\s+WHERE id = \$1\s+LIMIT 1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "provider", "provider_id", "phone", "address", "role", "created_at", "updated_at"}).
			AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, nil, "local", nil, nil, nil, "user", expectedUser.CreatedAt, expectedUser.UpdatedAt))

	user, err := adapter.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)
	assert.Equal(t, expectedUser.Email, user.Email)
	// Optionally check other fields

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestLegacyUserService_Structure tests the structure of the legacy user service.
// It checks that the adapter can be created (placeholder test).
func TestLegacyUserService_Structure(t *testing.T) {
	// Test that the legacy user service can be created (structure test)
	// This is a placeholder test - actual functionality would be tested with integration tests
	adapter := &legacyUserService{db: nil}
	assert.NotNil(t, adapter)
}

// TestLegacyLoggerService_WithError tests the WithError method of the legacy logger service.
// It checks that the adapter returns a logrus.Entry with the error attached.
func TestLegacyLoggerService_WithError(t *testing.T) {
	logger := logrus.New()
	adapter := &legacyLoggerService{logger: logger}

	err := assert.AnError

	result := adapter.WithError(err)

	assert.NotNil(t, result)
	// The result should be a logrus.Entry with the error
	entry, ok := result.(*logrus.Entry)
	assert.True(t, ok)
	assert.NotNil(t, entry)
}

// TestLegacyLoggerService_Error tests the Error method of the legacy logger service.
// It checks that calling Error does not panic.
func TestLegacyLoggerService_Error(t *testing.T) {
	logger := logrus.New()
	adapter := &legacyLoggerService{logger: logger}

	// This should not panic
	assert.NotPanics(t, func() {
		adapter.Error("test", "message")
	})
}

// TestLegacyMetadataService_GetIPAddress tests retrieval of the IP address from a request using the legacy metadata service.
// It checks that the adapter returns a non-empty IP address.
func TestLegacyMetadataService_GetIPAddress(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", testIPAddress)

	ip := adapter.GetIPAddress(req)

	assert.NotEmpty(t, ip)
}

// TestLegacyMetadataService_GetUserAgent tests retrieval of the user agent from a request using the legacy metadata service.
// It checks that the adapter returns the expected user agent string.
func TestLegacyMetadataService_GetUserAgent(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", testUserAgentAdapters)

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, testUserAgentAdapters, ua)
}

// TestLegacyMetadataService_GetUserAgent_Empty tests retrieval of the user agent when the header is missing.
// It checks that the adapter returns an empty string if the User-Agent header is not set.
func TestLegacyMetadataService_GetUserAgent_Empty(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	// No User-Agent header set

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, "", ua)
}
