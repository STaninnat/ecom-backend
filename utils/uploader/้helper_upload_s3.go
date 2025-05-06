package utilsuploaders

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Uploader struct {
	Client     *s3.Client
	BucketName string
}

func (u *S3Uploader) UploadFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	defer file.Seek(0, io.SeekStart) // reset pointer

	ext := filepath.Ext(fileHeader.Filename)
	key := fmt.Sprintf("uploads/%s_%d%s", uuid.New().String(), time.Now().Unix(), ext) // for dev
	// ideal >> key := fmt.Sprintf("uploads/%s_%s_%d%s", categoryID, productID, time.Now().Unix(), ext)

	contentType := fileHeader.Header.Get("Content-Type")

	_, err := u.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &u.BucketName,
		Key:         &key,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.BucketName, key)
	return url, nil
}

func DeleteFileFromS3IfExists(client *s3.Client, bucketName string, imageURL string) error {
	parts := strings.Split(imageURL, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid image URL")
	}
	key := strings.Join(parts[3:], "/") // remove file's key from url

	_, err := client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
