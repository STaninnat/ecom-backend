// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
)

// review_helper_test.go: Mocks for review service, logger, and dependencies used in unit tests.

// Mock Service
type mockReviewMongo struct{ mock.Mock }

func (m *mockReviewMongo) CreateReview(ctx context.Context, review *models.Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}
func (m *mockReviewMongo) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	args := m.Called(ctx, reviewID)
	return args.Get(0).(*models.Review), args.Error(1)
}
func (m *mockReviewMongo) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]*models.Review), args.Error(1)
}
func (m *mockReviewMongo) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Review), args.Error(1)
}

func (m *mockReviewMongo) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	args := m.Called(ctx, reviewID, updatedReview)
	return args.Error(0)
}
func (m *mockReviewMongo) DeleteReviewByID(ctx context.Context, reviewID string) error {
	args := m.Called(ctx, reviewID)
	return args.Error(0)
}
func (m *mockReviewMongo) GetReviewsByProductIDPaginated(ctx context.Context, productID string, opts *intmongo.PaginationOptions) (*intmongo.PaginatedResult[*models.Review], error) {
	args := m.Called(ctx, productID, opts)
	return args.Get(0).(*intmongo.PaginatedResult[*models.Review]), args.Error(1)
}
func (m *mockReviewMongo) GetReviewsByUserIDPaginated(ctx context.Context, userID string, opts *intmongo.PaginationOptions) (*intmongo.PaginatedResult[*models.Review], error) {
	args := m.Called(ctx, userID, opts)
	return args.Get(0).(*intmongo.PaginatedResult[*models.Review]), args.Error(1)
}

// Mock Wrapper
type mockLoggerWrapper struct{ mock.Mock }

func (m *mockLoggerWrapper) LogHandlerError(ctx context.Context, op, code, msg, ip, ua string, err error) {
	m.Called(ctx, op, code, msg, ip, ua, err)
}
func (m *mockLoggerWrapper) LogHandlerSuccess(ctx context.Context, op, msg, ip, ua string) {
	m.Called(ctx, op, msg, ip, ua)
}

type mockReviewService struct{}

func (m *mockReviewService) CreateReview(_ context.Context, _ *models.Review) error {
	return nil
}
func (m *mockReviewService) GetReviewByID(_ context.Context, _ string) (*models.Review, error) {
	return nil, nil
}
func (m *mockReviewService) GetReviewsByProductID(_ context.Context, _ string) ([]*models.Review, error) {
	return nil, nil
}
func (m *mockReviewService) GetReviewsByUserID(_ context.Context, _ string) ([]*models.Review, error) {
	return nil, nil
}
func (m *mockReviewService) UpdateReviewByID(_ context.Context, _ string, _ *models.Review) error {
	return nil
}
func (m *mockReviewService) DeleteReviewByID(_ context.Context, _ string) error { return nil }
func (m *mockReviewService) GetReviewsByProductIDPaginated(_ context.Context, _ string, _, _ int, _, _, _ *int, _, _ *time.Time, _ *bool, _ string) (any, error) {
	return nil, nil
}
func (m *mockReviewService) GetReviewsByUserIDPaginated(_ context.Context, _ string, _, _ int, _, _, _ *int, _, _ *time.Time, _ *bool, _ string) (any, error) {
	return nil, nil
}

// Mock Create Review
type MockReviewService struct{ mock.Mock }

func (m *MockReviewService) CreateReview(ctx context.Context, review *models.Review) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}
func (m *MockReviewService) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	args := m.Called(ctx, reviewID)
	return args.Get(0).(*models.Review), args.Error(1)
}
func (m *MockReviewService) GetReviewsByProductID(ctx context.Context, productID string) ([]*models.Review, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]*models.Review), args.Error(1)
}
func (m *MockReviewService) GetReviewsByUserID(ctx context.Context, userID string) ([]*models.Review, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Review), args.Error(1)
}
func (m *MockReviewService) UpdateReviewByID(ctx context.Context, reviewID string, updatedReview *models.Review) error {
	args := m.Called(ctx, reviewID, updatedReview)
	return args.Error(0)
}
func (m *MockReviewService) DeleteReviewByID(ctx context.Context, reviewID string) error {
	args := m.Called(ctx, reviewID)
	return args.Error(0)
}
func (m *MockReviewService) GetReviewsByProductIDPaginated(ctx context.Context, productID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	args := m.Called(ctx, productID, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
	return args.Get(0), args.Error(1)
}
func (m *MockReviewService) GetReviewsByUserIDPaginated(ctx context.Context, userID string, page, pageSize int, rating, minRating, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) (any, error) {
	args := m.Called(ctx, userID, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
	return args.Get(0), args.Error(1)
}

// ... other methods omitted for brevity ...

type MockLogger struct{ mock.Mock }

func (m *MockLogger) LogHandlerError(ctx context.Context, op, code, msg, ip, ua string, err error) {
	m.Called(ctx, op, code, msg, ip, ua, err)
}
func (m *MockLogger) LogHandlerSuccess(ctx context.Context, op, msg, ip, ua string) {
	m.Called(ctx, op, msg, ip, ua)
}
