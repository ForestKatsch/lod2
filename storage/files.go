package storage

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go.jetify.com/typeid"
)

// Given a source path (on the filesystem) and a dest path (within storage), copies it in.
func ImportFile(sourcePath, destPath string) error {
	destPath, err := DangerousFilesystemPath(destPath)
	if err != nil {
		return err
	}

	inputFile, err := os.Open(sourcePath)
	if err != nil {
		log.Printf("Couldn't open source file: %s", err)
		return err
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		log.Printf("Couldn't open dest file: %s", err)
		return err
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		log.Printf("Writing to output file failed: %s", err)
		return err
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		log.Printf("Failed removing original file: %s", err)
		return err
	}
	return nil
}

func MoveFile(sourcePath, destPath string) error {
	sourcePath, err := DangerousFilesystemPath(sourcePath)
	if err != nil {
		return err
	}

	destPath, err = DangerousFilesystemPath(destPath)
	if err != nil {
		return err
	}

	err = os.Rename(sourcePath, destPath)
	if err != nil {
		log.Printf("Failed moving file: %s", err)
		return err
	}

	return nil
}

func DeleteFile(path string) error {
	// Generate ISO 8601 timestamp + random suffix for unique trash filename
	trashId, _ := typeid.WithPrefix("trash")
	trashPath := fmt.Sprintf("/.trash/%s%s", trashId.String(), path)

	dir := filepath.Dir(trashPath)
	log.Printf("creating directory %s (moving file %s to %s)", dir, path, trashPath)

	CreateDirectory(dir)

	log.Printf("moving file %s to %s", path, trashPath)

	return MoveFile(path, trashPath)
}

func ServeFile(w http.ResponseWriter, r *http.Request, path string) {
	filesystemPath, err := DangerousFilesystemPath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, filesystemPath)
}
