package auth

import (
	"errors"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

func IssueAccessToken(refreshToken jwt.Token) (string, error) {
	builder := getTokenBuilder(time.Now().Add(AccessTokenExpirationDuration))

	subject, exists := refreshToken.Subject()
	if !exists {
		return "", errors.New("unable to extract subject from refresh token")
	}

	builder.Subject(subject)

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}
