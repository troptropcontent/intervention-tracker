package utils

import (
	"os"
	"path/filepath"
)

// GetProjectRootPath finds and returns the project root directory
func GetProjectRootPath() (string, error) {
	// Start from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for go.mod file to identify project root
	dir := currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			break
		}
		dir = parent
	}

	// Fallback to current directory if go.mod not found
	return currentDir, nil
}

// GetStaticFilePath returns the absolute path to a static file
func GetStaticFilePath(relativePath string) (string, error) {
	root, err := GetProjectRootPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, "static", relativePath), nil
}

// GetStaticFilePath returns the absolute path to a static file
func GetPathFromRoot(pathFromRoot string) (string, error) {
	root, err := GetProjectRootPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, pathFromRoot), nil
}

// GetStaticFilePath returns the absolute path to a static file
func MustGetPathFromRoot(pathFromRoot string) string {
	root, err := GetProjectRootPath()
	if err != nil {
		panic("could not find root path")
	}

	return filepath.Join(root, pathFromRoot)
}
