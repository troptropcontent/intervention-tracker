package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name         string
		content      []byte
		filename     string
		expectedType string
		expectedErr  bool
		description  string
	}{
		{
			name:         "JPEG image",
			content:      []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			filename:     "photo.jpg",
			expectedType: "image/jpeg",
			expectedErr:  false,
			description:  "Should detect JPEG from magic bytes",
		},
		{
			name:         "PNG image",
			content:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			filename:     "image.png",
			expectedType: "image/png",
			expectedErr:  false,
			description:  "Should detect PNG from magic bytes",
		},
		{
			name:         "PDF document",
			content:      []byte("%PDF-1.4"),
			filename:     "document.pdf",
			expectedType: "application/pdf",
			expectedErr:  false,
			description:  "Should detect PDF from header",
		},
		{
			name:         "Plain text",
			content:      []byte("Hello, this is plain text content"),
			filename:     "note.txt",
			expectedType: "text/plain; charset=utf-8",
			expectedErr:  false,
			description:  "Should detect plain text",
		},
		{
			name:         "Fallback to extension - Go file",
			content:      []byte("package main\n\nfunc main() {}"),
			filename:     "main.go",
			expectedType: "text/plain; charset=utf-8",
			expectedErr:  false,
			description:  "Should detect Go source from content",
		},
		{
			name:         "Fallback to extension - ZIP file",
			content:      []byte{0x50, 0x4B, 0x03, 0x04}, // ZIP magic bytes
			filename:     "archive.zip",
			expectedType: "application/zip",
			expectedErr:  false,
			description:  "Should detect ZIP from magic bytes",
		},
		{
			name:         "Generic binary with extension",
			content:      []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			filename:     "data.bin",
			expectedType: "application/octet-stream",
			expectedErr:  false,
			description:  "Should return generic type for unknown binary",
		},
		{
			name:         "Unknown extension fallback",
			content:      []byte("some content"),
			filename:     "file.unknown",
			expectedType: "text/plain; charset=utf-8",
			expectedErr:  false,
			description:  "Should detect text for unknown extension",
		},
		{
			name:         "HTML content",
			content:      []byte("<!DOCTYPE html><html><body>Test</body></html>"),
			filename:     "index.html",
			expectedType: "text/html; charset=utf-8",
			expectedErr:  false,
			description:  "Should detect HTML from content",
		},
		{
			name:         "JSON content",
			content:      []byte(`{"key": "value", "number": 123}`),
			filename:     "data.json",
			expectedType: "text/plain; charset=utf-8",
			expectedErr:  false,
			description:  "JSON is detected as text/plain by http.DetectContentType",
		},
		{
			name:         "Empty file",
			content:      []byte{},
			filename:     "empty.txt",
			expectedType: "text/plain; charset=utf-8",
			expectedErr:  false,
			description:  "Should handle empty files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader from the test content
			reader := bytes.NewReader(tt.content)

			// Detect content type
			contentType, err := DetectContentType(reader, tt.filename)

			// Check error expectation
			if (err != nil) != tt.expectedErr {
				t.Errorf("DetectContentType() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			// Check content type
			if contentType != tt.expectedType {
				t.Errorf("DetectContentType() = %v, expected %v (%s)", contentType, tt.expectedType, tt.description)
			}

			// Verify file pointer was reset to beginning
			pos, err := reader.Seek(0, io.SeekCurrent)
			if err != nil {
				t.Errorf("Failed to get current position: %v", err)
			}
			if pos != 0 {
				t.Errorf("File pointer not reset to beginning, position = %d", pos)
			}
		})
	}
}

func TestDetectContentType_LargeFile(t *testing.T) {
	// Create content larger than 512 bytes
	largeContent := make([]byte, 1024)
	// Set JPEG magic bytes at the beginning
	copy(largeContent[:10], []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46})

	reader := bytes.NewReader(largeContent)
	contentType, err := DetectContentType(reader, "large.jpg")

	if err != nil {
		t.Fatalf("DetectContentType() error = %v", err)
	}

	if contentType != "image/jpeg" {
		t.Errorf("DetectContentType() = %v, expected image/jpeg", contentType)
	}

	// Verify file pointer was reset
	pos, _ := reader.Seek(0, io.SeekCurrent)
	if pos != 0 {
		t.Errorf("File pointer not reset after detecting large file, position = %d", pos)
	}
}

func TestGetFileSize(t *testing.T) {
	tests := []struct {
		name         string
		content      []byte
		expectedSize int64
		description  string
	}{
		{
			name:         "Empty file",
			content:      []byte{},
			expectedSize: 0,
			description:  "Should return 0 for empty file",
		},
		{
			name:         "Small file",
			content:      []byte("Hello, World!"),
			expectedSize: 13,
			description:  "Should return correct size for small file",
		},
		{
			name:         "Large file",
			content:      bytes.Repeat([]byte("a"), 10000),
			expectedSize: 10000,
			description:  "Should return correct size for large file",
		},
		{
			name:         "Binary file",
			content:      []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
			expectedSize: 7,
			description:  "Should return correct size for binary file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)

			size, err := GetFileSize(reader)
			if err != nil {
				t.Fatalf("GetFileSize() error = %v", err)
			}

			if size != tt.expectedSize {
				t.Errorf("GetFileSize() = %v, expected %v (%s)", size, tt.expectedSize, tt.description)
			}

			// Verify file pointer was reset to beginning
			pos, err := reader.Seek(0, io.SeekCurrent)
			if err != nil {
				t.Errorf("Failed to get current position: %v", err)
			}
			if pos != 0 {
				t.Errorf("File pointer not reset to beginning, position = %d", pos)
			}
		})
	}
}

func TestGetFileSize_PreservesContent(t *testing.T) {
	content := []byte("Test content that should be readable after GetFileSize")
	reader := bytes.NewReader(content)

	// Get size
	size, err := GetFileSize(reader)
	if err != nil {
		t.Fatalf("GetFileSize() error = %v", err)
	}

	if size != int64(len(content)) {
		t.Errorf("GetFileSize() = %v, expected %v", size, len(content))
	}

	// Verify we can still read the full content
	readContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read content after GetFileSize: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content changed after GetFileSize(). Got %q, expected %q", readContent, content)
	}
}

func TestCreateTempFile(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()

	tests := []struct {
		name          string
		content       string
		originalName  string
		expectedSize  int64
		shouldSucceed bool
		description   string
	}{
		{
			name:          "Simple text file",
			content:       "Hello, World!",
			originalName:  "test.txt",
			expectedSize:  13,
			shouldSucceed: true,
			description:   "Should create temp file with correct content",
		},
		{
			name:          "Empty file",
			content:       "",
			originalName:  "empty.txt",
			expectedSize:  0,
			shouldSucceed: true,
			description:   "Should handle empty files",
		},
		{
			name:          "Large file",
			content:       strings.Repeat("a", 10000),
			originalName:  "large.txt",
			expectedSize:  10000,
			shouldSucceed: true,
			description:   "Should handle large files",
		},
		{
			name:          "File with path in name",
			content:       "content",
			originalName:  "/path/to/file.txt",
			expectedSize:  7,
			shouldSucceed: true,
			description:   "Should extract basename from path",
		},
		{
			name:          "File with special characters",
			content:       "special content",
			originalName:  "file with spaces.txt",
			expectedSize:  15,
			shouldSucceed: true,
			description:   "Should handle filenames with special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)

			path, size, err := CreateTempFile(testDir, tt.originalName, reader)

			if !tt.shouldSucceed {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}

			// Clean up after test
			defer os.Remove(path)

			// Verify size
			if size != tt.expectedSize {
				t.Errorf("CreateTempFile() size = %v, expected %v (%s)", size, tt.expectedSize, tt.description)
			}

			// Verify file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Error("Temporary file was not created")
			}

			// Verify file is in the correct directory
			if filepath.Dir(path) != testDir {
				t.Errorf("File created in wrong directory. Got %s, expected %s", filepath.Dir(path), testDir)
			}

			// Verify file contains correct content
			readContent, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read temp file: %v", err)
			}

			if string(readContent) != tt.content {
				t.Errorf("File content mismatch. Got %q, expected %q", readContent, tt.content)
			}

			// Verify filename contains original basename
			expectedBasename := filepath.Base(tt.originalName)
			if !strings.Contains(filepath.Base(path), expectedBasename) {
				t.Errorf("Temp filename %q does not contain original basename %q", filepath.Base(path), expectedBasename)
			}
		})
	}
}

func TestCreateTempFile_CreatesDirectory(t *testing.T) {
	// Use a non-existent directory
	testDir := filepath.Join(t.TempDir(), "nested", "directory")

	reader := strings.NewReader("test content")
	path, _, err := CreateTempFile(testDir, "test.txt", reader)

	if err != nil {
		t.Fatalf("CreateTempFile() failed to create directory: %v", err)
	}
	defer os.Remove(path)

	// Verify directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
}

func TestCreateTempFile_UniqueNames(t *testing.T) {
	testDir := t.TempDir()

	// Create multiple temp files with the same original name
	var paths []string
	for i := 0; i < 5; i++ {
		reader := strings.NewReader("content")
		path, _, err := CreateTempFile(testDir, "same-name.txt", reader)
		if err != nil {
			t.Fatalf("CreateTempFile() error = %v", err)
		}
		paths = append(paths, path)
		defer os.Remove(path)
	}

	// Verify all paths are unique
	pathSet := make(map[string]bool)
	for _, path := range paths {
		if pathSet[path] {
			t.Errorf("Duplicate temp file path generated: %s", path)
		}
		pathSet[path] = true
	}
}

func TestDetectContentTypeFromPath(t *testing.T) {
	testDir := t.TempDir()

	tests := []struct {
		name         string
		content      []byte
		filename     string
		expectedType string
		description  string
	}{
		{
			name:         "JPEG file",
			content:      []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			filename:     "photo.jpg",
			expectedType: "image/jpeg",
			description:  "Should detect JPEG from file path",
		},
		{
			name:         "Text file",
			content:      []byte("Hello, World!"),
			filename:     "test.txt",
			expectedType: "text/plain; charset=utf-8",
			description:  "Should detect text from file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(testDir, tt.filename)
			if err := os.WriteFile(filePath, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Detect content type
			contentType, err := DetectContentTypeFromPath(filePath)
			if err != nil {
				t.Fatalf("DetectContentTypeFromPath() error = %v", err)
			}

			if contentType != tt.expectedType {
				t.Errorf("DetectContentTypeFromPath() = %v, expected %v (%s)", contentType, tt.expectedType, tt.description)
			}
		})
	}
}

func TestDetectContentTypeFromPath_NonExistentFile(t *testing.T) {
	_, err := DetectContentTypeFromPath("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// Benchmark tests
func BenchmarkDetectContentType(b *testing.B) {
	content := bytes.Repeat([]byte("test content"), 1000)
	reader := bytes.NewReader(content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		DetectContentType(reader, "test.txt")
	}
}

func BenchmarkGetFileSize(b *testing.B) {
	content := bytes.Repeat([]byte("test content"), 1000)
	reader := bytes.NewReader(content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Seek(0, io.SeekStart)
		GetFileSize(reader)
	}
}

func BenchmarkCreateTempFile(b *testing.B) {
	testDir := b.TempDir()
	content := "benchmark test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		path, _, err := CreateTempFile(testDir, "bench.txt", reader)
		if err != nil {
			b.Fatalf("CreateTempFile() error = %v", err)
		}
		os.Remove(path)
	}
}
