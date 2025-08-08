package auth

import (
	"github.com/go-chi/chi/v5"
)

// chi v5 router
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", getLogin)
	r.Post("/login", postLogin)

	r.Get("/logout", getLogout)
	r.Post("/logout/confirm", postLogoutConfirm)

	return r
}
