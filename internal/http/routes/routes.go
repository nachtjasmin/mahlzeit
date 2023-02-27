package routes

import (
	"net/http"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/http/httpreq"
	"codeberg.org/mahlzeit/mahlzeit/internal/zaphelper"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// All returns the [chi.Mux] that is going to be used for our HTTP handlers.
// It's extracted into this function to see quickly which routes exist and where
// they are registered.
func All(c *app.Application) *chi.Mux {
	r := chi.NewMux()

	// The default middleware stack
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		zaphelper.InjectLogger(c.Logger),
		zaphelper.RequestLogger(),
		middleware.Recoverer,
		middleware.CleanPath,
	)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Add a static file server for the assets.
	// In the future, those assets should be embedded into the binary to simplify the deployment.
	fileServer := http.FileServer(http.Dir("./web/static/"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	w := appWrapper{c}
	r.Route("/recipes", func(r chi.Router) {
		r.Get("/", w.getAllRecipes)
		r.Route("/{id}", func(r chi.Router) {
			// Add simple middleware to validate the recipe ID.
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, err := httpreq.IDParam(r, "id")
					if err != nil {
						app.HandleClientError(w, r, err, http.StatusBadRequest)
						return
					}

					next.ServeHTTP(w, r)
				})
			})

			r.Get("/", w.getSingleRecipe)
			r.Get("/edit", w.getEditSingleRecipe)
			r.Post("/edit", w.postEditSingleRecipe)
			r.Post("/edit/add_step", w.postAddStepToRecipe)
			r.Route("/steps/{stepID}", func(r chi.Router) {
				r.Post("/", w.postNewRecipeStep)
				r.Delete("/", w.deleteRecipeStep)
				r.Post("/add_ingredient", w.postAddNewRecipeStepIngredient)
				r.Post("/ingredients", w.postAddRecipeStepIngredient)
				r.Delete("/ingredients/{ingredientID}", w.deleteRecipeStepIngredient)
			})
		})
	})

	return r
}

// appWrapper is the struct that all HTTP handlers should attach to.
// Example:
//
//	func (w *appWrapper) getAllEntities(w http.ResponseWriter, r *http.Request) {...}
type appWrapper struct{ app *app.Application }
