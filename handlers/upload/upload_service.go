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

// UploadService defines the business logic interface for uploads (local or S3).
// It provides methods to upload and update product images.
type UploadService interface {
	UploadProductImage(ctx context.Context, userID string, r *http.Request) (string, error)
	UpdateProductImage(ctx context.Context, productID string, userID string, r *http.Request) (string, error)
}

// ProductDB defines the database operations needed for product image uploads.
// It is implemented by ProductDBAdapter.
type ProductDB interface {
	GetProductByID(ctx context.Context, id string) (Product, error)
	UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error
}

// ProductDBAdapter implements ProductDB using *database.Queries.
// This adapter is used by the upload service for DB operations
// and is constructed via NewProductDBAdapter.
type ProductDBAdapter struct {
	Queries *database.Queries
}

// NewProductDBAdapter creates a new ProductDBAdapter with the given queries.
func NewProductDBAdapter(queries *database.Queries) ProductDB {
	return &ProductDBAdapter{Queries: queries}
}

// GetProductByID retrieves a product by its ID from the database.
// It maps the database product to the local Product type.
func (a *ProductDBAdapter) GetProductByID(ctx context.Context, id string) (Product, error) {
	dbProduct, err := a.Queries.GetProductByID(ctx, id)
	if err != nil {
		return Product{}, err
	}
	return Product{
		ID: dbProduct.ID,
		ImageUrl: struct {
			String string
			Valid  bool
		}{
			String: dbProduct.ImageUrl.String,
			Valid:  dbProduct.ImageUrl.Valid,
		},
	}, nil
}

// UpdateProductImageURL updates the image URL for a product in the database.
func (a *ProductDBAdapter) UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error {
	return a.Queries.UpdateProductImageURL(ctx, database.UpdateProductImageURLParams{
		ID:        params.ID,
		ImageUrl:  utils.ToNullString(params.ImageUrl),
		UpdatedAt: time.Unix(params.UpdatedAt, 0),
	})
}

// Product represents a product with an optional image URL.
type Product struct {
	ID       string
	ImageUrl struct {
		String string
		Valid  bool
	}
}

// UpdateProductImageURLParams contains parameters for updating a product's image URL.
type UpdateProductImageURLParams struct {
	ID        string
	ImageUrl  string
	UpdatedAt int64 // Unix timestamp for simplicity
}

// uploadServiceImpl implements the UploadService interface.
// It handles the business logic for uploading and updating product images.
type uploadServiceImpl struct {
	db        ProductDB
	uploadDir string
	storage   FileStorage
}

// NewUploadService creates a new UploadService with the given dependencies.
//
// Parameters:
//   - db: ProductDB for database operations
//   - uploadDir: string path for uploads
//   - storage: FileStorage implementation (local or S3)
func NewUploadService(db ProductDB, uploadDir string, storage FileStorage) UploadService {
	return &uploadServiceImpl{db: db, uploadDir: uploadDir, storage: storage}
}

// UploadProductImage handles uploading a new product image.
// It validates the form, checks MIME type, saves the file, and returns the image URL.
//
// Parameters:
//   - ctx: context.Context for request-scoped values
//   - userID: string user identifier
//   - r: *http.Request containing the multipart form
//
// Returns the image URL or an error.
func (s *uploadServiceImpl) UploadProductImage(ctx context.Context, userID string, r *http.Request) (string, error) {
	file, fileHeader, err := ParseAndGetImageFile(r)
	if err != nil {
		return "", &handlers.AppError{Code: "invalid_form", Message: err.Error(), Err: err}
	}
	defer file.Close()

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
// It retrieves the product, validates the form, deletes the old image, saves the new file, updates the DB, and returns the new image URL.
//
// Parameters:
//   - ctx: context.Context for request-scoped values
//   - productID: string product identifier
//   - userID: string user identifier
//   - r: *http.Request containing the multipart form
//
// Returns the new image URL or an error.
func (s *uploadServiceImpl) UpdateProductImage(ctx context.Context, productID string, userID string, r *http.Request) (string, error) {
	product, err := s.db.GetProductByID(ctx, productID)
	if err != nil {
		return "", &handlers.AppError{Code: "not_found", Message: "Product not found", Err: err}
	}

	file, fileHeader, err := ParseAndGetImageFile(r)
	if err != nil {
		return "", &handlers.AppError{Code: "invalid_form", Message: err.Error(), Err: err}
	}
	defer file.Close()

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
	if product.ImageUrl.Valid && product.ImageUrl.String != "" {
		_ = s.storage.Delete(product.ImageUrl.String, s.uploadDir)
	}

	filename, err := s.storage.Save(file, fileHeader, s.uploadDir)
	if err != nil {
		return "", &handlers.AppError{Code: "file_save_failed", Message: err.Error(), Err: err}
	}
	imageURL := "/static/" + filename[strings.LastIndex(filename, "/")+1:]

	params := UpdateProductImageURLParams{
		ID:        productID,
		ImageUrl:  imageURL,
		UpdatedAt: time.Now().Unix(),
	}
	if err := s.db.UpdateProductImageURL(ctx, params); err != nil {
		return "", &handlers.AppError{Code: "db_error", Message: "Failed to update product image", Err: err}
	}
	return imageURL, nil
}

// --- FileStorage interface will be in storage_local.go ---
// --- ParseAndGetImageFile will be in storage_local.go ---
