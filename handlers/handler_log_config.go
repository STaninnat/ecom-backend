package handlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

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

func ErrMsgOrNil(err error) string {
	if err != nil {
		return err.Error()
	}

	return ""
}

func GetRequestMetadata(r *http.Request) (ip string, userAgent string) {
	ip = middlewares.GetIPAddress(r)
	userAgent = r.UserAgent()
	return
}
