// Package zaphelpers provides several helpers for the usage with zap.
package zaphelpers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var (
	// The internal context key for the logger that can be fetched with [FromContext].
	contextKey = struct{}{}

	nopLogger = zap.NewNop()
)

// HTTPMiddleware defines the type for the usual net/http middleware.
type HTTPMiddleware func(handler http.Handler) http.Handler

// noopMiddleware is a middleware that does nothing.
var noopMiddleware HTTPMiddleware = func(next http.Handler) http.Handler { return next }

// InjectLogger returns a middleware that injects the given logger into each request.
func InjectLogger(logger *zap.Logger) HTTPMiddleware {
	if logger == nil {
		return noopMiddleware
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Attach several pre-known request fields to the logger.
			logger = logger.With(
				zap.String("proto", r.Proto),
				zap.String("path", r.URL.Path),
				zap.String("request_id", middleware.GetReqID(r.Context())),
			)

			ctx := context.WithValue(r.Context(), contextKey, logger)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs each HTTP request. It should be added to the middleware stack
// after InjectLogger.
func RequestLogger() HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := FromContext(r.Context())

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				reqLogger := logger.With(
					zap.Duration("duration", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("response_bytes", ww.BytesWritten()),
				)
				reqLogger.Info("HTTP request served")
			}()
			next.ServeHTTP(ww, r)
		})
	}
}

// FromContext extracts a logger from the context, if it's injected there.
// Otherwise, a no-op logger is returned.
func FromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(contextKey).(*zap.Logger)
	if !ok {
		return nopLogger
	}

	return logger
}

