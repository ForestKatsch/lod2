package auth

import "time"


// The `iss` field of our JWTs.
const issuer = "https://lod2.zip"
const audience = "lod2.zip"

const refreshTokenExpirationDuration = time.Hour * 6 * 30 * 24


const refreshTokenCookieName = "lod2.refresh"
