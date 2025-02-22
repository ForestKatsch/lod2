package auth

import (
	"context"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

const AccessTokenContextKey = "accessToken"

func GetCurrentUsername(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	token, ok := ctx.Value(AccessTokenContextKey).(jwt.Token)

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
