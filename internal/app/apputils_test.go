package app

import (
	"testing"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"codeberg.org/mahlzeit/mahlzeit/internal/testhelper"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap/zaptest"
)

type testApplication struct {
	*Application
	DB *pgxpool.Pool
}

// newApp returns a new application that logs to t and has a database attached to it.
func newApp(t *testing.T) testApplication {
	t.Helper()

	db := testhelper.Database(t)
	return testApplication{
		Application: &Application{
			Queries: queries.New(db),
			Logger:  zaptest.NewLogger(t),
		},
		DB: db,
	}
}
