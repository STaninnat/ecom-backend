package utilsuploaders

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	putErr       error
	deleteErr    error
	putCalled    bool
	deleteCalled bool
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalled = true
	return &s3.PutObjectOutput{}, m.putErr
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteCalled = true
	return &s3.DeleteObjectOutput{}, m.deleteErr
}

// Satisfy S3Client interface using s3 types
func (m *mockS3Client) PutObjectS3(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalled = true
	return &s3.PutObjectOutput{}, m.putErr
}
func (m *mockS3Client) DeleteObjectS3(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteCalled = true
	return &s3.DeleteObjectOutput{}, m.deleteErr
}

func TestUploadFileToS3(t *testing.T) {
	uploader := &S3Uploader{Client: &mockS3Client{}, BucketName: "bucket"}
	file := &fakeFile{}
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
