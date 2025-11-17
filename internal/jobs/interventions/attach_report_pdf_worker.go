package interventions

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/riverqueue/river"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs/types"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/interventions"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type AttachReportPdfWorker struct {
	river.WorkerDefaults[types.AttachReportPdfArgs]
	DB             *gorm.DB
	StorageService storage.StorageService
	EmailService   email.EmailService
	JobRunner      jobs.BackgroundJobRunner
}

// Work generates a PDF report for an intervention, uploads it to S3,
// creates an Attachment record, and sends a notification email to the user.
func (w *AttachReportPdfWorker) Work(ctx context.Context, job *river.Job[types.AttachReportPdfArgs]) error {
	interventionID := job.Args.InterventionID

	// Load intervention with all relationships
	var intervention models.Intervention
	if err := w.DB.Preload("User").Preload("Attachments", "kind = ?", "photo").Preload("Portal").Preload("Controls").First(&intervention, interventionID).Error; err != nil {
		return fmt.Errorf("failed to load intervention %d: %w", interventionID, err)
	}

	// Get signed URLs for all photo attachments
	for i, photo := range intervention.Attachments {
		url, err := w.StorageService.GetFileURL(
			ctx,
			photo.StorageKey,
			15*time.Minute, // URL valid for 15 minutes
		)
		if err != nil {
			return fmt.Errorf("failed to retrieve signed URL for %s: %w", photo.FileName, err)
		}
		intervention.Attachments[i].SignedUrl = url
	}

	// Generate PDF report
	pdfService := interventions.NewPDFService()
	pdfFile, err := pdfService.GenerateReportPDF(&intervention)
	if err != nil {
		return fmt.Errorf("failed to generate PDF for intervention %d: %w", interventionID, err)
	}
	defer pdfFile.Close()
	defer os.Remove(pdfFile.Name()) // Clean up temporary file

	// Get file info for metadata
	fileInfo, err := pdfFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for intervention %d: %w", interventionID, err)
	}

	attachmentService := attachments.NewAttachmentService(w.DB, w.StorageService, w.JobRunner)

	// Attach the PDF report (this will upload synchronously)
	if _, err := attachmentService.Attach(ctx, pdfFile, fileInfo.Name(), intervention.ID, "interventions", "report"); err != nil {
		return fmt.Errorf("failed to attach PDF for intervention %d: %w", interventionID, err)
	}

	// Send notification email
	notificationService := interventions.NewNotificationService(pdfService, w.EmailService)
	if err := notificationService.SendInterventionReport(&intervention); err != nil {
		// Log but don't fail - the important part (PDF generation and attachment) is already done
		log.Printf("WARNING: Failed to send notification email for intervention %d: %v", interventionID, err)
	} else {
		log.Printf("Successfully sent notification email for intervention %d", interventionID)
	}

	return nil
}
