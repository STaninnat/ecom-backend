// Package categoryhandlers provides HTTP handlers and services for managing product categories.
package categoryhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// category_helper_test.go: Implements category CRUD HTTP handlers with supporting mocks for unit and integration testing.

// MockHandlersConfig is a mock implementation of handlers.HandlerLogger for testing
// and can be embedded in test configs for handler tests.
type MockHandlersConfig struct {
	mock.Mock
}

func (m *MockHandlersConfig) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

func (m *MockHandlersConfig) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

// TestHandlersCategoryConfig is a test configuration that embeds the mock
// and provides the GetCategoryService method for handler tests.
type TestHandlersCategoryConfig struct {
	*MockHandlersConfig
	Logger          handlers.HandlerLogger
	categoryService CategoryService
}

func (cfg *TestHandlersCategoryConfig) GetCategoryService() CategoryService {
	return cfg.categoryService
}

// HandlerCreateCategory handles category creation requests
func (cfg *TestHandlersCategoryConfig) HandlerCreateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryRequest(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		cfg.handleCategoryError,
		"create_category",
		func(ctx context.Context, service CategoryService, params CategoryRequest) (string, error) {
			return service.CreateCategory(ctx, params)
		},
		"Category created successfully",
		http.StatusCreated,
	)
}

// HandlerUpdateCategory handles category update requests
func (cfg *TestHandlersCategoryConfig) HandlerUpdateCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryRequest(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		cfg.handleCategoryError,
		"update_category",
		func(ctx context.Context, service CategoryService, params CategoryRequest) (string, error) {
			return "", service.UpdateCategory(ctx, params)
		},
		"Category updated successfully",
		http.StatusOK,
	)
}

// HandlerDeleteCategory handles category deletion requests
func (cfg *TestHandlersCategoryConfig) HandlerDeleteCategory(w http.ResponseWriter, r *http.Request, user database.User) {
	HandleCategoryDelete(
		w, r, user,
		cfg.Logger,
		cfg.GetCategoryService,
		SharedHandleCategoryError,
	)
}

// HandlerGetAllCategories handles category retrieval requests
func (cfg *TestHandlersCategoryConfig) HandlerGetAllCategories(w http.ResponseWriter, r *http.Request, _ *database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	categoryService := cfg.GetCategoryService()
	categories, err := categoryService.GetAllCategories(ctx)
	if err != nil {
		cfg.handleCategoryError(w, r, err, "get_all_categories", ip, userAgent)
		return
	}

	// Convert database structs to response format with lowercase field names
	type CategoryResponse struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description,omitempty"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	responseCategories := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		responseCategories[i] = CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description.String,
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		}
	}

	err = json.NewEncoder(w).Encode(responseCategories)
	if err != nil {
		println("json.Encode failed:", err.Error())
	}
}

// handleCategoryError handles category-specific errors with proper logging and responses
func (cfg *TestHandlersCategoryConfig) handleCategoryError(w http.ResponseWriter, r *http.Request, err error, operation, ip, userAgent string) {
	SharedHandleCategoryError(cfg.Logger, w, r, err, operation, ip, userAgent)
}

// Mock implementations for testing
type MockCategoryDBQueries struct {
	mock.Mock
}

func (m *MockCategoryDBQueries) WithTx(tx CategoryDBTx) CategoryDBQueries {
	args := m.Called(tx)
	return args.Get(0).(CategoryDBQueries)
}

func (m *MockCategoryDBQueries) CreateCategory(ctx context.Context, params database.CreateCategoryParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCategoryDBQueries) UpdateCategories(ctx context.Context, params database.UpdateCategoriesParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCategoryDBQueries) DeleteCategory(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryDBQueries) GetAllCategories(ctx context.Context) ([]database.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Category), args.Error(1)
}

type MockCategoryDBConn struct {
	mock.Mock
}

func (m *MockCategoryDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (CategoryDBTx, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(CategoryDBTx), args.Error(1)
}

type MockCategoryDBTx struct {
	mock.Mock
}

func (m *MockCategoryDBTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCategoryDBTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// Mock logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}

func (m *MockLogger) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}

// Mock category service for testing
type MockCategoryService struct {
	mock.Mock
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, params CategoryRequest) (string, error) {
	args := m.Called(ctx, params)
	return args.String(0), args.Error(1)
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, params CategoryRequest) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	args := m.Called(ctx, categoryID)
	return args.Error(0)
}

func (m *MockCategoryService) GetAllCategories(ctx context.Context) ([]database.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Category), args.Error(1)
}

// MockCategoryService for integration tests
type MockCategoryServiceForGetIntegration struct {
	mock.Mock
}

func (m *MockCategoryServiceForGetIntegration) CreateCategory(ctx context.Context, params CategoryRequest) (string, error) {
	args := m.Called(ctx, params)
	return args.String(0), args.Error(1)
}
func (m *MockCategoryServiceForGetIntegration) UpdateCategory(ctx context.Context, params CategoryRequest) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}
func (m *MockCategoryServiceForGetIntegration) DeleteCategory(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockCategoryServiceForGetIntegration) GetAllCategories(ctx context.Context) ([]database.Category, error) {
	println("DEBUG: mock called with ctx:", ctx)
	args := m.Called(ctx)
	err := args.Error(1)
	if err != nil {
		println("DEBUG: mock returning error:", err.Error())
	}
	return args.Get(0).([]database.Category), err
}

// MockLogger for integration tests
type MockLoggerForGetIntegration struct {
	mock.Mock
}

func (m *MockLoggerForGetIntegration) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	m.Called(ctx, action, details, ip, ua)
}
func (m *MockLoggerForGetIntegration) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	m.Called(ctx, action, details, logMsg, ip, ua, err)
}
