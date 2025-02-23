package auth

import (
	"lod2/internal/auth"
	"lod2/internal/page"
	"log"
	"net/http"
)

const nextParam = "next"

func getNextUrl(r *http.Request, defaultUrl string) string {
	nextUrl := r.URL.Query().Get(nextParam)

	if nextUrl == "" {
		nextUrl = r.Referer()
	}

	if nextUrl == "" {
		nextUrl = defaultUrl
	}

	return nextUrl
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	nextUrl := getNextUrl(r, "/account")

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

	var errorMessage = ""

	refreshTokenString, err := auth.IssueRefreshToken(username, password)
	var accessTokenString string

	if err != nil {
		errorMessage = err.Error()
	} else {
		refreshToken, _ := auth.ParseToken(refreshTokenString)

		accessTokenString, err = auth.IssueAccessToken(refreshToken)

		if err != nil {
			log.Printf("unable to issue access token from refresh token %v", err)
			errorMessage = "Unexpected error. Check logs for more information"
		}
	}

	if errorMessage != "" {
		w.WriteHeader(http.StatusUnauthorized)

		page.Render(w, r, "auth/login.html", map[string]interface{}{
			"Username": username,
			"Password": password,
			"Error":    errorMessage})
	}

	// should never happen, in theory.
	if next == "" {
		next = "/"
	}

	auth.SetCookie(w, auth.RefreshTokenCookieName, refreshTokenString, auth.RefreshTokenExpirationDuration)
	auth.SetCookie(w, auth.AccessTokenCookieName, accessTokenString, auth.AccessTokenExpirationDuration)

	http.Redirect(w, r, next, http.StatusSeeOther)
	return
}
