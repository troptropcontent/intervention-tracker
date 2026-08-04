package qrcodes

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

const (
	MinBatchCount = 1
	MaxBatchCount = 500
)

type GenerateService struct {
	DB *gorm.DB
}

func NewGenerateService(db *gorm.DB) (*GenerateService, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	return &GenerateService{DB: db}, nil
}

// GenerateBatch creates a new batch of QR codes, each with a fresh UUID and
// status "available", ready to be printed and later associated to a portal.
func (s *GenerateService) GenerateBatch(count int, label *string) (*models.QRCodeBatch, error) {
	if count < MinBatchCount || count > MaxBatchCount {
		return nil, fmt.Errorf("count must be between %d and %d", MinBatchCount, MaxBatchCount)
	}

	batch := models.QRCodeBatch{Label: label, Count: count}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to create batch: %w", err)
		}

		codes := make([]models.QRCode, count)
		for i := range codes {
			codes[i] = models.QRCode{
				UUID:    uuid.New().String(),
				Status:  models.QRCodeStatusAvailable,
				BatchID: &batch.ID,
			}
		}

		if err := tx.Create(&codes).Error; err != nil {
			return fmt.Errorf("failed to create qr codes: %w", err)
		}

		batch.QRCodes = codes
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &batch, nil
}
