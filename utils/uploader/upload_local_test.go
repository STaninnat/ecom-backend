package utilsuploaders

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAndGetImageFile(t *testing.T) {
	t.Helper()
	// Valid image
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("image", "test.jpg")
	fw.Write([]byte("data"))
	w.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	file, fileHeader, err := ParseAndGetImageFile(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if file == nil || fileHeader == nil {
		t.Error("expected file and fileHeader to be non-nil")
	}
	if file != nil {
		file.Close()
	}

	// Unsupported extension
	body.Reset()
	w = multipart.NewWriter(body)
	fw, _ = w.CreateFormFile("image", "test.txt")
	fw.Write([]byte("data"))
	w.Close()
	req = httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	file, _, err = ParseAndGetImageFile(req)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected unsupported file extension error, got %v", err)
	}
	if file != nil {
		file.Close()
	}

	// Missing file
	body.Reset()
	w = multipart.NewWriter(body)
	w.Close()
	req = httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	_, _, err = ParseAndGetImageFile(req)
	if err == nil || !strings.Contains(err.Error(), "failed to retrieve image file") {
		t.Errorf("expected error for missing file, got %v", err)
	}

	// Parse error
	req = httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=bad")
	_, _, err = ParseAndGetImageFile(req)
	if err == nil || !strings.Contains(err.Error(), "failed to parse multipart form") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestSaveUploadedFile(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	file := &fakeFile{data: []byte("hello"), readErr: false, closeErr: false}
	fh := &multipart.FileHeader{Filename: "test.jpg"}
	path, err := SaveUploadedFile(file, fh, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(path, dir) {
		t.Errorf("expected path to start with dir, got %q", path)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "hello" {
		t.Errorf("file not written correctly: %v, %q", err, string(b))
	}

	// Directory creation error
	badDir := "/dev/null/shouldfail"
	file = &fakeFile{data: []byte("x"), readErr: false, closeErr: false}
	_, err = SaveUploadedFile(file, fh, badDir)
	if err == nil {
		t.Errorf("expected error for bad dir")
	}
}

type fakeFile struct {
	data     []byte
	readErr  bool
	closeErr bool
	readPos  int
}

func (f *fakeFile) Read(p []byte) (int, error) {
	if f.readErr {
		return 0, errors.New("read error")
	}
	n := copy(p, f.data[f.readPos:])
	f.readPos += n
	if f.readPos >= len(f.data) {
		return n, io.EOF
	}
	return n, nil
}

func (f *fakeFile) Close() error {
	if f.closeErr {
		return errors.New("close error")
	}
	return nil
}

func (f *fakeFile) ReadAt(p []byte, off int64) (int, error) {
	if int(off) >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *fakeFile) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(f.readPos) + offset
	case io.SeekEnd:
		abs = int64(len(f.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("negative position")
	}
	f.readPos = int(abs)
	return abs, nil
}

func TestDeleteFileIfExists(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	// File exists
	filePath := filepath.Join(dir, "test.jpg")
	os.WriteFile(filePath, []byte("x"), 0644)
	imageURL := "/static/test.jpg"
	err := DeleteFileIfExists(imageURL, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("file not deleted")
	}

	// File does not exist
	err = DeleteFileIfExists("/static/doesnotexist.jpg", dir)
	if err != nil {
		t.Errorf("unexpected error for non-existent file: %v", err)
	}

	// Empty URL
	err = DeleteFileIfExists("", dir)
	if err != nil {
		t.Errorf("expected nil for empty url, got %v", err)
	}

	// Invalid prefix
	err = DeleteFileIfExists("/badprefix/test.jpg", dir)
	if err == nil || !strings.Contains(err.Error(), "invalid image URL format") {
		t.Errorf("expected error for invalid prefix, got %v", err)
	}

	// Invalid file path
	badURL := "/static/../../evil.jpg"
	err = DeleteFileIfExists(badURL, dir)
	if err == nil || !strings.Contains(err.Error(), "invalid file path") {
		t.Errorf("expected error for invalid file path, got %v", err)
	}
}
