package httpreq

import (
	"net/http"
	"strconv"

	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi/v5"
)

// IDParam calls [chi.URLParam] and if the parameter is set,
// returns the value as int. If the value is not an int or not positive, an error with a helpful message is returned.
func IDParam(r *http.Request, name string) (int, error) {
	res, err := strconv.Atoi(chi.URLParam(r, name))
	if err != nil {
		return 0, resperr.WithUserMessagef(err, "invalid ID in parameter %q provided", name)
	}

	if res <= 0 {
		return 0, resperr.WithUserMessagef(err, "invalid ID in parameter %q provided", name)
	}

	return res, nil
}

// MustIDParam calls [chi.URLParam] and if the parameter is set,
// returns the value as int. If the value is not an int or not positive, it panics.
func MustIDParam(r *http.Request, name string) int {
	res, err := IDParam(r, name)
	if err != nil {
		panic(err)
	}
	return res
}
