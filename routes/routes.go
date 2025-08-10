package routes

import (
	"lod2/page"
	accountRoutes "lod2/routes/account"
	adminRoutes "lod2/routes/admin"
	authRoutes "lod2/routes/auth"
	storageRoutes "lod2/routes/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/admin", adminRoutes.Router())
	r.Mount("/account", accountRoutes.Router())
	r.Mount("/auth", authRoutes.Router())
	r.Mount("/files", storageRoutes.Router())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "index.html", nil)
	})

	r.NotFound(page.NotFound)

	return r
}
