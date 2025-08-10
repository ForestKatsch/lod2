package storage

import (
	"os"
	"sort"
	"strings"
	"time"
)

// All public functions operate on _unsafe paths_ provided by the user. All functions
// here are expected to verify paths and return errors if appropriate.
func Exists(path string) (bool, error) {
	filesystemPath, err := DangerousFilesystemPath(path)
	if err != nil {
		return false, err
	}

	// good case - check if it actually exists
	if _, err := os.Stat(filesystemPath); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func IsDirectory(path string) (bool, error) {
	filesystemPath, err := DangerousFilesystemPath(path)
	if err != nil {
		return false, err
	}

	exists, err := Exists(path)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	fi, err := os.Stat(filesystemPath)
	if err != nil {
		return false, err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true, nil
	case mode.IsRegular():
		return false, nil
	default:
		return false, nil
	}
}

type Entry struct {
	// The name of this entry.
	Name string

	IsDirectory bool

	// Size, in bytes; 0 if not a file.
	Size int64

	// Last modified time.
	LastModified time.Time
}

func ListContents(path string) ([]Entry, error) {
	filesystemPath, err := DangerousFilesystemPath(path)

	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filesystemPath)
	if err != nil {
		return nil, err
	}

	// Filter out .trash directory only at root level
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		// Hide .trash only when we're at the root directory
		if path == "/" && entry.Name() == ".trash" {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}

	results := make([]Entry, len(filteredEntries))

	for i, entry := range filteredEntries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		results[i] = Entry{
			Name:         entry.Name(),
			IsDirectory:  entry.IsDir(),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		}
	}

	// Sort: directories first, then files, both alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDirectory != results[j].IsDirectory {
			return results[i].IsDirectory // directories first
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return results, nil
}

func GetMetadata(path string) (Entry, error) {
	filesystemPath, err := DangerousFilesystemPath(path)
	if err != nil {
		return Entry{}, err
	}

	fi, err := os.Stat(filesystemPath)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		Name:         fi.Name(),
		IsDirectory:  fi.IsDir(),
		Size:         fi.Size(),
		LastModified: fi.ModTime(),
	}, nil
}
