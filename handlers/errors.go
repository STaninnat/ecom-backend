// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import "fmt"

// errors.go: Defines common error and response types used across API handler modules.

// AppError is a generic error type for all handler modules.
// Can be used directly or embedded/aliased for module-specific errors.
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface for AppError, returning the error message.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error for AppError, supporting error unwrapping.
func (e *AppError) Unwrap() error {
	return e.Err
}

// APIResponse is a standard response struct for all API handlers.
// Use Data for success payloads, Error for error messages, and Code for error codes.
type APIResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
}
