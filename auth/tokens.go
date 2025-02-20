package auth

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"

	"lod2/config"
)

// Responsible for managing tokens used for user authentication.
// Primarily, this includes the private/public key pair used to sign and verify JWTs.
// The private key is loaded from disk and used to sign JWTs, while the public key is derived from the private key and used to verify JWTs.

// The global private key used to sign JWTs.
var privkey jwk.Key

// The global public key used to verify JWTs.
var pubkey jwk.Key

// Needs to be called before using any of the functions in this package.
func initTokens() {
	initAuthTokenKeys()
}

// Loads the private and public keys used for JWT signing and verification.
func initAuthTokenKeys() {
	// First, read public/private keys from disk.
	privkeyPath := filepath.Join(config.Config.ConfigPath, "keys/auth/private.jwk.json")

	var err error
	privkey, err = loadPrivateKey(privkeyPath)

	if err != nil {
		log.Printf("unable to load private key: %s\n", err)
	}

	pubkey, err = jwk.PublicKeyOf(privkey)

	if err != nil {
		log.Printf("unable to derive public key from private key: %s\n", err)
	}
}

// Loads private key from disk and returns the key.
func loadPrivateKey(privkeyFilename string) (jwk.Key, error) {
	// Read the public key file
	bytes, err := os.ReadFile(privkeyFilename)
	if err != nil {
		return nil, err
	}

	// Parse, serialize, slice and dice JWKs!
	privkey, err := jwk.ParseKey(bytes)
	if err != nil {
		log.Printf("failed to parse JWK: %s", err)
		return nil, err
	}

	return privkey, nil
}

// Issue a new token with the provided audience and expiration time.
func getTokenBuilder(exp time.Time) *jwt.Builder {
	return jwt.NewBuilder().
		Issuer(issuer).
		IssuedAt(time.Now()).
		Expiration(exp)
}

func signToken(builder *jwt.Builder) (string, error) {
	tok, err := builder.Build()

	if err != nil {
		log.Printf("unable to build token: %s", err)
		return "", err
	}

	// Sign the JWT.
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), privkey))
	if err != nil {
		log.Printf("unable to sign token: %s", err)
		return "", err
	}

	return string(signed), nil
}

func issueRefreshToken(username string, password string) (string, error) {
	if username != "admin" {
		return "", errors.New("Invalid username")
	}

	if password != "admin" {
		return "", errors.New("Wrong password")
	}

	builder := getTokenBuilder(time.Now().Add(refreshTokenExpirationDuration))
	//builder.Audience(audience)

	return signToken(builder)
}

// Returns true if the token is valid and issued by us. Does not validate anything about the token's claims.
func VerifyToken(signedToken string) bool {
	token, err := jwt.Parse([]byte(signedToken), jwt.WithKey(jwa.RS256(), pubkey))

	if err != nil {
		log.Printf("unable to verify JWT was signed by us: %s", err)
		return false
	}

	err = jwt.Validate(
		token,
		jwt.
			WithIssuer(issuer))

	if err != nil {
		log.Printf("unable to validate JWT or issuer: %s", err)
		return false
	}

	// TODO: validate claims?
	return false
}
