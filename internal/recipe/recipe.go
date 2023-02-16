package recipe

import (
	"net/http"

	"codeberg.org/mahlzeit/mahlzeit/internal/app"
	"github.com/go-chi/chi/v5"
)

func Handler(c *app.Application) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_ = c.Templates.Render(w, "home.tmpl", nil)
		})
	}
}
