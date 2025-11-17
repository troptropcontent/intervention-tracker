package attachments

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/riverqueue/river"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs/types"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type UploadWorker struct {
	river.WorkerDefaults[types.UploadArgs]
	DB             *gorm.DB
	StorageService storage.StorageService
}

// Work processes an attachment upload by:
// 1. Reading the file from the temporary path stored in attachment.StorageKey
// 2. Uploading to S3
// 3. Updating the attachment record with the final S3 key
// 4. Cleaning up the temporary file
func (w *UploadWorker) Work(ctx context.Context, job *river.Job[types.UploadArgs]) error {
	// Fetch the attachment record
	var attachment models.Attachment
	if err := w.DB.First(&attachment, job.Args.AttachmentID).Error; err != nil {
		return fmt.Errorf("failed to find attachment %d: %w", job.Args.AttachmentID, err)
	}

	// Skip if already completed (idempotency check)
	if attachment.UploadStatus == models.AttachmentUploadStatusCompleted {
		// File already uploaded, just clean up temp file if it still exists
		if _, err := os.Stat(attachment.StorageKey); err == nil {
			// Temp file exists, try to remove it
			if err := os.Remove(attachment.StorageKey); err != nil {
				fmt.Printf("WARNING: failed to remove already-processed temp file %s: %v\n", attachment.StorageKey, err)
			}
		}
		return nil
	}

	// Open the temporary file (StorageKey contains the temp file path before upload)
	tempFilePath := attachment.StorageKey
	file, err := os.Open(tempFilePath)
	if err != nil {
		// Return error to trigger River retry
		return fmt.Errorf("failed to open temporary file %s: %w", tempFilePath, err)
	}
	defer file.Close()

	// Generate S3 storage key
	// Format: {holder_type}/{holder_id}/{kind}/{timestamp}_{filename}
	timestamp := time.Now().Unix()
	storageKey := fmt.Sprintf("%s/%d/%s/%d_%s",
		attachment.HolderType,
		attachment.HolderID,
		attachment.Kind,
		timestamp,
		attachment.FileName,
	)

	// Upload to S3 (this is NOT transactional - S3 is external)
	_, err = w.StorageService.UploadFile(ctx, storageKey, file, attachment.ContentType)
	if err != nil {
		// Return error to trigger River retry
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Update attachment record with S3 key and mark as completed
	now := time.Now()
	if err := w.DB.Model(&attachment).Updates(map[string]any{
		"storage_key":   storageKey,
		"upload_status": models.AttachmentUploadStatusCompleted,
		"uploaded_at":   sql.NullTime{Time: now, Valid: true},
	}).Error; err != nil {
		// The file is uploaded but we couldn't update the record
		// This is a critical error - the file exists in S3 but the DB doesn't reflect it
		// Return error to trigger River retry (which will try to upload again with new timestamp)
		return fmt.Errorf("failed to update attachment record after successful upload: %w", err)
	}

	// Clean up temporary file only after successful DB update
	// We do this outside the transaction since file operations can't be rolled back
	// If this fails, we log but don't return error - the upload was successful
	if err := os.Remove(tempFilePath); err != nil {
		fmt.Printf("WARNING: failed to remove temporary file %s: %v\n", tempFilePath, err)
	}

	return nil
}
