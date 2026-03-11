package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := "postgres://dashpoint:dashpoint_dev@localhost:5432/dashpoint_dev?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	pinHash, err := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Hash failed: %v\n", err)
		os.Exit(1)
	}

	result, err := pool.Exec(context.Background(),
		"UPDATE users SET pin_hash = $1, updated_at = NOW() WHERE email = 'cashier@dashpoint.local'",
		string(pinHash))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Updated %d row - Cashier PIN is now 1234\n", result.RowsAffected())
}