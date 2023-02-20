package main

import (
	"net/http"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/recipe"
	"codeberg.org/mahlzeit/mahlzeit/internal/zaphelpers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes returns the [chi.Mux] that is going to be used for our HTTP handlers.
// It's extracted into this function to see quickly which routes exist and where
// they are registered.
func routes(c *app.Application) *chi.Mux {
	r := chi.NewMux()

	// The default middleware stack
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		zaphelpers.InjectLogger(c.Logger),
		zaphelpers.RequestLogger(),
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

	r.Route("/recipes", recipe.ChiHandler(c))

	return r
}
