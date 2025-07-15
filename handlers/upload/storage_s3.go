package uploadhandlers

import (
	"context"
	"fmt"
	"mime/multipart"

	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
)

// S3FileStorage implements FileStorage for AWS S3.
// S3Client must satisfy utilsuploaders.S3Client.
// BucketName is the S3 bucket to use.
// S3Uploader is a helper for S3 operations.
type S3FileStorage struct {
	S3Client   utilsuploaders.S3Client
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
	uploader := &utilsuploaders.S3Uploader{
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
	return utilsuploaders.DeleteFileFromS3IfExists(s.S3Client, s.BucketName, imageURL)
}
