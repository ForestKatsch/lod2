package auth

import (
	"errors"
	"log"
	"time"
)

func IssueRefreshToken(username string, password string) (string, error) {
	// TODO: should probably remove this lmao
	if username != "admin" {
		return "", errors.New("Invalid username")
	}

	if password != "admin" {
		return "", errors.New("Invalid password")
	}

	builder := getTokenBuilder(time.Now().Add(RefreshTokenExpirationDuration))
	builder.Audience([]string{"refresh"})
	builder.Subject(username)

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}
