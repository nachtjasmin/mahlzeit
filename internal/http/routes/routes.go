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
		stripMultipleQueryParameters,
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
		r.Get("/", errorWrapper(w.getAllRecipes))
		r.Route("/{id}", func(r chi.Router) {
			r.Use(validateID("id"))

			r.Get("/", errorWrapper(w.getSingleRecipe))
			r.Get("/edit", errorWrapper(w.getEditSingleRecipe))
			r.Post("/edit", errorWrapper(w.postEditSingleRecipe))
			r.Post("/edit/add_step", errorWrapper(w.postAddStepToRecipe))
			r.Route("/steps/{stepID}", func(r chi.Router) {
				r.Post("/", errorWrapper(w.postNewRecipeStep))
				r.Delete("/", errorWrapper(w.deleteRecipeStep))
				r.Post("/add_ingredient", errorWrapper(w.postAddNewRecipeStepIngredient))
				r.Post("/ingredients", errorWrapper(w.postAddRecipeStepIngredient))
				r.Delete("/ingredients/{ingredientID}", errorWrapper(w.deleteRecipeStepIngredient))
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

// ErrHandlerFunc is an adapted version of the http.HandlerFunc which allows to return an error.
// This is especially helpful to avoid the pattern of:
//
//	 err := someMethod()
//	 if err != nil {
//			app.HandleError(w, r, err)
//			return
//	 }
//
// Instead, the error can be returned, reducing the possibility of forgetting a return statement.
type ErrHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// errorWrapper takes a ErrHandlerFunc and forwards all errors to [app.HandleError].
func errorWrapper(fn ErrHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			app.HandleError(w, r, err)
		}
	}
}

// stripMultipleQueryParameters strips all query parameters that occur multiple times on a URL.
// Only the last query parameter is kept.
// This is implemented, because requests sent by HTMX can be erroneous with multiple params with the same name.
// For example, given the URL "localhost/?a=1&b=2&a=3", stripMultipleQueryParameters would remove the "a=1" from the URL.
func stripMultipleQueryParameters(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			handler.ServeHTTP(w, r)
			return
		}

		vals := r.URL.Query()
		for k, v := range vals {
			if len(v) <= 1 {
				continue
			}

			// Set the value to the last element
			vals.Set(k, v[len(v)-1])
		}
		r.URL.RawQuery = vals.Encode()

		handler.ServeHTTP(w, r)
	})
}

// validateID gets the route parameter associated with idParam and validates
// whether it's a valid ID or not, as determined by [httpreq.IDParam].
func validateID(idParam string) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := httpreq.IDParam(r, idParam)
			if err != nil {
				app.HandleError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
