package auth

import "time"

// The `iss` field of our JWTs.
const issuer = "https://lod2.zip"
const audience = "lod2.zip"

const RefreshTokenExpirationDuration = time.Hour * 6 * 30 * 24
const AccessTokenExpirationDuration = time.Minute * 15

const RefreshTokenCookieName = "lod2.refresh"
const AccessTokenCookieName = "lod2.access"
