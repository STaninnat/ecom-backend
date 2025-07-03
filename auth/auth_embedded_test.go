package auth

import (
	"encoding/json"
	"testing"
)

func TestClaimsJSONTags(t *testing.T) {
	claims := Claims{UserID: "user1"}
	b, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) == "" || !json.Valid(b) {
		t.Error("invalid JSON for Claims")
	}
}

func TestRefreshTokenDataJSONTags(t *testing.T) {
	data := RefreshTokenData{Token: "tok", Provider: "local"}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) == "" || !json.Valid(b) {
		t.Error("invalid JSON for RefreshTokenData")
	}
}
