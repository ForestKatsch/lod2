package routes

import (
	"lod2/internal/page"
	accountRoutes "lod2/routes/account"
	authRoutes "lod2/routes/auth"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/auth", authRoutes.Router())
	r.Mount("/account", accountRoutes.Router())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "index.html", nil)
	})

	return r
}
