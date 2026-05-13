package models

import (
	"testing"
	"time"
)

func TestRefreshTokenValidity(t *testing.T) {
	valid := &RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}
	if !valid.IsValid() {
		t.Fatalf("expected token to be valid")
	}

	expired := &RefreshToken{ExpiresAt: time.Now().Add(-time.Hour)}
	if expired.IsValid() {
		t.Fatalf("expected expired token to be invalid")
	}

	now := time.Now()
	revoked := &RefreshToken{ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &now}
	if revoked.IsValid() {
		t.Fatalf("expected revoked token to be invalid")
	}
}
