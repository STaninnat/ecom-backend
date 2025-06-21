package utilsuploaders_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
)

func CreateMultipartRequest(t *testing.T, fieldName, filename string, content []byte) *http.Request {
	t.Helper()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	part.Write(content)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestParseAndGetImageFile(t *testing.T) {
	content := []byte("fake image content")
	req := CreateMultipartRequest(t, "image", "image.png", content)

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if file == nil || fileHeader == nil {
		t.Fatalf("expected file and fileHeader, got nil")
	}
	defer file.Close()
}

func TestParseAndGetImageFile_MissingField(t *testing.T) {
	req := CreateMultipartRequest(t, "notimage", "image.png", []byte("data"))

	_, _, err := utilsuploaders.ParseAndGetImageFile(req)
	if err == nil || !strings.Contains(err.Error(), "failed to retrieve image file") {
		t.Fatalf("expected error for missing image field, got %v", err)
	}
}

func TestSaveUploadedFile(t *testing.T) {
	tempDir := t.TempDir()

	content := []byte("test content")
	req := CreateMultipartRequest(t, "image", "test.jpg", content)

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	if err != nil {
		t.Fatalf("ParseAndGetImageFile failed: %v", err)
	}
	defer file.Close()

	filename, err := utilsuploaders.SaveUploadedFile(file, fileHeader, tempDir)
	if err != nil {
		t.Fatalf("SaveUploadedFile failed: %v", err)
	}

	fullPath := filepath.Join(tempDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Fatalf("expected file to exist at %s", fullPath)
	}
}

func TestDeleteFileIfExists(t *testing.T) {
	tempDir := t.TempDir()
	filename := "image.jpg"
	fullPath := filepath.Join(tempDir, filename)
	if err := os.WriteFile(fullPath, []byte("delete me"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	imageURL := "/static/" + filename
	err := utilsuploaders.DeleteFileIfExists(imageURL, tempDir)
	if err != nil {
		t.Fatalf("DeleteFileIfExists failed: %v", err)
	}

	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted")
	}
}

func TestDeleteFileIfExists_FileNotExist(t *testing.T) {
	tempDir := t.TempDir()
	imageURL := "/static/not_exist.jpg"

	err := utilsuploaders.DeleteFileIfExists(imageURL, tempDir)
	if err != nil {
		t.Fatalf("expected no error for non-existing file, got %v", err)
	}
}

func TestDeleteFileIfExists_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	err := utilsuploaders.DeleteFileIfExists("invalid.jpg", tempDir)
	if err == nil || !strings.Contains(err.Error(), "invalid image URL format") {
		t.Fatalf("expected format error, got %v", err)
	}
}

func TestDeleteFileIfExists_TraversalAttempt(t *testing.T) {
	tempDir := t.TempDir()
	evilFile := "../evil.jpg"
	imageURL := "/static/" + evilFile

	err := utilsuploaders.DeleteFileIfExists(imageURL, tempDir)
	if err == nil || !strings.Contains(err.Error(), "invalid file path") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

func TestDeleteFileIfExists_EmptyURL(t *testing.T) {
	tempDir := t.TempDir()
	err := utilsuploaders.DeleteFileIfExists("", tempDir)
	if err != nil {
		t.Fatalf("expected no error for empty input, got %v", err)
	}
}
