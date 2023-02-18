package main

import (
	"net/http"
	"time"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"codeberg.org/mahlzeit/mahlzeit/internal/recipe"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
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
		zapLoggingMiddleware(c.Logger),
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

func zapLoggingMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	if logger == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				reqLogger := logger.With(
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.String("request_id", middleware.GetReqID(r.Context())),
					zap.Duration("duration", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("response_bytes", ww.BytesWritten()),
				)
				reqLogger.Info("HTTP request served")
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
