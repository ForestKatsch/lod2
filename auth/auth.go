package auth

import (
	"errors"
	"log"
	"net/http"
)

func Init() {
	initTokens()
	PostMigrationSetup()
}

func SetTokenCookies(w http.ResponseWriter, r *http.Request, username string, password string) error {
	refreshTokenString, err := IssueRefreshToken(username, password)

	var accessTokenString string

	if err != nil {
		return err
	}
	refreshToken, _ := ParseToken(refreshTokenString)

	accessTokenString, err = IssueAccessToken(refreshToken)

	if err != nil {
		log.Printf("unable to issue access token from refresh token: %v", err)
		return errors.New("Unexpected error. Check logs for more information")
	}

	SetCookie(w, RefreshTokenCookieName, refreshTokenString, RefreshTokenExpirationDuration)
	SetCookie(w, AccessTokenCookieName, accessTokenString, AccessTokenExpirationDuration)

	return nil
}
