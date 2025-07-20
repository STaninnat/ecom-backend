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
	LocalProvider  = "local"
	GoogleProvider = "google"

	// User roles
	UserRole = "user"

	// Redis key prefixes
	RefreshTokenKeyPrefix = "refresh_token:"
	OAuthStateKeyPrefix   = "oauth_state:"

	// OAuth state value
	OAuthStateValid = "valid"
)

// AuthService defines the business logic interface for authentication.
// Provides methods for signup, signin, signout, token refresh, Google OAuth, and auth URL generation.
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

// UserGoogleInfo represents user information retrieved from Google OAuth
type UserGoogleInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

// DBQueries defines the interface for database query operations needed by AuthServiceImpl.
type DBQueries interface {
	CheckUserExistsByName(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	UpdateUserStatusByID(ctx context.Context, params database.UpdateUserStatusByIDParams) error
	WithTx(tx DBTx) DBQueries
	CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error)
	UpdateUserSigninStatusByEmail(ctx context.Context, params database.UpdateUserSigninStatusByEmailParams) error
}

// DBConn defines the interface for database connection operations needed by AuthServiceImpl.
type DBConn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (DBTx, error)
}

// DBTx defines the interface for database transaction operations needed by AuthServiceImpl.
type DBTx interface {
	Commit() error
	Rollback() error
}

// MinimalRedis defines the minimal Redis operations needed by AuthServiceImpl.
type MinimalRedis interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

// AuthConfig defines the interface for authentication configuration and token operations needed by AuthServiceImpl.
type AuthConfig interface {
	HashPassword(password string) (string, error)
	GenerateTokens(userID string, expiresAt time.Time) (string, string, error)
	StoreRefreshTokenInRedis(ctx context.Context, userID, refreshToken, provider string, ttl time.Duration) error
	GenerateAccessToken(userID string, expiresAt time.Time) (string, error)
}

// OAuth2Exchanger abstracts all OAuth2 operations needed by authServiceImpl
type OAuth2Exchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

// AuthServiceImpl implements AuthService, providing authentication business logic.
type AuthServiceImpl struct {
	db          DBQueries
	dbConn      DBConn
	auth        AuthConfig
	redisClient MinimalRedis
	oauth       OAuth2Exchanger
}

// NewAuthService creates a new AuthService instance with the given dependencies.
func NewAuthService(
	db DBQueries,
	dbConn DBConn,
	auth AuthConfig,
	redisClient MinimalRedis,
	oauth OAuth2Exchanger,
) AuthService {
	return &AuthServiceImpl{
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

// SignUp handles user registration with local authentication.
func (s *AuthServiceImpl) SignUp(ctx context.Context, params SignUpParams) (*AuthResult, error) {
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
	hashedPassword, err := s.auth.HashPassword(params.Password)
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
	authResult, err := s.generateAndStoreTokens(ctx, userID.String(), LocalProvider, timeNow, true)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return authResult, nil
}

// SignIn handles user authentication with local credentials.
func (s *AuthServiceImpl) SignIn(ctx context.Context, params SignInParams) (*AuthResult, error) {
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
	authResult, err := s.generateAndStoreTokens(ctx, userID.String(), LocalProvider, timeNow, false)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return authResult, nil
}

// SignOut handles user signout, revoking tokens and cleaning up session state.
func (s *AuthServiceImpl) SignOut(ctx context.Context, userID string, provider string) error {
	// Delete refresh token from Redis
	err := s.redisClient.Del(ctx, RefreshTokenKeyPrefix+userID).Err()
	if err != nil {
		return &handlers.AppError{Code: "redis_error", Message: "Error deleting refresh token", Err: err}
	}

	return nil
}

// RefreshToken handles refresh token logic, issuing new tokens for the user.
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, userID string, provider string, refreshToken string) (*AuthResult, error) {
	timeNow := time.Now().UTC()

	if provider == "google" {
		return s.refreshGoogleToken(ctx, userID, refreshToken, timeNow)
	}

	return s.refreshLocalToken(ctx, userID, timeNow)
}

// GenerateGoogleAuthURL generates the Google OAuth authorization URL for the given state.
func (s *AuthServiceImpl) GenerateGoogleAuthURL(state string) (string, error) {
	// Store state in Redis
	err := s.redisClient.Set(context.Background(), OAuthStateKeyPrefix+state, OAuthStateValid, OAuthStateTTL).Err()
	if err != nil {
		return "", &handlers.AppError{Code: "redis_error", Message: "Error storing state", Err: err}
	}

	authURL := s.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, nil
}

// HandleGoogleAuth processes the Google OAuth callback, exchanges code for tokens, and authenticates the user.
func (s *AuthServiceImpl) HandleGoogleAuth(ctx context.Context, code string, state string) (*AuthResult, error) {
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
	userInfo, err := s.getUserInfoFromGoogle(token, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, &handlers.AppError{Code: "google_api_error", Message: "Failed to get user info", Err: err}
	}

	// Handle user authentication
	return s.handleGoogleUserAuth(ctx, userInfo, token)
}

// Helper methods

// generateAndStoreTokens generates access and refresh tokens and stores the refresh token
func (s *AuthServiceImpl) generateAndStoreTokens(ctx context.Context, userID, provider string, timeNow time.Time, isNewUser bool) (*AuthResult, error) {
	accessTokenExpiresAt := timeNow.Add(AccessTokenTTL)
	refreshTokenExpiresAt := timeNow.Add(RefreshTokenTTL)

	accessToken, refreshToken, err := s.auth.GenerateTokens(userID, accessTokenExpiresAt)
	if err != nil {
		return nil, &handlers.AppError{Code: "token_generation_error", Message: "Error generating tokens", Err: err}
	}

	// Store refresh token
	err = s.auth.StoreRefreshTokenInRedis(ctx, userID, refreshToken, provider, refreshTokenExpiresAt.Sub(timeNow))
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
func (s *AuthServiceImpl) refreshGoogleToken(ctx context.Context, userID, refreshToken string, timeNow time.Time) (*AuthResult, error) {
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
func (s *AuthServiceImpl) refreshLocalToken(ctx context.Context, userID string, timeNow time.Time) (*AuthResult, error) {
	// Delete old refresh token
	err := s.redisClient.Del(ctx, RefreshTokenKeyPrefix+userID).Err()
	if err != nil {
		return nil, &handlers.AppError{Code: "redis_error", Message: "Error deleting old refresh token", Err: err}
	}

	// Generate new tokens and store refresh token
	return s.generateAndStoreTokens(ctx, userID, LocalProvider, timeNow, false)
}

// getUserInfoFromGoogle retrieves user information from Google API
func (s *AuthServiceImpl) getUserInfoFromGoogle(token *oauth2.Token, userInfoURL string, clientOpt ...*http.Client) (*UserGoogleInfo, error) {
	if userInfoURL == "" {
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	}
	var client *http.Client
	if len(clientOpt) > 0 && clientOpt[0] != nil {
		client = clientOpt[0]
	} else {
		client = s.oauth.Client(context.Background(), token)
	}
	resp, err := client.Get(userInfoURL)
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
func (s *AuthServiceImpl) handleGoogleUserAuth(ctx context.Context, user *UserGoogleInfo, token *oauth2.Token) (*AuthResult, error) {
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
	err = s.auth.StoreRefreshTokenInRedis(ctx, userID, refreshToken, "google", RefreshTokenTTL)
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

	// Get cart service from the embedded HandlersCartConfig
	cartService := apicfg.GetCartService()
	if cartService == nil {
		apicfg.LogHandlerError(ctx, "merge_cart", "cart_service_not_available", "Cart service not available", "", "", nil)
		return
	}

	guestCart, err := cartService.GetGuestCart(ctx, sessionID)
	if err != nil {
		// Log error but don't fail the authentication process
		apicfg.LogHandlerError(ctx, "merge_cart", "get_guest_cart_failed", "Failed to get guest cart", "", "", err)
		return
	}

	if len(guestCart.Items) == 0 {
		// No items to merge, just clean up the guest cart
		if err := cartService.DeleteGuestCart(ctx, sessionID); err != nil {
			apicfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete empty guest cart", "", "", err)
		}
		return
	}

	// For each item in the guest cart, add it to the user's cart
	for _, item := range guestCart.Items {
		if err := cartService.AddItemToUserCart(ctx, userID, item.ProductID, item.Quantity); err != nil {
			apicfg.LogHandlerError(ctx, "merge_cart", "add_item_to_user_cart_failed", "Failed to add item to user cart", "", "", err)
			return
		}
	}

	// Clean up guest cart after successful merge
	if err := cartService.DeleteGuestCart(ctx, sessionID); err != nil {
		apicfg.LogHandlerError(ctx, "merge_cart", "delete_guest_cart_failed", "Failed to delete guest cart after merge", "", "", err)
	}
}
