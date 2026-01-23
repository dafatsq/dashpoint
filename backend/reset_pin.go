package backend
package main



































}	fmt.Printf("✓ Updated %d row - Cashier PIN is now 1234\n", result.RowsAffected())	}		os.Exit(1)		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)	if err != nil {		string(pinHash))		"UPDATE users SET pin_hash = $1, updated_at = NOW() WHERE email = 'cashier@dashpoint.local'",	result, err := pool.Exec(context.Background(),	}		os.Exit(1)		fmt.Fprintf(os.Stderr, "Hash failed: %v\n", err)	if err != nil {	pinHash, err := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)	defer pool.Close()	}		os.Exit(1)		fmt.Fprintf(os.Stderr, "Unable to connect: %v\n", err)	if err != nil {	pool, err := pgxpool.New(context.Background(), dbURL)	dbURL := "postgres://dashpoint:dashpoint_dev@localhost:5432/dashpoint_dev?sslmode=disable"func main() {)	"golang.org/x/crypto/bcrypt"	"github.com/jackc/pgx/v5/pgxpool"	"os"	"fmt"	"context"import (