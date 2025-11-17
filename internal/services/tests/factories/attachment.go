package factories

import (
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type AttachmentBuilder struct {
	attachment *models.Attachment
}

func NewAttachment() *AttachmentBuilder {
	return &AttachmentBuilder{
		attachment: &models.Attachment{
			HolderType:   "interventions",
			HolderID:     1,
			Kind:         "photo",
			StorageKey:   "test/default.jpg",
			FileName:     "default.jpg",
			ContentType:  "image/jpeg",
			FileSize:     1024,
			UploadStatus: models.AttachmentUploadStatusPending,
		},
	}
}

func (b *AttachmentBuilder) WithHolderType(holderType string) *AttachmentBuilder {
	b.attachment.HolderType = holderType
	return b
}

func (b *AttachmentBuilder) WithHolderID(holderID uint) *AttachmentBuilder {
	b.attachment.HolderID = holderID
	return b
}

func (b *AttachmentBuilder) WithKind(kind string) *AttachmentBuilder {
	b.attachment.Kind = kind
	return b
}

func (b *AttachmentBuilder) WithStorageKey(storageKey string) *AttachmentBuilder {
	b.attachment.StorageKey = storageKey
	return b
}

func (b *AttachmentBuilder) WithFileName(fileName string) *AttachmentBuilder {
	b.attachment.FileName = fileName
	return b
}

func (b *AttachmentBuilder) WithContentType(contentType string) *AttachmentBuilder {
	b.attachment.ContentType = contentType
	return b
}

func (b *AttachmentBuilder) WithFileSize(fileSize int64) *AttachmentBuilder {
	b.attachment.FileSize = fileSize
	return b
}

func (b *AttachmentBuilder) WithUploadStatus(status models.AttachmentUploadStatus) *AttachmentBuilder {
	b.attachment.UploadStatus = status
	return b
}

func (b *AttachmentBuilder) Build() *models.Attachment {
	return b.attachment
}

func (b *AttachmentBuilder) Create(db *gorm.DB) *models.Attachment {
	db.Create(b.attachment)
	return b.attachment
}
