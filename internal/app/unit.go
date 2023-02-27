package app

import (
	"context"
	"fmt"
)

func (app *Application) GetAllUnits(ctx context.Context) ([]Unit, error) {
	units, err := app.Queries.GetAllUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("querying all units: %w", err)
	}

	var res []Unit
	for _, u := range units {
		res = append(res, Unit{
			ID:   int(u.ID),
			Name: u.Name,
		})
	}

	return res, nil
}
