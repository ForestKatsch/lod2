package storage

import (
	"errors"
	"lod2/config"
	"path/filepath"
	"strings"
)

// VerifyPath validates an untrusted path and ensures it stays within the storage root.
// It strictly rejects any path that would resolve outside the storage root, including
// directory traversal attempts. Both absolute and relative paths are supported.
//
// The function treats the input as a path within the storage root filesystem.
// For example, if storage root is "/home/user/storage":
//   - VerifyPath("/foo/bar") -> "/foo/bar" (verified safe)
//   - VerifyPath("foo/bar") -> "foo/bar" (verified safe, same as above)
//   - VerifyPath("/") -> "/" (verified safe)
//   - VerifyPath("/../etc/passwd") -> ERROR (would escape storage root)
//   - VerifyPath("/folder/../../../etc") -> ERROR (would escape storage root)
//
// Security features:
//   - Accepts both absolute and relative paths
//   - Strictly rejects any path that would resolve outside storage root
//   - Normalizes path separators for cross-platform compatibility
//   - No path cleaning/sanitization - rejection only
//
// Usage:
//
//	verifiedPath, err := storage.VerifyPath("/foo/bar/baz")
//	if err != nil {
//	    // Handle security violation or invalid path - path is REJECTED
//	}
//	// verifiedPath is now safe to use with FilesystemPath()
func VerifyPath(untrustedPath string) (string, error) {
	// Handle empty path - should return root "/"
	if untrustedPath == "" {
		return "/", nil
	}

	// Normalize path - add leading slash if not present to treat as absolute within storage
	normalizedInput := untrustedPath
	if !strings.HasPrefix(normalizedInput, "/") {
		normalizedInput = "/" + normalizedInput
	}

	// Normalize Windows-style backslashes for consistency
	normalizedPath := strings.ReplaceAll(normalizedInput, "\\", "/")

	// Clean the path to resolve . and .. components
	cleanPath := filepath.Clean(normalizedPath)
	
	// Check if the cleaned path tries to escape the root directory
	// After cleaning, any remaining ".." at the start means it would escape
	if !strings.HasPrefix(cleanPath, "/") || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return "", errors.New("path resolves outside storage root directory")
	}
	
	// Special case: if the clean path is just "." it should become "/"
	if cleanPath == "." {
		cleanPath = "/"
	}
	
	// IMPORTANT: Check if the input path tried to escape by simulating path traversal
	// Split the path and simulate directory navigation
	components := strings.Split(strings.Trim(normalizedPath, "/"), "/")
	currentDepth := 0 // Start at root level (depth 0)
	
	for _, comp := range components {
		if comp == "" || comp == "." {
			continue // Skip empty and current directory references
		} else if comp == ".." {
			currentDepth-- // Go up one level
			if currentDepth < 0 {
				// We've gone above the root - this is an escape attempt
				return "", errors.New("path resolves outside storage root directory")
			}
		} else {
			currentDepth++ // Go down into a directory
		}
	}

	// Remove trailing slash if present, except for root path "/"
	if len(cleanPath) > 1 && strings.HasSuffix(cleanPath, "/") {
		cleanPath = strings.TrimSuffix(cleanPath, "/")
	}

	// Always return the verified clean path with leading slash
	return cleanPath, nil
}

func DangerousFilesystemPath(path string) (string, error) {
	verifiedPath, err := VerifyPath(path)
	if err != nil {
		return "", err
	}

	return filepath.Join(config.Config.StoragePath, verifiedPath), nil
}

type PathBreadcrumb struct {
	// The relative path up to (and including) this point
	Path string

	// Just this component
	Component string

	// is this the final component?
	Final bool
}

// Returns a list of breadcrumbs for the given path.
// Example:
//   - /foo/bar/baz -> [
//     {"Path": "foo", "Component": "foo", "Final": false},
//     {"Path": "foo/bar", "Component": "bar", "Final": false},
//     {"Path": "foo/bar/baz", "Component": "baz", "Final": true},
//
// ]
func GetPathBreadcrumbs(path string) []PathBreadcrumb {
	verifiedPath, err := VerifyPath(path)
	if err != nil {
		return []PathBreadcrumb{}
	}

	// Handle root path
	if verifiedPath == "/" {
		return []PathBreadcrumb{}
	}

	// Split and filter out empty components
	allComponents := strings.Split(verifiedPath, "/")
	components := make([]string, 0, len(allComponents))
	for _, comp := range allComponents {
		if comp != "" {
			components = append(components, comp)
		}
	}

	if len(components) == 0 {
		return []PathBreadcrumb{}
	}

	breadcrumbs := make([]PathBreadcrumb, len(components))
	for i, component := range components {
		// Build path progressively: foo, foo/bar, foo/bar/baz
		pathComponents := components[:i+1]
		breadcrumbPath := strings.Join(pathComponents, "/")

		breadcrumbs[i] = PathBreadcrumb{
			Path:      breadcrumbPath,
			Component: component,
			Final:     i == len(components)-1,
		}
	}

	return breadcrumbs
}
