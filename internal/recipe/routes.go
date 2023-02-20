package recipe

import (
	"net/http"
	"strconv"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/htmx"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgtype"
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
			if err := c.Templates.RenderPage(w, "recipes/index.tmpl", recipes); err != nil {
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

			if err := c.Templates.RenderPage(w, "recipes/single.tmpl", res); err != nil {
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

			if err := c.Templates.RenderPage(w, "recipes/edit.tmpl", res); err != nil {
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
		r.Route("/{id}/steps/{stepID}", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				idStr := chi.URLParam(r, "stepID")
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
					Instruction string
					Time        string
				}{}
				if err := bind.Request(r).Field(&data.Instruction, "instruction"); err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}
				if err := bind.Request(r).Field(&data.Time, "time"); err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}

				var pgTime pgtype.Interval
				dur, _ := time.ParseDuration(data.Time)
				_ = pgTime.Set(dur)
				if err := c.Queries.UpdateStepByID(r.Context(), queries.UpdateStepByIDParams{
					ID:          int64(id),
					Instruction: data.Instruction,
					Time:        pgTime,
				}); err != nil {
					app.HandleServerError(w, err)
					return
				}

				if htmx.IsHTMXRequest(r) {
					if err := c.Templates.RenderTemplate(w, "recipes/edit.tmpl", "single_step", Step{
						ID:          id,
						RecipeID:    0,
						Instruction: data.Instruction,
						Time:        dur,
						Ingredients: nil,
					}); err != nil {
						app.HandleServerError(w, err)
						return
					}
				} else {
					http.Redirect(w, r, "/recipes/"+chi.URLParam(r, "id"), http.StatusFound)
				}
			})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				idStr := chi.URLParam(r, "stepID")
				id, err := strconv.Atoi(idStr)
				if err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}

				if err := c.Queries.DeleteStepByID(r.Context(), int64(id)); err != nil {
					app.HandleServerError(w, err)
					return
				}

				w.WriteHeader(200)
			})
		})
	}
}
