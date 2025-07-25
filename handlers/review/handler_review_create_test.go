// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
)

// handler_review_create_test.go: Tests for review creation handler and request validation logic.

// TestHandlerCreateReview_Success tests the successful creation of a review via the handler.
// It verifies that the handler returns HTTP 201, the correct response message, and success code when the service succeeds.
func TestHandlerCreateReview_Success(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reqBody := ReviewCreateRequest{
		ProductID: "p1",
		Rating:    5,
		Comment:   "Great!",
		MediaURLs: []string{"url1"},
	}
	jsonBody, _ := json.Marshal(reqBody)
	review := &models.Review{
		UserID:    user.ID,
		ProductID: reqBody.ProductID,
		Rating:    reqBody.Rating,
		Comment:   reqBody.Comment,
		MediaURLs: reqBody.MediaURLs,
	}
	mockService.On("CreateReview", mock.Anything, review).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "create_review", "Review created successfully", mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateReview(w, req, user)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp handlers.APIResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Review created successfully", resp.Message)
	assert.Equal(t, "success", resp.Code)
	assert.NotNil(t, resp.Data)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerCreateReview_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerCreateReview_InvalidPayload(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLogger.On("LogHandlerError", mock.Anything, "create_review", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateReview(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerCreateReview_ValidationError tests the handler's response when the review request fails validation.
// It verifies that the handler returns HTTP 400 with the appropriate validation error message.
func TestHandlerCreateReview_ValidationError(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	// Missing ProductID
	reqBody := ReviewCreateRequest{
		ProductID: "",
		Rating:    5,
		Comment:   "Great!",
	}
	jsonBody, _ := json.Marshal(reqBody)
	mockLogger.On("LogHandlerError", mock.Anything, "create_review", "invalid_request", "Product ID is required", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateReview(w, req, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLogger.AssertExpectations(t)
}

// TestHandlerCreateReview_ServiceError tests the handler's behavior when the review service returns an error.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerCreateReview_ServiceError(t *testing.T) {
	mockService := new(MockReviewService)
	mockLogger := new(MockLogger)
	cfg := &HandlersReviewConfig{
		Config:        &handlers.Config{},
		Logger:        mockLogger,
		ReviewService: mockService,
	}
	user := database.User{ID: "u1"}
	reqBody := ReviewCreateRequest{
		ProductID: "p1",
		Rating:    5,
		Comment:   "Great!",
	}
	jsonBody, _ := json.Marshal(reqBody)
	review := &models.Review{
		UserID:    user.ID,
		ProductID: reqBody.ProductID,
		Rating:    reqBody.Rating,
		Comment:   reqBody.Comment,
	}
	err := &handlers.AppError{Code: "internal_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("CreateReview", mock.Anything, review).Return(err)
	mockLogger.On("LogHandlerError", mock.Anything, "create_review", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	req := httptest.NewRequest("POST", "/reviews", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreateReview(w, req, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestReviewCreateRequest_Validate_EdgeCases tests the validation logic for review creation requests with various edge cases.
// It covers rating boundaries (0, 6), empty comments, and valid boundary conditions to ensure robust validation.
func TestReviewCreateRequest_Validate_EdgeCases(t *testing.T) {
	cases := []struct {
		name    string
		request ReviewCreateRequest
		wantErr bool
		errCode string
	}{
		{
			name:    "rating too low",
			request: ReviewCreateRequest{ProductID: "p1", Rating: 0, Comment: "test"},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "rating too high",
			request: ReviewCreateRequest{ProductID: "p1", Rating: 6, Comment: "test"},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "empty comment",
			request: ReviewCreateRequest{ProductID: "p1", Rating: 5, Comment: ""},
			wantErr: true,
			errCode: "invalid_request",
		},
		{
			name:    "valid rating boundaries",
			request: ReviewCreateRequest{ProductID: "p1", Rating: 1, Comment: "test"},
			wantErr: false,
		},
		{
			name:    "valid rating boundaries max",
			request: ReviewCreateRequest{ProductID: "p1", Rating: 5, Comment: "test"},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				require.Error(t, err)
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
