package account

import (
	"lod2/middleware"
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.AuthRequiredMiddleware())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		page.Render(w, r, "account/index.html", map[string]interface{}{})
	})

	r.Get("/invite-link", getInviteLinkFragment)
	r.Get("/change-password", getChangePassword)
	r.Post("/change-password", postChangePassword)

	return r
}
