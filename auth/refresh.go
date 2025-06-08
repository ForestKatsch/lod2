package auth

import (
	"log"
	"time"
)

func IssueRefreshToken(username string, password string) (string, error) {
	userId, err := getUserLogin(username, password)

	if err != nil {
		return "", err
	}

	sessionId, err := createUserSession(userId)

	if err != nil {
		return "", err
	}

	builder := getTokenBuilder(time.Now().Add(RefreshTokenExpirationDuration))
	builder.Audience([]string{"refresh"})
	builder.Subject(sessionId)
	builder.Claim("username", username)

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}
