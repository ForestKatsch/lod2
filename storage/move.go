package storage

import (
	"io"
	"log"
	"os"
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
