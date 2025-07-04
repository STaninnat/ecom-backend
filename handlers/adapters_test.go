package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test HandlerConfig Adapters
// ===========================

func TestHandlerConfigAuthAdapter_ValidateAccessToken_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	adapter := &handlerConfigAuthAdapter{authService: mockAuthService}

	tokenString := "test-token"
	secret := "test-secret"
	expectedClaims := &Claims{UserID: "user123"}

	mockAuthService.On("ValidateAccessToken", tokenString, secret).Return(expectedClaims, nil)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerConfigAuthAdapter_ValidateAccessToken_Error(t *testing.T) {
	mockAuthService := &MockAuthService{}
	adapter := &handlerConfigAuthAdapter{authService: mockAuthService}

	tokenString := "invalid-token"
	secret := "test-secret"
	expectedError := assert.AnError

	mockAuthService.On("ValidateAccessToken", tokenString, secret).Return(nil, expectedError)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, expectedError, err)
	mockAuthService.AssertExpectations(t)
}

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

func TestHandlerConfigLoggerAdapter_Error(t *testing.T) {
	mockLoggerService := &MockLoggerService{}
	adapter := &handlerConfigLoggerAdapter{loggerService: mockLoggerService}

	mockLoggerService.On("Error", "test", "message").Return()

	adapter.Error("test", "message")

	mockLoggerService.AssertExpectations(t)
}

func TestHandlerConfigMetadataAdapter_GetIPAddress(t *testing.T) {
	mockMetadataService := &MockRequestMetadataService{}
	adapter := &handlerConfigMetadataAdapter{metadataService: mockMetadataService}

	req, _ := http.NewRequest("GET", "/test", nil)
	expectedIP := "192.168.1.1"

	mockMetadataService.On("GetIPAddress", req).Return(expectedIP)

	ip := adapter.GetIPAddress(req)

	assert.Equal(t, expectedIP, ip)
	mockMetadataService.AssertExpectations(t)
}

func TestHandlerConfigMetadataAdapter_GetUserAgent(t *testing.T) {
	mockMetadataService := &MockRequestMetadataService{}
	adapter := &handlerConfigMetadataAdapter{metadataService: mockMetadataService}

	req, _ := http.NewRequest("GET", "/test", nil)
	expectedUA := "test-user-agent"

	mockMetadataService.On("GetUserAgent", req).Return(expectedUA)

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, expectedUA, ua)
	mockMetadataService.AssertExpectations(t)
}

// Test Legacy Adapters
// ===================

// MockLegacyAuth for testing legacy auth adapter
type MockLegacyAuth struct {
	mock.Mock
}

func (m *MockLegacyAuth) ValidateAccessToken(tokenString, secret string) (*auth.Claims, error) {
	args := m.Called(tokenString, secret)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func TestLegacyAuthService_ValidateAccessToken_Success(t *testing.T) {
	mockAuth := &MockLegacyAuth{}
	adapter := &legacyAuthService{auth: mockAuth}

	tokenString := "test-token"
	secret := "test-secret"
	expectedAuthClaims := &auth.Claims{UserID: "user123"}

	mockAuth.On("ValidateAccessToken", tokenString, secret).Return(expectedAuthClaims, nil)

	claims, err := adapter.ValidateAccessToken(tokenString, secret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	mockAuth.AssertExpectations(t)
}

func TestLegacyAuthService_ValidateAccessToken_Error(t *testing.T) {
	mockAuth := &MockLegacyAuth{}
	adapter := &legacyAuthService{auth: mockAuth}

	tokenString := "invalid-token"
	secret := "test-secret"
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

func TestLegacyUserService_Structure(t *testing.T) {
	// Test that the legacy user service can be created (structure test)
	// This is a placeholder test - actual functionality would be tested with integration tests
	adapter := &legacyUserService{db: nil}
	assert.NotNil(t, adapter)
}

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

func TestLegacyLoggerService_Error(t *testing.T) {
	logger := logrus.New()
	adapter := &legacyLoggerService{logger: logger}

	// This should not panic
	assert.NotPanics(t, func() {
		adapter.Error("test", "message")
	})
}

func TestLegacyMetadataService_GetIPAddress(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	ip := adapter.GetIPAddress(req)

	assert.NotEmpty(t, ip)
}

func TestLegacyMetadataService_GetUserAgent(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-user-agent")

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, "test-user-agent", ua)
}

func TestLegacyMetadataService_GetUserAgent_Empty(t *testing.T) {
	adapter := &legacyMetadataService{}

	req, _ := http.NewRequest("GET", "/test", nil)
	// No User-Agent header set

	ua := adapter.GetUserAgent(req)

	assert.Equal(t, "", ua)
}
