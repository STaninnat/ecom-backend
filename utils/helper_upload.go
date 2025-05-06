package utils

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

	return file, fileHeader, nil
}

func SaveUploadedFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Create uploads folder if not exists
	uploadPath := "./uploads"
	if err := os.MkdirAll(uploadPath, 0750); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
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

	return filename, nil
}

func DeleteFileIfExists(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	const staticPrefix = "/static/"
	if len(imageURL) <= len(staticPrefix) || imageURL[:len(staticPrefix)] != staticPrefix {
		return fmt.Errorf("invalid image URL format")
	}

	filename := imageURL[len(staticPrefix):]
	fullPath := filepath.Join("./uploads", filename)

	if _, err := os.Stat(fullPath); err == nil {
		return os.Remove(fullPath)
	}

	return nil
}
