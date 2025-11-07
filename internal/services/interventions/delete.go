package interventions

import (
	"context"
	"fmt"
	"log"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type DeleteInterventionService struct {
	DB             *gorm.DB
	StorageService storage.StorageService
}

// Delete removes an intervention and all its associated attachments.
// It first deletes database records within a transaction (which can be rolled back),
// then deletes the physical files from storage after the transaction commits successfully.
// This ensures data consistency even if storage operations fail.
func (s *DeleteInterventionService) Delete(ctx context.Context, interventionID uint) error {
	// Load intervention with attachments
	var intervention models.Intervention
	if err := s.DB.Preload("Attachments").First(&intervention, interventionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("intervention with ID %d not found", interventionID)
		}
		return fmt.Errorf("failed to load intervention: %w", err)
	}

	// Collect storage keys before deleting DB records
	storageKeys := make([]string, 0, len(intervention.Attachments))
	for _, attachment := range intervention.Attachments {
		storageKeys = append(storageKeys, attachment.StorageKey)
	}

	// Delete database records in transaction (rollback-safe)
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Delete all attachment records
		for _, attachment := range intervention.Attachments {
			if err := tx.Delete(&attachment).Error; err != nil {
				return fmt.Errorf("failed to delete attachment record %d: %w", attachment.ID, err)
			}
		}

		// Delete intervention record
		if err := tx.Delete(&intervention).Error; err != nil {
			return fmt.Errorf("failed to delete intervention record %d: %w", intervention.ID, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// After successful DB transaction, delete files from storage
	// We log errors but don't fail the operation since DB records are already deleted
	for _, storageKey := range storageKeys {
		if err := s.StorageService.DeleteFile(ctx, storageKey); err != nil {
			// Log the error but continue - consider implementing a cleanup job for orphaned files
			log.Printf("WARNING: failed to delete file from storage (key: %s): %v", storageKey, err)
		}
	}

	return nil
}
