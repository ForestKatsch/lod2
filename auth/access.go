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

	var username string
	err := refreshToken.Get("username", &username)

	if err != nil {
		return "", err
	}

	builder.Claim("username", username)

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}
