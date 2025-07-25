// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
)

// upload_service.go: Defines upload service interface, database adapter, and implementation for handling product image uploads and updates,
// including file validation, storage, and DB updates with error handling.

// UploadService defines the business logic interface for uploads (local or S3).
// Provides methods to upload and update product images with validation and error handling.
type UploadService interface {
	UploadProductImage(ctx context.Context, userID string, r *http.Request) (string, error)
	UpdateProductImage(ctx context.Context, productID string, userID string, r *http.Request) (string, error)
}

// ProductDB defines the database operations needed for product image uploads.
// Provides data access layer methods for product retrieval and image URL updates.
// Implemented by ProductDBAdapter.
type ProductDB interface {
	GetProductByID(ctx context.Context, id string) (Product, error)
	UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error
}

// ProductDBAdapter implements ProductDB using *database.Queries.
// Adapter used by the upload service for DB operations.
// Constructed via NewProductDBAdapter for dependency injection.
type ProductDBAdapter struct {
	Queries *database.Queries
}

// NewProductDBAdapter creates a new ProductDBAdapter with the given queries.
// Factory function for creating database adapter instances.
// Parameters:
//   - queries: *database.Queries for database operations
//
// Returns:
//   - ProductDB: configured database adapter instance
func NewProductDBAdapter(queries *database.Queries) ProductDB {
	return &ProductDBAdapter{Queries: queries}
}

// GetProductByID retrieves a product by its ID from the database.
// Maps the database product to the local Product type for upload service consumption.
// Parameters:
//   - ctx: context.Context for the operation
//   - id: string identifier of the product to retrieve
//
// Returns:
//   - Product: the found product with image URL information
//   - error: nil on success, database error on failure
func (a *ProductDBAdapter) GetProductByID(ctx context.Context, id string) (Product, error) {
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

// UpdateProductImageURL updates the image URL for a product in the database.
// Converts parameters to database format and delegates to the underlying queries.
// Parameters:
//   - ctx: context.Context for the operation
//   - params: UpdateProductImageURLParams containing the update data
//
// Returns:
//   - error: nil on success, database error on failure
func (a *ProductDBAdapter) UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error {
	return a.Queries.UpdateProductImageURL(ctx, database.UpdateProductImageURLParams{
		ID:        params.ID,
		ImageUrl:  utils.ToNullString(params.ImageURL),
		UpdatedAt: time.Unix(params.UpdatedAt, 0),
	})
}

// Product represents a product with an optional image URL.
// Local type for upload service operations, mapped from database models.
type Product struct {
	ID       string
	ImageURL struct {
		String string
		Valid  bool
	}
}

// UpdateProductImageURLParams contains parameters for updating a product's image URL.
// Structured parameters for database update operations with timestamp tracking.
type UpdateProductImageURLParams struct {
	ID        string
	ImageURL  string
	UpdatedAt int64 // Unix timestamp for simplicity
}

// uploadServiceImpl implements the UploadService interface.
// Handles the business logic for uploading and updating product images.
// Manages file validation, storage operations, and database updates.
type uploadServiceImpl struct {
	db        ProductDB
	uploadDir string
	storage   FileStorage
}

// NewUploadService creates a new UploadService with the given dependencies.
// Factory function for creating upload service instances with configurable storage backends.
// Parameters:
//   - db: ProductDB for database operations
//   - uploadDir: string path for uploads
//   - storage: FileStorage implementation (local or S3)
//
// Returns:
//   - UploadService: configured upload service instance
func NewUploadService(db ProductDB, uploadDir string, storage FileStorage) UploadService {
	return &uploadServiceImpl{db: db, uploadDir: uploadDir, storage: storage}
}

// UploadProductImage handles uploading a new product image.
// Validates the form, checks MIME type, saves the file, and returns the image URL.
// Supports JPEG, PNG, GIF, and WebP image formats with proper error handling.
// Parameters:
//   - ctx: context.Context for request-scoped values
//   - userID: string user identifier
//   - r: *http.Request containing the multipart form
//
// Returns:
//   - string: the generated image URL on success
//   - error: AppError with appropriate code and message on failure
func (s *uploadServiceImpl) UploadProductImage(_ context.Context, _ string, r *http.Request) (string, error) {
	file, fileHeader, err := ParseAndGetImageFile(r)
	if err != nil {
		return "", &handlers.AppError{Code: "invalid_form", Message: err.Error(), Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = file.Close()
	}()

	// MIME type validation
	allowedMIMEs := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/gif":  {},
		"image/webp": {},
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if _, ok := allowedMIMEs[mimeType]; !ok {
		return "", &handlers.AppError{Code: "invalid_image", Message: "Unsupported image MIME type", Err: nil}
	}

	filename, err := s.storage.Save(file, fileHeader, s.uploadDir)
	if err != nil {
		return "", &handlers.AppError{Code: "file_save_failed", Message: err.Error(), Err: err}
	}
	imageURL := "/static/" + filename[strings.LastIndex(filename, "/")+1:]
	return imageURL, nil
}

// UpdateProductImage handles updating a product's image.
// Retrieves the product, validates the form, deletes the old image, saves the new file,
// updates the DB, and returns the new image URL. Supports JPEG, PNG, GIF, and WebP formats.
// Parameters:
//   - ctx: context.Context for request-scoped values
//   - productID: string product identifier
//   - userID: string user identifier
//   - r: *http.Request containing the multipart form
//
// Returns:
//   - string: the new image URL on success
//   - error: AppError with appropriate code and message on failure
func (s *uploadServiceImpl) UpdateProductImage(ctx context.Context, productID string, _ string, r *http.Request) (string, error) {
	product, err := s.db.GetProductByID(ctx, productID)
	if err != nil {
		return "", &handlers.AppError{Code: "not_found", Message: "Product not found", Err: err}
	}

	file, fileHeader, err := ParseAndGetImageFile(r)
	if err != nil {
		return "", &handlers.AppError{Code: "invalid_form", Message: err.Error(), Err: err}
	}
	defer func() {
		// Log error but don't return it since we're in defer
		_ = file.Close()
	}()

	// MIME type validation
	allowedMIMEs := map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/gif":  {},
		"image/webp": {},
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if _, ok := allowedMIMEs[mimeType]; !ok {
		return "", &handlers.AppError{Code: "invalid_image", Message: "Unsupported image MIME type", Err: nil}
	}

	// Delete old image if exists
	if product.ImageURL.Valid && product.ImageURL.String != "" {
		_ = s.storage.Delete(product.ImageURL.String, s.uploadDir)
	}

	filename, err := s.storage.Save(file, fileHeader, s.uploadDir)
	if err != nil {
		return "", &handlers.AppError{Code: "file_save_failed", Message: err.Error(), Err: err}
	}
	imageURL := "/static/" + filename[strings.LastIndex(filename, "/")+1:]

	params := UpdateProductImageURLParams{
		ID:        productID,
		ImageURL:  imageURL,
		UpdatedAt: time.Now().Unix(),
	}
	if err := s.db.UpdateProductImageURL(ctx, params); err != nil {
		return "", &handlers.AppError{Code: "db_error", Message: "Failed to update product image", Err: err}
	}
	return imageURL, nil
}

// --- FileStorage interface will be in storage_local.go ---
// --- ParseAndGetImageFile will be in storage_local.go ---
