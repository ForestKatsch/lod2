package storage

import (
	"lod2/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestStorageRoot(t *testing.T) (string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "safepath_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Store original config
	originalStoragePath := config.Config.StoragePath

	// Set config to use temp directory
	config.Config.StoragePath = tempDir

	// Return cleanup function
	cleanup := func() {
		config.Config.StoragePath = originalStoragePath
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestVerifyPath_ValidPaths(t *testing.T) {
	storageRoot, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	tests := []struct {
		name      string
		inputPath string
		expected  string // What the clean path should be
	}{
		{
			name:      "root path",
			inputPath: "/",
			expected:  "/",
		},
		{
			name:      "simple file",
			inputPath: "/file.txt",
			expected:  "/file.txt",
		},
		{
			name:      "nested file",
			inputPath: "/folder/file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "deeply nested file",
			inputPath: "/a/b/c/d/file.txt",
			expected:  "/a/b/c/d/file.txt",
		},
		{
			name:      "path with redundant separators",
			inputPath: "/folder//file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "path with current dir references",
			inputPath: "/folder/./file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "relative path - simple file",
			inputPath: "file.txt",
			expected:  "/file.txt",
		},
		{
			name:      "relative path - nested file",
			inputPath: "folder/file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "relative path - deeply nested",
			inputPath: "a/b/c/d/file.txt",
			expected:  "/a/b/c/d/file.txt",
		},
		{
			name:      "relative path with redundant separators",
			inputPath: "folder//file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "relative path with current dir references",
			inputPath: "folder/./file.txt",
			expected:  "/folder/file.txt",
		},
		{
			name:      "empty path should return root",
			inputPath: "",
			expected:  "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyPath(tt.inputPath)
			if err != nil {
				t.Errorf("VerifyPath(%q) returned error: %v", tt.inputPath, err)
				return
			}

			if result != tt.expected {
				t.Errorf("VerifyPath(%q) = %q, want %q", tt.inputPath, result, tt.expected)
			}

			// Verify filesystemPath result is within storage root
			fsPath, err := DangerousFilesystemPath(result)
			if err != nil {
				t.Errorf("filesystemPath(%q) failed: %v", result, err)
				return
			}
			if !strings.HasPrefix(fsPath, storageRoot) {
				t.Errorf("filesystemPath(VerifyPath(%q)) result %q is not within storage root %q", tt.inputPath, fsPath, storageRoot)
			}
		})
	}
}

func TestVerifyPath_RelativePathsEquivalent(t *testing.T) {
	storageRoot, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	// Test that relative and absolute paths behave identically
	testCases := []struct {
		absolute string
		relative string
	}{
		{"/file.txt", "file.txt"},
		{"/folder/file.txt", "folder/file.txt"},
		{"/a/b/c/file.txt", "a/b/c/file.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.absolute, func(t *testing.T) {
			absResult, absErr := VerifyPath(tc.absolute)
			relResult, relErr := VerifyPath(tc.relative)

			// Both should succeed
			if absErr != nil {
				t.Errorf("VerifyPath(%q) failed: %v", tc.absolute, absErr)
			}
			if relErr != nil {
				t.Errorf("VerifyPath(%q) failed: %v", tc.relative, relErr)
			}

			// filesystemPath should produce identical results
			absFS, absFSErr := DangerousFilesystemPath(absResult)
			relFS, relFSErr := DangerousFilesystemPath(relResult)

			if absFSErr != nil {
				t.Errorf("filesystemPath(%q) failed: %v", absResult, absFSErr)
				return
			}
			if relFSErr != nil {
				t.Errorf("filesystemPath(%q) failed: %v", relResult, relFSErr)
				return
			}

			if absFS != relFS {
				t.Errorf("filesystemPath mismatch: abs=%q, rel=%q", absFS, relFS)
			}

			// Both should be within storage root
			if !strings.HasPrefix(absFS, storageRoot) {
				t.Errorf("Absolute path result %q not within storage root %q", absFS, storageRoot)
			}
			if !strings.HasPrefix(relFS, storageRoot) {
				t.Errorf("Relative path result %q not within storage root %q", relFS, storageRoot)
			}
		})
	}
}

func TestVerifyPath_MaliciousPaths(t *testing.T) {
	_, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	// Only test actual attacks that should escape the storage root
	maliciousPaths := []struct {
		name string
		path string
	}{
		{
			name: "relative parent path with traversal",
			path: "../relative",
		},
		{
			name: "absolute parent path with traversal",
			path: "/../etc/passwd",
		},
	}

	for _, tt := range maliciousPaths {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyPath(tt.path)
			if err == nil {
				t.Errorf("VerifyPath(%q) should have returned an error but got result: %q", tt.path, result)
				return
			}

			expectedErrors := []string{
				"path resolves outside storage root directory",
				"empty path not allowed",
				"path contains directory traversal patterns",
			}

			errorFound := false
			errorMsg := err.Error()
			for _, expectedErr := range expectedErrors {
				if strings.Contains(errorMsg, expectedErr) {
					errorFound = true
					break
				}
			}

			if !errorFound {
				t.Errorf("VerifyPath(%q) returned unexpected error: %v", tt.path, err)
			}
		})
	}
}

func TestVerifyPath_TraversalAttacksRejected(t *testing.T) {
	_, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	// These paths contain traversal patterns that would escape storage root and MUST be rejected
	maliciousPaths := []struct {
		name string
		path string
	}{
		{
			name: "traversal attempting to reach parent",
			path: "/../",
		},
		{
			name: "multiple traversals attempting escape",
			path: "/../../../../",
		},
		{
			name: "traversal with target file",
			path: "/../../../etc/passwd",
		},
		{
			name: "mixed traversal with valid path",
			path: "/folder/../../../etc/passwd",
		},
		{
			name: "complex nested traversal attack",
			path: "/a/b/c/../../../../../../../etc/passwd",
		},
		{
			name: "windows style traversal attack",
			path: "/..\\..\\..\\windows\\system32",
		},
	}

	for _, tt := range maliciousPaths {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyPath(tt.path)
			if err == nil {
				t.Errorf("VerifyPath(%q) should have been REJECTED but got result: %q", tt.path, result)
				return
			}

			expectedErrors := []string{
				"path contains directory traversal patterns",
			}

			errorFound := false
			errorMsg := err.Error()
			for _, expectedErr := range expectedErrors {
				if strings.Contains(errorMsg, expectedErr) {
					errorFound = true
					break
				}
			}

			if !errorFound {
				t.Errorf("VerifyPath(%q) returned unexpected error: %v", tt.path, err)
			}
		})
	}
}

func TestVerifyPath_EdgeCases(t *testing.T) {
	_, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	edgeCases := []struct {
		name          string
		path          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "very long path within bounds",
			path:        "/" + strings.Repeat("a/", 100) + "file.txt",
			expectError: false,
		},
		{
			name:        "path with spaces",
			path:        "/folder with spaces/file with spaces.txt",
			expectError: false,
		},
		{
			name:        "path with unicode",
			path:        "/folder/файл.txt",
			expectError: false,
		},
		{
			name:        "path with special chars",
			path:        "/folder/file-name_123.txt",
			expectError: false,
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("VerifyPath(%q) expected error but got result: %q", tt.path, result)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("VerifyPath(%q) error %q should contain %q", tt.path, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("VerifyPath(%q) unexpected error: %v", tt.path, err)
				}
			}
		})
	}
}

func TestVerifyPath_SecurityInvariant(t *testing.T) {
	storageRoot, cleanup := setupTestStorageRoot(t)
	defer cleanup()

	// Test many different malicious inputs to ensure none can escape
	maliciousInputs := []string{
		"/../",
		"/../../",
		"/../../../",
		"/../../../../",
		"/../../../../../",
		"/../../../../../../",
		"/../../../../../../../",
		"/../../../../../../../../",
		"/../../../../../../../..",
		"/folder/../..",
		"/a/b/c/../../../..",
		"/./../../..",
		"/folder/../../..",
		"/../folder/../..",
		"/" + strings.Repeat("../", 50),
		"/" + strings.Repeat("../", 100),
		"/..\\..\\..\\Windows\\System32",
	}

	for i, maliciousInput := range maliciousInputs {
		t.Run(filepath.Base(maliciousInput)+string(rune(i)), func(t *testing.T) {
			result, err := VerifyPath(maliciousInput)

			if err == nil {
				// If no error, verify the result is still safe
				absResult, absErr := filepath.Abs(result)
				if absErr != nil {
					t.Errorf("Could not get absolute path of result %q: %v", result, absErr)
					return
				}

				absRoot, absErr := filepath.Abs(storageRoot)
				if absErr != nil {
					t.Errorf("Could not get absolute path of storage root %q: %v", storageRoot, absErr)
					return
				}

				if !strings.HasPrefix(absResult, absRoot) {
					t.Errorf("SECURITY VIOLATION: VerifyPath(%q) returned %q which escapes storage root %q", maliciousInput, result, storageRoot)
				}

				// Allow exact root match since "/" should map to storage root
				if absResult != absRoot && !strings.HasPrefix(absResult, absRoot+string(filepath.Separator)) {
					t.Errorf("SECURITY VIOLATION: VerifyPath(%q) returned %q which escapes storage root %q", maliciousInput, result, storageRoot)
				}
			}
		})
	}
}

// Benchmark to ensure the function is performant
func BenchmarkVerifyPath(b *testing.B) {
	storageRoot, cleanup := setupTestStorageRoot(nil)
	defer cleanup()

	// Set a dummy config for benchmarking
	config.Config.StoragePath = storageRoot

	testPath := "/folder/subfolder/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyPath(testPath)
	}
}

func BenchmarkVerifyPathMalicious(b *testing.B) {
	storageRoot, cleanup := setupTestStorageRoot(nil)
	defer cleanup()

	config.Config.StoragePath = storageRoot

	testPath := "/../../../../../../../../etc/passwd"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyPath(testPath)
	}
}
