package auth

import (
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Init() {
	initTokens()
}

const nextParam = "next"

// chi v5 router
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		// the next url to go to
		nextUrl := r.URL.Query().Get(nextParam)

		if nextUrl == "" {
			nextUrl = r.Referer()
		}

		page.Render(w, r, "auth/login.html", map[string]interface{}{
			"Username": "",
			"Password": "",
			"Redirect": nextUrl,
		})
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		username := r.Form.Get("username")
		password := r.Form.Get("password")
		next := r.Form.Get("nextRedirectUrl")

		var error = ""

		if username != "admin" {
			error = "Invalid username or password"
		}

		// todo: actual auth
		if error == "" {
			if next == "" {
				next = "/"
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "lod2.refresh",
				Value:    "empty cookie",
				HttpOnly: true,
			})
			http.Redirect(w, r, next, http.StatusSeeOther)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		page.Render(w, r, "auth/login.html", map[string]interface{}{
			"Username": username,
			"Password": password,
			"Error":    error,
		})
	})

	return r
}
