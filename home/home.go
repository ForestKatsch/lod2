package home

import (
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "index.html", nil)
	})

	r.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "index.html", nil)
	})

	return r
}
