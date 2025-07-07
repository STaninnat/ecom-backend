package authhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// HandlerSignIn handles user authentication requests for testing
func (cfg *TestHandlersAuthConfig) HandlerSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	// Parse and validate request
	params, err := auth.DecodeAndValidate[struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		cfg.LogHandlerError(
			ctx,
			"signin-local",
			"invalid_request",
			"Invalid signin payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Call business logic service
	result, err := cfg.GetAuthService().SignIn(ctx, SignInParams{
		Email:    params.Email,
		Password: params.Password,
	})

	if err != nil {
		cfg.handleAuthError(w, r, err, "signin-local", ip, userAgent)
		return
	}

	// Merge cart if needed
	cfg.MergeCart(ctx, r, result.UserID)

	// Set cookies
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)

	// Log success
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, result.UserID)
	cfg.LogHandlerSuccess(ctxWithUserID, "signin-local", "Local signin success", ip, userAgent)

	// Respond
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Signin successful",
	})
}

func TestHandlerSignIn_Success(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	expectedResult := &AuthResult{
		UserID:              "user123",
		AccessToken:         "access_token_123",
		RefreshToken:        "refresh_token_123",
		AccessTokenExpires:  time.Now().Add(30 * time.Minute),
		RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
		IsNewUser:           false,
	}

	// Setup expectations
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(expectedResult, nil)

	mockCartConfig.On("MergeCart", mock.Anything, mock.Anything, "user123").Return()
	mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "signin-local", "Local signin success", mock.Anything, mock.Anything).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Signin successful", response.Message)

	// Verify cookies were set
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 2) // access and refresh token cookies

	// Verify mocks
	mockAuthService.AssertExpectations(t)
	mockCartConfig.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignIn_InvalidRequest(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data - invalid JSON
	invalidJSON := `{"email": "test@example.com", "password": "password123"`

	// Setup expectations
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_request", "Invalid signin payload", mock.Anything, mock.Anything, mock.Anything).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])

	// Verify mocks
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "SignIn")
}

func TestHandlerSignIn_MissingFields(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data - missing password
	requestBody := map[string]string{
		"email": "test@example.com",
		// password missing
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Setup expectations - the service will be called with empty password
	appError := &handlers.AppError{Code: "invalid_password", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "",
	}).Return(nil, appError)

	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_password", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	// Verify mocks
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignIn_UserNotFound(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data
	requestBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Setup expectations
	appError := &handlers.AppError{Code: "user_not_found", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}).Return(nil, appError)

	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "user_not_found", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_InvalidPassword(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Setup expectations
	appError := &handlers.AppError{Code: "invalid_password", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}).Return(nil, appError)

	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_password", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_DatabaseError(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Setup expectations
	dbError := errors.New("database connection failed")
	appError := &handlers.AppError{Code: "database_error", Message: "Database error", Err: dbError}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(nil, appError)

	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "database_error", "Database error", mock.Anything, mock.Anything, dbError).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_UnknownError(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}

	// Test data
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// Setup expectations
	unknownError := errors.New("unknown error occurred")
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(nil, unknownError)

	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	// Create request
	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	cfg.HandlerSignIn(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	// Verify mocks
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}
