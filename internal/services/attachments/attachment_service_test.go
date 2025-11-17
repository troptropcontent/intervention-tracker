package attachments

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"gorm.io/gorm"
)

// Tests for Delete method

func TestAttachmentService_Delete_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	// Create an attachment using factory
	attachment := factories.NewAttachment().
		WithStorageKey("test/file.jpg").
		WithFileName("file.jpg").
		Create(db)

	ctx := context.Background()

	// Delete the attachment
	err := service.Delete(ctx, attachment.ID)

	// Verify success
	require.NoError(t, err)

	// Verify attachment was deleted from DB
	var deleted models.Attachment
	err = db.First(&deleted, attachment.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Attachment should be deleted from database")

	// Verify storage service was called
	assert.True(t, mockStorage.DeleteCalled, "Storage DeleteFile should be called")
	assert.Contains(t, mockStorage.DeletedFiles, "test/file.jpg")
}

func TestAttachmentService_Delete_NotFound(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	ctx := context.Background()

	// Try to delete non-existent attachment
	err := service.Delete(ctx, 99999)

	// Verify error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Verify storage was not called
	assert.False(t, mockStorage.DeleteCalled)
}

func TestAttachmentService_Delete_StorageFailure(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{
		DeleteError: errors.New("storage unavailable"),
	}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	// Create an attachment
	attachment := factories.NewAttachment().
		WithStorageKey("test/file.jpg").
		Create(db)

	ctx := context.Background()

	// Delete the attachment (should succeed despite storage failure)
	err := service.Delete(ctx, attachment.ID)

	// Verify success (storage errors are logged, not returned)
	require.NoError(t, err)

	// Verify attachment was deleted from DB
	var deleted models.Attachment
	err = db.First(&deleted, attachment.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// Verify storage was called (but failed)
	assert.True(t, mockStorage.DeleteCalled)
}

// Tests for DeleteByHolder method

func TestAttachmentService_DeleteByHolder_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	// Create multiple attachments for the same holder using factory
	factories.NewAttachment().
		WithHolderID(1).
		WithKind("photo").
		WithStorageKey("test/file1.jpg").
		WithFileName("file1.jpg").
		Create(db)

	factories.NewAttachment().
		WithHolderID(1).
		WithKind("photo").
		WithStorageKey("test/file2.jpg").
		WithFileName("file2.jpg").
		WithFileSize(2048).
		Create(db)

	factories.NewAttachment().
		WithHolderID(1).
		WithKind("report").
		WithStorageKey("test/report.pdf").
		WithFileName("report.pdf").
		WithContentType("application/pdf").
		WithFileSize(4096).
		Create(db)

	ctx := context.Background()

	// Delete all attachments for the holder
	err := service.DeleteByHolder(ctx, "interventions", 1)

	// Verify success
	require.NoError(t, err)

	// Verify all attachments were deleted from DB
	var remaining []models.Attachment
	db.Where("holder_type = ? AND holder_id = ?", "interventions", 1).Find(&remaining)
	assert.Empty(t, remaining, "All attachments should be deleted")

	// Verify storage service was called for all files
	assert.True(t, mockStorage.DeleteCalled)
	assert.Len(t, mockStorage.DeletedFiles, 3)
	assert.Contains(t, mockStorage.DeletedFiles, "test/file1.jpg")
	assert.Contains(t, mockStorage.DeletedFiles, "test/file2.jpg")
	assert.Contains(t, mockStorage.DeletedFiles, "test/report.pdf")
}

func TestAttachmentService_DeleteByHolder_NoAttachments(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	ctx := context.Background()

	// Delete attachments for non-existent holder
	err := service.DeleteByHolder(ctx, "interventions", 999)

	// Verify success (no-op)
	require.NoError(t, err)

	// Verify storage was not called
	assert.False(t, mockStorage.DeleteCalled)
}

func TestAttachmentService_DeleteByHolder_PartialStorageFailure(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{
		DeleteErrors: map[string]error{
			"test/file2.jpg": errors.New("permission denied"),
		},
	}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	// Create multiple attachments
	factories.NewAttachment().
		WithHolderID(1).
		WithStorageKey("test/file1.jpg").
		Create(db)

	factories.NewAttachment().
		WithHolderID(1).
		WithStorageKey("test/file2.jpg").
		Create(db)

	ctx := context.Background()

	// Delete attachments (should succeed despite partial storage failure)
	err := service.DeleteByHolder(ctx, "interventions", 1)

	// Verify success (storage errors are logged, not returned)
	require.NoError(t, err)

	// Verify all attachments were deleted from DB
	var remaining []models.Attachment
	db.Where("holder_type = ? AND holder_id = ?", "interventions", 1).Find(&remaining)
	assert.Empty(t, remaining)

	// Verify storage was called for all files
	assert.Len(t, mockStorage.DeletedFiles, 2)
}

func TestAttachmentService_DeleteByHolder_MultipleHolders(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	service := NewAttachmentService(db, mockStorage, mockRunner)

	// Create attachments for two different holders
	factories.NewAttachment().
		WithHolderID(1).
		WithStorageKey("test/intervention1.jpg").
		Create(db)

	factories.NewAttachment().
		WithHolderID(2).
		WithStorageKey("test/intervention2.jpg").
		Create(db)

	ctx := context.Background()

	// Delete attachments for holder 1 only
	err := service.DeleteByHolder(ctx, "interventions", 1)
	require.NoError(t, err)

	// Verify only holder 1's attachments were deleted
	var remaining []models.Attachment
	db.Where("holder_type = ? AND holder_id = ?", "interventions", 1).Find(&remaining)
	assert.Empty(t, remaining, "Holder 1 attachments should be deleted")

	db.Where("holder_type = ? AND holder_id = ?", "interventions", 2).Find(&remaining)
	assert.Len(t, remaining, 1, "Holder 2 attachments should remain")

	// Verify only one file was deleted from storage
	assert.Len(t, mockStorage.DeletedFiles, 1)
	assert.Contains(t, mockStorage.DeletedFiles, "test/intervention1.jpg")
}

func TestAttachmentService_DeleteByHolder_WithTransaction(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockRunner := &tests.MockBackgroundJobRunner{}

	// Create an attachment
	attachment := factories.NewAttachment().
		WithHolderID(1).
		WithStorageKey("test/file.jpg").
		Create(db)

	ctx := context.Background()

	// Use a transaction and roll it back
	err := db.Transaction(func(tx *gorm.DB) error {
		service := NewAttachmentService(db, mockStorage, mockRunner)

		// Delete within transaction
		if err := service.DeleteByHolder(ctx, "interventions", 1); err != nil {
			return err
		}

		// Force rollback
		return errors.New("intentional rollback")
	})

	// Verify transaction was rolled back
	require.Error(t, err)

	// Verify attachment still exists in DB (rollback worked)
	var stillExists models.Attachment
	err = db.First(&stillExists, attachment.ID).Error
	require.NoError(t, err, "Attachment should still exist after rollback")

	// Note: Storage files may have been deleted (can't rollback storage operations)
	// This is expected behavior and why we log storage errors
}
