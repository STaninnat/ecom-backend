package reviewhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// makeDeleteRequestWithID creates a DELETE HTTP request with the specified review ID in the URL path.
// It sets up the chi router context with the review ID parameter for testing the delete handler.
//
// Parameters:
//   - id: string representing the review ID to be included in the request URL
//
// Returns:
//   - *http.Request: configured DELETE request with the review ID in the URL parameters
func makeDeleteRequestWithID(id string) *http.Request {
	r := httptest.NewRequest("DELETE", "/reviews/"+id, nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

// TestHandlerDeleteReviewByID_Success tests the successful deletion of a review via the handler.
// It verifies that the handler returns HTTP 200, the correct response message, and properly logs the success event
// when the review service successfully deletes the review.
func TestHandlerDeleteReviewByID_Success(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLogger,
		ReviewService:  mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := "r1"
	review := &models.Review{ID: reviewID, UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(review, nil)
	mockService.On("DeleteReviewByID", mock.Anything, reviewID).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "delete_review_by_id", "Review deleted successfully", mock.Anything, mock.Anything).Return()

	r := makeDeleteRequestWithID(reviewID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteReviewByID(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.APIResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "success", resp.Code)
	assert.Equal(t, "Review deleted successfully", resp.Message)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteReviewByID_MissingID tests the handler's response when no review ID is provided in the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error for missing review ID.
func TestHandlerDeleteReviewByID_MissingID(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLogger,
		ReviewService:  mockService,
	}
	user := database.User{ID: "u1"}
	mockLogger.On("LogHandlerError", mock.Anything, "delete_review_by_id", "invalid_request", "Review ID is required", mock.Anything, mock.Anything, nil).Return()

	r := makeDeleteRequestWithID("")
	w := httptest.NewRecorder()

	cfg.HandlerDeleteReviewByID(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteReviewByID_ReviewNotFound tests the handler's behavior when the review service cannot find the specified review.
// It ensures the handler returns HTTP 404 and logs the appropriate error when the review does not exist.
func TestHandlerDeleteReviewByID_ReviewNotFound(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLogger,
		ReviewService:  mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := "r1"
	err := &handlers.AppError{Code: "not_found", Message: "Review not found"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return((*models.Review)(nil), err)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_review_by_id", "not_found", "Review not found", mock.Anything, mock.Anything, err.Err).Return()

	r := makeDeleteRequestWithID(reviewID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteReviewByID(w, r, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteReviewByID_Unauthorized tests the handler's response when a user attempts to delete a review they don't own.
// It verifies that the handler returns HTTP 403 and logs the appropriate error for unauthorized access.
func TestHandlerDeleteReviewByID_Unauthorized(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLogger,
		ReviewService:  mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := "r1"
	review := &models.Review{ID: reviewID, UserID: "other", ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(review, nil)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_review_by_id", "unauthorized", "You can only delete your own reviews", mock.Anything, mock.Anything, nil).Return()

	r := makeDeleteRequestWithID(reviewID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteReviewByID(w, r, user)
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerDeleteReviewByID_ServiceError tests the handler's behavior when the review service encounters an error during deletion.
// It ensures the handler returns HTTP 500 and logs the service error correctly when the deletion operation fails.
func TestHandlerDeleteReviewByID_ServiceError(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLogger,
		ReviewService:  mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := "r1"
	review := &models.Review{ID: reviewID, UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(review, nil)
	err := &handlers.AppError{Code: "internal_error", Message: "fail"}
	mockService.On("DeleteReviewByID", mock.Anything, reviewID).Return(err)
	mockLogger.On("LogHandlerError", mock.Anything, "delete_review_by_id", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := makeDeleteRequestWithID(reviewID)
	w := httptest.NewRecorder()

	cfg.HandlerDeleteReviewByID(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
