// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_log_config.go: Provides logging and metadata utilities for handler operations, supporting both HandlerConfig and Config.

// LogHandlerError logs an error with structured logging and user action tracking for HandlerConfig.
func (cfg *HandlerConfig) LogHandlerError(_ context.Context, _ string, _ string, logMsg, _ string, _ string, err error) {
	if cfg.LoggerService == nil {
		return
	}
	if err != nil {
		cfg.LoggerService.WithError(err).Error(logMsg)
	} else {
		cfg.LoggerService.Error(logMsg)
	}

	// Note: Currently using legacy logging method. Future improvement: Create adapter or modify utils.LogUserAction to accept interfaces for better testability.
}

// LogHandlerSuccess logs a successful operation with structured logging and user action tracking for HandlerConfig.
func (cfg *HandlerConfig) LogHandlerSuccess(_ context.Context, _ string, _ string, _ string, _ string) {
	// Note: Currently using legacy logging method. Future improvement: Create adapter or modify utils.LogUserAction to accept interfaces for better testability.
}

// ErrMsgOrNil returns the error message or empty string if error is nil.
func ErrMsgOrNil(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetRequestMetadata extracts IP address and user agent from the request using the configured RequestMetadataService.
func (cfg *HandlerConfig) GetRequestMetadata(r *http.Request) (ip string, userAgent string) {
	if cfg.RequestMetadataService == nil {
		return "", ""
	}
	ip = cfg.RequestMetadataService.GetIPAddress(r)
	userAgent = cfg.RequestMetadataService.GetUserAgent(r)
	return
}

// LogHandlerError logs an error with structured logging and user action tracking for Config.
// It logs the error message and user action details.
func (apicfg *Config) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	if err != nil {
		apicfg.Logger.WithError(err).Error(logMsg)
	} else {
		apicfg.Logger.Error(logMsg)
	}

	utils.LogUserAction(utils.ActionLogParams{
		Logger:    apicfg.Logger,
		Ctx:       ctx,
		Action:    action,
		Status:    "fail",
		Details:   details,
		ErrorMsg:  ErrMsgOrNil(err),
		UserAgent: ua,
		IP:        ip,
	})
}

// LogHandlerSuccess logs a successful operation with structured logging and user action tracking for Config.
func (apicfg *Config) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	utils.LogUserAction(utils.ActionLogParams{
		Logger:    apicfg.Logger,
		Ctx:       ctx,
		Action:    action,
		Status:    "success",
		Details:   details,
		UserAgent: ua,
		IP:        ip,
	})
}

// GetRequestMetadata extracts IP address and user agent from the request using middlewares.GetIPAddress and r.UserAgent().
func GetRequestMetadata(r *http.Request) (ip string, userAgent string) {
	ip = middlewares.GetIPAddress(r)
	userAgent = r.UserAgent()
	return
}

// Ensure HandlersConfig implements HandlerLogger
var _ HandlerLogger = (*Config)(nil)
