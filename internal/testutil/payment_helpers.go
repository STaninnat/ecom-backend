package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// RunHandlerMissingOrderIDTest is a generic helper for testing missing order_id param in payment handlers.
func RunHandlerMissingOrderIDTest(
	t *testing.T,
	handler func(cfg any, w http.ResponseWriter, r *http.Request, user any),
	cfg any,
	user any,
	method, url, logAction string,
	mockLogger *mock.Mock,
) {
	mockLogger.On("LogHandlerError", mock.Anything, logAction, "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest(method, url, nil)
	rctx := chi.NewRouteContext() // no order_id param
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler(cfg, w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}
