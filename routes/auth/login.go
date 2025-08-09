package auth

import (
	"lod2/auth"
	"lod2/page"
	"lod2/utils"
	"log"
	"net/http"
)

func getLogin(w http.ResponseWriter, r *http.Request) {
	nextUrl := utils.GetNextUrl(r, "/account")

	// check if the user is already logged in
	if auth.IsUserLoggedIn(r.Context()) {
		print("already logged in!")
		http.Redirect(w, r, nextUrl, http.StatusSeeOther)
		return
	}

	page.Render(w, r, "auth/login.html", map[string]interface{}{
		"Username": "",
		"Password": "",
		"Redirect": nextUrl,
	})
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	next := r.Form.Get("nextRedirectUrl")

	// should never happen, in theory.
	if next == "" {
		next = "/"
	}

	err := auth.SetTokenCookies(w, r, username, password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		page.Render(w, r, "auth/login.html", map[string]interface{}{
			"Username": username,
			"Password": password,
			"Redirect": next,
			"Error":    err.Error(),
		})
		return
	}

	log.Printf("sign-in successful, redirecting to %s", next)
	http.Redirect(w, r, next, http.StatusSeeOther)
	return
}
