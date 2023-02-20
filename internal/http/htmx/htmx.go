package htmx

import "net/http"

const (
	htmxRequestHeaderName = "HX-Request"
)

func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get(htmxRequestHeaderName) == "true"
}
