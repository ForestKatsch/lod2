package auth

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
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

func SignOut(w http.ResponseWriter, r *http.Request) {
	log.Println("signing out user")
	_deleteAuthCookie(w, AccessTokenCookieName)
	_deleteAuthCookie(w, RefreshTokenCookieName)
}

const accessTokenContextKey = "accessToken"

func validateAuth(w http.ResponseWriter, r *http.Request) context.Context {
	refreshTokenCookie, refreshErr := r.Cookie(RefreshTokenCookieName)
	accessTokenCookie, accessErr := r.Cookie(AccessTokenCookieName)

	if refreshErr == nil && accessErr != nil {
		log.Println("signing out user (has access token but no refresh token?)")
		SignOut(w, r)

		return r.Context()
	}

	if refreshTokenCookie == nil && accessTokenCookie == nil {
		return r.Context()
	}

	refreshToken, _ := ParseToken(refreshTokenCookie.Value)

	accessToken, accessErr := ParseToken(accessTokenCookie.Value)

	// at this point, an expired access token is not a problem and means we need to refresh it
	if accessErr != nil {
		accessTokenString, err := IssueAccessToken(refreshToken)

		if err != nil {
			log.Println("error reissuing access token:", err)
			SignOut(w, r)
			return r.Context()
		}

		accessToken, _ = ParseToken(accessTokenString)
		log.Println("access token was expired; refreshed")
		SetCookie(w, AccessTokenCookieName, accessTokenString, AccessTokenExpirationDuration)
	}

	ctx := context.WithValue(r.Context(), accessTokenContextKey, accessToken)

	return ctx
}

func GetCurrentUsername(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, ok := ctx.Value(accessTokenContextKey).(jwt.Token)

	if !ok {
		return ""
	}

	subject, ok := token.Subject()
	if !ok {
		return ""
	}

	return subject
}

func IsUserLoggedIn(ctx context.Context) bool {
	if GetCurrentUsername(ctx) == "" {
		return false
	}

	return true
}

func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := validateAuth(w, r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
