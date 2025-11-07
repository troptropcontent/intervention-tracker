package attachments

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type AttachmentService struct {
	db             *gorm.DB
	storageService storage.StorageService
}

// NewAttachmentService creates a new attachment service instance
func NewAttachmentService(db *gorm.DB, storageService storage.StorageService) *AttachmentService {
	return &AttachmentService{
		db:             db,
		storageService: storageService,
	}
}

// Attach uploads a file to storage and creates an attachment record in the database.
// The caller is responsible for closing the file after this method returns.
// The content type is automatically detected from the file contents.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - file: The file to upload (must implement io.ReadSeeker, caller must close)
//   - filename: The original filename (with extension)
//   - holder_id: ID of the record to attach the file to
//   - holder_type: Type of the record (e.g., "interventions", "portals")
//   - kind: The kind/category of attachment (e.g., "report", "photo")
//
// Returns an error if the upload or database operation fails.
// If database creation fails, the uploaded file is automatically cleaned up.
func (a *AttachmentService) Attach(ctx context.Context, file io.ReadSeeker, filename string, holder_id uint, holder_type string, kind string) error {
	// STEP 1: Get file size
	// We move the bookmark to the very end of the file and ask "what position am I at?"
	// That position IS the file size (e.g., if the bookmark is at byte 1024, the file is 1024 bytes)
	fileSize, err := file.Seek(0, io.SeekEnd) // Move bookmark to end
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	// STEP 2: Move bookmark back to the beginning
	// We need to read from the start to detect content type
	// Without this, we'd be reading from the END (where there's nothing to read!)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer to start: %w", err)
	}

	// STEP 3: Read first 512 bytes to detect content type
	// The bookmark starts at position 0 (beginning)
	// After this Read(), the bookmark will be at position 512 (or less if file is smaller)
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file for content type detection: %w", err)
	}

	// STEP 4: Reset bookmark back to beginning AGAIN
	// The upload function needs to read the ENTIRE file from the start
	// If we don't reset, it would only upload from byte 512 onwards (missing the first 512 bytes!)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer for upload: %w", err)
	}

	// Detect content type from the first 512 bytes we read
	contentType := http.DetectContentType(buffer[:n])

	// If DetectContentType returns generic "application/octet-stream",
	// try to get a more specific type from the file extension
	if contentType == "application/octet-stream" {
		ext := filepath.Ext(filename)
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			contentType = mimeType
		}
	}

	// Generate storage key with file extension
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("%s/%d/%s/%s%s", holder_type, holder_id, kind, uuid.New().String(), ext)

	// Upload the file (this will read from current position - which is 0, the start)
	_, err = a.storageService.UploadFile(ctx, key, file, contentType)
	if err != nil {
		return err
	}

	// Create database record
	attachment := models.Attachment{
		HolderType:  holder_type,
		Kind:        kind,
		HolderID:    holder_id,
		StorageKey:  key,
		FileName:    filename,
		ContentType: contentType,
		FileSize:    fileSize,
		UploadedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}

	if err := a.db.Create(&attachment).Error; err != nil {
		log.Printf("Failed to create attachment record for %s %d: %v", holder_type, holder_id, err)
		// Attempt to clean up uploaded file
		if cleanupErr := a.storageService.DeleteFile(ctx, key); cleanupErr != nil {
			log.Printf("Failed to cleanup storage file after attachment creation failure: %v", cleanupErr)
		}
		return err
	}
	return nil
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
