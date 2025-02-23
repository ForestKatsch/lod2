package account

import (
	"lod2/auth"
	"lod2/page"
	"lod2/utils"
	"net/http"
)

func getChangePassword(w http.ResponseWriter, r *http.Request) {
	page.Render(w, r, "account/change-password.html", map[string]interface{}{
		"Redirect": utils.GetNextUrl(r, "/account"),
	})
}

func postChangePassword(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userInfo := auth.GetCurrentUserInfo(r.Context())

	current := r.Form.Get("current")
	new := r.Form.Get("new")
	verify := r.Form.Get("verify")

	next := r.Form.Get("nextRedirectUrl")

	err := auth.ChangePassword(userInfo.UserId, current, new, verify)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		page.Render(w, r, "account/change-password.html", map[string]interface{}{
			"Current":  current,
			"New":      new,
			"Verify":   verify,
			"Redirect": next,
			"Error":    err.Error()})
		return
	}

	// should never happen, in theory.
	if next == "" {
		next = "/"
	}

	page.Render(w, r, "account/change-password-success.html", map[string]interface{}{
		"Redirect": next})
}
