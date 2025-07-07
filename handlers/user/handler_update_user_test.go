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

type mockUpdateHandlerLogger struct {
	mock.Mock
}

func (m *mockUpdateHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockUpdateHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

func TestHandlerUpdateUser_Success(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
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

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.HandlerResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Updated user info successful", resp.Message)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_ValidationError(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	// Send invalid JSON that will fail auth.DecodeAndValidate
	invalidJSON := `{"name": "A", "email": "a@b.com", "phone": 123}` // phone should be string, not number
	body := []byte(invalidJSON)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	// Expect error logging for invalid JSON
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", "invalid_request", "Invalid update payload", mock.Anything, mock.Anything, mock.Anything).Return()

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
	mockService.AssertNotCalled(t, "UpdateUser")
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_InvalidJSON(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
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
		"Invalid update payload",
		mock.Anything, // ip
		mock.Anything, // userAgent
		mock.Anything, // error (should be an error like unexpected EOF)
	).Return()

	r := httptest.NewRequest("PUT", "/user", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
	mockService.AssertNotCalled(t, "UpdateUser")
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_ServiceError(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
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

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.NotEqual(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerUpdateUser_UserNotFoundInContext(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}

	// Mock error logging for user not found
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", "user_not_found", "User not found in context", mock.Anything, mock.Anything, mock.Anything).Return()

	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Don't set user in context - this should trigger the error path

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errorResp map[string]any
	json.NewDecoder(w.Body).Decode(&errorResp)
	assert.Contains(t, errorResp, "error")
	mockLogger.AssertExpectations(t)
}
