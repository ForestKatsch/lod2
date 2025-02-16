package auth

import (
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Init() {
	initTokens()
}

// chi v5 router
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		err := page.Render(w, "auth/login.html", nil)

		if err != nil {
			page.RenderError(w, r, err)
		}
	})

	return r
}
