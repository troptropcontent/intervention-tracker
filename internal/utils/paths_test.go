package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectRootPath(t *testing.T) {
	rootPath, err := GetProjectRootPath()
	require.NoError(t, err, "Should be able to find project root")

	// Verify the path exists
	assert.DirExists(t, rootPath, "Project root path should exist")

	// Verify go.mod exists in the root path
	goModPath := filepath.Join(rootPath, "go.mod")
	assert.FileExists(t, goModPath, "go.mod should exist in project root")

	// Verify the path is absolute
	assert.True(t, filepath.IsAbs(rootPath), "Project root path should be absolute")

	// Verify the path contains our project name (basic sanity check)
	assert.True(t, strings.Contains(rootPath, "qr_code_maintenance"),
		"Project root should contain project directory name")
}

func TestGetStaticFilePath(t *testing.T) {
	tests := []struct {
		name         string
		relativePath string
		expectExists bool
	}{
		{
			name:         "CSS output file",
			relativePath: "css/output.css",
			expectExists: true,
		},
		{
			name:         "CSS input file",
			relativePath: "css/input.css",
			expectExists: true,
		},
		{
			name:         "Non-existent file",
			relativePath: "nonexistent/file.txt",
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			staticPath, err := GetStaticFilePath(tt.relativePath)
			require.NoError(t, err, "Should be able to resolve static file path")

			// Verify the path is absolute
			assert.True(t, filepath.IsAbs(staticPath), "Static file path should be absolute")

			// Verify the path contains "static" directory
			assert.True(t, strings.Contains(staticPath, "static"),
				"Static file path should contain 'static' directory")

			// Verify the relative path is correctly appended
			assert.True(t, strings.HasSuffix(staticPath, tt.relativePath),
				"Static file path should end with the relative path")

			// Check file existence based on test expectation
			if tt.expectExists {
				assert.FileExists(t, staticPath, "Expected file should exist")
			} else {
				assert.NoFileExists(t, staticPath, "Non-existent file should not exist")
			}
		})
	}
}

func TestGetStaticFilePath_WithNestedPaths(t *testing.T) {
	nestedPath := "images/icons/favicon.ico"
	staticPath, err := GetStaticFilePath(nestedPath)
	require.NoError(t, err, "Should handle nested paths")

	// Verify the nested structure is preserved
	assert.True(t, strings.HasSuffix(staticPath, filepath.Join("static", "images", "icons", "favicon.ico")),
		"Should preserve nested directory structure")
}

func TestGetStaticFilePath_WithEmptyPath(t *testing.T) {
	staticPath, err := GetStaticFilePath("")
	require.NoError(t, err, "Should handle empty path")

	// Should resolve to just the static directory
	assert.True(t, strings.HasSuffix(staticPath, "static"),
		"Empty relative path should resolve to static directory")
}

// Test behavior when working directory changes
func TestGetProjectRootPath_FromDifferentDirectories(t *testing.T) {
	// Get original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Should be able to get current working directory")
	defer func() {
		// Restore original working directory
		os.Chdir(originalWd)
	}()

	// Get project root from current location
	rootPath1, err := GetProjectRootPath()
	require.NoError(t, err, "Should find project root from current directory")

	// Change to a subdirectory (if it exists)
	testDirs := []string{"internal", "cmd", "static"}
	for _, dir := range testDirs {
		dirPath := filepath.Join(originalWd, dir)
		if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
			// Change to subdirectory
			err = os.Chdir(dirPath)
			require.NoError(t, err, "Should be able to change to subdirectory")

			// Get project root from subdirectory
			rootPath2, err := GetProjectRootPath()
			require.NoError(t, err, "Should find project root from subdirectory")

			// Both should resolve to the same root path
			assert.Equal(t, rootPath1, rootPath2,
				"Project root should be the same regardless of current working directory")

			break // Only test with first available directory
		}
	}
}

func TestGetProjectRootPath_Integration(t *testing.T) {
	// This is an integration test that verifies the project structure
	rootPath, err := GetProjectRootPath()
	require.NoError(t, err, "Should find project root")

	// Verify key project directories exist
	expectedDirs := []string{"internal", "static", "cmd"}
	for _, dir := range expectedDirs {
		dirPath := filepath.Join(rootPath, dir)
		assert.DirExists(t, dirPath, "Expected project directory should exist: %s", dir)
	}

	// Verify key files exist
	expectedFiles := []string{"go.mod", "go.sum"}
	for _, file := range expectedFiles {
		filePath := filepath.Join(rootPath, file)
		if file == "go.sum" {
			// go.sum might not exist in all cases, so just check it exists if present
			if _, err := os.Stat(filePath); err == nil {
				assert.FileExists(t, filePath, "go.sum should exist if present")
			}
		} else {
			assert.FileExists(t, filePath, "Expected project file should exist: %s", file)
		}
	}
}
