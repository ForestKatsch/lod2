package home

import (
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// chi v5 router
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := page.Execute(w, "index.html", nil)

		if err != nil {
			page.RenderError(w, r, err)
		}
	})

	return r
}
