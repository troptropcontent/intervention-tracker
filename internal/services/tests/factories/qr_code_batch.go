package factories

import (
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type QRCodeBatchBuilder struct {
	batch *models.QRCodeBatch
}

func NewQRCodeBatch() *QRCodeBatchBuilder {
	return &QRCodeBatchBuilder{
		batch: &models.QRCodeBatch{
			Count: 10,
		},
	}
}

func (b *QRCodeBatchBuilder) WithLabel(label string) *QRCodeBatchBuilder {
	b.batch.Label = &label
	return b
}

func (b *QRCodeBatchBuilder) WithCount(count int) *QRCodeBatchBuilder {
	b.batch.Count = count
	return b
}

func (b *QRCodeBatchBuilder) Build() *models.QRCodeBatch {
	return b.batch
}

func (b *QRCodeBatchBuilder) Create(db *gorm.DB) *models.QRCodeBatch {
	db.Create(b.batch)
	return b.batch
}
