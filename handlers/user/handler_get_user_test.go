package userhandlers

import (
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

type mockGetHandlerLogger struct {
	mock.Mock
}

func (m *mockGetHandlerLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *mockGetHandlerLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

func TestHandlerGetUser_Success(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
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

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var got UserResponse
	json.NewDecoder(w.Body).Decode(&got)
	assert.Equal(t, *resp, got)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerGetUser_ServiceError(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
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

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.NotEqual(t, http.StatusOK, w.Code)
	// Check that we get an error response
	var errorResp map[string]any
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	assert.NoError(t, err)
	assert.Contains(t, errorResp, "error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandlerGetUser_UserNotFoundInContext(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}

	// Mock error logging for user not found
	mockLogger.On("LogHandlerError", mock.Anything, "get_user", "user_not_found", "User not found in context", mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Don't set user in context - this should trigger the error path

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errorResp map[string]any
	json.NewDecoder(w.Body).Decode(&errorResp)
	assert.Contains(t, errorResp, "error")
	mockLogger.AssertExpectations(t)
}

func TestHandlerGetUser_EmptyUserData(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{Logger: logrus.New()},
		Logger:         mockLogger,
		userService:    mockService,
	}
	user := database.User{ID: "u1"}
	// Return user with empty/null fields
	resp := &UserResponse{ID: "u1", Name: "", Email: "", Phone: "", Address: ""}
	mockService.On("GetUser", mock.Anything, user).Return(resp, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_user", "Get user info success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u1", response.ID)
	assert.Equal(t, "", response.Name)
	assert.Equal(t, "", response.Email)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
