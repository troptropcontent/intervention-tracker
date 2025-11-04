package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// Attachment represents a file stored in external storage (S3)
// Uses polymorphic associations to attach to any model
type Attachment struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	HolderType  string         `json:"holder_type" gorm:"not null;index:idx_holder"` // e.g., "interventions", "portals"
	HolderID    uint           `json:"holder_id" gorm:"not null;index:idx_holder"`   // ID of the parent record
	Kind        string         `json:"kind" gorm:"not null;index:idx_kind"`          // Kind of attachement ie report, photo
	StorageKey  string         `json:"storage_key" gorm:"not null"`                  // S3 key/path
	FileName    string         `json:"file_name" gorm:"not null"`                    // Original filename
	ContentType string         `json:"content_type" gorm:"type:varchar(255)"`        // MIME type (e.g., "application/pdf")
	FileSize    int64          `json:"file_size"`                                    // Size in bytes
	UploadedAt  sql.NullTime   `json:"uploaded_at"`                                  // When the file was uploaded
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Attachment) TableName() string {
	return "attachments"
}
