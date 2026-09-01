package database

// Demo deployment only (dashpoint-demo branch). The public demo must always
// have working, active accounts regardless of database state, so this branch
// upserts them on every backend start. The passwords ship in the public
// frontend bundle by design — never put real accounts here.

import (
	"context"
	"fmt"
	"time"

	"dashpoint/backend/internal/auth"

	"github.com/rs/zerolog/log"
)

const demoAccountPassword = "demo1234"

type demoAccount struct {
	email string
	name  string
	role  string
}

var demoAccounts = []demoAccount{
	{email: "demo-owner@dashpoint.local", name: "Demo Owner", role: "owner"},
	{email: "demo-manager@dashpoint.local", name: "Demo Manager", role: "manager"},
	{email: "demo-cashier@dashpoint.local", name: "Demo Cashier", role: "cashier"},
}

// EnsureDemoAccounts upserts the demo accounts so the public demo always has
// active logins.
func EnsureDemoAccounts(ctx context.Context, db *DB) error {
	hash, err := auth.HashPassword(demoAccountPassword)
	if err != nil {
		return fmt.Errorf("hash demo account password: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, account := range demoAccounts {
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO users (email, name, password_hash, role_id, is_active, updated_at)
			VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4), true, now())
			ON CONFLICT (email) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    is_active = true,
			    updated_at = now()
		`, account.email, account.name, hash, account.role); err != nil {
			return fmt.Errorf("ensure demo account %s: %w", account.email, err)
		}
		log.Info().Str("email", account.email).Str("role", account.role).Msg("Demo account ensured")
	}

	return nil
}
