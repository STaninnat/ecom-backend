package utils

import (
	"context"

	"github.com/sirupsen/logrus"
)

type ContextKey string

const ContextKeyUserID ContextKey = "userID"
const ContextKeyRequestID ContextKey = "reqestID"

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

	if p.Status == "pending" {
		entry.Info("User action pending")
	} else if p.Status == "fail" {
		entry.Error("User action failed")
	} else {
		entry.Info("User action success")
	}
}
