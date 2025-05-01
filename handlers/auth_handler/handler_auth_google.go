package auth_handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const oauthStatePrefix = "oauth_state:"

type UserGoogleInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (apicfg *HandlersAuthConfig) HandlerGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	state := auth.GenerateState()
	err := apicfg.RedisClient.Set(r.Context(), oauthStatePrefix+state, "valid", 10*time.Minute).Err()
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "signin-google", "store state failed", "Error storing state to Redis", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store state")
		return
	}

	authURL := apicfg.OAuth.Google.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (apicfg *HandlersAuthConfig) HandlerGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)

	state := r.URL.Query().Get("state")
	if state == "" {
		apicfg.LogHandlerError(r.Context(), "callback-google", "state parameter failed", "Error getting state from URL", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing state parameter")
		return
	}

	redisState, err := apicfg.RedisClient.Get(r.Context(), oauthStatePrefix+state).Result()
	if redisState == "" {
		apicfg.LogHandlerError(r.Context(), "callback-google", "get state failed", "Error getting state from Redis, redirecting to signin", ip, userAgent, nil)
		http.Redirect(w, r, "/v1/auth/google/signin", http.StatusTemporaryRedirect)
		return
	}
	if err != nil || redisState != "valid" {
		apicfg.LogHandlerError(r.Context(), "callback-google", "get state failed", "Error getting state from Redis, invalid state", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid state parameter")
		return
	}

	_ = apicfg.RedisClient.Expire(r.Context(), "oauth_state:"+state, 1*time.Minute).Err()

	code := r.URL.Query().Get("code")
	if code == "" {
		apicfg.LogHandlerError(r.Context(), "callback-google", "authorization code failed", "Error getting authorization code from URL", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing authorization code")
		return
	}

	token, err := apicfg.exchangeGoogleToken(code)
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "callback-google", "exchange token failed", "Error exchange Google token", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to exchange token")
		return
	}

	user, err := apicfg.getUserInfoFromGoogle(token)
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "callback-google", "retrieve user failed", "Error retrieving user info from Google", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user info")
		return
	}

	accessToken, refreshToken, userID, err := apicfg.handleUserAuthentication(w, r, user, token, state)
	if err != nil {
		apicfg.LogHandlerError(r.Context(), "callback-google", "authentication failed", "Error during Google authentication", ip, userAgent, err)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Authentication error")
		return
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(handlers.AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(handlers.RefreshTokenTTL)

	auth.SetTokensAsCookies(w, accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	ctxWithUserID := context.WithValue(r.Context(), utils.ContextKeyUserID, userID)

	apicfg.LogHandlerSuccess(ctxWithUserID, "callback-google", "Google signin success", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Google signin successful",
	})
}

// exchangeGoogleToken exchanges authorization code for access token
func (apicfg *HandlersAuthConfig) exchangeGoogleToken(code string) (*oauth2.Token, error) {
	return apicfg.OAuth.Google.Exchange(context.Background(), code)
}

// getUserInfoFromGoogle retrieves user information from Google API
func (apicfg *HandlersAuthConfig) getUserInfoFromGoogle(token *oauth2.Token) (*UserGoogleInfo, error) {
	client := apicfg.OAuth.Google.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user UserGoogleInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// handleUserAuthentication handles user signin/signup
func (apicfg *HandlersAuthConfig) handleUserAuthentication(w http.ResponseWriter, r *http.Request, user *UserGoogleInfo, token *oauth2.Token, state string) (string, string, string, error) {
	ctx := r.Context()
	existingUser, err := apicfg.DB.CheckExistsAndGetIDByEmail(ctx, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return "", "", "", err
	}
	if err == sql.ErrNoRows || !existingUser.Exists {
		existingUser.ID = ""
	}

	var userID string
	provider := "google"
	providerID := sql.NullString{String: user.ID, Valid: true}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(handlers.AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(handlers.RefreshTokenTTL)

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	if existingUser.Exists {
		userID = existingUser.ID
	} else {
		userID = uuid.New().String()
		err = queries.CreateUser(ctx, database.CreateUserParams{
			ID:         userID,
			Name:       user.Name,
			Email:      user.Email,
			Password:   sql.NullString{},
			Provider:   provider,
			ProviderID: providerID,
			CreatedAt:  timeNow,
			UpdatedAt:  timeNow,
		})
		if err != nil {
			return "", "", "", err
		}
	}

	accessToken, err := apicfg.Auth.GenerateAccessToken(userID, accessTokenExpiresAt)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := apicfg.RedisClient.Get(ctx, "refresh_token:"+userID).Result()
	if err != nil || refreshToken == "" {
		refreshToken = token.RefreshToken

		if refreshToken != "" {
			err = apicfg.Auth.StoreRefreshTokenInRedis(r, userID, refreshToken, "google", refreshTokenExpiresAt.Sub(timeNow))
			if err != nil {
				return "", "", "", err
			}
		} else {
			authURL := apicfg.OAuth.Google.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
			http.Redirect(w, r, authURL, http.StatusFound)
			return "", "", "", err
		}
	}

	err = queries.UpdateUserSigninStatusByEmail(ctx, database.UpdateUserSigninStatusByEmailParams{
		UpdatedAt:  timeNow,
		Provider:   provider,
		ProviderID: providerID,
		Email:      user.Email,
	})
	if err != nil {
		return "", "", "", err
	}

	if err = tx.Commit(); err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, userID, nil
}
