package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type UserGoogleInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (apicfg *HandlersConfig) HandlerGoogleSignIn(w http.ResponseWriter, r *http.Request) {
	state := auth.GenerateState()
	err := apicfg.RedisClient.Set(r.Context(), "oauth_state:"+state, "valid", 10*time.Minute).Err()
	if err != nil {
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to store state")
		return
	}
	fmt.Println("Generated State:", state)

	authURL := apicfg.OAuth.Google.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (apicfg *HandlersConfig) HandlerGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing state parameter")
		return
	}

	redisState, err := apicfg.RedisClient.Get(r.Context(), "oauth_state:"+state).Result()
	fmt.Println("Received State:", state, "Redis State:", redisState)
	if redisState == "" {
		fmt.Println("State missing from Redis, redirecting to signin")
		http.Redirect(w, r, "/v1/auth/google/signin", http.StatusTemporaryRedirect)
		return
	}
	if err != nil || redisState != "valid" {
		middlewares.RespondWithError(w, http.StatusUnauthorized, "Invalid state parameter")
		return
	}

	fmt.Println("Expire state from Redis:", state)
	_ = apicfg.RedisClient.Expire(r.Context(), "oauth_state:"+state, 1*time.Minute).Err()

	code := r.URL.Query().Get("code")
	if code == "" {
		middlewares.RespondWithError(w, http.StatusBadRequest, "Missing authorization code")
		return
	}

	token, err := apicfg.exchangeGoogleToken(code)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to exchange token")
		return
	}

	user, err := apicfg.getUserInfoFromGoogle(token)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user info")
		return
	}

	accessToken, refreshToken, err := apicfg.handleUserAuthentication(w, r, user, token, state)
	if err != nil {
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Authentication error")
		return
	}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	auth.SetTokensAsCookies(w, accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	middlewares.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Google login successful",
	})
}

// exchangeGoogleToken exchanges authorization code for access token
func (apicfg *HandlersConfig) exchangeGoogleToken(code string) (*oauth2.Token, error) {
	return apicfg.OAuth.Google.Exchange(context.Background(), code)
}

// getUserInfoFromGoogle retrieves user information from Google API
func (apicfg *HandlersConfig) getUserInfoFromGoogle(token *oauth2.Token) (*UserGoogleInfo, error) {
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
func (apicfg *HandlersConfig) handleUserAuthentication(w http.ResponseWriter, r *http.Request, user *UserGoogleInfo, token *oauth2.Token, state string) (string, string, error) {
	fmt.Println("Authenticating user:", user.Email)

	ctx := r.Context()
	existingUser, err := apicfg.DB.CheckExistsAndGetIDByEmail(ctx, user.Email)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("Error checking user existence:", err)
		return "", "", err
	}
	if err == sql.ErrNoRows || !existingUser.Exists {
		fmt.Println("User not found, creating new user:", user.Email)
		existingUser.ID = ""
	}

	var userID string
	provider := "google"
	providerID := sql.NullString{String: user.ID, Valid: true}

	timeNow := time.Now().UTC()
	accessTokenExpiresAt := timeNow.Add(30 * time.Minute)
	refreshTokenExpiresAt := timeNow.Add(7 * 24 * time.Hour)

	tx, err := apicfg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("Error starting DB transaction:", err)
		return "", "", err
	}
	defer tx.Rollback()

	queries := apicfg.DB.WithTx(tx)

	if existingUser.Exists {
		userID = existingUser.ID
		fmt.Println("Existing user found:", userID)
	} else {
		userID = uuid.New().String()
		fmt.Println("Creating new user:", userID)
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
			fmt.Println("Error creating user:", err)
			return "", "", err
		}
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		fmt.Println("Error parsing user UUID:", err)
		return "", "", err
	}

	accessToken, err := apicfg.Auth.GenerateAccessToken(userUUID, accessTokenExpiresAt)
	if err != nil {
		fmt.Println("Error generating access token:", err)
		return "", "", err
	}

	refreshToken, err := apicfg.RedisClient.Get(ctx, "refresh_token:"+userID).Result()
	if err != nil || refreshToken == "" {
		refreshToken = token.RefreshToken
		fmt.Println("New refresh token from Google:", refreshToken)

		if refreshToken != "" {
			err = apicfg.Auth.StoreRefreshTokenInRedis(r, userID, refreshToken, "google", refreshTokenExpiresAt.Sub(timeNow))
			if err != nil {
				fmt.Println("Error storing refresh token in Redis:", err)
				return "", "", err
			}
		} else {
			authURL := apicfg.OAuth.Google.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
			http.Redirect(w, r, authURL, http.StatusFound)
			return "", "", err
		}
	}

	err = queries.UpdateUserSigninStatusByEmail(ctx, database.UpdateUserSigninStatusByEmailParams{
		UpdatedAt:  timeNow,
		Provider:   provider,
		ProviderID: providerID,
		Email:      user.Email,
	})
	if err != nil {
		fmt.Println("Error updating user signin status:", err)
		return "", "", err
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("Error committing transaction:", err)
		return "", "", err
	}

	fmt.Println("Authentication successful for:", user.Email)
	return accessToken, refreshToken, nil
}
