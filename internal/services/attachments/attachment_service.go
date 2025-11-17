package attachments

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/troptropcontent/qr_code_maintenance/internal/jobs"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs/types"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
	"gorm.io/gorm"
)

type AttachmentService struct {
	db             *gorm.DB
	storageService storage.StorageService
	jobRunner      jobs.BackgroundJobRunner
}

// NewAttachmentService creates a new attachment service instance
func NewAttachmentService(db *gorm.DB, storageService storage.StorageService, jobRunner jobs.BackgroundJobRunner) *AttachmentService {
	return &AttachmentService{
		db:             db,
		storageService: storageService,
		jobRunner:      jobRunner,
	}
}

// Attach saves the file to a temporary directory and enqueues a background job to upload it to S3.
// This method is useful for large files or when you want to return a response quickly without waiting for S3 upload.
//
// Workflow:
// 1. Saves the file to a temporary directory
// 2. Creates an attachment record with the temp file path in StorageKey
// 3. Enqueues a background job to upload the file to S3
// 4. Returns immediately (the upload happens asynchronously)
// 5. Worker uploads to S3, updates StorageKey, and marks UploadStatus as "completed"
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - file: The file to upload (must implement io.Reader, caller must close)
//   - filename: The original filename (with extension)
//   - holder_id: ID of the record to attach the file to
//   - holder_type: Type of the record (e.g., "interventions", "portals")
//   - kind: The kind/category of attachment (e.g., "report", "photo")
//
// Returns the created attachment ID and an error if the operation fails.
func (a *AttachmentService) Attach(ctx context.Context, file io.ReadSeeker, filename string, holder_id uint, holder_type string, kind string) (uint, error) {
	// Get temporary directory path from environment or use default
	tempDir := utils.GetEnv("UPLOAD_TEMP_DIR", "/tmp/qr_code_uploads")

	// Create temporary file using utility function
	tempFilePath, fileSize, err := utils.CreateTempFile(tempDir, filename, file)
	if err != nil {
		return 0, err
	}

	// Detect content type from the temporary file
	contentType, err := utils.DetectContentTypeFromPath(tempFilePath)
	if err != nil {
		return 0, err
	}

	// Create attachment record with temp file path
	// Note: StorageKey will be updated to the S3 key by the worker
	// UploadStatus starts as empty/NULL and will be set to "completed" by the worker
	attachment := models.Attachment{
		HolderType:   holder_type,
		Kind:         kind,
		HolderID:     holder_id,
		StorageKey:   tempFilePath, // Store temp path - will be replaced with S3 key by worker
		FileName:     filename,
		ContentType:  contentType,
		FileSize:     fileSize,
		UploadStatus: models.AttachmentUploadStatusPending,
	}

	if err := a.db.Create(&attachment).Error; err != nil {
		return 0, fmt.Errorf("failed to create attachment record: %w", err)
	}

	job := types.UploadArgs{
		AttachmentID: attachment.ID,
	}

	// Enqueue background job to upload to S3
	if err := a.jobRunner.Enqueue(ctx, job); err != nil {
		// If job enqueue fails, mark attachment as failed but don't delete it
		// This allows for manual retry or inspection
		a.db.Model(&attachment).Updates(map[string]interface{}{
			"upload_status": models.AttachmentUploadStatusFailed,
			"error_message": fmt.Sprintf("failed to enqueue upload job: %v", err),
		})
		return 0, fmt.Errorf("failed to enqueue upload job: %w", err)
	}

	return attachment.ID, nil
}

// Delete removes a single attachment by ID.
// It first deletes the database record, then deletes the file from storage.
// Storage deletion errors are logged but do not cause the operation to fail.
func (a *AttachmentService) Delete(ctx context.Context, attachmentID uint) error {
	// Load the attachment to get storage key
	var attachment models.Attachment
	if err := a.db.First(&attachment, attachmentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("attachment with ID %d not found", attachmentID)
		}
		return fmt.Errorf("failed to load attachment: %w", err)
	}

	// Delete from database first (rollback-safe)
	if err := a.db.Delete(&attachment).Error; err != nil {
		return fmt.Errorf("failed to delete attachment record %d: %w", attachmentID, err)
	}

	// Delete from storage after successful DB deletion
	if err := a.storageService.DeleteFile(ctx, attachment.StorageKey); err != nil {
		log.Printf("WARNING: failed to delete file from storage (key: %s): %v", attachment.StorageKey, err)
	}

	return nil
}

// DeleteByHolder deletes all attachments for a specific holder (e.g., all attachments for an intervention).
// It deletes database records using the current DB connection (which may be a transaction),
// then deletes files from storage. Storage deletion errors are logged but do not fail the operation.
// NOTE: If called with a transaction DB, this will participate in that transaction.
// If called with a regular DB, it will operate directly without creating a new transaction.
func (a *AttachmentService) DeleteByHolder(ctx context.Context, holderType string, holderID uint) error {
	// Load all attachments for this holder
	var attachments []models.Attachment
	if err := a.db.Where("holder_type = ? AND holder_id = ?", holderType, holderID).Find(&attachments).Error; err != nil {
		return fmt.Errorf("failed to load attachments for %s %d: %w", holderType, holderID, err)
	}

	// If no attachments, nothing to do
	if len(attachments) == 0 {
		return nil
	}

	// Collect storage keys before deleting DB records
	storageKeys := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		storageKeys = append(storageKeys, attachment.StorageKey)
	}

	// Delete database records using the provided DB connection (may be a transaction)
	for _, attachment := range attachments {
		if err := a.db.Delete(&attachment).Error; err != nil {
			return fmt.Errorf("failed to delete attachment record %d: %w", attachment.ID, err)
		}
	}

	// After successful DB deletion, delete files from storage
	// We log errors but don't fail the operation since DB records are already deleted
	for _, storageKey := range storageKeys {
		if err := a.storageService.DeleteFile(ctx, storageKey); err != nil {
			log.Printf("WARNING: failed to delete file from storage (key: %s): %v", storageKey, err)
		}
	}

	return nil
}
