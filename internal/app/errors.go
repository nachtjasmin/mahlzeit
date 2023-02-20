package app

import (
	"log"
	"net/http"
)

// HandleClientError handles all errors that occurred due to the users fault.
// This can be anything from a [http.StatusBadRequest] up to a [http.StatusUnprocessableEntity].
// Errors are logged for debugging.
func HandleClientError(w http.ResponseWriter, err error, httpStatusCode int) {
	w.WriteHeader(httpStatusCode)
	log.Printf("client error: %s", err.Error())
}

// HandleServerError handles server-side errors with a [http.StatusInternalServerError].
// Errors are logged for debugging.
func HandleServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Printf("server error: %s", err.Error())
}
