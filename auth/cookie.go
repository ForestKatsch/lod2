package auth

import (
	"log"
	"net/http"
	"time"
)

func SetCookie(w http.ResponseWriter, name, value string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(duration),
	})
}

func _deleteAuthCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func tryInvalidateSession(w http.ResponseWriter, r *http.Request) {
	refreshTokenCookie, refreshErr := r.Cookie(RefreshTokenCookieName)

	if refreshErr != nil || refreshTokenCookie == nil || refreshTokenCookie.Value == "" {
		return
	}

	refreshToken, err := ParseToken(refreshTokenCookie.Value)

	if err != nil {
		return
	}

	var sessionId string
	refreshToken.Get("sid", &sessionId)

	if sessionId == "" {
		return
	}

	invalidateSession(sessionId)
}

func SignOut(w http.ResponseWriter, r *http.Request) {
	log.Println("signing out user")
	tryInvalidateSession(w, r)
	_deleteAuthCookie(w, AccessTokenCookieName)
	_deleteAuthCookie(w, RefreshTokenCookieName)
}
