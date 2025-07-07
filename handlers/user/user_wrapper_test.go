package userhandlers

import (
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

func TestGetUserService_InitializesIfNil(t *testing.T) {
	cfg := &HandlersUserConfig{
		HandlersConfig: &handlers.HandlersConfig{APIConfig: &config.APIConfig{}},
		userService:    nil,
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
	err := &UserError{Code: "update_failed", Message: "fail", Err: errors.New("db")}
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
	err := errors.New("unexpected")
	w := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	r := httptest.NewRequest("POST", "/", nil)
	mockLogger.On("LogHandlerError", mock.Anything, "op", "unknown_error", "Unknown error occurred", "ip", "ua", err).Return()

	cfg.handleUserError(w, r, err, "op", "ip", "ua")
	assert.Equal(t, http.StatusInternalServerError, w.status)
	assert.Contains(t, w.body, "Internal server error")
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
	err := &UserError{Code: "other_error", Message: "fail", Err: errors.New("db")}
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
