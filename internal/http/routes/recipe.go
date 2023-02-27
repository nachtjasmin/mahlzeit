package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/db/queries"
	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/htmx"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/httpreq"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgtype"
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

	// TODO: Add input validation
	// TODO: Refactor into service
	if err := a.app.Queries.UpdateBasicRecipeInformation(r.Context(), queries.UpdateBasicRecipeInformationParams{
		ID:          int64(id),
		Name:        data.Name,
		Servings:    int32(data.Servings),
		Description: data.Description,
	}); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	http.Redirect(w, r, "/recipes/"+strconv.Itoa(id), http.StatusFound)
}
func (a appWrapper) postAddStepToRecipe(w http.ResponseWriter, r *http.Request) {
	id := httpreq.MustIDParam(r, "id")
	emptyStep, err := a.app.Queries.AddNewEmptyStep(r.Context(), int64(id))
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	s := app.Step{
		ID:          int(emptyStep.ID),
		RecipeID:    int(emptyStep.RecipeID),
		Instruction: emptyStep.Instruction,
		Ingredients: nil,
	}
	_ = emptyStep.Time.AssignTo(&s.Time)

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "single_step", s); err != nil {
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

	var pgTime pgtype.Interval
	dur, _ := time.ParseDuration(data.Time)
	_ = pgTime.Set(dur)
	if err := a.app.Queries.UpdateStepByID(r.Context(), queries.UpdateStepByIDParams{
		ID:          int64(id),
		Instruction: data.Instruction,
		Time:        pgTime,
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
	if err := a.app.Queries.DeleteStepByID(r.Context(), int64(id)); err != nil {
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

	ingredients, err := a.app.Queries.GetAllIngredients(r.Context())
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	units, err := a.app.Queries.GetAllUnits(r.Context())
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	for _, i := range ingredients {
		data.Ingredients = append(data.Ingredients, app.Ingredient{
			ID:   int(i.ID),
			Name: i.Name,
		})
	}
	for _, u := range units {
		data.Units = append(data.Units, app.Unit{
			ID:   int(u.ID),
			Name: u.Name,
		})
	}

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
	stepID, err := httpreq.IDParam(r, "stepID")
	if err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
		return
	}

	recipeID := httpreq.MustIDParam(r, "id")

	if err := r.ParseForm(); err != nil {
		app.HandleClientError(w, r, err, http.StatusBadRequest)
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

	if err := a.app.Queries.AddIngredientToStep(r.Context(), queries.AddIngredientToStepParams{
		StepID:        int64(stepID),
		IngredientsID: int64(data.Ingredient),
		UnitID:        unit,
		Amount:        amount,
		Note:          data.Note,
	}); err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	ingredientName, err := a.app.Queries.GetIngredientNameByID(r.Context(), int64(data.Ingredient))
	if err != nil {
		app.HandleServerError(w, r, err)
		return
	}

	if htmx.IsHTMXRequest(r) {
		if err := a.app.Templates.RenderTemplate(w, "recipes/edit.tmpl", "ingredient", app.Ingredient{
			Name:     ingredientName,
			Amount:   float64(data.Amount),
			Note:     data.Note,
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

	if err := a.app.Queries.DeleteIngredientFromStep(r.Context(), queries.DeleteIngredientFromStepParams{
		StepID:        int64(stepID),
		IngredientsID: int64(ingredientID),
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
