package attachments

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

func TestUploadWorker_Work_Success(t *testing.T) {
	// Setup test database and storage
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.jpg")
	testContent := []byte("test file content")
	err := os.WriteFile(tempFile, testContent, 0644)
	require.NoError(t, err)

	// Create attachment record with temp file path
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("test.jpg").
		WithContentType("image/jpeg").
		WithFileSize(int64(len(testContent))).
		WithUploadStatus(models.AttachmentUploadStatusPending).
		WithHolderType("interventions").
		WithHolderID(1).
		WithKind("photo").
		Create(db)

	// Verify temp file exists
	_, err = os.Stat(tempFile)
	require.NoError(t, err, "temp file should exist before worker runs")

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	require.NoError(t, err, "Worker should complete successfully")

	// Verify attachment was updated
	var updated models.Attachment
	err = db.First(&updated, attachment.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.AttachmentUploadStatusCompleted, updated.UploadStatus)
	assert.NotEqual(t, tempFile, updated.StorageKey, "StorageKey should be updated to S3 path")
	assert.Contains(t, updated.StorageKey, "interventions/1/photo/", "StorageKey should follow pattern")
	assert.Contains(t, updated.StorageKey, "test.jpg", "StorageKey should contain filename")
	assert.True(t, updated.UploadedAt.Valid)

	// Verify file was uploaded to storage
	assert.True(t, mockStorage.UploadCalled)
	assert.Len(t, mockStorage.UploadedFiles, 1)
	assert.Equal(t, updated.StorageKey, mockStorage.UploadedFiles[0].Key)

	// Verify temp file was deleted
	_, err = os.Stat(tempFile)
	assert.True(t, os.IsNotExist(err), "temp file should be deleted after successful upload")
}

func TestUploadWorker_Work_AlreadyCompleted(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.jpg")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create already-completed attachment
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("test.jpg").
		WithUploadStatus(models.AttachmentUploadStatusCompleted).
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify no upload happened (idempotency)
	assert.False(t, mockStorage.UploadCalled, "should not upload if already completed")

	// Verify temp file was cleaned up anyway
	_, err = os.Stat(tempFile)
	assert.True(t, os.IsNotExist(err), "temp file should be cleaned up even if already completed")
}

func TestUploadWorker_Work_AttachmentNotFound(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker with non-existent attachment ID
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: 99999,
		},
	}

	err := worker.Work(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find attachment")
}

func TestUploadWorker_Work_TempFileNotFound(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create attachment with non-existent temp file path
	attachment := factories.NewAttachment().
		WithStorageKey("/tmp/non-existent-file.jpg").
		WithFileName("test.jpg").
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open temporary file")

	// Verify upload status not changed
	var updated models.Attachment
	db.First(&updated, attachment.ID)
	assert.NotEqual(t, models.AttachmentUploadStatusCompleted, updated.UploadStatus)
}

func TestUploadWorker_Work_S3UploadFails(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Configure mock to fail uploads
	mockStorage.UploadError = fmt.Errorf("S3 connection timeout")

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.jpg")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create attachment
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("test.jpg").
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload file to S3")

	// Verify attachment status not changed
	var updated models.Attachment
	db.First(&updated, attachment.ID)
	assert.NotEqual(t, models.AttachmentUploadStatusCompleted, updated.UploadStatus)

	// Verify temp file still exists (for retry)
	_, err = os.Stat(tempFile)
	assert.NoError(t, err, "temp file should remain for retry")
}

func TestUploadWorker_Work_TempFileCleanupFails(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file in a directory we'll make read-only
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.jpg")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create attachment
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("test.jpg").
		Create(db)

	// Make directory read-only to prevent file deletion
	// Note: This may not work on all systems, so we'll skip if it doesn't
	err = os.Chmod(tempDir, 0444)
	if err != nil {
		t.Skip("Cannot make directory read-only on this system")
	}
	defer os.Chmod(tempDir, 0755) // Restore for cleanup

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	assert.Error(t, err, "should fail when temp file cannot be deleted")

	// Verify attachment status NOT updated (transaction rolled back)
	var updated models.Attachment
	db.First(&updated, attachment.ID)
	assert.NotEqual(t, models.AttachmentUploadStatusCompleted, updated.UploadStatus)

	// Restore permissions
	os.Chmod(tempDir, 0755)
}

func TestUploadWorker_Work_StorageKeyPattern(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "my-photo.jpg")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create attachment with specific holder info
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("my-photo.jpg").
		WithHolderType("interventions").
		WithHolderID(42).
		WithKind("photo").
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify storage key pattern
	var updated models.Attachment
	db.First(&updated, attachment.ID)

	// Should be: interventions/42/photo/{timestamp}_my-photo.jpg
	assert.Regexp(t, `^interventions/42/photo/\d+_my-photo\.jpg$`, updated.StorageKey)
}

func TestUploadWorker_Work_MultipleRetries(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.jpg")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create attachment
	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName("test.jpg").
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	// Simulate first attempt fails
	mockStorage.UploadError = fmt.Errorf("network error")
	err = worker.Work(context.Background(), job)
	assert.Error(t, err)

	// Verify temp file still exists
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	// Simulate second attempt succeeds
	mockStorage.UploadError = nil
	err = worker.Work(context.Background(), job)
	assert.NoError(t, err)

	// Verify successful completion
	var updated models.Attachment
	db.First(&updated, attachment.ID)
	assert.Equal(t, models.AttachmentUploadStatusCompleted, updated.UploadStatus)

	// Verify temp file deleted
	_, err = os.Stat(tempFile)
	assert.True(t, os.IsNotExist(err))
}

func TestUploadWorker_Work_PreservesFileMetadata(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create temporary test file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "document.pdf")
	testContent := []byte("PDF content here")
	err := os.WriteFile(tempFile, testContent, 0644)
	require.NoError(t, err)

	// Create attachment with specific metadata
	originalFileName := "important-document.pdf"
	originalContentType := "application/pdf"
	originalSize := int64(len(testContent))

	attachment := factories.NewAttachment().
		WithStorageKey(tempFile).
		WithFileName(originalFileName).
		WithContentType(originalContentType).
		WithFileSize(originalSize).
		Create(db)

	// Create worker
	worker := &UploadWorker{
		DB:             db,
		StorageService: mockStorage,
	}

	// Execute worker
	job := &river.Job[UploadArgs]{
		Args: UploadArgs{
			AttachmentID: attachment.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify metadata preserved
	var updated models.Attachment
	db.First(&updated, attachment.ID)

	assert.Equal(t, originalFileName, updated.FileName)
	assert.Equal(t, originalContentType, updated.ContentType)
	assert.Equal(t, originalSize, updated.FileSize)
}
