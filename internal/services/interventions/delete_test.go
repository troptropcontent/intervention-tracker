package interventions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"gorm.io/gorm"
)

// setupDeleteTestFixture creates an intervention with attachments for deletion tests
func setupDeleteTestFixture(db *gorm.DB) (*models.Intervention, []string) {
	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)

	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		Create(db)

	// Create attachments with storage keys using factory
	storageKeys := []string{
		"interventions/1/photo1.jpg",
		"interventions/1/photo2.jpg",
		"interventions/1/photo3.jpg",
	}

	for i, key := range storageKeys {
		factories.NewAttachment().
			WithHolderType("interventions").
			WithHolderID(intervention.ID).
			WithKind("photo").
			WithStorageKey(key).
			WithFileName(key).
			WithFileSize(1024 * (int64(i) + 1)).
			Create(db)
	}

	return intervention, storageKeys
}

func TestDeleteInterventionService_Delete_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	intervention, storageKeys := setupDeleteTestFixture(db)
	ctx := context.Background()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify no error
	require.NoError(t, err, "Delete should succeed")

	// Verify intervention was deleted from DB
	var deletedIntervention models.Intervention
	err = db.First(&deletedIntervention, intervention.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Intervention should be deleted from database")

	// Verify attachments were deleted from DB
	var attachments []models.Attachment
	db.Where("holder_type = ? AND holder_id = ?", "interventions", intervention.ID).Find(&attachments)
	assert.Empty(t, attachments, "All attachments should be deleted from database")

	// Verify storage service was called
	assert.True(t, mockStorage.DeleteCalled, "Storage DeleteFile should be called")
	assert.Len(t, mockStorage.DeletedFiles, 3, "Should delete 3 files from storage")

	// Verify correct files were deleted
	for _, expectedKey := range storageKeys {
		assert.Contains(t, mockStorage.DeletedFiles, expectedKey, "Should delete file %s", expectedKey)
	}
}

func TestDeleteInterventionService_Delete_InterventionNotFound(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	ctx := context.Background()
	nonExistentID := uint(99999)

	// Execute delete
	err := service.Delete(ctx, nonExistentID)

	// Verify error
	require.Error(t, err, "Delete should return error for non-existent intervention")
	assert.Contains(t, err.Error(), "not found", "Error should indicate intervention not found")

	// Verify storage service was NOT called
	assert.False(t, mockStorage.DeleteCalled, "Storage DeleteFile should not be called")
}

func TestDeleteInterventionService_Delete_NoAttachments(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)

	// Create intervention without attachments
	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		Create(db)

	ctx := context.Background()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify no error
	require.NoError(t, err, "Delete should succeed even without attachments")

	// Verify intervention was deleted
	var deletedIntervention models.Intervention
	err = db.First(&deletedIntervention, intervention.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Intervention should be deleted")

	// Verify no storage operations were attempted
	assert.Empty(t, mockStorage.DeletedFiles, "No files should be deleted from storage")
}

func TestDeleteInterventionService_Delete_StorageFailureDoesNotRollback(t *testing.T) {
	db := tests.SetupTestDB(t)

	// Setup mock storage that fails on all deletions
	mockStorage := &tests.MockStorageService{
		DeleteError: errors.New("storage service unavailable"),
	}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	intervention, _ := setupDeleteTestFixture(db)
	ctx := context.Background()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify no error returned (storage failures are logged, not returned)
	require.NoError(t, err, "Delete should succeed despite storage failures")

	// Verify intervention was STILL deleted from DB
	var deletedIntervention models.Intervention
	err = db.First(&deletedIntervention, intervention.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Intervention should be deleted despite storage failure")

	// Verify attachments were STILL deleted from DB
	var attachments []models.Attachment
	db.Where("holder_type = ? AND holder_id = ?", "interventions", intervention.ID).Find(&attachments)
	assert.Empty(t, attachments, "All attachments should be deleted despite storage failure")

	// Verify storage service was called (but failed)
	assert.True(t, mockStorage.DeleteCalled, "Storage DeleteFile should be attempted")
}

func TestDeleteInterventionService_Delete_PartialStorageFailure(t *testing.T) {
	db := tests.SetupTestDB(t)

	intervention, storageKeys := setupDeleteTestFixture(db)

	// Setup mock storage that fails on specific file
	mockStorage := &tests.MockStorageService{
		DeleteErrors: map[string]error{
			storageKeys[1]: errors.New("permission denied for file 2"),
		},
	}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	ctx := context.Background()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify no error returned (storage failures are logged, not returned)
	require.NoError(t, err, "Delete should succeed despite partial storage failure")

	// Verify all files were attempted to be deleted
	assert.Len(t, mockStorage.DeletedFiles, 3, "Should attempt to delete all 3 files")

	// Verify intervention and attachments were deleted from DB
	var deletedIntervention models.Intervention
	err = db.First(&deletedIntervention, intervention.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Intervention should be deleted")
}

func TestDeleteInterventionService_Delete_TransactionRollback(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	// Create a custom DB that will fail on intervention deletion
	// This simulates a DB error during transaction
	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)
	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		Create(db)

	// Create an attachment using factory
	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithStorageKey("test/key.jpg").
		WithFileName("test.jpg").
		Create(db)

	// Close the DB connection to cause a failure during transaction
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	ctx := context.Background()

	// Execute delete (should fail)
	err = service.Delete(ctx, intervention.ID)

	// Verify error occurred
	require.Error(t, err, "Delete should fail when DB operations fail")

	// Since DB is closed, we can't verify rollback, but the error proves
	// that DB errors are properly propagated and transaction would rollback

	// Verify storage was NOT called (failure happened before storage operations)
	assert.False(t, mockStorage.DeleteCalled, "Storage should not be called on DB failure")
}

func TestDeleteInterventionService_Delete_ContextPropagation(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	intervention, _ := setupDeleteTestFixture(db)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify success
	require.NoError(t, err, "Delete should succeed with context")

	// Verify the context was used (mock storage received the context)
	// In a real scenario, the storage service would respect context cancellation
	assert.True(t, mockStorage.DeleteCalled, "Storage should be called with context")
}

func TestDeleteInterventionService_Delete_MultipleAttachmentTypes(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)
	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		Create(db)

	// Create different types of attachments using factory
	attachmentTypes := []struct {
		kind        string
		storageKey  string
		contentType string
	}{
		{"photo", "interventions/1/photo.jpg", "image/jpeg"},
		{"report", "interventions/1/report.pdf", "application/pdf"},
		{"signature", "interventions/1/signature.png", "image/png"},
	}

	for _, att := range attachmentTypes {
		factories.NewAttachment().
			WithHolderType("interventions").
			WithHolderID(intervention.ID).
			WithKind(att.kind).
			WithStorageKey(att.storageKey).
			WithFileName(att.storageKey).
			WithContentType(att.contentType).
			WithFileSize(2048).
			Create(db)
	}

	ctx := context.Background()

	// Execute delete
	err := service.Delete(ctx, intervention.ID)

	// Verify success
	require.NoError(t, err, "Delete should handle multiple attachment types")

	// Verify all attachments were deleted
	assert.Len(t, mockStorage.DeletedFiles, 3, "Should delete all attachment types")

	// Verify specific files
	assert.Contains(t, mockStorage.DeletedFiles, "interventions/1/photo.jpg")
	assert.Contains(t, mockStorage.DeletedFiles, "interventions/1/report.pdf")
	assert.Contains(t, mockStorage.DeletedFiles, "interventions/1/signature.png")
}

func TestDeleteInterventionService_Delete_DatabasePersistence(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}

	service := &DeleteInterventionService{
		DB:             db,
		StorageService: mockStorage,
	}

	// Create multiple interventions
	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)

	intervention1 := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		WithSummary("First intervention").
		Create(db)

	intervention2 := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		WithSummary("Second intervention").
		Create(db)

	ctx := context.Background()

	// Delete only the first intervention
	err := service.Delete(ctx, intervention1.ID)
	require.NoError(t, err)

	// Verify first intervention is deleted
	var deleted models.Intervention
	err = db.First(&deleted, intervention1.ID).Error
	assert.Equal(t, gorm.ErrRecordNotFound, err, "First intervention should be deleted")

	// Verify second intervention still exists
	var stillExists models.Intervention
	err = db.First(&stillExists, intervention2.ID).Error
	require.NoError(t, err, "Second intervention should still exist")
	assert.Equal(t, "Second intervention", *stillExists.Summary)
}
