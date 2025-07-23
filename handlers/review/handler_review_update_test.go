// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

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
	"github.com/STaninnat/ecom-backend/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_review_update_test.go: Tests for HandlerUpdateReviewByID covering success, error cases, validation, authorization, and service behavior.

// makeUpdateRequestWithID creates a PUT HTTP request for updating a review by ID.
// It sets up the chi router context with the review ID parameter and includes the provided request body
// for testing the update review handler.
//
// Parameters:
//   - id: string representing the review ID to be included in the request URL
//   - body: []byte containing the JSON request body for the update operation
//
// Returns:
//   - *http.Request: configured PUT request with the review ID in the URL parameters and the provided body
func makeUpdateRequestWithID(id string, body []byte) *http.Request {
	r := httptest.NewRequest("PUT", "/reviews/"+id, bytes.NewBuffer(body))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

// TestHandlerUpdateReviewByID_Success tests the successful update of a review via the handler.
// It verifies that the handler returns HTTP 200, the correct response message, and properly logs
// the success event when the review service successfully updates the review.
func TestHandlerUpdateReviewByID_Success(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	existing := &models.Review{UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(existing, nil)
	updateReq := ReviewUpdateRequest{Rating: 4, Comment: "Updated!", MediaURLs: []string{"url1"}}
	jsonBody, _ := json.Marshal(updateReq)
	update := &models.Review{UserID: user.ID, ProductID: existing.ProductID, Rating: updateReq.Rating, Comment: updateReq.Comment, MediaURLs: updateReq.MediaURLs}
	mockService.On("UpdateReviewByID", mock.Anything, reviewID, update).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_review_by_id", "Review updated successfully", mock.Anything, mock.Anything).Return()

	r := makeUpdateRequestWithID(reviewID, jsonBody)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "Review updated successfully", resp.Message)
	assert.Equal(t, "success", resp.Code)
	assert.NotNil(t, resp.Data)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_MissingID tests the handler's response when no review ID is provided in the request.
// It checks that the handler returns HTTP 400 and logs the appropriate error for missing review ID.
func TestHandlerUpdateReviewByID_MissingID(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "invalid_request", "Review ID is required", mock.Anything, mock.Anything, nil).Return()

	r := makeUpdateRequestWithID("", []byte(`{}`))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_ReviewNotFound tests the handler's behavior when the review service cannot find the specified review.
// It ensures the handler returns HTTP 404 and logs the appropriate error when the review does not exist.
func TestHandlerUpdateReviewByID_ReviewNotFound(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	err := &handlers.AppError{Code: "not_found", Message: "Review not found"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return((*models.Review)(nil), err)
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "not_found", "Review not found", mock.Anything, mock.Anything, err.Err).Return()

	r := makeUpdateRequestWithID(reviewID, []byte(`{}`))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_Unauthorized tests the handler's response when a user attempts to update a review they don't own.
// It verifies that the handler returns HTTP 403 and logs the appropriate error for unauthorized access.
func TestHandlerUpdateReviewByID_Unauthorized(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	existing := &models.Review{UserID: "other", ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(existing, nil)
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "unauthorized", "You can only update your own reviews", mock.Anything, mock.Anything, nil).Return()

	r := makeUpdateRequestWithID(reviewID, []byte(`{}`))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error for malformed JSON.
func TestHandlerUpdateReviewByID_InvalidPayload(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	existing := &models.Review{UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(existing, nil)
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	r := makeUpdateRequestWithID(reviewID, []byte(`{"bad":}`))
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_ValidationError tests the handler's behavior when the update request fails validation.
// It ensures the handler returns HTTP 400 and logs the appropriate error when the request data is invalid.
func TestHandlerUpdateReviewByID_ValidationError(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	existing := &models.Review{UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(existing, nil)
	updateReq := ReviewUpdateRequest{Rating: 0, Comment: ""} // Invalid
	jsonBody, _ := json.Marshal(updateReq)
	err := &handlers.AppError{}
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "invalid_request", "Rating must be between 1 and 5", mock.Anything, mock.Anything, err.Err).Return()

	r := makeUpdateRequestWithID(reviewID, jsonBody)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateReviewByID_ServiceError tests the handler's behavior when the review service encounters an error during update.
// It ensures the handler returns HTTP 500 and logs the service error correctly when the update operation fails.
func TestHandlerUpdateReviewByID_ServiceError(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reviewID := testReviewID
	existing := &models.Review{UserID: user.ID, ProductID: "p1"}
	mockService.On("GetReviewByID", mock.Anything, reviewID).Return(existing, nil)
	updateReq := ReviewUpdateRequest{Rating: 4, Comment: "Updated!"}
	jsonBody, _ := json.Marshal(updateReq)
	update := &models.Review{UserID: user.ID, ProductID: existing.ProductID, Rating: updateReq.Rating, Comment: updateReq.Comment}
	err := &handlers.AppError{Code: "internal_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("UpdateReviewByID", mock.Anything, reviewID, update).Return(err)
	mockLogger.On("LogHandlerError", mock.Anything, "update_review_by_id", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := makeUpdateRequestWithID(reviewID, jsonBody)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateReviewByID(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestReviewUpdateRequest_Validate_EdgeCases tests the ReviewUpdateRequest.Validate method with various edge cases and boundary conditions.
// It verifies that the validation function correctly handles invalid ratings, empty comments, and valid boundary values,
// ensuring robust validation of update request data.
func TestReviewUpdateRequest_Validate_EdgeCases(t *testing.T) {
	cases := []struct {
		name    string
		request ReviewUpdateRequest
		wantErr bool
		errCode string
	}{
		{
			name:    "rating too low",
			request: ReviewUpdateRequest{Rating: 0, Comment: "test"},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "rating too high",
			request: ReviewUpdateRequest{Rating: 6, Comment: "test"},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "empty comment",
			request: ReviewUpdateRequest{Rating: 5, Comment: ""},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "valid rating boundaries min",
			request: ReviewUpdateRequest{Rating: 1, Comment: "test"},
			wantErr: false,
		},
		{
			name:    "valid rating boundaries max",
			request: ReviewUpdateRequest{Rating: 5, Comment: "test"},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				appErr := &handlers.AppError{}
				ok := errors.As(err, &appErr)
				assert.True(t, ok)
				assert.Equal(t, tc.errCode, appErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
