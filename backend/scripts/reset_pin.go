package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := flag.String("email", "cashier@dashpoint.local", "email of the user whose PIN should be reset")
	pin := flag.String("pin", "1234", "new PIN value")
	databaseURL := flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.Parse()

	if *databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL or --database-url is required")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), *databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	pinHash, err := bcrypt.GenerateFromPassword([]byte(*pin), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Hash failed: %v\n", err)
		os.Exit(1)
	}

	result, err := pool.Exec(
		context.Background(),
		"UPDATE users SET pin_hash = $1, updated_at = NOW() WHERE email = $2",
		string(pinHash),
		*email,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %d row(s) - %s PIN has been reset\n", result.RowsAffected(), *email)
}
