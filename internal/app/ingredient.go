package app

import (
	"context"
	"fmt"
)

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
		return Ingredient{}, fmt.Errorf("querying ingredient %d: %w", id, err)
	}

	return Ingredient{
		ID:   id,
		Name: name,
	}, nil
}
