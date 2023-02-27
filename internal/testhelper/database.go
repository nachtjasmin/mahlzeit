package testhelper

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database provides a working database if a valid database connection is provided with the TEST_DATABASE_DSN
// environment variable. If not, the test is marked as failed.
func Database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Fatalf("no database connection provided, please provide one using the %s environment variable", "TEST_DATABASE_DSN")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("database connection failed: %s", err)
	}

	return pool
}
