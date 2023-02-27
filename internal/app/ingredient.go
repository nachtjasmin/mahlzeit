package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgx/v4"
)

// AddIngredient adds a new ingredient.
func (app *Application) AddIngredient(ctx context.Context, name string) (Ingredient, error) {
	id, err := app.Queries.AddIngredient(ctx, name)
	if err != nil {
		return Ingredient{}, fmt.Errorf("adding new ingredient: %w", err)
	}

	return Ingredient{
		ID:   int(id),
		Name: name,
	}, nil
}

func (app *Application) GetAllIngredients(ctx context.Context) ([]Ingredient, error) {
	ingredients, err := app.Queries.GetAllIngredients(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all ingredients: %w", err)
	}

	var res []Ingredient
	for _, i := range ingredients {
		res = append(res, Ingredient{
			ID:   int(i.ID),
			Name: i.Name,
		})
	}

	return res, nil
}

func (app *Application) GetIngredient(ctx context.Context, id int) (Ingredient, error) {
	name, err := app.Queries.GetIngredientNameByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Ingredient{}, resperr.New(http.StatusNotFound, "ingredient %d not found", id)
		}
		return Ingredient{}, fmt.Errorf("querying ingredient %d: %w", id, err)
	}

	return Ingredient{
		ID:   id,
		Name: name,
	}, nil
}
