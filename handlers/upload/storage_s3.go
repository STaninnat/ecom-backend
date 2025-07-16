package uploadhandlers

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Define interface for mocking
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3Uploader struct {
	Client     S3Client
	BucketName string
}

// S3FileStorage implements FileStorage for AWS S3.
// S3Client must satisfy utilsuploaders.S3Client.
// BucketName is the S3 bucket to use.
// S3Uploader is a helper for S3 operations.
type S3FileStorage struct {
	S3Client   S3Client
	BucketName string
}

// Save uploads the provided file to AWS S3 using the configured S3 client and bucket.
//
// Parameters:
//   - file: multipart.File representing the uploaded file
//   - fileHeader: *multipart.FileHeader containing file metadata
//   - _: string (unused, for interface compatibility)
//
// Returns the image URL and an error, if any.
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
//
// Parameters:
//   - imageURL: string URL of the image to delete
//   - _: string (unused, for interface compatibility)
//
// Returns an error, if any.
func (s *S3FileStorage) Delete(imageURL, _ string) error {
	return DeleteFileFromS3IfExists(s.S3Client, s.BucketName, imageURL)
}

func (u *S3Uploader) UploadFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, string, error) {
	defer file.Seek(0, io.SeekStart) // reset pointer

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := AllowedImageExtensions[ext]; !ok {
		return "", "", fmt.Errorf("unsupported file extension: %s", ext)
	}

	key := fmt.Sprintf("uploads/%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
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
