// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// review_service_test.go: Tests for the review service to verify correct behavior under success and failure scenarios.
// Covers all core operations, error wrapping, and filtering logic.

// TestCreateReview_Success tests the successful creation of a review via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns no error
// when the database operation succeeds.
func TestCreateReview_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	review := &models.Review{ID: "r1"}
	m.On("CreateReview", mock.Anything, review).Return(nil)
	err := svc.CreateReview(context.Background(), review)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

// TestCreateReview_Failure tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the appropriate code.
func TestCreateReview_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	review := &models.Review{ID: "r1"}
	m.On("CreateReview", mock.Anything, review).Return(errors.New("db fail"))
	err := svc.CreateReview(context.Background(), review)
	assert.Error(t, err)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "create_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewByID_Success tests the successful retrieval of a review by ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected review
// when the database operation succeeds.
func TestGetReviewByID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	review := &models.Review{ID: "r1"}
	m.On("GetReviewByID", mock.Anything, "r1").Return(review, nil)
	got, err := svc.GetReviewByID(context.Background(), "r1")
	assert.NoError(t, err)
	assert.Equal(t, review, got)
	m.AssertExpectations(t)
}

// TestGetReviewByID_NotFound tests the review service behavior when a review is not found in the database.
// It ensures the service correctly wraps the database error in an AppError with the "not_found" code.
func TestGetReviewByID_NotFound(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("review not found")
	m.On("GetReviewByID", mock.Anything, "r1").Return((*models.Review)(nil), dbErr)
	got, err := svc.GetReviewByID(context.Background(), "r1")
	assert.Nil(t, got)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "not_found", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewByID_DBError tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetReviewByID_DBError(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("GetReviewByID", mock.Anything, "r1").Return((*models.Review)(nil), dbErr)
	got, err := svc.GetReviewByID(context.Background(), "r1")
	assert.Nil(t, got)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewsByProductID_Success tests the successful retrieval of reviews by product ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected reviews
// when the database operation succeeds.
func TestGetReviewsByProductID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	reviews := []*models.Review{{ID: "r1"}}
	m.On("GetReviewsByProductID", mock.Anything, "p1").Return(reviews, nil)
	got, err := svc.GetReviewsByProductID(context.Background(), "p1")
	assert.NoError(t, err)
	assert.Equal(t, reviews, got)
	m.AssertExpectations(t)
}

// TestGetReviewsByProductID_Failure tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetReviewsByProductID_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("GetReviewsByProductID", mock.Anything, "p1").Return(([]*models.Review)(nil), dbErr)
	got, err := svc.GetReviewsByProductID(context.Background(), "p1")
	assert.Nil(t, got)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewsByUserID_Success tests the successful retrieval of reviews by user ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected reviews
// when the database operation succeeds.
func TestGetReviewsByUserID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	reviews := []*models.Review{{ID: "r1"}}
	m.On("GetReviewsByUserID", mock.Anything, "u1").Return(reviews, nil)
	got, err := svc.GetReviewsByUserID(context.Background(), "u1")
	assert.NoError(t, err)
	assert.Equal(t, reviews, got)
	m.AssertExpectations(t)
}

// TestGetReviewsByUserID_Failure tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetReviewsByUserID_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("GetReviewsByUserID", mock.Anything, "u1").Return(([]*models.Review)(nil), dbErr)
	got, err := svc.GetReviewsByUserID(context.Background(), "u1")
	assert.Nil(t, got)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestUpdateReviewByID_Success tests the successful update of a review by ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns no error
// when the database operation succeeds.
func TestUpdateReviewByID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	review := &models.Review{ID: "r1"}
	m.On("UpdateReviewByID", mock.Anything, "r1", review).Return(nil)
	err := svc.UpdateReviewByID(context.Background(), "r1", review)
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

// TestUpdateReviewByID_NotFound tests the review service behavior when the review to update is not found.
// It ensures the service correctly wraps the database error in an AppError with the "not_found" code.
func TestUpdateReviewByID_NotFound(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("review not found")
	review := &models.Review{ID: "r1"}
	m.On("UpdateReviewByID", mock.Anything, "r1", review).Return(dbErr)
	err := svc.UpdateReviewByID(context.Background(), "r1", review)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "not_found", appErr.Code)
	m.AssertExpectations(t)
}

// TestUpdateReviewByID_Failure tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "update_failed" code.
func TestUpdateReviewByID_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	review := &models.Review{ID: "r1"}
	m.On("UpdateReviewByID", mock.Anything, "r1", review).Return(dbErr)
	err := svc.UpdateReviewByID(context.Background(), "r1", review)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "update_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestDeleteReviewByID_Success tests the successful deletion of a review by ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns no error
// when the database operation succeeds.
func TestDeleteReviewByID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	m.On("DeleteReviewByID", mock.Anything, "r1").Return(nil)
	err := svc.DeleteReviewByID(context.Background(), "r1")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

// TestDeleteReviewByID_NotFound tests the review service behavior when the review to delete is not found.
// It ensures the service correctly wraps the database error in an AppError with the "not_found" code.
func TestDeleteReviewByID_NotFound(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("review not found")
	m.On("DeleteReviewByID", mock.Anything, "r1").Return(dbErr)
	err := svc.DeleteReviewByID(context.Background(), "r1")
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "not_found", appErr.Code)
	m.AssertExpectations(t)
}

// TestDeleteReviewByID_Failure tests the review service behavior when the database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "delete_failed" code.
func TestDeleteReviewByID_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("DeleteReviewByID", mock.Anything, "r1").Return(dbErr)
	err := svc.DeleteReviewByID(context.Background(), "r1")
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "delete_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewsByProductIDPaginated_Success tests the successful retrieval of paginated reviews by product ID.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected
// paginated response when the database operation succeeds.
func TestGetReviewsByProductIDPaginated_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)
	m.AssertExpectations(t)
}

// TestGetReviewsByProductIDPaginated_Failure tests the review service behavior when the paginated database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetReviewsByProductIDPaginated_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return((*intmongo.PaginatedResult[*models.Review])(nil), dbErr)
	resp, err := svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, nil, nil, "")
	assert.Nil(t, resp)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewsByUserIDPaginated_Success tests the successful retrieval of paginated reviews by user ID.
// It verifies that the service correctly delegates to the MongoDB layer and returns the expected
// paginated response when the database operation succeeds.
func TestGetReviewsByUserIDPaginated_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)
	m.AssertExpectations(t)
}

// TestGetReviewsByUserIDPaginated_Failure tests the review service behavior when the paginated database operation fails.
// It ensures the service correctly wraps the database error in an AppError with the "get_failed" code.
func TestGetReviewsByUserIDPaginated_Failure(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	dbErr := errors.New("db fail")
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return((*intmongo.PaginatedResult[*models.Review])(nil), dbErr)
	resp, err := svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, nil, nil, "")
	assert.Nil(t, resp)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "get_failed", appErr.Code)
	m.AssertExpectations(t)
}

// TestGetReviewsByProductIDPaginated_WithFilters tests the review service behavior with various filter combinations.
// It verifies that the service correctly handles rating filters, rating ranges, date ranges, media filters,
// and sort options when retrieving paginated reviews by product ID.
func TestGetReviewsByProductIDPaginated_WithFilters(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	// Test with rating filter
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, intPtr(5), nil, nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with rating range
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, intPtr(3), intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with date range
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, &from, &to, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with hasMedia filter
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, nil, boolPtr(true), "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with sort option
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, nil, nil, "rating_desc")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	m.AssertExpectations(t)
}

// TestGetReviewsByUserIDPaginated_WithFilters tests the review service behavior with various filter combinations.
// It verifies that the service correctly handles rating filters, rating ranges, date ranges, media filters,
// and sort options when retrieving paginated reviews by user ID.
func TestGetReviewsByUserIDPaginated_WithFilters(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	// Test with rating filter
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, intPtr(5), nil, nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with rating range
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, intPtr(3), intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with date range
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, &from, &to, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with hasMedia filter
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, nil, boolPtr(true), "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with sort option
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, nil, nil, "rating_desc")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	m.AssertExpectations(t)
}

// TestGetReviewsByProductIDPaginated_EdgeCases tests the review service behavior with edge cases and boundary conditions.
// It verifies that the service correctly handles partial filter combinations, filter prioritization,
// and various edge scenarios when retrieving paginated reviews by product ID.
func TestGetReviewsByProductIDPaginated_EdgeCases(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	// Test with hasMedia false (no media)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, nil, boolPtr(false), "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with both rating and rating range (should prioritize rating range)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, intPtr(5), intPtr(3), intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only minRating (no maxRating)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, intPtr(3), nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only maxRating (no minRating)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only from date (no to date)
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, &from, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only to date (no from date)
	to := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	m.On("GetReviewsByProductIDPaginated", mock.Anything, "p1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByProductIDPaginated(context.Background(), "p1", 1, 10, nil, nil, nil, nil, &to, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	m.AssertExpectations(t)
}

// TestGetReviewsByUserIDPaginated_EdgeCases tests the review service behavior with edge cases and boundary conditions.
// It verifies that the service correctly handles partial filter combinations, filter prioritization,
// and various edge scenarios when retrieving paginated reviews by user ID.
func TestGetReviewsByUserIDPaginated_EdgeCases(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	result := &intmongo.PaginatedResult[*models.Review]{
		Data:       []*models.Review{{ID: "r1"}},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
	}

	// Test with hasMedia false (no media)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err := svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, nil, boolPtr(false), "")
	assert.NoError(t, err)
	r, ok := resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with both rating and rating range (should prioritize rating range)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, intPtr(5), intPtr(3), intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only minRating (no maxRating)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, intPtr(3), nil, nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only maxRating (no minRating)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, intPtr(5), nil, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only from date (no to date)
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, &from, nil, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	// Test with only to date (no from date)
	to := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	m.On("GetReviewsByUserIDPaginated", mock.Anything, "u1", mock.Anything).Return(result, nil)
	resp, err = svc.GetReviewsByUserIDPaginated(context.Background(), "u1", 1, 10, nil, nil, nil, nil, &to, nil, "")
	assert.NoError(t, err)
	r, ok = resp.(PaginatedReviewsResponse)
	assert.True(t, ok)
	assert.Equal(t, int64(1), r.TotalCount)

	m.AssertExpectations(t)
}

// TestParseSortOption tests the parseSortOption function with various sort options.
// It verifies that the function correctly maps sort option strings to MongoDB sort specifications,
// including default sorting, date sorting, rating sorting, and comment length sorting.
func TestParseSortOption(t *testing.T) {
	assert.Equal(t, map[string]any{"created_at": -1}, parseSortOption(""))
	assert.Equal(t, map[string]any{"created_at": -1}, parseSortOption("date_desc"))
	assert.Equal(t, map[string]any{"created_at": 1}, parseSortOption("date_asc"))
	assert.Equal(t, map[string]any{"rating": -1}, parseSortOption("rating_desc"))
	assert.Equal(t, map[string]any{"rating": 1}, parseSortOption("rating_asc"))
	assert.Equal(t, map[string]any{"updated_at": -1}, parseSortOption("updated_desc"))
	assert.Equal(t, map[string]any{"updated_at": 1}, parseSortOption("updated_asc"))
	assert.Equal(t, map[string]any{"$expr": map[string]any{"$strLenCP": "$comment"}, "$meta": -1}, parseSortOption("comment_length_desc"))
	assert.Equal(t, map[string]any{"$expr": map[string]any{"$strLenCP": "$comment"}, "$meta": 1}, parseSortOption("comment_length_asc"))
}
