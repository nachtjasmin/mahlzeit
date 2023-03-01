package app

import (
	"context"
	"testing"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"codeberg.org/mahlzeit/mahlzeit/internal/testhelper"
	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap/zaptest"
)

type testApplication struct {
	t *testing.T
	*Application
	DB *pgxpool.Pool
}

// newApp returns a new application that logs to t and has a database attached to it.
func newApp(t *testing.T) testApplication {
	t.Helper()

	db := testhelper.Database(t)
	return testApplication{
		t: t,
		Application: &Application{
			Queries: queries.New(db),
			Logger:  zaptest.NewLogger(t),
		},
		DB: db,
	}
}

// AddEmptyRecipe adds a new empty recipe to the database.
func (app *testApplication) AddEmptyRecipe(ctx context.Context) Recipe {
	app.t.Helper()

	var t pgtype.Interval
	_ = t.Set(time.Minute * 10)

	userID, err := app.Queries.AddDemoUser(ctx)
	assert.NoError(app.t, err)

	params := queries.AddRecipeParams{
		Name:        testhelper.RandomString(20),
		Description: app.t.Name(),
		Servings:    2,
		WaitingTime: t,
		WorkingTime: t,
		CreatedBy:   userID,
	}
	recipe, err := app.Queries.AddRecipe(ctx, params)
	assert.NoError(app.t, err)

	return Recipe{
		ID:           int(recipe.ID),
		Name:         params.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		BaseServings: 2,
		Servings:     2,
	}
}
