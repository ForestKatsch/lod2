package middleware

import (
	"context"
	"lod2/internal/auth"
	"lod2/internal/page"
	"log"
	"net/http"
)

func validateAuth(w http.ResponseWriter, r *http.Request) context.Context {
	refreshTokenCookie, refreshErr := r.Cookie(auth.RefreshTokenCookieName)
	accessTokenCookie, accessErr := r.Cookie(auth.AccessTokenCookieName)

	if refreshErr != nil && accessErr == nil {
		log.Println("signing out user (has access token but no refresh token?)")
		auth.SignOut(w, r)

		return r.Context()
	}

	if refreshErr != nil || accessErr != nil {
		auth.SignOut(w, r)
		return r.Context()
	}

	refreshToken, _ := auth.ParseToken(refreshTokenCookie.Value)

	accessToken, accessErr := auth.ParseToken(accessTokenCookie.Value)

	// at this point, an expired access token is not a problem and means we need to refresh it
	if accessErr != nil {
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

	ctx := context.WithValue(r.Context(), auth.AccessTokenContextKey, accessToken)

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
