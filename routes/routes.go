package routes

import (
	"lod2/internal/page"
	authRoutes "lod2/routes/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/auth", authRoutes.Router())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "index.html", nil)
	})

	return r
}
