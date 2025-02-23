package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDirForFile(filename string) error {
	dir := filepath.Dir(filename) // Extract the directory path
	return os.MkdirAll(dir, 0755) // Create the directory if it doesn’t exist
}

func ExpandHomePath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("unexpectedly unable to expand home path: %v; ignoring and returning original path: ", err)
			return path
		}
		return filepath.Join(home, path[1:])
	}

	return path
}
