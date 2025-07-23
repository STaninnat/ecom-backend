// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"database/sql"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// upload_service_test.go: Tests for UploadService and ProductDBAdapter covering success and failure cases of image upload, update,
// validation, storage, deletion, and DB operations, including mocks and error handling.

// TestUploadServiceImpl_UploadProductImage_Success tests successful product image upload via the service.
// It verifies that a valid image is saved and the correct URL is returned.
func TestUploadServiceImpl_UploadProductImage_Success(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.jpg", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()

	// Set Content-Type header for the file
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("/tmp/uploads/test.jpg", nil)

	// Patch ParseAndGetImageFile to use the real function (since it is pure)
	ctx := context.Background()
	imageURL, err := service.UploadProductImage(ctx, "user123", req)
	assert.NoError(t, err)
	assert.Equal(t, "/static/test.jpg", imageURL)
	mockStorage.AssertExpectations(t)
}

// TestUploadServiceImpl_UploadProductImage_InvalidForm tests the service's behavior with an invalid form.
// It ensures an error is returned and the error code is "invalid_form".
func TestUploadServiceImpl_UploadProductImage_InvalidForm(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	req := httptest.NewRequest("POST", "/upload", nil) // No body
	ctx := context.Background()
	imageURL, err := service.UploadProductImage(ctx, "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_form", appErr.Code)
}

// TestUploadServiceImpl_UploadProductImage_InvalidMIME tests the service's behavior with an unsupported MIME type.
// It ensures an error is returned and the error code is "invalid_image".
func TestUploadServiceImpl_UploadProductImage_InvalidMIME(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.jpg", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "application/pdf") // Not allowed

	// Patch Save should not be called
	ctx := context.Background()
	imageURL, err := service.UploadProductImage(ctx, "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_image", appErr.Code)
}

// TestUploadServiceImpl_UploadProductImage_SaveError tests the service's behavior when file saving fails.
// It ensures an error is returned and the error code is "file_save_failed".
func TestUploadServiceImpl_UploadProductImage_SaveError(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.jpg", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	saveErr := errors.New("disk full")
	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("", saveErr)

	ctx := context.Background()
	imageURL, err := service.UploadProductImage(ctx, "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "file_save_failed", appErr.Code)
	assert.Equal(t, saveErr, appErr.Err)
	mockStorage.AssertExpectations(t)
}

// TestUploadServiceImpl_UpdateProductImage_Success tests successful product image update via the service.
// It verifies the image is saved, DB is updated, and the correct URL is returned.
func TestUploadServiceImpl_UpdateProductImage_Success(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.png", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "image/png")

	product := Product{ID: "prod123"}
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("/tmp/uploads/test.png", nil)
	mockDB.On("UpdateProductImageURL", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.NoError(t, err)
	assert.Equal(t, "/static/test.png", imageURL)
	mockDB.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestUpdateProductImage_ProductNotFound tests the service's behavior when the product is not found.
// It ensures an error is returned and the error code is "not_found".
func TestUpdateProductImage_ProductNotFound(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	mockDB.On("GetProductByID", mock.Anything, "prod404").Return(Product{}, errors.New("not found"))
	imgContent := []byte("fake image data")
	req, _ := newMultipartImageRequest(t, "image", "test.png", imgContent)
	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod404", "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "not_found", appErr.Code)
}

// TestUpdateProductImage_InvalidForm tests the service's behavior with an invalid form during update.
// It ensures an error is returned and the error code is "invalid_form".
func TestUpdateProductImage_InvalidForm(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	product := Product{ID: "prod123"}
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	// No body in request
	req := httptest.NewRequest("POST", "/update", nil)
	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_form", appErr.Code)
}

// TestUpdateProductImage_InvalidMIME tests the service's behavior with an unsupported MIME type during update.
// It ensures an error is returned and the error code is "invalid_image".
func TestUpdateProductImage_InvalidMIME(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	product := Product{ID: "prod123"}
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.png", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "application/pdf")

	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "invalid_image", appErr.Code)
}

// TestUpdateProductImage_SaveError tests the service's behavior when file saving fails during update.
// It ensures an error is returned and the error code is "file_save_failed".
func TestUpdateProductImage_SaveError(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	product := Product{ID: "prod123"}
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.png", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "image/png")

	saveErr := errors.New("disk full")
	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("", saveErr)

	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "file_save_failed", appErr.Code)
	assert.Equal(t, saveErr, appErr.Err)
	mockStorage.AssertExpectations(t)
}

// TestUpdateProductImage_DBUpdateError tests the service's behavior when the DB update fails.
// It ensures an error is returned and the error code is "db_error".
func TestUpdateProductImage_DBUpdateError(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	product := Product{ID: "prod123"}
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.png", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "image/png")

	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("/tmp/uploads/test.png", nil)
	dbErr := errors.New("db error")
	mockDB.On("UpdateProductImageURL", mock.Anything, mock.Anything).Return(dbErr)

	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.Error(t, err)
	assert.Empty(t, imageURL)
	appErr := &handlers.AppError{}
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, "db_error", appErr.Code)
	assert.Equal(t, dbErr, appErr.Err)
	mockDB.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestUpdateProductImage_DeletesOldImage tests that the service deletes the old image when updating.
// It verifies the old image is deleted if present.
func TestUpdateProductImage_DeletesOldImage(t *testing.T) {
	mockDB := new(mockProductDB)
	mockStorage := new(mockFileStorage)
	service := NewUploadService(mockDB, "/tmp/uploads", mockStorage)

	product := Product{ID: "prod123"}
	product.ImageURL.String = "/static/old.png"
	product.ImageURL.Valid = true
	mockDB.On("GetProductByID", mock.Anything, "prod123").Return(product, nil)
	imgContent := []byte("fake image data")
	req, fileHeader := newMultipartImageRequest(t, "image", "test.png", imgContent)
	file, _, _ := req.FormFile("image")
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}()
	fileHeader.Header.Set("Content-Type", "image/png")

	mockStorage.On("Delete", "/static/old.png", "/tmp/uploads").Return(nil)
	mockStorage.On("Save", mock.Anything, fileHeader, "/tmp/uploads").Return("/tmp/uploads/test.png", nil)
	mockDB.On("UpdateProductImageURL", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	imageURL, err := service.UpdateProductImage(ctx, "prod123", "user123", req)
	assert.NoError(t, err)
	assert.Equal(t, "/static/test.png", imageURL)
	mockStorage.AssertCalled(t, "Delete", "/static/old.png", "/tmp/uploads")
}

// --- ProductDBAdapter direct unit tests ---
// Use the ProductDB interface for the adapter, not *database.Queries, so we can inject mocks.

// mockQueries is a test double for database.Queries, allowing custom DB behavior in tests.
type mockQueries struct {
	getProductByIDFunc        func(ctx context.Context, id string) (database.Product, error)
	updateProductImageURLFunc func(ctx context.Context, arg database.UpdateProductImageURLParams) error
}

// GetProductByID mocks the database GetProductByID method.
func (m *mockQueries) GetProductByID(ctx context.Context, id string) (database.Product, error) {
	return m.getProductByIDFunc(ctx, id)
}

// UpdateProductImageURL mocks the database UpdateProductImageURL method.
func (m *mockQueries) UpdateProductImageURL(ctx context.Context, arg database.UpdateProductImageURLParams) error {
	return m.updateProductImageURLFunc(ctx, arg)
}

// ProductDBAdapter now accepts a ProductDB interface for testing
// Redefine a local test adapter type for the test only

// testProductDBAdapter is a test adapter for the ProductDB interface, using injected queries.
type testProductDBAdapter struct {
	Queries interface {
		GetProductByID(ctx context.Context, id string) (database.Product, error)
		UpdateProductImageURL(ctx context.Context, arg database.UpdateProductImageURLParams) error
	}
}

// GetProductByID implements ProductDBAdapter's GetProductByID for tests.
func (a *testProductDBAdapter) GetProductByID(ctx context.Context, id string) (Product, error) {
	dbProduct, err := a.Queries.GetProductByID(ctx, id)
	if err != nil {
		return Product{}, err
	}
	return Product{
		ID: dbProduct.ID,
		ImageURL: struct {
			String string
			Valid  bool
		}{
			String: dbProduct.ImageUrl.String,
			Valid:  dbProduct.ImageUrl.Valid,
		},
	}, nil
}

// UpdateProductImageURL implements ProductDBAdapter's UpdateProductImageURL for tests.
func (a *testProductDBAdapter) UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error {
	return a.Queries.UpdateProductImageURL(ctx, database.UpdateProductImageURLParams{
		ID:        params.ID,
		ImageUrl:  sql.NullString{String: params.ImageURL, Valid: true},
		UpdatedAt: time.Unix(params.UpdatedAt, 0),
	})
}

// newTestProductDBAdapter creates a new testProductDBAdapter with the given queries.
func newTestProductDBAdapter(queries interface {
	GetProductByID(ctx context.Context, id string) (database.Product, error)
	UpdateProductImageURL(ctx context.Context, arg database.UpdateProductImageURLParams) error
}) *testProductDBAdapter {
	return &testProductDBAdapter{Queries: queries}
}

// TestNewProductDBAdapter tests the construction of a new ProductDBAdapter.
// It verifies the adapter is created with the provided queries.
func TestNewProductDBAdapter(t *testing.T) {
	q := &database.Queries{}
	a := NewProductDBAdapter(q)
	assert.NotNil(t, a)
}

// TestProductDBAdapter_GetProductByID_Success tests ProductDBAdapter's GetProductByID for the success case.
// It verifies the returned Product is mapped correctly from the DB.
func TestProductDBAdapter_GetProductByID_Success(t *testing.T) {
	fakeProduct := database.Product{
		ID:       "p1",
		ImageUrl: sql.NullString{String: "img.png", Valid: true},
	}
	q := &mockQueries{
		getProductByIDFunc: func(_ context.Context, id string) (database.Product, error) {
			assert.Equal(t, "p1", id)
			return fakeProduct, nil
		},
	}
	a := newTestProductDBAdapter(q)
	prod, err := a.GetProductByID(context.Background(), "p1")
	assert.NoError(t, err)
	assert.Equal(t, "p1", prod.ID)
	assert.Equal(t, "img.png", prod.ImageURL.String)
	assert.True(t, prod.ImageURL.Valid)
}

// TestProductDBAdapter_GetProductByID_Error tests ProductDBAdapter's GetProductByID for the error case.
// It ensures the error is propagated.
func TestProductDBAdapter_GetProductByID_Error(t *testing.T) {
	dbErr := errors.New("not found")
	q := &mockQueries{
		getProductByIDFunc: func(_ context.Context, _ string) (database.Product, error) {
			return database.Product{}, dbErr
		},
	}
	a := newTestProductDBAdapter(q)
	prod, err := a.GetProductByID(context.Background(), "badid")
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	assert.Empty(t, prod.ID)
}

// TestProductDBAdapter_UpdateProductImageURL_Success tests ProductDBAdapter's UpdateProductImageURL for the success case.
// It verifies the DB method is called and no error is returned.
func TestProductDBAdapter_UpdateProductImageURL_Success(t *testing.T) {
	q := &mockQueries{
		updateProductImageURLFunc: func(_ context.Context, arg database.UpdateProductImageURLParams) error {
			assert.Equal(t, "p1", arg.ID)
			assert.Equal(t, "img.png", arg.ImageUrl.String)
			assert.True(t, arg.ImageUrl.Valid)
			return nil
		},
	}
	a := newTestProductDBAdapter(q)
	params := UpdateProductImageURLParams{ID: "p1", ImageURL: "img.png", UpdatedAt: 1234567890}
	err := a.UpdateProductImageURL(context.Background(), params)
	assert.NoError(t, err)
}

// TestProductDBAdapter_UpdateProductImageURL_Error tests ProductDBAdapter's UpdateProductImageURL for the error case.
// It ensures the error is propagated.
func TestProductDBAdapter_UpdateProductImageURL_Error(t *testing.T) {
	dbErr := errors.New("db fail")
	q := &mockQueries{
		updateProductImageURLFunc: func(_ context.Context, _ database.UpdateProductImageURLParams) error {
			return dbErr
		},
	}
	a := newTestProductDBAdapter(q)
	params := UpdateProductImageURLParams{ID: "p1"}
	err := a.UpdateProductImageURL(context.Background(), params)
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
}
