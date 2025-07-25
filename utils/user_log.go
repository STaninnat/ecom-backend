// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"context"

	"github.com/sirupsen/logrus"
)

// user_log.go: Provides utilities for logging user actions with contextual information.

// ContextKey is a custom type for context keys used in user action logging.
type ContextKey string

// ContextKeyUserID and ContextKeyRequestID are context keys for storing user and request IDs in context.Context.
const (
	ContextKeyUserID    ContextKey = "userID"
	ContextKeyRequestID ContextKey = "requestID"
)

// ActionLogParams holds parameters for logging a user action, including logger, context, action details, status, and metadata.
type ActionLogParams struct {
	Logger    *logrus.Logger
	Ctx       context.Context
	Action    string
	Status    string
	Details   string
	ErrorMsg  string
	UserAgent string
	IP        string
}

// LogUserAction logs a user action with contextual information and status.
// Logs at Info level for "pending" and "success" (or default), and at Error level for "fail".
// If ErrorMsg is provided, it is included in the log fields.
func LogUserAction(p ActionLogParams) {
	userID := p.Ctx.Value(ContextKeyUserID)
	requestID := p.Ctx.Value(ContextKeyRequestID)

	fields := logrus.Fields{
		"userID":     userID,
		"action":     p.Action,
		"status":     p.Status,
		"details":    p.Details,
		"userAgent":  p.UserAgent,
		"ip":         p.IP,
		"request_id": requestID,
	}

	if p.ErrorMsg != "" {
		fields["error"] = p.ErrorMsg
	}

	entry := p.Logger.WithFields(fields)

	switch p.Status {
	case "pending":
		entry.Info("User action pending")
	case "fail":
		entry.Error("User action failed")
	default:
		entry.Info("User action success")
	}
}
