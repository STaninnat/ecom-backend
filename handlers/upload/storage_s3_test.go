package uploadhandlers

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
	"testing"
)

// TestS3FileStorage_Save_Success tests the successful saving of a file to S3 storage.
// It verifies that the S3 client is called and a non-empty URL is returned on success.
func TestS3FileStorage_Save_Success(t *testing.T) {
	client := &mockS3Client{}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	file := &s3FakeFile{data: []byte("imgdata")}
	fh := &multipart.FileHeader{Filename: "test.jpg", Header: make(map[string][]string)}
	fh.Header.Set("Content-Type", "image/jpeg")
	url, err := storage.Save(file, fh, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty url")
	}
	if !client.putCalled {
		t.Error("expected PutObject to be called")
	}
}

// TestS3FileStorage_Save_S3Error tests the behavior when the S3 client returns an error during file upload.
// It ensures that an error is returned and the URL is empty.
func TestS3FileStorage_Save_S3Error(t *testing.T) {
	client := &mockS3Client{putErr: errors.New("s3 error")}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	file := &s3FakeFile{data: []byte("imgdata")}
	fh := &multipart.FileHeader{Filename: "test.jpg", Header: make(map[string][]string)}
	fh.Header.Set("Content-Type", "image/jpeg")
	url, err := storage.Save(file, fh, "")
	if err == nil || !strings.Contains(err.Error(), "failed to upload file to S3") {
		t.Errorf("expected S3 error, got: %v", err)
	}
	if url != "" {
		t.Error("expected empty url on error")
	}
}

// TestS3FileStorage_Save_UnsupportedExtension tests the behavior when an unsupported file extension is uploaded.
// It ensures that an error is returned and the URL is empty.
func TestS3FileStorage_Save_UnsupportedExtension(t *testing.T) {
	client := &mockS3Client{}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	file := &s3FakeFile{data: []byte("imgdata")}
	fh := &multipart.FileHeader{Filename: "test.exe", Header: make(map[string][]string)}
	fh.Header.Set("Content-Type", "application/octet-stream")
	url, err := storage.Save(file, fh, "")
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Errorf("expected unsupported extension error, got: %v", err)
	}
	if url != "" {
		t.Error("expected empty url on error")
	}
}

// TestS3FileStorage_Delete_Success tests the successful deletion of a file from S3 storage.
// It verifies that the S3 client is called and no error is returned on success.
func TestS3FileStorage_Delete_Success(t *testing.T) {
	client := &mockS3Client{}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	url := "https://bucket.s3.amazonaws.com/uploads/test.jpg"
	err := storage.Delete(url, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !client.deleteCalled {
		t.Error("expected DeleteObject to be called")
	}
}

// TestS3FileStorage_Delete_InvalidURL tests the behavior when an invalid image URL is provided for deletion.
// It ensures that an error is returned for invalid URLs.
func TestS3FileStorage_Delete_InvalidURL(t *testing.T) {
	client := &mockS3Client{}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	err := storage.Delete(":badurl", "")
	if err == nil || !strings.Contains(err.Error(), "invalid image URL") {
		t.Errorf("expected invalid image URL error, got: %v", err)
	}
}

// TestS3FileStorage_Delete_S3Error tests the behavior when the S3 client returns an error during file deletion.
// It ensures that an error is returned for S3 deletion failures.
func TestS3FileStorage_Delete_S3Error(t *testing.T) {
	client := &mockS3Client{deleteErr: errors.New("s3 error")}
	storage := &S3FileStorage{S3Client: client, BucketName: "bucket"}
	url := "https://bucket.s3.amazonaws.com/uploads/test.jpg"
	err := storage.Delete(url, "")
	if err == nil || !strings.Contains(err.Error(), "failed to delete file from S3") {
		t.Errorf("expected S3 error, got: %v", err)
	}
}

// TestUploadFileToS3 tests the UploadFileToS3 method of S3Uploader for:
// - Successful upload
// - Unsupported extension
// - S3 error
func TestUploadFileToS3(t *testing.T) {
	uploader := &S3Uploader{Client: &mockS3Client{}, BucketName: "bucket"}
	file := &s3FakeFile{data: []byte("imgdata")}
	fh := &multipart.FileHeader{Filename: "test.jpg", Header: make(map[string][]string)}
	fh.Header.Set("Content-Type", "image/jpeg")
	key, url, err := uploader.UploadFileToS3(context.Background(), file, fh)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if key == "" || url == "" {
		t.Errorf("expected non-empty key and url")
	}

	// Unsupported extension
	fh = &multipart.FileHeader{Filename: "test.txt", Header: make(map[string][]string)}
	_, _, err = uploader.UploadFileToS3(context.Background(), file, fh)
	if err == nil || !errors.Is(err, err) {
		t.Errorf("expected error for unsupported extension")
	}

	// S3 error
	uploader.Client = &mockS3Client{putErr: errors.New("s3 error")}
	fh = &multipart.FileHeader{Filename: "test.jpg", Header: make(map[string][]string)}
	fh.Header.Set("Content-Type", "image/jpeg")
	_, _, err = uploader.UploadFileToS3(context.Background(), file, fh)
	if err == nil || !errors.Is(err, uploader.Client.(*mockS3Client).putErr) {
		t.Errorf("expected s3 error, got %v", err)
	}
}

// TestDeleteFileFromS3IfExists tests the DeleteFileFromS3IfExists function for:
// - Successful delete
// - Invalid URL
// - Missing key
// - S3 error
func TestDeleteFileFromS3IfExists(t *testing.T) {
	client := &mockS3Client{}
	bucket := "bucket"
	url := "https://bucket.s3.amazonaws.com/uploads/test.jpg"
	err := DeleteFileFromS3IfExists(client, bucket, url)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !client.deleteCalled {
		t.Errorf("expected delete to be called")
	}

	// Invalid URL
	err = DeleteFileFromS3IfExists(client, bucket, ":badurl")
	if err == nil {
		t.Errorf("expected error for invalid url")
	}

	// Missing key
	err = DeleteFileFromS3IfExists(client, bucket, "https://bucket.s3.amazonaws.com/")
	if err == nil || !errors.Is(err, err) {
		t.Errorf("expected error for missing key")
	}

	// S3 error
	client = &mockS3Client{deleteErr: errors.New("s3 error")}
	err = DeleteFileFromS3IfExists(client, bucket, url)
	if err == nil || !errors.Is(err, client.deleteErr) {
		t.Errorf("expected s3 error, got %v", err)
	}
}
