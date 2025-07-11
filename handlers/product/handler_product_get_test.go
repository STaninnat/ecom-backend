package producthandlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerGetAllProducts_Success tests the successful retrieval of all products via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns products without error.
func TestHandlerGetAllProducts_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1", Role: "admin"}
	products := []database.Product{{ID: "p1"}, {ID: "p2"}}
	mockService.On("GetAllProducts", mock.Anything, true).Return(products, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "get_products", "Get all products success", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllProducts(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetAllProducts_ServiceError tests the handler's behavior when the product service returns an error during retrieval of all products.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerGetAllProducts_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := &database.User{ID: "u1", Role: "admin"}
	err := &handlers.AppError{Code: "transaction_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("GetAllProducts", mock.Anything, true).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_products", "transaction_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetAllProducts(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetProductByID_Success tests the successful retrieval of a product by ID via the handler.
// It verifies that the handler returns HTTP 200 and logs success when the service returns the product without error.
func TestHandlerGetProductByID_Success(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1", Role: "admin"}
	product := database.Product{ID: "p1"}
	mockService.On("GetProductByID", mock.Anything, "p1", true).Return(product, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "get_product_by_id", "Get products success", mock.Anything, mock.Anything).Return()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "p1")
	req := httptest.NewRequest("GET", "/products/p1", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	w := httptest.NewRecorder()

	cfg.HandlerGetProductByID(w, req, user)
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetProductByID_MissingID tests the handler's response when the product ID is missing from the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerGetProductByID_MissingID(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1", Role: "admin"}
	mockLog.On("LogHandlerError", mock.Anything, "get_product_by_id", "invalid_request", "Missing product ID", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("GET", "/products/", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetProductByID(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetProductByID_ServiceError tests the handler's behavior when the product service returns an error during retrieval by ID.
// It ensures the handler returns HTTP 404 and logs the service error correctly.
func TestHandlerGetProductByID_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	mockLog := new(mockLogger)
	cfg := &HandlersProductConfig{
		DB:             nil,
		DBConn:         nil,
		Logger:         mockLog,
		productService: mockService,
	}
	user := database.User{ID: "u1", Role: "admin"}
	err := &handlers.AppError{Code: "product_not_found", Message: "fail", Err: errors.New("fail")}
	mockService.On("GetProductByID", mock.Anything, "p1", true).Return(database.Product{}, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_product_by_id", "product_not_found", "fail", mock.Anything, mock.Anything, err.Err).Return()

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "p1")
	req := httptest.NewRequest("GET", "/products/p1", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	w := httptest.NewRecorder()

	cfg.HandlerGetProductByID(w, req, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
