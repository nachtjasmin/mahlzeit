package recipe

import (
	"net/http"
	"strconv"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/bind"
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
		r.Get("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
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

			if err := c.Templates.Render(w, "recipes/edit.tmpl", res); err != nil {
				app.HandleServerError(w, err)
				return
			}
		})
		r.Post("/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				app.HandleClientError(w, http.StatusBadRequest, err)
				return
			}

			if err := r.ParseForm(); err != nil {
				app.HandleClientError(w, http.StatusBadRequest, err)
				return
			}

			data := struct {
				Name        string
				Servings    int
				Description string
			}{}
			if err := bind.Request(r).All(&data); err != nil {
				app.HandleClientError(w, http.StatusBadRequest, err)
				return
			}

			// TODO: Add input validation
			// TODO: Refactor into service
			if err := c.Queries.UpdateBasicRecipeInformation(r.Context(), queries.UpdateBasicRecipeInformationParams{
				ID:          int64(id),
				Name:        data.Name,
				Servings:    int32(data.Servings),
				Description: data.Description,
			}); err != nil {
				app.HandleServerError(w, err)
				return
			}

			http.Redirect(w, r, "/recipes/"+idStr, http.StatusFound)
		})
	}
}
