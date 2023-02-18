package recipe

import (
	"net/http"
	"strconv"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"github.com/go-chi/chi/v5"
)

func ChiHandler(c *app.Application) func(r chi.Router) {
	h := Handler{app: c}

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			recipes, err := h.GetAllRecipes(r.Context())
			if err != nil {
				app.HandleServerError(w, err)
				return
			}
			if err := c.Templates.Render(w, "recipes/index.tmpl", recipes); err != nil {
				app.HandleServerError(w, err)
				return
			}
		})
	}
}
