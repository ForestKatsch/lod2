package config

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"flag"
	"lod2/utils"
	"log"
	"os"
	"path/filepath"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var Config struct {
	Http struct {
		Host string
		Port int
	}

	// configuration directory used for relatively long-term persistent configuration. read-only.
	ConfigPath string

	// dynamic data directory for things such as DBs and files. read-write.
	DataPath string

	StoragePath string
}

func Init(autocreate bool) {
	flag.StringVar(&Config.Http.Host, "host", "localhost", "host to listen on")
	flag.IntVar(&Config.Http.Port, "port", 10800, "port to listen on")

	flag.StringVar(&Config.ConfigPath, "config", "~/.config/lod2/", "path to configuration directory")
	flag.StringVar(&Config.DataPath, "data", "~/.local/share/lod2/", "path to data directory")
	flag.StringVar(&Config.StoragePath, "storage", "~/storage/", "path to huge storage directory")

	Config.ConfigPath = utils.ExpandHomePath(Config.ConfigPath)
	Config.DataPath = utils.ExpandHomePath(Config.DataPath)
	Config.StoragePath = utils.ExpandHomePath(Config.StoragePath)

	flag.Parse()

	if autocreate {
		if err := ensureBaseDirs(); err != nil {
			log.Fatalf("failed to create base dirs: %v", err)
		}
		if err := createAuthPrivateKeyIfNotExists(); err != nil {
			log.Fatalf("failed to create auth key: %v", err)
		}
	}
}

// Kickstart initializes the local installation by creating required directories
// and generating a private JWK for JWT signing if it does not already exist.
// Previously public; now internalized into Init(autocreate).

// ensureBaseDirs creates the base config and data directories if missing.
func ensureBaseDirs() error {
	if err := os.MkdirAll(Config.ConfigPath, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(Config.DataPath, 0o755); err != nil {
		return err
	}
	return nil
}

// authPrivateKeyPath returns the expected path to the private JWK file.
func authPrivateKeyPath() string {
	return filepath.Join(Config.ConfigPath, "keys/auth/private.jwk.json")
}

// createAuthPrivateKeyIfNotExists generates and writes an RSA private JWK if missing.
func createAuthPrivateKeyIfNotExists() error {
	privKeyPath := authPrivateKeyPath()

	if err := utils.EnsureDirForFile(privKeyPath); err != nil {
		return err
	}

	if _, err := os.Stat(privKeyPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	key, err := generateRSAPrivateJWK()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(privKeyPath, bytes, 0o600); err != nil {
		return err
	}
	return nil
}

// generateRSAPrivateJWK creates a 2048-bit RSA key and wraps it as a JWK with metadata.
func generateRSAPrivateJWK() (jwk.Key, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	key, err := jwk.FromRaw(rsaKey)
	if err != nil {
		return nil, err
	}
	_ = key.Set(jwk.KeyUsageKey, "sig")
	_ = key.Set(jwk.AlgorithmKey, jwa.RS256)
	return key, nil
}
