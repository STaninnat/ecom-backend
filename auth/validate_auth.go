package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func DecodeAndValidate[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	defer r.Body.Close()

	var params T
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Println("Decode error: ", err)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return nil, false
	}

	v := reflect.ValueOf(params)
	for i := range v.NumField() {
		if v.Field(i).Interface() == "" {
			log.Println("Invalid request format: missing fields")
			middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
			return nil, false
		}
	}

	return &params, true
}

func IsValidUserNameFormat(name string) bool {
	nameRegex := `^[a-zA-Z0-9]+([-._]?[a-zA-Z0-9]+)*$`

	re := regexp.MustCompile(nameRegex)

	return len(name) >= 3 && len(name) <= 30 && re.MatchString(name)
}

func IsValidEmailFormat(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	re := regexp.MustCompile(emailRegex)

	return re.MatchString(email)
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (cfg *AuthConfig) ValidateAccessToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("invalid issuer: got '%s'", claims.Issuer)
	}

	if !slices.Contains(claims.Audience, cfg.Audience) {
		return nil, fmt.Errorf("invalid audience: got '%s'", claims.Audience)
	}

	timeNow := time.Now().UTC()
	if claims.ExpiresAt.Time.Before(timeNow) {
		return nil, fmt.Errorf("token expired")
	}

	if claims.NotBefore.Time.After(timeNow) {
		return nil, fmt.Errorf("token not valid yet")
	}

	return claims, nil
}

func (cfg *AuthConfig) ValidateRefreshToken(refreshToken string) (uuid.UUID, error) {
	parts := strings.Split(refreshToken, ":")
	if len(parts) != 3 {
		return uuid.Nil, fmt.Errorf("invalid refresh token format")
	}

	userIDStr, rawUUID, signature := parts[0], parts[1], parts[2]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid userID in refresh token")
	}

	message := fmt.Sprintf("%s:%s", userIDStr, rawUUID)
	h := hmac.New(sha256.New, []byte(cfg.RefreshSecret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return uuid.Nil, fmt.Errorf("invalid refresh token signature")
	}

	return userID, nil
}
