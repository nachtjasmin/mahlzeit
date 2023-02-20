package app

import (
	"net/http"

	"codeberg.org/mahlzeit/mahlzeit/internal/zaphelpers"
	"go.uber.org/zap"
)

// HandleClientError handles all errors that occurred due to the users fault.
// This can be anything from a [http.StatusBadRequest] up to a [http.StatusUnprocessableEntity].
// Errors are logged for debugging.
func HandleClientError(w http.ResponseWriter, r *http.Request, err error, httpStatusCode int) {
	w.WriteHeader(httpStatusCode)
	zaphelpers.FromRequest(r).
		Info("client-side error", zap.Error(err))
}

// HandleServerError handles server-side errors with a [http.StatusInternalServerError].
// Errors are logged for debugging.
func HandleServerError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	zaphelpers.FromRequest(r).
		Info("server-side error", zap.Error(err))
}
