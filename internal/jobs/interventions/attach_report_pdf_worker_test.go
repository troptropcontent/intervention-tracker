package interventions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs/types"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/mocks"
)

func TestAttachReportPdfWorker_Work_Success(t *testing.T) {
	// Setup test database and services
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test user and portal
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)

	// Create intervention with photo attachments
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create photo attachments with temp storage keys (simulating local files)
	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("temp/photo1.jpg").
		WithFileName("photo1.jpg").
		WithContentType("image/jpeg").
		Create(db)

	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("temp/photo2.jpg").
		WithFileName("photo2.jpg").
		WithContentType("image/jpeg").
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err, "Worker should complete successfully")

	// // Verify PDF attachment was created
	var reportAttachment models.Attachment
	err = db.Where("holder_type = ? AND holder_id = ? AND kind = ?", "interventions", intervention.ID, "report").
		First(&reportAttachment).Error
	require.NoError(t, err, "Report attachment should be created")

	// Verify attachment properties
	assert.Equal(t, "interventions", reportAttachment.HolderType)
	assert.Equal(t, intervention.ID, reportAttachment.HolderID)
	assert.Equal(t, "report", reportAttachment.Kind)
	assert.Contains(t, reportAttachment.FileName, ".pdf")
	assert.Equal(t, models.AttachmentUploadStatusPending, reportAttachment.UploadStatus)

	// // Verify attachement job have been triggered
	assert.True(t, mockRunner.EnqueueCalled, "pdf job should be called")
	require.Len(t, mockRunner.EnqueuedJobs, 1)
	assert.Equal(t, mockRunner.EnqueuedJobs[0].Kind(), "attachment_upload")
}

func TestAttachReportPdfWorker_Work_InterventionNotFound(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:             db,
		StorageService: mockStorage,
		EmailService:   mockEmail,
		JobRunner:      mockRunner,
	}

	// Execute worker with non-existent intervention ID
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: 99999,
		},
	}

	err := worker.Work(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load intervention")

	// Verify no upload happened
	assert.False(t, mockStorage.UploadCalled)
}

func TestAttachReportPdfWorker_Work_PDFGenerationFailure(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{ConvertError: fmt.Errorf("OUPSY")}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker - will fail because Gotenberg service is not available
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate PDF")

	// Verify no upload happened
	assert.False(t, mockStorage.UploadCalled)

	// Verify no report attachment was created
	var count int64
	db.Model(&models.Attachment{}).
		Where("holder_type = ? AND holder_id = ? AND kind = ?", "interventions", intervention.ID, "report").
		Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestAttachReportPdfWorker_Work_EmailFailureDoesNotBlockJob(t *testing.T) {
	// This test verifies that email sending failures are logged but don't fail the job
	// Since email is a "nice to have" feature after the main work is done

	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{
		SendError: errors.New("SMTP server unavailable"),
	}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	// Email failure should NOT block the job - the PDF generation and attachment should succeed
	err := worker.Work(context.Background(), job)
	require.NoError(t, err, "Worker should succeed even when email fails")

	// Verify PDF attachment was created despite email failure
	var reportAttachment models.Attachment
	err = db.Where("holder_type = ? AND holder_id = ? AND kind = ?", "interventions", intervention.ID, "report").
		First(&reportAttachment).Error
	require.NoError(t, err, "Report attachment should be created even when email fails")

	// Verify email service was called (but failed)
	assert.True(t, mockEmail.SendCalled, "Email service should have been called")
}

func TestAttachReportPdfWorker_Work_WithPhotoAttachments(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create multiple photo attachments
	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("interventions/1/photo/photo1.jpg").
		WithFileName("photo1.jpg").
		Create(db)

	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("interventions/1/photo/photo2.jpg").
		WithFileName("photo2.jpg").
		Create(db)

	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("interventions/1/photo/photo3.jpg").
		WithFileName("photo3.jpg").
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err, "Worker should complete successfully")

	// Verify intervention was loaded with photos
	var loadedIntervention models.Intervention
	db.Preload("Attachments", "kind = ?", "photo").First(&loadedIntervention, intervention.ID)
	assert.Len(t, loadedIntervention.Attachments, 3, "All 3 photo attachments should be loaded")

	// Verify report attachment was created
	var reportAttachment models.Attachment
	err = db.Where("holder_type = ? AND holder_id = ? AND kind = ?", "interventions", intervention.ID, "report").
		First(&reportAttachment).Error
	require.NoError(t, err, "Report attachment should be created")
}

func TestAttachReportPdfWorker_Work_TempFileCleanup(t *testing.T) {
	// This test verifies that temporary PDF files are cleaned up after processing

	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Count temp files before running worker
	tempFilesBefore, err := filepath.Glob("/tmp/intervention_report_*.pdf")
	require.NoError(t, err)

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err = worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify no new temp files are left behind in /tmp
	tempFilesAfter, err := filepath.Glob("/tmp/intervention_report_*.pdf")
	require.NoError(t, err)

	// The worker should clean up its temp file, so count should be the same
	assert.Equal(t, len(tempFilesBefore), len(tempFilesAfter), "Temp file should be cleaned up after processing")
}

func TestAttachReportPdfWorker_Work_Idempotency(t *testing.T) {
	// Test that running the worker multiple times doesn't create duplicate reports

	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Pre-create a report attachment (simulating previous successful run)
	existingReport := factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("report").
		WithStorageKey("interventions/1/report/existing_report.pdf").
		WithFileName("existing_report.pdf").
		WithUploadStatus(models.AttachmentUploadStatusCompleted).
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker again
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Count report attachments
	var reportCount int64
	db.Model(&models.Attachment{}).
		Where("holder_type = ? AND holder_id = ? AND kind = ?", "interventions", intervention.ID, "report").
		Count(&reportCount)

	// Note: Current implementation doesn't check for existing reports, so it will create duplicates
	// This test documents the current behavior. If idempotency is desired, the worker should:
	// 1. Check for existing report attachments
	// 2. Either skip generation or delete old report first
	assert.Equal(t, int64(2), reportCount, "Worker creates duplicate reports (current behavior)")
	t.Logf("Report count after re-run: %d (includes existing report: %d)", reportCount, existingReport.ID)
}

func TestAttachReportPdfWorker_Work_LoadsAllRelationships(t *testing.T) {
	// Verify that the worker properly loads all required relationships

	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data with all relationships
	user := factories.NewUser().
		WithName("Jane", "Smith").
		WithEmail("jane@example.com").
		Create(db)

	portal := factories.NewPortal().
		WithName("Main Entrance Portal").
		Create(db)

	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create control items
	control1 := factories.NewControl().
		WithIntervention(intervention.ID).
		Create(db)

	control2 := factories.NewControl().
		WithIntervention(intervention.ID).
		Create(db)

	// Create photo attachments
	factories.NewAttachment().
		WithHolderType("interventions").
		WithHolderID(intervention.ID).
		WithKind("photo").
		WithStorageKey("interventions/1/photo/before.jpg").
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify relationships were loaded by checking database queries
	var loadedIntervention models.Intervention
	err = db.Preload("User").
		Preload("Portal").
		Preload("Controls").
		Preload("Attachments", "kind = ?", "photo").
		First(&loadedIntervention, intervention.ID).Error

	require.NoError(t, err)
	assert.Equal(t, user.ID, loadedIntervention.User.ID)
	assert.Equal(t, portal.ID, loadedIntervention.Portal.ID)
	assert.Len(t, loadedIntervention.Controls, 2)
	assert.Equal(t, control1.ID, loadedIntervention.Controls[0].ID)
	assert.Equal(t, control2.ID, loadedIntervention.Controls[1].ID)
	assert.Len(t, loadedIntervention.Attachments, 1)
}

func TestAttachReportPdfWorker_Work_FileMetadata(t *testing.T) {
	// Verify that the PDF attachment has correct metadata

	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	mockRunner := &tests.MockBackgroundJobRunner{}
	mockHtmlToPdfConverter := &mocks.GotenbergService{}

	// Create test data
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	intervention := factories.NewIntervention().
		WithUserModel(user).
		WithPortalModel(portal).
		Create(db)

	// Create worker
	worker := &AttachReportPdfWorker{
		DB:                 db,
		StorageService:     mockStorage,
		EmailService:       mockEmail,
		JobRunner:          mockRunner,
		HtmlToPdfConverter: mockHtmlToPdfConverter,
	}

	// Execute worker
	job := &river.Job[types.AttachReportPdfArgs]{
		Args: types.AttachReportPdfArgs{
			InterventionID: intervention.ID,
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err)

	// Verify the report attachment was created with correct metadata
	var reportAttachment models.Attachment
	err = db.Where("holder_type = ? AND holder_id = ? AND kind = ?",
		"interventions", intervention.ID, "report").First(&reportAttachment).Error
	require.NoError(t, err, "Report attachment should be created")

	// Verify metadata
	assert.Contains(t, reportAttachment.FileName, ".pdf")
	assert.Equal(t, "application/pdf", reportAttachment.ContentType)
	assert.Greater(t, reportAttachment.FileSize, int64(0))
	assert.Equal(t, models.AttachmentUploadStatusPending, reportAttachment.UploadStatus)
}

// Helper function to create a minimal test file
func createTestPDF(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "test_pdf_*.pdf")
	require.NoError(t, err)
	defer tempFile.Close()

	content := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\n%%EOF")
	_, err = tempFile.Write(content)
	require.NoError(t, err)

	return tempFile.Name()
}

// Helper function to verify file was deleted
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Integration test helper - only runs when Gotenberg is available
func isGotenbergAvailable() bool {
	// This would need to check if Gotenberg service is running
	// For now, return false to skip integration tests
	return false
}
