// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// storage_local_test.go: Tests LocalFileStorage save/delete, multipart parsing with validation, and error cases for file operations and path security.

// TestLocalFileStorage_Save_And_Delete tests saving and deleting a file using LocalFileStorage.
// It verifies that the file is saved to disk, exists, and is deleted successfully.
func TestLocalFileStorage_Save_And_Delete(t *testing.T) {
	dir := t.TempDir()
	storage := &LocalFileStorage{}

	// Create a fake multipart file
	content := []byte("fake image data")
	file := &fakeFile{Reader: bytes.NewReader(content)}
	fileHeader := &multipart.FileHeader{Filename: "test.jpg"}

	path, err := storage.Save(file, fileHeader, dir)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if !strings.HasPrefix(path, dir) {
		t.Errorf("Saved file path should start with temp dir")
	}
	// File should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Saved file does not exist: %v", err)
	}

	// Delete should remove the file
	imageURL := "/static/" + filepath.Base(path)
	err = storage.Delete(imageURL, dir)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("File should be deleted, got: %v", err)
	}
}

type fakeFile struct{ *bytes.Reader }

func (f *fakeFile) Close() error { return nil }

// TestParseAndGetImageFile_Success tests successful parsing and retrieval of an image file from a multipart request.
// It verifies that the correct file and header are returned for a valid image upload.
func TestParseAndGetImageFile_Success(t *testing.T) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fw.Write([]byte("image data"))
	if err != nil {
		t.Errorf("fw.Write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	file, fileHeader, err := ParseAndGetImageFile(req)
	if err != nil {
		t.Fatalf("ParseAndGetImageFile failed: %v", err)
	}
	if fileHeader.Filename != "test.png" {
		t.Errorf("Expected filename 'test.png', got %q", fileHeader.Filename)
	}
	if err := file.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}

// TestParseAndGetImageFile_InvalidExtension tests the behavior when an unsupported file extension is uploaded.
// It ensures that an error is returned for invalid extensions and the file is closed.
func TestParseAndGetImageFile_InvalidExtension(t *testing.T) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("image", "test.exe")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fw.Write([]byte("image data"))
	if err != nil {
		t.Errorf("fw.Write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	file, fileHeader, err := ParseAndGetImageFile(req)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("Expected unsupported file extension error, got: %v", err)
	}
	if file != nil {
		if err := file.Close(); err != nil {
			t.Errorf("Failed to close file: %v", err)
		}
	}
	_ = fileHeader // may be nil
}

// TestSaveUploadedFile_InvalidPath tests the behavior when the upload path is invalid or cannot be created.
// It ensures that an error is returned for invalid upload paths.
func TestSaveUploadedFile_InvalidPath(t *testing.T) {
	file := &fakeFile{Reader: bytes.NewReader([]byte("data"))}
	fileHeader := &multipart.FileHeader{Filename: "test.jpg"}
	// Use a path that is guaranteed to fail
	_, err := SaveUploadedFile(file, fileHeader, "/this/path/should/not/exist/and/cannot/be/created")
	if err == nil {
		t.Error("Expected error for invalid upload path")
	}
}

// TestDeleteFileIfExists_InvalidURL tests the behavior when the image URL does not have the expected static prefix.
// It ensures that an error is returned for invalid image URL formats.
func TestDeleteFileIfExists_InvalidURL(t *testing.T) {
	dir := t.TempDir()
	err := DeleteFileIfExists("notstatic/test.jpg", dir)
	if err == nil || !strings.Contains(err.Error(), "invalid image URL format") {
		t.Errorf("Expected invalid image URL format error, got: %v", err)
	}
}

// TestDeleteFileIfExists_MissingFile tests the behavior when the file to delete does not exist.
// It ensures that no error is returned if the file is already missing.
func TestDeleteFileIfExists_MissingFile(t *testing.T) {
	dir := t.TempDir()
	imageURL := "/static/missing.jpg"
	err := DeleteFileIfExists(imageURL, dir)
	if err != nil {
		t.Errorf("Expected no error for missing file, got: %v", err)
	}
}

// --- Additional edge/error case tests ---
// TestSaveUploadedFile_CreateFails tests the behavior when file creation fails due to directory permissions.
// It ensures that an error is returned if os.Create fails.
func TestSaveUploadedFile_CreateFails(t *testing.T) {
	file := &fakeFile{Reader: bytes.NewReader([]byte("data"))}
	fileHeader := &multipart.FileHeader{Filename: "test.jpg"}
	dir := t.TempDir()
	// Make dir read-only to force os.Create to fail
	if err := os.Chmod(dir, 0500); err != nil { // nolint:gosec // Test file permissions for testing purposes
		t.Errorf("Failed to chmod directory: %v", err)
	}
	defer func() {
		if err := os.Chmod(dir, 0755); err != nil { // nolint:gosec // Test file permissions for testing purposes
			t.Errorf("Failed to restore directory permissions: %v", err)
		}
	}()
	_, err := SaveUploadedFile(file, fileHeader, dir)
	if err == nil || !strings.Contains(err.Error(), "failed to create file") {
		t.Errorf("Expected create file error, got: %v", err)
	}
}

// TestDeleteFileIfExists_EmptyURL tests the behavior when an empty image URL is provided.
// It ensures that no error is returned for an empty image URL.
func TestDeleteFileIfExists_EmptyURL(t *testing.T) {
	dir := t.TempDir()
	err := DeleteFileIfExists("", dir)
	if err != nil {
		t.Errorf("Expected nil for empty imageURL, got: %v", err)
	}
}
