package account

import (
	lod2Middleware "lod2/internal/middleware"
	"lod2/internal/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(lod2Middleware.AuthRequiredMiddleware())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "account/index.html", map[string]interface{}{})
	})

	return r
}
