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

func (cfg *TestHandlersAuthConfig) HandlerSignUp(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()
	params, err := auth.DecodeAndValidate[struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}](w, r)
	if err != nil {
		cfg.LogHandlerError(ctx, "signup-local", "invalid_request", "Invalid signup payload", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	result, err := cfg.GetAuthService().SignUp(ctx, SignUpParams{
		Name:     params.Name,
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		cfg.handleAuthError(w, r, err, "signup-local", ip, userAgent)
		return
	}
	cfg.MergeCart(ctx, r, result.UserID)
	auth.SetTokensAsCookies(w, result.AccessToken, result.RefreshToken, result.AccessTokenExpires, result.RefreshTokenExpires)
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, result.UserID)
	cfg.LogHandlerSuccess(ctxWithUserID, "signup-local", "Local signup success", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusCreated, handlers.HandlerResponse{Message: "Signup successful"})
}

func TestHandlerSignUp_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(requestBody)
	expectedResult := &AuthResult{
		UserID: "user123", AccessToken: "access_token_123", RefreshToken: "refresh_token_123",
		AccessTokenExpires: time.Now().Add(30 * time.Minute), RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour), IsNewUser: true,
	}
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: "password123"}).Return(expectedResult, nil)
	mockCartConfig.On("MergeCart", mock.Anything, mock.Anything, "user123").Return()
	mockHandlersConfig.On("LogHandlerSuccess", mock.Anything, "signup-local", "Local signup success", mock.Anything, mock.Anything).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Signup successful", response.Message)
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 2)
	mockAuthService.AssertExpectations(t)
	mockCartConfig.AssertExpectations(t)
	mockHandlersConfig.AssertExpectations(t)
}

func TestHandlerSignUp_InvalidRequest(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	invalidJSON := `{"name": "Test User", "email": "test@example.com", "password": "password123"`
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "invalid_request", "Invalid signup payload", mock.Anything, mock.Anything, mock.Anything).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "SignUp")
}

func TestHandlerSignUp_MissingFields(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com"}
	jsonBody, _ := json.Marshal(requestBody)
	appError := &handlers.AppError{Code: "hash_error", Message: "Error hashing password"}
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: ""}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "hash_error", "Error hashing password", mock.Anything, mock.Anything, nil).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignUp_DuplicateEmail(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(requestBody)
	appError := &handlers.AppError{Code: "email_exists", Message: "An account with this email already exists"}
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: "password123"}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "email_exists", "An account with this email already exists", mock.Anything, mock.Anything, nil).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "An account with this email already exists", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignUp_DuplicateName(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(requestBody)
	appError := &handlers.AppError{Code: "name_exists", Message: "An account with this name already exists"}
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: "password123"}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "name_exists", "An account with this name already exists", mock.Anything, mock.Anything, nil).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "An account with this name already exists", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignUp_DatabaseError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(requestBody)
	dbError := errors.New("database connection failed")
	appError := &handlers.AppError{Code: "database_error", Message: "Database error", Err: dbError}
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: "password123"}).Return(nil, appError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "database_error", "Database error", mock.Anything, mock.Anything, dbError).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestHandlerSignUp_UnknownError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockHandlersConfig := new(MockHandlersConfig)
	mockCartConfig := new(MockCartConfig)
	cfg := &TestHandlersAuthConfig{
		MockHandlersConfig: mockHandlersConfig,
		MockCartConfig:     mockCartConfig,
		authService:        mockAuthService,
	}
	requestBody := map[string]string{"name": "Test User", "email": "test@example.com", "password": "password123"}
	jsonBody, _ := json.Marshal(requestBody)
	unknownError := errors.New("unknown error occurred")
	mockAuthService.On("SignUp", mock.Anything, SignUpParams{Name: "Test User", Email: "test@example.com", Password: "password123"}).Return(nil, unknownError)
	mockHandlersConfig.On("LogHandlerError", mock.Anything, "signup-local", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()
	req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	cfg.HandlerSignUp(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])
	mockHandlersConfig.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
