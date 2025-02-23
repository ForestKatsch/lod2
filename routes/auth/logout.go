package auth

import (
	"lod2/auth"
	"lod2/page"
	"net/http"
)

func getLogout(w http.ResponseWriter, r *http.Request) {
	if !auth.IsUserLoggedIn(r.Context()) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	page.Render(w, r, "auth/logout.html", nil)
}

func getLogoutConfirm(w http.ResponseWriter, r *http.Request) {
	if auth.IsUserLoggedIn(r.Context()) {
		auth.SignOut(w, r)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
