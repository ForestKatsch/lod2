package middleware

import (
	"context"
	"lod2/auth"
	"lod2/page"
	"log"
	"net/http"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

func validateAuth(w http.ResponseWriter, r *http.Request) context.Context {
	refreshTokenCookie, refreshErr := r.Cookie(auth.RefreshTokenCookieName)
	accessTokenCookie, accessErr := r.Cookie(auth.AccessTokenCookieName)

	if refreshErr != nil && accessErr == nil {
		log.Println("signing out user (has access token but no refresh token?)")
		auth.SignOut(w, r)

		return r.Context()
	}

	if refreshTokenCookie == nil || refreshTokenCookie.Value == "" {
		return r.Context()
	}

	accessTokenString := ""

	if accessTokenCookie != nil {
		accessTokenString = accessTokenCookie.Value
	}

	refreshToken, _ := auth.ParseToken(refreshTokenCookie.Value)
	var accessToken jwt.Token

	if accessTokenString != "" {
		accessToken, _ = auth.ParseToken(accessTokenString)
	}

	// at this point, an expired access token is not a problem and means we need to refresh it
	if accessToken == nil {
		accessTokenString, err := auth.IssueAccessToken(refreshToken)

		if err != nil {
			log.Println("error reissuing access token:", err)
			auth.SignOut(w, r)
			return r.Context()
		}

		accessToken, _ = auth.ParseToken(accessTokenString)
		log.Println("access token was expired; refreshed")
		auth.SetCookie(w, auth.AccessTokenCookieName, accessTokenString, auth.AccessTokenExpirationDuration)
	}

	// Create the user info object that lives on the context
	userInfo := auth.UserInfo{}

	accessToken.Get("username", &userInfo.Username)

	subject, valid := accessToken.Subject()

	if !valid {
		log.Println("error getting subject from access token")
		auth.SignOut(w, r)
		return r.Context()
	}
	userInfo.UserId = subject

	ctx := context.WithValue(r.Context(), auth.UserInfoContextKey, userInfo)

	return ctx
}

func AuthRefreshMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := validateAuth(w, r)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthRequiredMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !auth.IsUserLoggedIn(r.Context()) {
				w.WriteHeader(http.StatusUnauthorized)
				page.Render(w, r, "/_error/401.html", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
