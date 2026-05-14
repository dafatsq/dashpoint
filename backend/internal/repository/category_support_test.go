package repository

import (
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCategoryStatus(t *testing.T) {
	if got := normalizeCategoryStatus(""); got != "active" {
		t.Fatalf("expected default active, got %q", got)
	}
	if got := normalizeCategoryStatus("archived"); got != "archived" {
		t.Fatalf("expected archived, got %q", got)
	}
	if got := normalizeCategoryStatus("all"); got != "all" {
		t.Fatalf("expected all, got %q", got)
	}
	if got := normalizeCategoryStatus("weird"); got != "active" {
		t.Fatalf("expected invalid status to normalize to active, got %q", got)
	}
}

func TestCollectCategoryIDs(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	got := collectCategoryIDs([]uuid.UUID{id1, id2})
	if len(got) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(got))
	}
	if got[0] != id1 || got[1] != id2 {
		t.Fatalf("unexpected ids: %+v", got)
	}
}
