package admin

import (
	"lod2/internal/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func userRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "admin/users/index.html", map[string]interface{}{})
	})

	return r
}
