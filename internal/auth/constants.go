package auth

import "time"

// The `iss` field of our JWTs.
const tokenIssuer = "https://lod2.zip"
const accessTokenAudience = "lod2.zip"

const RefreshTokenExpirationDuration = time.Hour * 6 * 30 * 24
const AccessTokenExpirationDuration = time.Second * 15

const RefreshTokenCookieName = "lod2.refresh"
const AccessTokenCookieName = "lod2.access"
