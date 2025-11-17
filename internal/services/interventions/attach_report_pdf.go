package interventions

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/jobs"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

// AttachReportPdf generates a PDF report for an intervention, uploads it to S3,
// creates an Attachment record, and sends a notification email to the user.
// This function is designed to be called asynchronously (in a goroutine).

type AttachReportPdfService struct {
	DB             *gorm.DB
	StorageService storage.StorageService
	EmailService   email.EmailService
	JobRunner      jobs.BackgroundJobRunner
}

func NewAttachReportPdfService(db *gorm.DB, storageService storage.StorageService, emailService email.EmailService, jobRunner jobs.BackgroundJobRunner) (*AttachReportPdfService, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	if storageService == nil {
		return nil, fmt.Errorf("storageService cannot be nil")
	}
	if emailService == nil {
		return nil, fmt.Errorf("emailService cannot be nil")
	}
	if jobRunner == nil {
		return nil, fmt.Errorf("jobRunner cannot be nil")
	}

	return &AttachReportPdfService{
		DB:             db,
		StorageService: storageService,
		EmailService:   emailService,
		JobRunner:      jobRunner,
	}, nil
}

func (s *AttachReportPdfService) AttachReportPdf(interventionId uint) error {
	ctx := context.Background()

	// Load intervention with all relationships
	var intervention models.Intervention
	if err := s.DB.Preload("User").Preload("Attachments", "kind = ?", "photo").Preload("Portal").Preload("Controls").First(&intervention, interventionId).Error; err != nil {
		return fmt.Errorf("failed to load intervention %d: %w", interventionId, err)
	}

	for i, photo := range intervention.Attachments {
		url, err := s.StorageService.GetFileURL(
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
	pdfService := NewPDFService()
	pdfFile, err := pdfService.GenerateReportPDF(&intervention)
	if err != nil {
		return fmt.Errorf("failed to generate PDF for intervention %d: %w", interventionId, err)
	}
	defer pdfFile.Close()
	defer os.Remove(pdfFile.Name()) // Clean up temporary file

	// Get file info for metadata
	fileInfo, err := pdfFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for intervention %d: %w", interventionId, err)
	}

	attachmentService := attachments.NewAttachmentService(s.DB, s.StorageService, s.JobRunner)

	// Attach the PDF report
	if _, err := attachmentService.Attach(ctx, pdfFile, fileInfo.Name(), intervention.ID, "interventions", "report"); err != nil {
		return fmt.Errorf("failed to attach PDF for intervention %d: %w", interventionId, err)
	}

	// Send notification email
	notificationService := NewNotificationService(pdfService, s.EmailService)
	if err := notificationService.SendInterventionReport(&intervention); err != nil {
		// Log but don't fail - the important part (PDF generation and attachment) is already done
		log.Printf("WARNING: Failed to send notification email for intervention %d: %v", interventionId, err)
	} else {
		log.Printf("Successfully sent notification email for intervention %d", interventionId)
	}

	return nil
}
