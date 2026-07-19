package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := flag.String("email", "", "email of the user whose PIN should be reset")
	pin := flag.String("pin", "", "new PIN value")
	databaseURL := flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
	flag.Parse()

	*email = strings.TrimSpace(*email)
	*pin = strings.TrimSpace(*pin)

	if *databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL or --database-url is required")
		os.Exit(1)
	}
	if *email == "" {
		fmt.Fprintln(os.Stderr, "--email is required")
		os.Exit(1)
	}
	if !strings.Contains(*email, "@") || len(*email) > 255 {
		fmt.Fprintln(os.Stderr, "--email must be a valid email address")
		os.Exit(1)
	}
	if !validPIN(*pin) {
		fmt.Fprintln(os.Stderr, "--pin must be 4 to 6 digits")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, *databaseURL)
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
		ctx,
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

func validPIN(pin string) bool {
	if len(pin) < 4 || len(pin) > 6 {
		return false
	}
	for _, r := range pin {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
