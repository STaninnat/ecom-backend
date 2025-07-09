package authhandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/handlers/cart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandlerSignIn_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

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

	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(expectedResult, nil)
	mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "signin-local", "Local signin success", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Signin successful", response.Message)
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 2)
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignIn_InvalidRequest(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	invalidJSON := `{"email": "test@example.com", "password": "password123"`
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_request", "Invalid signin payload", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "SignIn")
}

func TestHandlerSignIn_MissingFields(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	requestBody := map[string]string{
		"email": "test@example.com",
	}
	jsonBody, _ := json.Marshal(requestBody)

	appError := &handlers.AppError{Code: "invalid_password", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "",
	}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_password", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignIn_UserNotFound(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	requestBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	appError := &handlers.AppError{Code: "user_not_found", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "user_not_found", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_InvalidPassword(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(requestBody)

	appError := &handlers.AppError{Code: "invalid_password", Message: "Invalid credentials"}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "invalid_password", "Invalid credentials", mock.Anything, mock.Anything, nil).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_DatabaseError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	dbError := errors.New("database connection failed")
	appError := &handlers.AppError{Code: "database_error", Message: "Database error", Err: dbError}
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "database_error", "Database error", mock.Anything, mock.Anything, dbError).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}

func TestHandlerSignIn_UnknownError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)

	cfg := &HandlersAuthConfig{
		HandlersConfig:     &handlers.HandlersConfig{},
		HandlersCartConfig: &cart.HandlersCartConfig{},
		Logger:             mockHandlersConfig,
		authService:        mockAuthService,
	}

	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(requestBody)

	unknownError := errors.New("unknown error occurred")
	mockAuthService.On("SignIn", mock.Anything, SignInParams{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(nil, unknownError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signin-local", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	req := httptest.NewRequest("POST", "/signin", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerSignIn(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])
	mockAuthService.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
	mockCartConfig.AssertNotCalled(t, "MergeCart")
}
