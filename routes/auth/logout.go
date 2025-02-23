package auth

import (
	"lod2/internal/auth"
	"lod2/internal/page"
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
