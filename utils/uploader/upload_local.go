package utilsuploaders

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ParseAndGetImageFile parses the multipart form and retrieves the image file and header.
func ParseAndGetImageFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Get file from form-data
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve image file: %w", err)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := AllowedImageExtensions[ext]; !ok {
		file.Close()
		return nil, nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return file, fileHeader, nil
}

// SaveUploadedFile saves the uploaded file to disk and returns the full file path.
func SaveUploadedFile(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error) {
	defer file.Close()
	// Create uploads folder if not exists
	if err := os.MkdirAll(uploadPath, 0750); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Ensure file path is safe and does not allow path traversal
	cleanFilePath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanFilePath, filepath.Clean(uploadPath)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid file path: %s", filePath)
	}

	// Save file to disk
	dst, err := os.Create(cleanFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return cleanFilePath, nil
}

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
	if !strings.HasPrefix(cleanPath, filepath.Clean(uploadPath)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", fullPath)
	}

	if _, err := os.Stat(cleanPath); err == nil {
		return os.Remove(cleanPath)
	}

	return nil
}
