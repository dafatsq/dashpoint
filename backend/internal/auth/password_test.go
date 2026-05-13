package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !CheckPassword("password123", hash) {
		t.Fatal("expected password hash to validate")
	}
}

func TestHashTokenIsDeterministic(t *testing.T) {
	first := HashToken("refresh-token")
	second := HashToken("refresh-token")

	if first != second {
		t.Fatal("expected HashToken to be deterministic")
	}
}
