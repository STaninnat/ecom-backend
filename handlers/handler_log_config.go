package handlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// LogHandlerError logs an error with structured logging and user action tracking
func (cfg *HandlerConfig) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	if cfg.LoggerService == nil {
		return
	}
	if err != nil {
		cfg.LoggerService.WithError(err).Error(logMsg)
	} else {
		cfg.LoggerService.Error(logMsg)
	}

	// For now, we'll use the legacy method since utils.LogUserAction expects *logrus.Logger
	// TODO: Create an adapter or modify utils.LogUserAction to accept interfaces
}

// LogHandlerSuccess logs a successful operation with structured logging and user action tracking
func (cfg *HandlerConfig) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
	// For now, we'll use the legacy method since utils.LogUserAction expects *logrus.Logger
	// TODO: Create an adapter or modify utils.LogUserAction to accept interfaces
}

// ErrMsgOrNil returns the error message or empty string if error is nil
func ErrMsgOrNil(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// GetRequestMetadata extracts IP address and user agent from the request
func (cfg *HandlerConfig) GetRequestMetadata(r *http.Request) (ip string, userAgent string) {
	if cfg.RequestMetadataService == nil {
		return "", ""
	}
	ip = cfg.RequestMetadataService.GetIPAddress(r)
	userAgent = cfg.RequestMetadataService.GetUserAgent(r)
	return
}

// Legacy compatibility methods for existing HandlersConfig
func (apicfg *HandlersConfig) LogHandlerError(ctx context.Context, action, details, logMsg, ip, ua string, err error) {
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

func (apicfg *HandlersConfig) LogHandlerSuccess(ctx context.Context, action, details, ip, ua string) {
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

func GetRequestMetadata(r *http.Request) (ip string, userAgent string) {
	ip = middlewares.GetIPAddress(r)
	userAgent = r.UserAgent()
	return
}
