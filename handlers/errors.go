package handlers

import "fmt"

// AppError is a generic error type for all handler modules
// It can be used directly or embedded/aliased for module-specific errors
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// APIResponse is a standard response struct for all API handlers
// Use Data for success payloads, Error for error messages, and Code for error codes
type APIResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
}
