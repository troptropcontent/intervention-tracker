package utils

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// DetectContentType detects the MIME type of a file by reading its first 512 bytes.
// It falls back to extension-based detection if the content-based detection returns
// a generic "application/octet-stream" type.
//
// The file's read position is reset to the beginning after detection.
//
// Parameters:
//   - file: Must implement io.ReadSeeker (will be read and seeked)
//   - filename: Original filename (used for extension-based fallback detection)
//
// Returns the detected MIME type and any error encountered.
func DetectContentType(file io.ReadSeeker, filename string) (string, error) {
	// Read first 512 bytes for content type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for content type detection: %w", err)
	}

	// Reset file pointer to beginning for subsequent operations
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer after content type detection: %w", err)
	}

	// Detect content type from the bytes read
	contentType := http.DetectContentType(buffer[:n])

	// If DetectContentType returns generic "application/octet-stream",
	// try to get a more specific type from the file extension
	if contentType == "application/octet-stream" {
		ext := filepath.Ext(filename)
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			contentType = mimeType
		}
	}

	return contentType, nil
}

// GetFileSize returns the size of a file in bytes by seeking to the end.
// The file's read position is reset to the beginning after getting the size.
//
// Parameters:
//   - file: Must implement io.Seeker
//
// Returns the file size in bytes and any error encountered.
func GetFileSize(file io.Seeker) (int64, error) {
	// Seek to end to get file size
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, fmt.Errorf("failed to reset file pointer after getting size: %w", err)
	}

	return size, nil
}

// CreateTempFile creates a temporary file with a unique name in the specified directory.
// The content is copied from the provided reader to the temporary file.
//
// The caller is responsible for cleaning up the temporary file when done.
//
// Parameters:
//   - dir: Directory where the temp file should be created (will be created if it doesn't exist)
//   - originalFilename: Original filename (used as suffix for the temp file name)
//   - content: Reader containing the file content to write
//
// Returns:
//   - path: Absolute path to the created temporary file
//   - size: Number of bytes written to the file
//   - err: Any error encountered during the operation
//
// If an error occurs, any partially created files are cleaned up automatically.
func CreateTempFile(dir, originalFilename string, content io.Reader) (path string, size int64, err error) {
	// Ensure temp directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique temporary file path
	tempFileName := fmt.Sprintf("%s_%s", uuid.New().String(), filepath.Base(originalFilename))
	tempFilePath := filepath.Join(dir, tempFileName)

	// Create temporary file
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()

	// Copy content to temporary file
	written, err := io.Copy(tempFile, content)
	if err != nil {
		os.Remove(tempFilePath) // Clean up on error
		return "", 0, fmt.Errorf("failed to write to temporary file: %w", err)
	}

	return tempFilePath, written, nil
}

// DetectContentTypeFromPath detects the MIME type of a file at the given path.
// This is a convenience wrapper around DetectContentType for file paths.
//
// Parameters:
//   - filePath: Path to the file to analyze
//
// Returns the detected MIME type and any error encountered.
func DetectContentTypeFromPath(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return DetectContentType(file, filepath.Base(filePath))
}
