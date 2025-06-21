package utilsuploaders_test

import (
	"context"
	"errors"
	"testing"

	utilsuploaders "github.com/STaninnat/ecom-backend/utils/uploader"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock S3Client ---

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return &s3.PutObjectOutput{}, args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, input)
	return &s3.DeleteObjectOutput{}, args.Error(1)
}

// --- Unit Tests ---

func TestUploadFileToS3_Success_WithParseAndGetImageFile(t *testing.T) {
	req := CreateMultipartRequest(t, "image", "image.jpg", []byte("image data"))

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	assert.NoError(t, err)
	defer file.Close()

	mockClient := new(MockS3Client)
	mockClient.On("PutObject", mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	uploader := &utilsuploaders.S3Uploader{
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	url, err := uploader.UploadFileToS3(context.Background(), file, fileHeader)
	assert.NoError(t, err)
	assert.Contains(t, url, "https://test-bucket.s3.amazonaws.com/uploads/")
}

func TestUploadFileToS3_Failure_WithParseAndGetImageFile(t *testing.T) {
	req := CreateMultipartRequest(t, "image", "error.jpg", []byte("error content"))

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	assert.NoError(t, err)
	defer file.Close()

	mockClient := new(MockS3Client)
	mockClient.On("PutObject", mock.Anything, mock.Anything).Return(nil, errors.New("S3 error"))

	uploader := &utilsuploaders.S3Uploader{
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	url, err := uploader.UploadFileToS3(context.Background(), file, fileHeader)
	assert.Error(t, err)
	assert.Empty(t, url)
}

func TestUploadFileToS3_ContentTypeMissing_WithParseAndGetImageFile(t *testing.T) {
	req := CreateMultipartRequest(t, "image", "test.png", []byte("image"))

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	assert.NoError(t, err)
	defer file.Close()

	// ลบ Content-Type ออกจาก header
	fileHeader.Header.Del("Content-Type")

	mockClient := new(MockS3Client)
	mockClient.On("PutObject", mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	uploader := &utilsuploaders.S3Uploader{
		Client:     mockClient,
		BucketName: "bucket",
	}

	url, err := uploader.UploadFileToS3(context.Background(), file, fileHeader)
	assert.NoError(t, err)
	assert.Contains(t, url, "https://bucket.s3.amazonaws.com/uploads/")
}

func TestUploadFileToS3_InvalidExtension_WithParseAndGetImageFile(t *testing.T) {
	req := CreateMultipartRequest(t, "image", "noextension", []byte("data"))

	file, fileHeader, err := utilsuploaders.ParseAndGetImageFile(req)
	assert.NoError(t, err)
	defer file.Close()

	mockClient := new(MockS3Client)
	mockClient.On("PutObject", mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)

	uploader := &utilsuploaders.S3Uploader{
		Client:     mockClient,
		BucketName: "bucket",
	}

	url, err := uploader.UploadFileToS3(context.Background(), file, fileHeader)
	assert.NoError(t, err)
	assert.Contains(t, url, "https://bucket.s3.amazonaws.com/uploads/")
}

func TestDeleteFileFromS3IfExists_Success(t *testing.T) {
	mockClient := new(MockS3Client)
	mockClient.On("DeleteObject", mock.Anything, mock.Anything).Return(&s3.DeleteObjectOutput{}, nil)

	err := utilsuploaders.DeleteFileFromS3IfExists(mockClient, "my-bucket", "https://my-bucket.s3.amazonaws.com/uploads/abc.jpg")
	assert.NoError(t, err)
}

func TestDeleteFileFromS3IfExists_InvalidURL(t *testing.T) {
	mockClient := new(MockS3Client)
	err := utilsuploaders.DeleteFileFromS3IfExists(mockClient, "my-bucket", "invalid-url")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image URL")
}

func TestDeleteFileFromS3IfExists_Failure(t *testing.T) {
	mockClient := new(MockS3Client)
	mockClient.On("DeleteObject", mock.Anything, mock.Anything).Return(nil, errors.New("delete failed"))

	err := utilsuploaders.DeleteFileFromS3IfExists(mockClient, "my-bucket", "https://my-bucket.s3.amazonaws.com/uploads/abc.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete file from S3")
}
