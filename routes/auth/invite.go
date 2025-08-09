package auth

import (
	"lod2/auth"
	"lod2/page"
	"lod2/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getInviteAndUsername(w http.ResponseWriter, r *http.Request) (string, string, error) {
	inviteCode := chi.URLParam(r, "inviteCode")
	userId, err := auth.ValidateInviteCode(inviteCode)
	if err != nil {
		return "", "", err
	}
	user, err := auth.AdminGetUserById(userId)
	return inviteCode, user.Username, nil
}

func getInvite(w http.ResponseWriter, r *http.Request) {
	// check if the user is already logged in
	if auth.IsUserLoggedIn(r.Context()) {
		page.Render(w, r, "auth/invite-invalid.html", map[string]interface{}{
			"Title": "You are already logged in",
			"Error": "You cannot redeem an invitation while logged in. Please log out and try again.",
		})
		return
	}

	inviteCode, inviteUsername, err := getInviteAndUsername(w, r)
	if err != nil {
		page.Render(w, r, "auth/invite-invalid.html", map[string]interface{}{
			"Title": "Invalid or expired invite",
			"Error": err.Error(),
		})
		return
	}

	nextUrl := utils.GetNextUrl(r, "/account")

	page.Render(w, r, "auth/invite.html", map[string]interface{}{
		"InviteCode":     inviteCode,
		"InviteUsername": inviteUsername,
		"Username":       "",
		"Password":       "",
		"Redirect":       nextUrl,
	})
}

func postInvite(w http.ResponseWriter, r *http.Request) {
	// check if the user is already logged in
	if auth.IsUserLoggedIn(r.Context()) {
		page.Render(w, r, "auth/invite-invalid.html", map[string]interface{}{
			"Title": "You are already logged in",
			"Error": "You cannot redeem an invitation while logged in. Please log out and try again.",
		})
		return
	}

	r.ParseForm()

	inviteCode, inviteUsername, err := getInviteAndUsername(w, r)

	if err != nil {
		page.Render(w, r, "auth/invite-invalid.html", map[string]interface{}{
			"Title": "Invalid or expired invite",
			"Error": err.Error(),
		})
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	confirmPassword := r.Form.Get("confirm_password")
	nextUrl := r.Form.Get("redirect")

	if nextUrl == "" {
		nextUrl = "/account"
	}

	renderError := func(errorMsg string) {
		page.Render(w, r, "auth/invite.html", map[string]interface{}{
			"InviteCode":      inviteCode,
			"InviteUsername":  inviteUsername,
			"Username":        username,
			"Password":        password,
			"ConfirmPassword": confirmPassword,
			"Redirect":        nextUrl,
			"Error":           errorMsg,
		})
	}

	if username == "" {
		renderError("Username is required")
		return
	}

	if password == "" {
		renderError("Password is required")
		return
	}

	if password != confirmPassword {
		renderError("Passwords do not match")
		return
	}

	_, err = auth.RegisterUserWithInvite(inviteCode, username, password)
	if err != nil {
		renderError(err.Error())
		return
	}

	err = auth.SetTokenCookies(w, r, username, password)

	if err != nil {
		renderError(err.Error())
		return
	}

	// Log the user in after successful registration
	// We can't use IssueRefreshToken directly since it requires username/password
	// Instead, redirect to login page with success message
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}
