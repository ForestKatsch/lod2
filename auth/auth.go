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
		// pass down enough to do {{ .Username and .Password }}
		err := page.Render(w, "auth/login.html", map[string]interface{}{
			"Username": "",
			"Password": "",
		})

		if err != nil {
			page.RenderError(w, r, err)
		}
	})

	return r
}
