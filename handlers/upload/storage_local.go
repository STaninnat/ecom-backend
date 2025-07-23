// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/STaninnat/ecom-backend/utils"
)

// storage_local.go: Defines FileStorage interface and local filesystem implementation for secure file save/delete operations,
// including multipart parsing, file extension validation, path traversal protection, and filename generation.

// FileStorage abstracts file operations for uploads.
// Provides a common interface for saving and deleting files across different storage backends.
// Implemented by LocalFileStorage and S3FileStorage for local disk and cloud storage respectively.
type FileStorage interface {
	Save(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error)
	Delete(imageURL, uploadPath string) error
}

// LocalFileStorage implements FileStorage for local disk storage.
// Provides methods to save and delete files on the local filesystem with security checks.
type LocalFileStorage struct{}

// Save saves the uploaded file to local disk using SaveUploadedFile.
// Delegates to the underlying file save function with security validation.
// Parameters:
//   - file: multipart.File representing the uploaded file
//   - fileHeader: *multipart.FileHeader containing file metadata
//   - uploadPath: string path to the upload directory
//
// Returns:
//   - string: the full file path on success
//   - error: nil on success, error on failure
func (l *LocalFileStorage) Save(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error) {
	return SaveUploadedFile(file, fileHeader, uploadPath)
}

// Delete removes a file from local disk using DeleteFileIfExists.
// Delegates to the underlying file deletion function with security validation.
// Parameters:
//   - imageURL: string URL of the image to delete
//   - uploadPath: string path to the upload directory
//
// Returns:
//   - error: nil on success, error on failure
func (l *LocalFileStorage) Delete(imageURL, uploadPath string) error {
	return DeleteFileIfExists(imageURL, uploadPath)
}

// ParseAndGetImageFile parses the multipart form and retrieves the image file and header from the request.
// Validates the file extension against allowed types and handles form parsing errors.
// Parameters:
//   - r: *http.Request containing the multipart form data
//
// Returns:
//   - multipart.File: the uploaded file
//   - *multipart.FileHeader: file metadata
//   - error: nil on success, error on failure
func ParseAndGetImageFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		return nil, nil, err
	}
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		return nil, nil, err
	}
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := AllowedImageExtensions[ext]; !ok {
		if err := file.Close(); err != nil {
			return nil, nil, fmt.Errorf("unsupported file extension: %s (file close error: %w)", ext, err)
		}
		return nil, nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
	return file, fileHeader, nil
}

// SaveUploadedFile saves the uploaded file to disk and returns the full file path.
// Performs path traversal checks, creates secure filenames with UUIDs, and ensures the file is written securely.
// Parameters:
//   - file: multipart.File representing the uploaded file
//   - fileHeader: *multipart.FileHeader containing file metadata
//   - uploadPath: string path to the upload directory
//
// Returns:
//   - string: the full file path on success
//   - error: nil on success, error on failure
func SaveUploadedFile(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error) {
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("file.Close failed: %v\n", err)
		}
	}()
	if err := os.MkdirAll(uploadPath, 0750); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	filename := fmt.Sprintf("%s_%d%s", utils.NewUUIDString(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)
	cleanFilePath := filepath.Clean(filePath)
	// Strict path traversal check: cleanFilePath must be inside uploadPath
	absUploadPath, _ := filepath.Abs(uploadPath)
	absCleanFilePath, _ := filepath.Abs(cleanFilePath)
	if !strings.HasPrefix(absCleanFilePath, absUploadPath+string(os.PathSeparator)) && absCleanFilePath != absUploadPath {
		return "", fmt.Errorf("invalid file path: %s", filePath)
	}
	dst, err := os.Create(cleanFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			fmt.Printf("dst.Close failed: %v\n", err)
		}
	}()
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return cleanFilePath, nil
}

// DeleteFileIfExists deletes a file if it exists, given an image URL and upload path.
// Performs path traversal checks and only deletes files within the upload directory for security.
// Parameters:
//   - imageURL: string URL of the image to delete
//   - uploadPath: string path to the upload directory
//
// Returns:
//   - error: nil on success, error on failure
func DeleteFileIfExists(imageURL, uploadPath string) error {
	if imageURL == "" {
		return nil
	}
	const staticPrefix = "/static/"
	if !strings.HasPrefix(imageURL, staticPrefix) {
		return fmt.Errorf("invalid image URL format")
	}
	filename := imageURL[len(staticPrefix):]
	fullPath := filepath.Join(uploadPath, filename)
	cleanPath := filepath.Clean(fullPath)
	// Strict path traversal check: cleanPath must be inside uploadPath
	absUploadPath, _ := filepath.Abs(uploadPath)
	absCleanPath, _ := filepath.Abs(cleanPath)
	if !strings.HasPrefix(absCleanPath, absUploadPath+string(os.PathSeparator)) && absCleanPath != absUploadPath {
		return fmt.Errorf("invalid file path: %s", fullPath)
	}
	if _, err := os.Stat(cleanPath); err == nil {
		if err := os.Remove(cleanPath); err != nil {
			return err
		}
	}
	return nil
}

// AllowedImageExtensions is a set of allowed image file extensions for uploads.
// Defines the supported image formats: JPG, JPEG, PNG, GIF, and WebP.
var AllowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}
