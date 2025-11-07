package interventions

import (
	"context"
	"fmt"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type DeleteInterventionService struct {
	DB             *gorm.DB
	StorageService storage.StorageService
}

// Delete removes an intervention and all its associated attachments.
// It delegates attachment deletion to the AttachmentService using the transaction,
// ensuring atomicity across intervention and attachment deletions.
func (s *DeleteInterventionService) Delete(ctx context.Context, interventionID uint) error {
	// Verify intervention exists
	var intervention models.Intervention
	if err := s.DB.First(&intervention, interventionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("intervention with ID %d not found", interventionID)
		}
		return fmt.Errorf("failed to load intervention: %w", err)
	}

	// Delete in transaction for atomicity
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Create attachment service with the transaction DB
		attachmentService := attachments.NewAttachmentService(tx, s.StorageService)

		// Delete all attachments (DB + storage)
		if err := attachmentService.DeleteByHolder(ctx, "interventions", interventionID); err != nil {
			return fmt.Errorf("failed to delete attachments: %w", err)
		}

		// Delete intervention record
		if err := tx.Delete(&intervention).Error; err != nil {
			return fmt.Errorf("failed to delete intervention record %d: %w", intervention.ID, err)
		}

		return nil
	})
}
