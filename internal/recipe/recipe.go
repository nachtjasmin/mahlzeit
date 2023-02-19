package recipe

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"github.com/jackc/pgtype"
)

type Handler struct {
	app *app.Application
}

func (h *Handler) GetAllRecipes(ctx context.Context) ([]ListEntry, error) {
	dbResult, err := h.app.Queries.GetAllRecipesByName(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching recipes from database: %w", err)
	}

	var res []ListEntry
	for _, row := range dbResult {
		res = append(res, ListEntry{
			ID:   int(row.ID),
			Name: row.Name,
		})
	}

	return res, nil
}

// GetSingleRecipe returns a recipe by its ID. Optionally, the desired amount of servings
// can be provided. Any value less or equal to zero is ignored.
func (h *Handler) GetSingleRecipe(ctx context.Context, id, servings int) (*Recipe, error) {
	// TODO: execute the following queries in a transaction
	base, err := h.app.Queries.GetRecipeByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying recipe %d: %w", id, err)
	}

	res := &Recipe{
		ID:                  int(base.ID),
		Name:                base.Name,
		Description:         base.Description,
		CreatedAt:           base.CreatedAt,
		UpdatedAt:           base.UpdatedAt.Time,
		Source:              base.Source.String,
		Servings:            int(base.Servings),
		ServingsDescription: base.ServingsDescription,
	}
	_ = base.WorkingTime.AssignTo(&res.WorkingTime)
	_ = base.WaitingTime.AssignTo(&res.WaitingTime)

	ingredients, err := h.app.Queries.GetTotalIngredientsForRecipe(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying ingredients for recipe %d: %w", id, err)
	}

	for _, ingredient := range ingredients {
		res.Ingredients = append(res.Ingredients, Ingredient{
			Name:   ingredient.Name,
			Amount: float64(ingredient.TotalAmount),
		})
	}

	steps, err := h.app.Queries.GetStepsForRecipeByID(ctx, int64(id))
	if err != nil {
		return nil, fmt.Errorf("querying steps for recipe %d: %w", id, err)
	}

	for _, step := range steps {
		s := Step{Instruction: step.Instruction}
		_ = step.StepTime.AssignTo(&s.Time)

		if step.Ingredients.Status == pgtype.Present {
			if err := step.Ingredients.AssignTo(&s.Ingredients); err != nil {
				return nil, fmt.Errorf("scanning ingredients into struct: %w", err)
			}
		}

		// Because we might have empty ingredients, filter those out.
		// TODO: can we get rid of them in the SQL query?
		var ingredients []Ingredient
		for _, i := range s.Ingredients {
			if i.Name != "" {
				ingredients = append(ingredients, i)
			}
		}
		s.Ingredients = ingredients

		res.Steps = append(res.Steps, s)
	}

	// Let's calculate the actual amounts for the ingredients
	if servings > 0 {
		baseServings, newServings := float64(res.Servings), float64(servings)
		res.Servings = servings

		for i, ingredient := range res.Ingredients {
			ingredient.Amount = ingredient.Amount / baseServings * newServings
			res.Ingredients[i] = ingredient
		}
		for i, step := range res.Steps {
			for j, ingredient := range step.Ingredients {
				ingredient.Amount = ingredient.Amount / baseServings * newServings
				step.Ingredients[j] = ingredient
			}
			res.Steps[i] = step
		}
	}

	return res, nil
}

type ListEntry struct {
	ID   int
	Name string
}

type Recipe struct {
	ID                  int
	Name                string
	Description         string
	WorkingTime         time.Duration
	WaitingTime         time.Duration
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Source              string
	Servings            int
	ServingsDescription string

	Ingredients []Ingredient
	Steps       []Step
}
type Ingredient struct {
	Name   string
	Amount float64
	Note   string
}
type Step struct {
	Instruction string
	Time        time.Duration
	Ingredients []Ingredient
}
