package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestJWTManagerValidatesAccessAndRefreshTokens(t *testing.T) {
	manager := NewJWTManager("secret", 60, 24)

	pair, err := manager.GenerateTokenPair(
		uuid.New(),
		"user@example.com",
		"User",
		uuid.New(),
		"manager",
	)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	if _, err := manager.ValidateAccessToken(pair.AccessToken); err != nil {
		t.Fatalf("ValidateAccessToken returned error: %v", err)
	}

	if _, err := manager.ValidateRefreshToken(pair.RefreshToken); err != nil {
		t.Fatalf("ValidateRefreshToken returned error: %v", err)
	}
}

func TestJWTManagerRejectsWrongTokenType(t *testing.T) {
	manager := NewJWTManager("secret", 60, 24)

	pair, err := manager.GenerateTokenPair(
		uuid.New(),
		"user@example.com",
		"User",
		uuid.New(),
		"manager",
	)
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	if _, err := manager.ValidateAccessToken(pair.RefreshToken); err == nil {
		t.Fatal("expected refresh token to be rejected as access token")
	}
}
