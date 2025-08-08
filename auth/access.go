package auth

import (
	"errors"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

func IssueAccessToken(refreshToken jwt.Token) (string, error) {
	sessionId, exists := refreshToken.Subject()
	if !exists {
		return "", errors.New("unable to extract session ID from refresh token")
	}

	audience, exists := refreshToken.Audience()

	if !exists || audience[0] != "refresh" {
		return "", errors.New("unable to extract audience from refresh token")
	}

	userId, isValid := getUserSessionId(sessionId)

	if !isValid {
		return "", errors.New("invalid session")
	}

	updateUserSessionRefresh(sessionId)

	builder := getTokenBuilder(time.Now().Add(AccessTokenExpirationDuration))
	builder.Subject(userId)
	builder.Audience([]string{accessTokenAudience})

	if err := addAccessTokenClaims(builder, refreshToken, userId); err != nil {
		return "", err
	}

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}

// addAccessTokenClaims adds standard and custom claims to the access token builder.
func addAccessTokenClaims(builder *jwt.Builder, refreshToken jwt.Token, userId string) error {
	var username string
	if err := refreshToken.Get("username", &username); err != nil {
		return err
	}
	builder.Claim("username", username)

	if err := addRoleClaims(builder, userId); err != nil {
		return err
	}
	return nil
}

// addRoleClaims adds the user's roles to the token builder as the "roles" claim.
func addRoleClaims(builder *jwt.Builder, userId string) error {
	roles, err := GetUserRoles(userId)
	if err != nil {
		// If roles cannot be loaded (e.g., during migrations), skip silently
		log.Printf("unable to load roles for user %s; will be empty: %v", userId, err)
		return nil
	}
	// pack roles as array of {level, scope}
	type roleClaim struct {
		Level int `json:"level"`
		Scope int `json:"scope"`
	}
	claims := make([]roleClaim, 0, len(roles))
	for _, r := range roles {
		claims = append(claims, roleClaim{Level: int(r.Level), Scope: int(r.Scope)})
	}

	builder.Claim("roles", claims)
	return nil
}
