package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
)

// AddStepToRecipe adds an empty step to the recipe and returns it.
func (app *Application) AddStepToRecipe(ctx context.Context, id int) (Step, error) {
	step, err := app.Queries.AddNewEmptyStep(ctx, int64(id))
	if err != nil {
		return Step{}, fmt.Errorf("adding step in database for recipe %d: %w", id, err)
	}

	res := Step{
		ID:          int(step.ID),
		RecipeID:    int(step.RecipeID),
		Instruction: step.Instruction,
	}
	_ = step.Time.AssignTo(&res.Time)
	return res, nil
}

// UpdateStep updates an existing step.
func (app *Application) UpdateStep(ctx context.Context, s Step) error {
	var pgTime pgtype.Interval
	_ = pgTime.Set(s.Time)

	if err := app.Queries.UpdateStepByID(ctx, queries.UpdateStepByIDParams{
		ID:          int64(s.ID),
		Instruction: s.Instruction,
		Time:        pgTime,
	}); err != nil {
		return fmt.Errorf("updating step %d: %w", s.ID, err)
	}

	return nil
}

// DeleteRecipeStepByID deletes a recipe step by its ID.
// This is an idempotent action, if the step is already deleted, no error is returned.
func (app *Application) DeleteRecipeStepByID(ctx context.Context, id int) error {
	if err := app.Queries.DeleteStepByID(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting step %d: %w", id, err)
	}
	return nil
}

type AddIngredientToStepParams struct {
	StepID       int
	IngredientID int
	UnitID       *int
	Amount       float64
	Note         string
}

// AddIngredientToStep adds an ingredient to a step.
func (app *Application) AddIngredientToStep(ctx context.Context, params AddIngredientToStepParams) error {
	var amount pgtype.Numeric
	_ = amount.Set(params.Amount)

	if err := app.Queries.AddIngredientToStep(ctx, queries.AddIngredientToStepParams{
		StepID:        int64(params.StepID),
		IngredientsID: int64(params.IngredientID),
		UnitID:        int64(valueOrDefault(params.UnitID)),
		Amount:        amount,
		Note:          params.Note,
	}); err != nil {
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) {
			return fmt.Errorf("adding ingredient to step %d: %w", params.StepID, err)
		}

		switch pgerr.ConstraintName {
		case "step_ingredients_step_id_fkey":
			return resperr.WithCodeAndMessage(err, http.StatusNotFound, "step does not exist")
		case "step_ingredients_ingredients_id_fkey":
			return resperr.WithCodeAndMessage(err, http.StatusBadRequest, "ingredient does not exist")
		}

		return fmt.Errorf("adding ingredient to step %d: %w", params.StepID, err)
	}

	return nil
}

type DeleteIngredientFromStepParams struct {
	StepID       int
	IngredientID int
}

// DeleteIngredientFromStep deletes an ingredient from a step. The ingredient itself is unaffected.
// This action is idempotent, if the step is already deleted, no error is returned.
func (app *Application) DeleteIngredientFromStep(ctx context.Context, params DeleteIngredientFromStepParams) error {
	if err := app.Queries.DeleteIngredientFromStep(ctx, queries.DeleteIngredientFromStepParams{
		StepID:        int64(params.StepID),
		IngredientsID: int64(params.IngredientID),
	}); err != nil {
		return fmt.Errorf("deleting ingredient %d from step %d: %w", params.IngredientID, params.StepID, err)
	}

	return nil
}

// valueOrDefault returns the value for v or the default value of T otherwise.
func valueOrDefault[T comparable](v *T) T {
	var res T
	if v != nil {
		res = *v
	}
	return res
}
