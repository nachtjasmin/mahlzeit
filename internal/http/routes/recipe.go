package routes

import (
	"net/http"
	"strconv"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/htmx"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/httpreq"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/bind"
)

func (a appWrapper) getAllRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := a.app.GetAllRecipes(r.Context())
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}
	if err := a.app.Templates.RenderPage(w, "recipes/index.tmpl", recipes); err != nil {
		app.HandleServerError(w, r, err)
		return
	}
}

func (a appWrapper) getSingleRecipe(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	res, err := a.app.GetSingleRecipe(r.Context(), id)
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if servingsParam := r.URL.Query().Get("servings"); servingsParam != "" {
		// We deliberately ignore any errors, and "handle" them by checking whether we have a valid int.
		p, _ := strconv.Atoi(servingsParam)
		res.WithServings(p)
	}

	if err := a.app.Templates.RenderPage(w, "recipes/single.tmpl", res); err != nil {
		app.HandleServerError(w, r, err)
		return
	}
}

func (a appWrapper) getEditSingleRecipe(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")

	res, err := a.app.GetSingleRecipe(r.Context(), id)
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if err := a.app.Templates.RenderPage(w, "recipes/edit.tmpl", res); err != nil {
		app.HandleServerError(w, r, err)
		return
	}
}

func (a appWrapper) postEditSingleRecipe(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	if err := r.ParseForm(); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	data := struct {
		Name        string
		Servings    int
		Description string
	}{}
	if err := bind.Request(r).All(&data); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := a.app.UpdateRecipe(r.Context(), &app.Recipe{
		ID:           id,
		Name:         data.Name,
		Description:  data.Description,
		BaseServings: data.Servings,
		Servings:     data.Servings,
	}); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/recipes/"+strconv.Itoa(id), http.StatusFound)
}

func (a appWrapper) postAddStepToRecipe(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	step, err := a.app.AddStepToRecipe(r.Context(), id)
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "single_step", step); err != nil {
			app.HandleServerError(w, r, err)
			return
		}
	} else {
		http.Redirect(w, r, "", http.StatusFound)
	}
}

func (a appWrapper) postNewRecipeStep(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	if err := r.ParseForm(); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	data := struct {
		Instruction string
		Time        string
	}{}
	if err := bind.Request(r).Field(&data.Instruction, "instruction"); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}
	if err := bind.Request(r).Field(&data.Time, "time"); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	dur, _ := time.ParseDuration(data.Time)
	if err := a.app.UpdateStep(r.Context(), app.Step{
		ID:          id,
		Instruction: data.Instruction,
		Time:        dur,
	}); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "single_step", app.Step{
			ID:          id,
			RecipeID:    0,
			Instruction: data.Instruction,
			Time:        dur,
			Ingredients: nil,
		}); err != nil {
			app.HandleServerError(w, r, err)
			return
		}
	} else {
		http.Redirect(w, r, "/recipes/"+chi.URLParam(r, "id"), http.StatusFound)
	}
}

func (a appWrapper) deleteRecipeStep(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	if err := a.app.DeleteRecipeStepByID(r.Context(), id); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	w.WriteHeader(200)
}

func (a appWrapper) postAddNewRecipeStepIngredient(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Ingredients []app.Ingredient
		Units       []app.Unit
		RecipeID    int
		StepID      int
	}{}

	ingredients, err := a.app.GetAllIngredients(r.Context())
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	units, err := a.app.GetAllUnits(r.Context())
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	data.Ingredients = ingredients
	data.Units = units

	stepID, err := httpreq.IDParam(r, "stepID")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}
	recipeID, err := httpreq.IDParam(r, "id")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	data.RecipeID = recipeID
	data.StepID = stepID

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "new_ingredient", data); err != nil {
			app.HandleServerError(w, r, err)
			return
		}
	} else {
		panic("progressive enhancement not yet implemented")
	}
}

func (a appWrapper) postAddRecipeStepIngredient(w http.ResponseWriter, r *http.Request) {
	recipeID := httpreq.MustIDParam(r, "id")
	stepID, err := httpreq.IDParam(r, "stepID")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	params := app.AddIngredientToStepParams{
		StepID:       stepID,
		IngredientID: parseIntWithDefault(r.PostFormValue("Ingredient")),
		Amount:       float64(parseIntWithDefault(r.PostFormValue("Amount"))),
		Note:         r.PostFormValue("Note"),
	}

	if unit := parseIntWithDefault(r.PostFormValue("Unit")); unit > 0 {
		params.UnitID = &unit
	}

	if err := a.app.AddIngredientToStep(r.Context(), params); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	ingredient, err := a.app.GetIngredient(r.Context(), params.IngredientID)
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "ingredient", app.Ingredient{
			Name:     ingredient.Name,
			Amount:   params.Amount,
			Note:     params.Note,
			StepID:   stepID,
			RecipeID: recipeID,
		}); err != nil {
			app.HandleServerError(w, r, err)
			return
		}
	} else {
		panic("progressive enhancement not yet implemented")
	}
}

func (a appWrapper) deleteRecipeStepIngredient(w http.ResponseWriter, r *http.Request) {
	stepID, err := httpreq.IDParam(r, "stepID")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	ingredientID, err := httpreq.IDParam(r, "ingredientID")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	if err := a.app.DeleteIngredientFromStep(r.Context(), app.DeleteIngredientFromStepParams{
		StepID:       stepID,
		IngredientID: ingredientID,
	}); err != nil {
		app.HandleServerError(w, r, err)
		return
	}
}

func parseIntWithDefault(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
