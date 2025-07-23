// Package uploadhandlers manages product image uploads with local and S3 storage, including validation, error handling, and logging.
package uploadhandlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// storage_s3.go: Implements AWS S3 file storage with upload and delete operations, including file extension validation, unique key generation, and S3 URL parsing for secure object management.

// S3Client defines the interface for AWS S3 operations.
// Provides methods for uploading and deleting objects in S3 buckets.
// Used for mocking in tests and dependency injection.
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// S3Uploader provides helper methods for S3 file operations.
// Manages S3 client and bucket configuration for upload operations.
type S3Uploader struct {
	Client     S3Client
	BucketName string
}

// S3FileStorage implements FileStorage for AWS S3.
// Provides cloud storage operations using AWS S3 with proper error handling.
// S3Client must satisfy the S3Client interface for S3 operations.
// BucketName specifies the S3 bucket to use for file storage.
type S3FileStorage struct {
	S3Client   S3Client
	BucketName string
}

// Save uploads the provided file to AWS S3 using the configured S3 client and bucket.
// Validates file extensions, generates unique keys, and returns the S3 URL.
// Parameters:
//   - file: multipart.File representing the uploaded file
//   - fileHeader: *multipart.FileHeader containing file metadata
//   - _: string (unused, for interface compatibility)
//
// Returns:
//   - string: the S3 image URL on success
//   - error: nil on success, error on failure
func (s *S3FileStorage) Save(file multipart.File, fileHeader *multipart.FileHeader, _ string) (string, error) {
	uploader := &S3Uploader{
		Client:     s.S3Client,
		BucketName: s.BucketName,
	}
	_, imageURL, err := uploader.UploadFileToS3(context.Background(), file, fileHeader)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return imageURL, nil
}

// Delete removes a file from AWS S3 using the configured S3 client and bucket.
// Parses the S3 URL to extract the object key and deletes it from the bucket.
// Parameters:
//   - imageURL: string URL of the image to delete
//   - _: string (unused, for interface compatibility)
//
// Returns:
//   - error: nil on success, error on failure
func (s *S3FileStorage) Delete(imageURL, _ string) error {
	return DeleteFileFromS3IfExists(s.S3Client, s.BucketName, imageURL)
}

// UploadFileToS3 uploads a file to S3 with validation and unique key generation.
// Validates file extensions, creates unique S3 keys with UUIDs, and uploads with proper content type.
// Parameters:
//   - ctx: context.Context for the operation
//   - file: multipart.File representing the file to upload
//   - fileHeader: *multipart.FileHeader containing file metadata
//
// Returns:
//   - string: the S3 object key
//   - string: the S3 image URL
//   - error: nil on success, error on failure
func (u *S3Uploader) UploadFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, string, error) {
	defer func() {
		// reset pointer, ignore error on purpose
		_ = file.Close()
	}()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := AllowedImageExtensions[ext]; !ok {
		return "", "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	key := fmt.Sprintf("uploads/%s_%d%s", utils.NewUUIDString(), time.Now().Unix(), ext)
	contentType := fileHeader.Header.Get("Content-Type")

	_, err := u.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &u.BucketName,
		Key:         &key,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.BucketName, key)
	return key, url, nil
}

// DeleteFileFromS3IfExists deletes a file from S3 if it exists.
// Parses the S3 URL to extract the object key and deletes it from the specified bucket.
// Parameters:
//   - client: S3Client for S3 operations
//   - bucketName: string name of the S3 bucket
//   - imageURL: string URL of the image to delete
//
// Returns:
//   - error: nil on success, error on failure
func DeleteFileFromS3IfExists(client S3Client, bucketName string, imageURL string) error {
	u, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}
	key := strings.TrimPrefix(u.Path, "/")
	if key == "" {
		return fmt.Errorf("invalid image URL: missing key")
	}

	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
