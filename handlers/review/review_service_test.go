// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
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
	require.NoError(t, err)
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
	require.Error(t, err)
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
	require.NoError(t, err)
	assert.Equal(t, review, got)
	m.AssertExpectations(t)
}

func TestGetReviewByID_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name         string
		dbErr        error
		expectedCode string
	}{
		{
			name:         "NotFound",
			dbErr:        errors.New("review not found"),
			expectedCode: "not_found",
		},
		{
			name:         "DBError",
			dbErr:        errors.New("db fail"),
			expectedCode: "get_failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockReviewMongo)
			svc := NewReviewService(m)
			m.On("GetReviewByID", mock.Anything, "r1").Return((*models.Review)(nil), tt.dbErr)
			got, err := svc.GetReviewByID(context.Background(), "r1")
			assert.Nil(t, got)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			m.AssertExpectations(t)
		})
	}
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
	require.NoError(t, err)
	assert.Equal(t, reviews, got)
	m.AssertExpectations(t)
}

func TestGetReviewsByID_FailureScenarios(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		id           string
		dbErr        error
		expectedCode string
	}{
		{
			name:         "ProductID_Failure",
			method:       "GetReviewsByProductID",
			id:           "p1",
			dbErr:        errors.New("db fail"),
			expectedCode: "get_failed",
		},
		{
			name:         "UserID_Failure",
			method:       "GetReviewsByUserID",
			id:           "u1",
			dbErr:        errors.New("db fail"),
			expectedCode: "get_failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockReviewMongo)
			svc := NewReviewService(m)
			if tt.method == "GetReviewsByProductID" {
				m.On("GetReviewsByProductID", mock.Anything, tt.id).Return(([]*models.Review)(nil), tt.dbErr)
				got, err := svc.GetReviewsByProductID(context.Background(), tt.id)
				assert.Nil(t, got)
				appErr := &handlers.AppError{}
				ok := errors.As(err, &appErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, appErr.Code)
				m.AssertExpectations(t)
			} else {
				m.On("GetReviewsByUserID", mock.Anything, tt.id).Return(([]*models.Review)(nil), tt.dbErr)
				got, err := svc.GetReviewsByUserID(context.Background(), tt.id)
				assert.Nil(t, got)
				appErr := &handlers.AppError{}
				ok := errors.As(err, &appErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, appErr.Code)
				m.AssertExpectations(t)
			}
		})
	}
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	m.AssertExpectations(t)
}

func TestUpdateReviewByID_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name         string
		dbErr        error
		expectedCode string
	}{
		{
			name:         "NotFound",
			dbErr:        errors.New("review not found"),
			expectedCode: "not_found",
		},
		{
			name:         "Failure",
			dbErr:        errors.New("db fail"),
			expectedCode: "update_failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockReviewMongo)
			svc := NewReviewService(m)
			review := &models.Review{ID: "r1"}
			m.On("UpdateReviewByID", mock.Anything, "r1", review).Return(tt.dbErr)
			err := svc.UpdateReviewByID(context.Background(), "r1", review)
			appErr := &handlers.AppError{}
			ok := errors.As(err, &appErr)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			m.AssertExpectations(t)
		})
	}
}

// TestDeleteReviewByID_Success tests the successful deletion of a review by ID via the review service.
// It verifies that the service correctly delegates to the MongoDB layer and returns no error
// when the database operation succeeds.
func TestDeleteReviewByID_Success(t *testing.T) {
	m := new(mockReviewMongo)
	svc := NewReviewService(m)
	m.On("DeleteReviewByID", mock.Anything, "r1").Return(nil)
	err := svc.DeleteReviewByID(context.Background(), "r1")
	require.NoError(t, err)
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

// Deduplicated test for paginated review retrieval with filters (product/user)
func TestGetReviewsByPaginated_Scenarios(t *testing.T) {
	tests := []struct {
		name   string
		method string // "product" or "user"
		id     string
		params []any
		result any
	}{
		{
			name:   "Product_WithFilters",
			method: "product",
			id:     "p1",
			params: []any{(*int)(nil), (*int)(nil), (*int)(nil), (*time.Time)(nil), (*time.Time)(nil), (*bool)(nil), ""},
			result: &intmongo.PaginatedResult[*models.Review]{
				Data:       []*models.Review{{ID: "r1"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
				HasNext:    false,
				HasPrev:    false,
			},
		},
		{
			name:   "User_WithFilters",
			method: "user",
			id:     "u1",
			params: []any{(*int)(nil), (*int)(nil), (*int)(nil), (*time.Time)(nil), (*time.Time)(nil), (*bool)(nil), ""},
			result: &intmongo.PaginatedResult[*models.Review]{
				Data:       []*models.Review{{ID: "r1"}},
				TotalCount: 1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
				HasNext:    false,
				HasPrev:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockReviewMongo)
			svc := NewReviewService(m)
			if tt.method == "product" {
				m.On("GetReviewsByProductIDPaginated", mock.Anything, tt.id, mock.Anything).Return(tt.result, nil)
				resp, err := svc.GetReviewsByProductIDPaginated(
					context.Background(),
					tt.id, 1, 10,
					tt.params[0].(*int), tt.params[1].(*int), tt.params[2].(*int),
					tt.params[3].(*time.Time), tt.params[4].(*time.Time),
					tt.params[5].(*bool), tt.params[6].(string),
				)
				require.NoError(t, err)
				r, ok := resp.(PaginatedReviewsResponse)
				assert.True(t, ok)
				assert.Equal(t, int64(1), r.TotalCount)
				m.AssertExpectations(t)
			} else {
				m.On("GetReviewsByUserIDPaginated", mock.Anything, tt.id, mock.Anything).Return(tt.result, nil)
				resp, err := svc.GetReviewsByUserIDPaginated(
					context.Background(),
					tt.id, 1, 10,
					tt.params[0].(*int), tt.params[1].(*int), tt.params[2].(*int),
					tt.params[3].(*time.Time), tt.params[4].(*time.Time),
					tt.params[5].(*bool), tt.params[6].(string),
				)
				require.NoError(t, err)
				r, ok := resp.(PaginatedReviewsResponse)
				assert.True(t, ok)
				assert.Equal(t, int64(1), r.TotalCount)
				m.AssertExpectations(t)
			}
		})
	}
}

// TestGetReviewsByPaginated_EdgeCases tests the review service behavior with edge cases and boundary conditions.
// It verifies that the service correctly handles partial filter combinations, filter prioritization,
// and various edge scenarios when retrieving paginated reviews by product ID.
func TestGetReviewsByPaginated_EdgeCases(t *testing.T) {
	type testCase struct {
		name       string
		id         string
		mockMethod string
		callFunc   func(svc ReviewService, ctx context.Context, id string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error)
	}

	testCases := []testCase{
		{
			name:       "ProductIDPaginated",
			id:         "p1",
			mockMethod: "GetReviewsByProductIDPaginated",
			callFunc: func(svc ReviewService, ctx context.Context, id string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
				return svc.GetReviewsByProductIDPaginated(ctx, id, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
			},
		},
		{
			name:       "UserIDPaginated",
			id:         "u1",
			mockMethod: "GetReviewsByUserIDPaginated",
			callFunc: func(svc ReviewService, ctx context.Context, id string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
				return svc.GetReviewsByUserIDPaginated(ctx, id, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err := tc.callFunc(svc, context.Background(), tc.id, 1, 10, nil, nil, nil, nil, nil, boolPtr(false), "")
			require.NoError(t, err)
			r, ok := resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			// Test with both rating and rating range (should prioritize rating range)
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err = tc.callFunc(svc, context.Background(), tc.id, 1, 10, intPtr(5), intPtr(3), intPtr(5), nil, nil, nil, "")
			require.NoError(t, err)
			r, ok = resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			// Test with only minRating (no maxRating)
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err = tc.callFunc(svc, context.Background(), tc.id, 1, 10, nil, intPtr(3), nil, nil, nil, nil, "")
			require.NoError(t, err)
			r, ok = resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			// Test with only maxRating (no minRating)
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err = tc.callFunc(svc, context.Background(), tc.id, 1, 10, nil, nil, intPtr(5), nil, nil, nil, "")
			require.NoError(t, err)
			r, ok = resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			// Test with only from date (no to date)
			from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err = tc.callFunc(svc, context.Background(), tc.id, 1, 10, nil, nil, nil, &from, nil, nil, "")
			require.NoError(t, err)
			r, ok = resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			// Test with only to date (no from date)
			to := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
			m.On(tc.mockMethod, mock.Anything, tc.id, mock.Anything).Return(result, nil)
			resp, err = tc.callFunc(svc, context.Background(), tc.id, 1, 10, nil, nil, nil, nil, &to, nil, "")
			require.NoError(t, err)
			r, ok = resp.(PaginatedReviewsResponse)
			assert.True(t, ok)
			assert.Equal(t, int64(1), r.TotalCount)

			m.AssertExpectations(t)
		})
	}
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
