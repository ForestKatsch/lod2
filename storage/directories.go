package storage

import (
	"log"
	"os"
)

func CreateDirectory(path string) error {
	filesystemPath, err := DangerousFilesystemPath(path)

	if err != nil {
		return err
	}

	log.Printf("creating directory %s", filesystemPath)

	return os.MkdirAll(filesystemPath, 0755)
}
