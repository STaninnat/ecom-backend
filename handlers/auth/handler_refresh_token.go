package authhandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"golang.org/x/oauth2"
)

func (apicfg *HandlersAuthConfig) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	userID, storedData, err := apicfg.Auth.ValidateCookieRefreshTokenData(w, r)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refresh_token",
			"validate cookie failed",
			"Error validating cookie",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, userID.String())

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(handlers.AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(handlers.RefreshTokenTTL)

	if storedData.Provider == "google" {
		refreshToken := storedData.Token

		newToken, err := apicfg.RefreshGoogleAccessToken(r, refreshToken)
		if err != nil {
			apicfg.LogHandlerError(
				ctx,
				"refresh_token",
				"refresh token failed",
				"Error refresh Google token",
				ip, userAgent, err,
			)
			middlewares.RespondWithError(w, http.StatusUnauthorized, "Failed to refresh Google token")
			return
		}

		auth.SetTokensAsCookies(w, newToken.AccessToken, refreshToken, newToken.Expiry, refreshTokenExpiresAt)

		apicfg.LogHandlerSuccess(ctxWithUserID, "refresh_token", "Refresh Google token success", ip, userAgent)

		middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
			Message: "Token refreshed successful",
		})

		return
	}

	err = apicfg.RedisClient.Del(ctx, auth.RedisRefreshTokenPrefix+userID.String()).Err()
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refresh_token",
			"delete token failed",
			"Error deleting refresh token from Redis",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to remove refresh token from Redis")
		return
	}

	accessToken, newRefreshToken, err := apicfg.Auth.GenerateTokens(userID.String(), accessTokenExpiresAt)
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refresh_token",
			"generate tokens failed",
			"Error generating tokens",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	err = apicfg.Auth.StoreRefreshTokenInRedis(r, userID.String(), newRefreshToken, "local", refreshTokenExpiresAt.Sub(timeNow))
	if err != nil {
		apicfg.LogHandlerError(
			ctx,
			"refresh_token",
			"store refresh token failed",
			"Error saving refresh token to Redis",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store session")
		return
	}

	auth.SetTokensAsCookies(w, accessToken, newRefreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	apicfg.LogHandlerSuccess(ctxWithUserID, "refresh_token", "Refresh token success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, handlers.HandlerResponse{
		Message: "Token refreshed successful",
	})
}

func (apicfg *HandlersAuthConfig) RefreshGoogleAccessToken(r *http.Request, refreshToken string) (*oauth2.Token, error) {
	var tokenSource oauth2.TokenSource
	if apicfg.CustomTokenSource != nil {
		tokenSource = apicfg.CustomTokenSource(r.Context(), refreshToken)
	} else {
		tokenSource = apicfg.OAuth.Google.TokenSource(r.Context(), &oauth2.Token{RefreshToken: refreshToken})
	}

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	return newToken, nil
}
