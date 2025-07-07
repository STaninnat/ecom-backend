package authhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

const (
	// Token TTLs
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour

	// OAuth state TTL
	OAuthStateTTL = 10 * time.Minute

	// Providers
	LocalProvider = "local"

	// User roles
	UserRole = "user"

	// Redis key prefixes
	RefreshTokenKeyPrefix = "refresh_token:"
	OAuthStateKeyPrefix   = "oauth_state:"

	// OAuth state value
	OAuthStateValid = "valid"
)

// AuthService defines the business logic interface for authentication
type AuthService interface {
	SignUp(ctx context.Context, params SignUpParams) (*AuthResult, error)
	SignIn(ctx context.Context, params SignInParams) (*AuthResult, error)
	SignOut(ctx context.Context, userID string, provider string) error
	RefreshToken(ctx context.Context, userID string, provider string, refreshToken string) (*AuthResult, error)
	HandleGoogleAuth(ctx context.Context, code string, state string) (*AuthResult, error)
	GenerateGoogleAuthURL(state string) (string, error)
}

// SignUpParams represents signup request parameters
type SignUpParams struct {
	Name     string
	Email    string
	Password string
}

// SignInParams represents signin request parameters
type SignInParams struct {
	Email    string
	Password string
}

// AuthResult represents the result of authentication operations
type AuthResult struct {
	UserID              string
	AccessToken         string
	RefreshToken        string
	AccessTokenExpires  time.Time
	RefreshTokenExpires time.Time
	IsNewUser           bool
}

// authServiceImpl implements AuthService
type authServiceImpl struct {
	db          *database.Queries
	dbConn      *sql.DB
	auth        *auth.AuthConfig
	redisClient redis.Cmdable
	oauth       *oauth2.Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	db *database.Queries,
	dbConn *sql.DB,
	auth *auth.AuthConfig,
	redisClient redis.Cmdable,
	oauth *oauth2.Config,
) AuthService {
	return &authServiceImpl{
		db:          db,
		dbConn:      dbConn,
		auth:        auth,
		redisClient: redisClient,
		oauth:       oauth,
	}
}

// AuthError represents authentication-specific errors
// Now aliases handlers.AppError for consistency
type AuthError = handlers.AppError

// SignUp handles user registration with local authentication
func (s *authServiceImpl) SignUp(ctx context.Context, params SignUpParams) (*AuthResult, error) {
	// Check if name exists
	nameExists, err := s.db.CheckUserExistsByName(ctx, params.Name)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Error checking name existence", Err: err}
	}
	if nameExists {
		return nil, &handlers.AppError{Code: "name_exists", Message: "An account with this name already exists"}
	}

	// Check if email exists
	emailExists, err := s.db.CheckUserExistsByEmail(ctx, params.Email)
	if err != nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "Error checking email existence", Err: err}
	}
	if emailExists {
		return nil, &handlers.AppError{Code: "email_exists", Message: "An account with this email already exists"}
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		return nil, &handlers.AppError{Code: "hash_error", Message: "Error hashing password", Err: err}
	}

	// Create user
	userID := uuid.New()
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	err = queries.CreateUser(ctx, database.CreateUserParams{
		ID:         userID.String(),
		Name:       params.Name,
		Email:      params.Email,
		Password:   utils.ToNullString(hashedPassword),
		Provider:   LocalProvider,
		ProviderID: sql.NullString{},
		Role:       UserRole,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "create_user_error", Message: "Error creating user", Err: err}
	}

	// Generate tokens and store refresh token
	authResult, err := s.generateAndStoreTokens(userID.String(), LocalProvider, timeNow, true)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return authResult, nil
}

// SignIn handles user authentication with local credentials
func (s *authServiceImpl) SignIn(ctx context.Context, params SignInParams) (*AuthResult, error) {
	// Get user by email
	user, err := s.db.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return nil, &handlers.AppError{Code: "user_not_found", Message: "Invalid credentials"}
	}

	// Check password
	err = auth.CheckPasswordHash(params.Password, user.Password.String)
	if err != nil {
		return nil, &handlers.AppError{Code: "invalid_password", Message: "Invalid credentials"}
	}

	// Parse user ID
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, &handlers.AppError{Code: "uuid_error", Message: "Invalid user ID", Err: err}
	}

	timeNow := time.Now().UTC()

	// Update user status and generate tokens
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	err = queries.UpdateUserStatusByID(ctx, database.UpdateUserStatusByIDParams{
		ID:        user.ID,
		Provider:  LocalProvider,
		UpdatedAt: timeNow,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "update_user_error", Message: "Error updating user status", Err: err}
	}

	// Generate tokens and store refresh token
	authResult, err := s.generateAndStoreTokens(userID.String(), LocalProvider, timeNow, false)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return authResult, nil
}

// SignOut handles user logout by clearing refresh tokens
func (s *authServiceImpl) SignOut(ctx context.Context, userID string, provider string) error {
	// Delete refresh token from Redis
	err := s.redisClient.Del(ctx, RefreshTokenKeyPrefix+userID).Err()
	if err != nil {
		return &handlers.AppError{Code: "redis_error", Message: "Error deleting refresh token", Err: err}
	}

	return nil
}

// RefreshToken handles token refresh for both local and Google authentication
func (s *authServiceImpl) RefreshToken(ctx context.Context, userID string, provider string, refreshToken string) (*AuthResult, error) {
	timeNow := time.Now().UTC()

	if provider == "google" {
		return s.refreshGoogleToken(ctx, userID, refreshToken, timeNow)
	}

	return s.refreshLocalToken(ctx, userID, timeNow)
}

// GenerateGoogleAuthURL generates the Google OAuth authorization URL
func (s *authServiceImpl) GenerateGoogleAuthURL(state string) (string, error) {
	// Store state in Redis
	err := s.redisClient.Set(context.Background(), OAuthStateKeyPrefix+state, OAuthStateValid, OAuthStateTTL).Err()
	if err != nil {
		return "", &handlers.AppError{Code: "redis_error", Message: "Error storing state", Err: err}
	}

	authURL := s.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, nil
}

// HandleGoogleAuth handles the Google OAuth callback and user authentication
func (s *authServiceImpl) HandleGoogleAuth(ctx context.Context, code string, state string) (*AuthResult, error) {
	// Validate state
	redisState, err := s.redisClient.Get(ctx, OAuthStateKeyPrefix+state).Result()
	if err != nil || redisState != OAuthStateValid {
		return nil, &handlers.AppError{Code: "invalid_state", Message: "Invalid state parameter"}
	}

	// Exchange code for token
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, &handlers.AppError{Code: "token_exchange_error", Message: "Failed to exchange token", Err: err}
	}

	// Get user info from Google
	userInfo, err := s.getUserInfoFromGoogle(token)
	if err != nil {
		return nil, &handlers.AppError{Code: "google_api_error", Message: "Failed to get user info", Err: err}
	}

	// Handle user authentication
	return s.handleGoogleUserAuth(ctx, userInfo, token)
}

// Helper methods

// generateAndStoreTokens generates access and refresh tokens and stores the refresh token
func (s *authServiceImpl) generateAndStoreTokens(userID, provider string, timeNow time.Time, isNewUser bool) (*AuthResult, error) {
	accessTokenExpiresAt := timeNow.Add(AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(RefreshTokenTTL)

	accessToken, refreshToken, err := s.auth.GenerateTokens(userID, accessTokenExpiresAt)
	if err != nil {
		return nil, &handlers.AppError{Code: "token_generation_error", Message: "Error generating tokens", Err: err}
	}

	// Store refresh token
	err = s.auth.StoreRefreshTokenInRedis(nil, userID, refreshToken, provider, refreshTokenExpiresAt.Sub(timeNow))
	if err != nil {
		return nil, &handlers.AppError{Code: "redis_error", Message: "Error storing refresh token", Err: err}
	}

	return &AuthResult{
		UserID:              userID,
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		AccessTokenExpires:  accessTokenExpiresAt,
		RefreshTokenExpires: refreshTokenExpiresAt,
		IsNewUser:           isNewUser,
	}, nil
}

// refreshGoogleToken handles Google OAuth token refresh
func (s *authServiceImpl) refreshGoogleToken(ctx context.Context, userID, refreshToken string, timeNow time.Time) (*AuthResult, error) {
	tokenSource := s.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, &handlers.AppError{Code: "google_token_error", Message: "Failed to refresh Google token", Err: err}
	}

	refreshTokenExpiresAt := timeNow.Add(RefreshTokenTTL)

	return &AuthResult{
		UserID:              userID,
		AccessToken:         newToken.AccessToken,
		RefreshToken:        refreshToken,
		AccessTokenExpires:  newToken.Expiry,
		RefreshTokenExpires: refreshTokenExpiresAt,
		IsNewUser:           false,
	}, nil
}

// refreshLocalToken handles local authentication token refresh
func (s *authServiceImpl) refreshLocalToken(ctx context.Context, userID string, timeNow time.Time) (*AuthResult, error) {
	// Delete old refresh token
	err := s.redisClient.Del(ctx, RefreshTokenKeyPrefix+userID).Err()
	if err != nil {
		return nil, &handlers.AppError{Code: "redis_error", Message: "Error deleting old refresh token", Err: err}
	}

	// Generate new tokens and store refresh token
	return s.generateAndStoreTokens(userID, LocalProvider, timeNow, false)
}

// getUserInfoFromGoogle retrieves user information from Google API
func (s *authServiceImpl) getUserInfoFromGoogle(token *oauth2.Token) (*UserGoogleInfo, error) {
	client := s.oauth.Client(context.Background(), token)
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

// handleGoogleUserAuth handles Google OAuth user authentication and account creation
func (s *authServiceImpl) handleGoogleUserAuth(ctx context.Context, user *UserGoogleInfo, token *oauth2.Token) (*AuthResult, error) {
	existingUser, err := s.db.CheckExistsAndGetIDByEmail(ctx, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, &handlers.AppError{Code: "database_error", Message: "Error checking user existence", Err: err}
	}

	var userID string
	timeNow := time.Now().UTC()
	isNewUser := err == sql.ErrNoRows || !existingUser.Exists

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	if isNewUser {
		// Create new user
		userID = uuid.New().String()
		err = queries.CreateUser(ctx, database.CreateUserParams{
			ID:         userID,
			Name:       user.Name,
			Email:      user.Email,
			Password:   sql.NullString{},
			Provider:   "google",
			ProviderID: sql.NullString{String: user.ID, Valid: true},
			CreatedAt:  timeNow,
			UpdatedAt:  timeNow,
		})
		if err != nil {
			return nil, &handlers.AppError{Code: "create_user_error", Message: "Error creating user", Err: err}
		}
	} else {
		userID = existingUser.ID
	}

	// Generate access token
	accessToken, err := s.auth.GenerateAccessToken(userID, timeNow.Add(AccessTokenTTL))
	if err != nil {
		return nil, &handlers.AppError{Code: "token_generation_error", Message: "Error generating access token", Err: err}
	}

	// Handle refresh token
	refreshToken := token.RefreshToken
	if refreshToken == "" {
		return nil, &handlers.AppError{Code: "no_refresh_token", Message: "No refresh token provided by Google"}
	}

	// Store refresh token
	err = s.auth.StoreRefreshTokenInRedis(nil, userID, refreshToken, "google", RefreshTokenTTL)
	if err != nil {
		return nil, &handlers.AppError{Code: "redis_error", Message: "Error storing refresh token", Err: err}
	}

	// Update user signin status
	err = queries.UpdateUserSigninStatusByEmail(ctx, database.UpdateUserSigninStatusByEmailParams{
		UpdatedAt:  timeNow,
		Provider:   "google",
		ProviderID: sql.NullString{String: user.ID, Valid: true},
		Email:      user.Email,
	})
	if err != nil {
		return nil, &handlers.AppError{Code: "update_user_error", Message: "Error updating user status", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return &AuthResult{
		UserID:              userID,
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		AccessTokenExpires:  timeNow.Add(AccessTokenTTL),
		RefreshTokenExpires: timeNow.Add(RefreshTokenTTL),
		IsNewUser:           isNewUser,
	}, nil
}

// MergeCart merges a guest cart with a user's cart after authentication
// It retrieves the session ID from the request, gets the guest cart,
// merges it with the user's cart, and cleans up the guest cart
func (apicfg *HandlersAuthConfig) MergeCart(ctx context.Context, r *http.Request, userID string) {
	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID == "" {
		return
	}

	guestCart, err := apicfg.GetGuestCart(ctx, sessionID)
	if err != nil {
		// Log error but don't fail the authentication process
		apicfg.LogHandlerError(ctx, "merge_cart", "get_guest_cart_failed", "Failed to get guest cart", "", "", err)
		return
	}

	if len(guestCart.Items) == 0 {
		// No items to merge, just clean up the guest cart
		if err := apicfg.DeleteGuestCart(ctx, sessionID); err != nil {
			apicfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete empty guest cart", "", "", err)
		}
		return
	}

	// Merge guest cart items to user cart
	if err := apicfg.CartMG.MergeGuestCartToUser(ctx, userID, guestCart.Items); err != nil {
		apicfg.LogHandlerError(ctx, "merge_cart", "merge_cart_failed", "Failed to merge guest cart to user", "", "", err)
		return
	}

	// Clean up guest cart after successful merge
	if err := apicfg.DeleteGuestCart(ctx, sessionID); err != nil {
		apicfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete guest cart after merge", "", "", err)
	}
}
