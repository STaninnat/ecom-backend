package userhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetUser(ctx context.Context, user database.User) (*UserResponse, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResponse), args.Error(1)
}

func (m *mockUserService) UpdateUser(ctx context.Context, user database.User, params UpdateUserParams) error {
	args := m.Called(ctx, user, params)
	return args.Error(0)
}

type mockHandlerLogger struct {
	mock.Mock
}

func (m *mockHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

func TestHandlerGetUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	resp := &UserResponse{ID: "u1", Name: "Test"}
	mockService.On("GetUser", mock.Anything, user).Return(resp, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_user", "Get user info success", mock.Anything, mock.Anything).Return()
	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetUser(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var got UserResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, *resp, got)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerGetUser_ServiceError(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	mockService.On("GetUser", mock.Anything, user).Return(nil, errors.New("fail"))
	mockLogger.On("LogHandlerError", mock.Anything, "get_user", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetUser(w, r, user)
	assert.NotEqual(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_Success(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	mockService.On("UpdateUser", mock.Anything, user, params).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_user", "User info updated", mock.Anything, mock.Anything).Return()

	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, r, user)
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_ValidationError(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	params := map[string]string{"name": "", "email": ""}
	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "UpdateUser")
}

func TestHandlerUpdateUser_InvalidJSON(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	invalidJSON := `{"name": "A", "email": "a@b.com"` // missing closing brace

	// Expect error logging for invalid JSON
	mockLogger.On(
		"LogHandlerError",
		mock.Anything, // ctx
		"update_user",
		"invalid_request",
		"Failed to parse body",
		mock.Anything, // ip
		mock.Anything, // userAgent
		mock.Anything, // error (should be an error like unexpected EOF)
	).Return()

	r := httptest.NewRequest("PUT", "/user", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "UpdateUser")
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_ServiceError(t *testing.T) {
	mockService := new(mockUserService)
	mockLogger := new(mockHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	mockService.On("UpdateUser", mock.Anything, user, params).Return(errors.New("fail"))
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateUser(w, r, user)
	assert.NotEqual(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
