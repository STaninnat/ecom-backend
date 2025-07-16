package uploadhandlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Handler Local ---
type contextKey string

type mockLogger struct{ mock.Mock }

func (m *mockLogger) LogHandlerError(ctx context.Context, operation, code, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, code, message, ip, userAgent, err)
}
func (m *mockLogger) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}

type mockUploadService struct{ mock.Mock }

func (m *mockUploadService) UploadProductImage(ctx context.Context, userID string, r *http.Request) (string, error) {
	args := m.Called(ctx, userID, r)
	return args.String(0), args.Error(1)
}
func (m *mockUploadService) UpdateProductImage(ctx context.Context, productID string, userID string, r *http.Request) (string, error) {
	args := m.Called(ctx, productID, userID, r)
	return args.String(0), args.Error(1)
}

// --- Mock Handler S3 ---
type mockS3Logger struct{ mock.Mock }

func (m *mockS3Logger) LogHandlerError(ctx context.Context, operation, code, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, code, message, ip, userAgent, err)
}
func (m *mockS3Logger) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	m.Called(ctx, operation, message, ip, userAgent)
}

type mockS3UploadService struct{ mock.Mock }

func (m *mockS3UploadService) UploadProductImage(ctx context.Context, userID string, r *http.Request) (string, error) {
	args := m.Called(ctx, userID, r)
	return args.String(0), args.Error(1)
}
func (m *mockS3UploadService) UpdateProductImage(ctx context.Context, productID string, userID string, r *http.Request) (string, error) {
	args := m.Called(ctx, productID, userID, r)
	return args.String(0), args.Error(1)
}

// --- Mock Storage S3 ---
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

// PutObjectS3 satisfies the S3Client interface for PutObject using s3 types.
func (m *mockS3Client) PutObjectS3(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalled = true
	return &s3.PutObjectOutput{}, m.putErr
}

// DeleteObjectS3 satisfies the S3Client interface for DeleteObject using s3 types.
func (m *mockS3Client) DeleteObjectS3(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteCalled = true
	return &s3.DeleteObjectOutput{}, m.deleteErr
}

type s3FakeFile struct {
	data    []byte
	readPos int
}

func (f *s3FakeFile) Read(p []byte) (int, error) {
	n := copy(p, f.data[f.readPos:])
	f.readPos += n
	if f.readPos >= len(f.data) {
		return n, io.EOF
	}
	return n, nil
}
func (f *s3FakeFile) Close() error { return nil }

func (f *s3FakeFile) ReadAt(p []byte, off int64) (int, error) {
	if int(off) >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *s3FakeFile) Seek(offset int64, whence int) (int64, error) {
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

// --- Mocks Upload Service ---
type mockProductDB struct{ mock.Mock }

func (m *mockProductDB) GetProductByID(ctx context.Context, id string) (Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Product), args.Error(1)
}
func (m *mockProductDB) UpdateProductImageURL(ctx context.Context, params UpdateProductImageURLParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

type mockFileStorage struct{ mock.Mock }

func (m *mockFileStorage) Save(file multipart.File, fileHeader *multipart.FileHeader, uploadPath string) (string, error) {
	args := m.Called(file, fileHeader, uploadPath)
	return args.String(0), args.Error(1)
}
func (m *mockFileStorage) Delete(imageURL, uploadPath string) error {
	args := m.Called(imageURL, uploadPath)
	return args.Error(0)
}

// --- Helper to create a multipart request with an image file ---
func newMultipartImageRequest(t *testing.T, fieldName, fileName string, fileContent []byte) (*http.Request, *multipart.FileHeader) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile(fieldName, fileName)
	assert.NoError(t, err)
	_, err = fw.Write(fileContent)
	assert.NoError(t, err)
	w.Close()

	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Parse to get the file header for mocking
	err = req.ParseMultipartForm(10 << 20)
	assert.NoError(t, err)
	_, fileHeader, err := req.FormFile(fieldName)
	assert.NoError(t, err)
	return req, fileHeader
}

// MockLogger mocks handlers.HandlerLogger
// Only LogHandlerError is needed for these tests
type MockLogger struct{ mock.Mock }

func (m *MockLogger) LogHandlerError(ctx context.Context, operation, code, message, ip, userAgent string, err error) {
	m.Called(ctx, operation, code, message, ip, userAgent, err)
}
func (m *MockLogger) LogHandlerSuccess(ctx context.Context, operation, message, ip, userAgent string) {
	// not used in these tests
}
