package recipe

import (
	"database/sql"
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
		r.Post("/{id}/edit/add_step", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				app.HandleClientError(w, http.StatusBadRequest, err)
				return
			}

			emptyStep, err := c.Queries.AddNewEmptyStep(r.Context(), int64(id))
			if err != nil {
				app.HandleServerError(w, err)
				return
			}

			s := Step{
				ID:          int(emptyStep.ID),
				RecipeID:    int(emptyStep.RecipeID),
				Instruction: emptyStep.Instruction,
				Ingredients: nil,
			}
			_ = emptyStep.Time.AssignTo(&s.Time)

			if htmx.IsHTMXRequest(r) {
				if err := c.Templates.RenderTemplate(w, "recipes/edit.tmpl", "single_step", s); err != nil {
					app.HandleServerError(w, err)
					return
				}
			} else {
				http.Redirect(w, r, "", http.StatusFound)
			}
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
			r.Post("/add_ingredient", func(w http.ResponseWriter, r *http.Request) {
				data := struct {
					Ingredients []Ingredient
					Units       []Unit
					RecipeID    int
					StepID      int
				}{}

				ingredients, err := c.Queries.GetAllIngredients(r.Context())
				if err != nil {
					app.HandleServerError(w, err)
					return
				}

				units, err := c.Queries.GetAllUnits(r.Context())
				if err != nil {
					app.HandleServerError(w, err)
					return
				}

				for _, i := range ingredients {
					data.Ingredients = append(data.Ingredients, Ingredient{
						ID:   int(i.ID),
						Name: i.Name,
					})
				}
				for _, u := range units {
					data.Units = append(data.Units, Unit{
						ID:   int(u.ID),
						Name: u.Name,
					})
				}

				idStr := chi.URLParam(r, "stepID")
				stepID, err := strconv.Atoi(idStr)
				if err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}

				recipeIDStr := chi.URLParam(r, "id")
				recipeID, err := strconv.Atoi(recipeIDStr)
				if err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}

				data.RecipeID = recipeID
				data.StepID = stepID

				if htmx.IsHTMXRequest(r) {
					if err := c.Templates.RenderTemplate(w, "recipes/edit.tmpl", "new_ingredient", data); err != nil {
						app.HandleServerError(w, err)
						return
					}
				} else {
					panic("progressive enhancement not yet implemented")
				}
			})
			r.Post("/ingredients", func(w http.ResponseWriter, r *http.Request) {
				idStr := chi.URLParam(r, "stepID")
				stepID, err := strconv.Atoi(idStr)
				if err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}

				if err := r.ParseForm(); err != nil {
					app.HandleClientError(w, http.StatusBadRequest, err)
					return
				}
				data := struct {
					Ingredient int
					Amount     int
					Unit       int
					Note       string
				}{
					Ingredient: parseIntWithDefault(r.PostFormValue("Ingredient")),
					Amount:     parseIntWithDefault(r.PostFormValue("Amount")),
					Unit:       parseIntWithDefault(r.PostFormValue("Unit")),
					Note:       r.PostFormValue("Note"),
				}
				var (
					amount pgtype.Numeric
					unit   sql.NullInt64
				)
				if data.Unit > 0 {
					unit = sql.NullInt64{
						Int64: int64(data.Unit),
						Valid: true,
					}
				}
				_ = amount.Set(data.Amount)

				if err := c.Queries.AddIngredientToStep(r.Context(), queries.AddIngredientToStepParams{
					StepID:        int64(stepID),
					IngredientsID: int64(data.Ingredient),
					UnitID:        unit,
					Amount:        amount,
					Note:          data.Note,
				}); err != nil {
					app.HandleServerError(w, err)
					return
				}

				ingredientName, err := c.Queries.GetIngredientNameByID(r.Context(), int64(data.Ingredient))
				if err != nil {
					app.HandleServerError(w, err)
					return
				}

				if htmx.IsHTMXRequest(r) {
					if err := c.Templates.RenderTemplate(w, "recipes/edit.tmpl", "ingredient", map[string]any{
						"Name":   ingredientName,
						"Amount": data.Amount,
						"Note":   data.Note,
					}); err != nil {
						app.HandleServerError(w, err)
						return
					}
				} else {
					panic("progressive enhancement not yet implemented")
				}
			})
		})
	}
}

func parseIntWithDefault(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
