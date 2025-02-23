package auth

import (
	"lod2/internal/db"
	"log"
	"time"
)

func IssueRefreshToken(username string, password string) (string, error) {
	userId, err := db.GetUserLogin(username, password)

	if err != nil {
		return "", err
	}

	builder := getTokenBuilder(time.Now().Add(RefreshTokenExpirationDuration))
	builder.Audience([]string{"refresh"})
	builder.Subject(userId)
	builder.Claim("username", username)

	signed, err := signToken(builder)

	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return signed, nil
}
