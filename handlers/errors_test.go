package handlers

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name   string
		appErr *AppError
		want   string
	}{
		{
			name: "error with underlying error",
			appErr: &AppError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input",
				Err:     errors.New("field is required"),
			},
			want: "Invalid input: field is required",
		},
		{
			name: "error without underlying error",
			appErr: &AppError{
				Code:    "NOT_FOUND",
				Message: "User not found",
				Err:     nil,
			},
			want: "User not found",
		},
		{
			name: "error with empty message and underlying error",
			appErr: &AppError{
				Code:    "INTERNAL_ERROR",
				Message: "",
				Err:     errors.New("database connection failed"),
			},
			want: ": database connection failed",
		},
		{
			name: "error with empty message and no underlying error",
			appErr: &AppError{
				Code:    "UNKNOWN",
				Message: "",
				Err:     nil,
			},
			want: "",
		},
		{
			name: "error with nil underlying error",
			appErr: &AppError{
				Code:    "AUTH_ERROR",
				Message: "Authentication failed",
				Err:     nil,
			},
			want: "Authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			if got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	tests := []struct {
		name   string
		appErr *AppError
		want   error
	}{
		{
			name: "unwrap with underlying error",
			appErr: &AppError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input",
				Err:     errors.New("field is required"),
			},
			want: errors.New("field is required"),
		},
		{
			name: "unwrap without underlying error",
			appErr: &AppError{
				Code:    "NOT_FOUND",
				Message: "User not found",
				Err:     nil,
			},
			want: nil,
		},
		{
			name: "unwrap with custom error type",
			appErr: &AppError{
				Code:    "DATABASE_ERROR",
				Message: "Database operation failed",
				Err:     &AppError{Message: "Connection timeout"},
			},
			want: &AppError{Message: "Connection timeout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Unwrap()
			if tt.want == nil {
				if got != nil {
					t.Errorf("AppError.Unwrap() = %v, want nil", got)
				}
			} else {
				if got == nil || got.Error() != tt.want.Error() {
					t.Errorf("AppError.Unwrap() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestAppError_Fields(t *testing.T) {
	// Test that AppError fields are accessible and properly set
	underlyingErr := errors.New("test error")
	appErr := &AppError{
		Code:    "TEST_CODE",
		Message: "Test message",
		Err:     underlyingErr,
	}

	if appErr.Code != "TEST_CODE" {
		t.Errorf("AppError.Code = %v, want TEST_CODE", appErr.Code)
	}

	if appErr.Message != "Test message" {
		t.Errorf("AppError.Message = %v, want Test message", appErr.Message)
	}

	if appErr.Err != underlyingErr {
		t.Errorf("AppError.Err = %v, want %v", appErr.Err, underlyingErr)
	}
}

func TestAPIResponse_Structure(t *testing.T) {
	// Test that APIResponse can be created with various field combinations
	response := APIResponse{
		Message: "Success",
		Data:    map[string]string{"key": "value"},
		Error:   "",
		Code:    "",
	}

	if response.Message != "Success" {
		t.Errorf("APIResponse.Message = %v, want Success", response.Message)
	}

	data, ok := response.Data.(map[string]string)
	if !ok {
		t.Error("APIResponse.Data type assertion failed")
	}

	if data["key"] != "value" {
		t.Errorf("APIResponse.Data[key] = %v, want value", data["key"])
	}

	// Test error response
	errorResponse := APIResponse{
		Message: "",
		Data:    nil,
		Error:   "Something went wrong",
		Code:    "INTERNAL_ERROR",
	}

	if errorResponse.Error != "Something went wrong" {
		t.Errorf("APIResponse.Error = %v, want Something went wrong", errorResponse.Error)
	}

	if errorResponse.Code != "INTERNAL_ERROR" {
		t.Errorf("APIResponse.Code = %v, want INTERNAL_ERROR", errorResponse.Code)
	}
}
