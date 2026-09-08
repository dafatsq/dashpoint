package handlers

import (
	"strings"
	"testing"
)

func TestValidateUserPasswordEnforcesMinimumLength(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"one character", "x", true},
		{"seven characters", strings.Repeat("x", 7), true},
		{"exactly eight", strings.Repeat("x", 8), false},
		{"long", strings.Repeat("x", 64), false},
	}
	for _, tc := range cases {
		pw := tc.password
		if got := validateUserPassword(&pw, true); (got != "") != tc.wantErr {
			t.Fatalf("%s: wantErr=%v, got %q", tc.name, tc.wantErr, got)
		}
	}
}

func TestValidateUserPasswordOptionalSemanticsUnchanged(t *testing.T) {
	absent := (*string)(nil)
	if got := validateUserPassword(absent, false); got != "" {
		t.Fatalf("optional absent password must pass, got %q", got)
	}
	if got := validateUserPassword(absent, true); got == "" {
		t.Fatalf("required absent password must be rejected")
	}
}

func TestValidateUserPasswordMatchesSetupMinimum(t *testing.T) {
	// The setup path and every change path must share one minimum so a
	// password rejected at setup cannot be re-introduced later.
	pw := strings.Repeat("x", 7)
	if msg := validateUserPassword(&pw, true); msg == "" {
		t.Fatalf("expected shared minimum to reject 7-char password")
	}
	if userMinPasswordLen != setupMinPasswordLen {
		t.Fatalf("minimum mismatch: change paths %d vs setup %d", userMinPasswordLen, setupMinPasswordLen)
	}
}
