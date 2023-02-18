package recipe

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
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

type ListEntry struct {
	ID   int
	Name string
}
}
