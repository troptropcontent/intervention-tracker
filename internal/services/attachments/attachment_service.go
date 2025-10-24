package attachments

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
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
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - file: The file to upload (caller must close this file)
//   - record_id: ID of the record to attach the file to
//   - record_type: Type of the record (e.g., "interventions", "portals")
//   - contentType: MIME type of the file (e.g., "application/pdf")
//
// Returns an error if the upload or database operation fails.
// If database creation fails, the uploaded file is automatically cleaned up.
func (a *AttachmentService) Attach(ctx context.Context, file *os.File, record_id uint, record_type string, contentType string) error {
	// Get file info for metadata
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	ext := filepath.Ext(fileInfo.Name())
	key := fmt.Sprintf("%s/%d/%s%s", record_type, record_id, uuid.New().String(), ext)

	_, err = a.storageService.UploadFile(ctx, key, file, contentType)
	if err != nil {
		return err
	}

	attachment := models.Attachment{
		HolderType:  record_type,
		HolderID:    record_id,
		StorageKey:  key,
		FileName:    fileInfo.Name(),
		ContentType: contentType,
		FileSize:    fileInfo.Size(),
		UploadedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}

	if err := a.db.Create(&attachment).Error; err != nil {
		log.Printf("Failed to create attachment record for %s %d: %v", record_type, record_id, err)
		// Attempt to clean up uploaded file
		if cleanupErr := a.storageService.DeleteFile(ctx, key); cleanupErr != nil {
			log.Printf("Failed to cleanup storage file after attachment creation failure: %v", cleanupErr)
		}
		return err
	}
	return nil
}
