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
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				app.HandleClientError(w, http.StatusBadRequest, err)
				return
			}

			res, err := h.GetSingleRecipe(r.Context(), id)
			if err != nil {
				app.HandleServerError(w, err)
				return
			}

			if servingsParam := r.URL.Query().Get("servings"); servingsParam != "" {
				// We deliberately ignore any errors, and "handle" them by checking whether we have a valid int.
				p, _ := strconv.Atoi(servingsParam)
				res.WithServings(p)
			}

			if err := c.Templates.Render(w, "recipes/single.tmpl", res); err != nil {
				app.HandleServerError(w, err)
				return
			}
		})
	}
}
