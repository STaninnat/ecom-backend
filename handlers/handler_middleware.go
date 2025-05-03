package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
)

type authhandler func(http.ResponseWriter, *http.Request, database.User)

func (apicfg *HandlersConfig) HandlerMiddleware(handler authhandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, userAgent := GetRequestMetadata(r)

		tokenString, err := r.Cookie("access_token")
		if err != nil {
			apicfg.LogHandlerError(
				r.Context(),
				"auth middleware",
				"missing access token cookie",
				"Access token cookie not found",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusUnauthorized, "Couldn't find token")
			return
		}

		token := tokenString.Value

		claims, err := apicfg.Auth.ValidateAccessToken(token, apicfg.JWTSecret)
		if err != nil {
			apicfg.LogHandlerError(
				r.Context(),
				"auth middleware",
				"invalid access token",
				"Access token validation failed",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		user, err := apicfg.DB.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			apicfg.LogHandlerError(
				r.Context(),
				"auth middleware",
				"user lookup failed",
				"Failed to fetch user from database",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusInternalServerError, "Couldn't get user info")
			return
		}

		handler(w, r, user)
	}
}
