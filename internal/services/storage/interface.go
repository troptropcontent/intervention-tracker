package storage

import (
	"context"
	"io"
	"time"
)

// StorageService defines the interface for file storage operations.
// This interface abstracts storage operations to allow for different implementations
// (S3, local filesystem, etc.) while maintaining a consistent API.
type StorageService interface {
	// UploadFile uploads a file to storage
	// Parameters:
	//   - ctx: context for cancellation and timeout
	//   - key: the path/key where the file will be stored (e.g., "interventions/2024/report.pdf")
	//   - file: the file content as an io.Reader
	//   - contentType: MIME type of the file (e.g., "application/pdf", "image/jpeg")
	// Returns the storage URL and an error if the upload fails
	UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error)

	// DownloadFile retrieves a file from storage
	// Parameters:
	//   - ctx: context for cancellation and timeout
	//   - key: the path/key of the file to download
	// Returns an io.ReadCloser for the file content (caller must close it) and an error if download fails
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)

	// DeleteFile removes a file from storage
	// Parameters:
	//   - ctx: context for cancellation and timeout
	//   - key: the path/key of the file to delete
	// Returns an error if deletion fails
	DeleteFile(ctx context.Context, key string) error

	// GetFileURL generates a presigned URL for direct file access
	// Parameters:
	//   - ctx: context for cancellation and timeout
	//   - key: the path/key of the file
	//   - expiration: how long the URL should remain valid
	// Returns a presigned URL and an error if generation fails
	GetFileURL(ctx context.Context, key string, expiration time.Duration) (string, error)

	// FileExists checks if a file exists in storage
	// Parameters:
	//   - ctx: context for cancellation and timeout
	//   - key: the path/key of the file to check
	// Returns true if file exists, false otherwise, and an error if check fails
	FileExists(ctx context.Context, key string) (bool, error)
}
