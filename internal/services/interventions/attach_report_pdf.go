package interventions

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

const GotenbergUrl = "http://gotemberg:3000"

// AttachReportPdf generates a PDF report for an intervention, uploads it to S3,
// creates an Attachment record, and sends a notification email to the user.
// This function is designed to be called asynchronously (in a goroutine).
func AttachReportPdf(db *gorm.DB, storageService storage.StorageService, emailService email.EmailService, interventionId uint) {
	ctx := context.Background()

	// Load intervention with all relationships
	var intervention models.Intervention
	if err := db.Preload("User").Preload("Attachments", "kind = ?", "photo").Preload("Portal").Preload("Controls").First(&intervention, interventionId).Error; err != nil {
		log.Printf("Failed to load intervention %d: %v", interventionId, err)
		return
	}

	for i, photo := range intervention.Attachments {
		url, err := storageService.GetFileURL(
			ctx,
			photo.StorageKey,
			15*time.Minute, // URL valid for 15 minutes
		)
		if err != nil {
			log.Printf("Could not retrieve signed url for %v: %v", photo.FileName, err)
			return
		}
		intervention.Attachments[i].SignedUrl = url
	}

	// Generate PDF report
	pdfService := NewPDFService(GotenbergUrl)
	pdfFile, err := pdfService.GenerateReportPDF(&intervention)
	if err != nil {
		log.Printf("Failed to generate PDF for intervention %d: %v", interventionId, err)
		return
	}
	defer pdfFile.Close()
	defer os.Remove(pdfFile.Name()) // Clean up temporary file

	// Get file info for metadata
	fileInfo, err := pdfFile.Stat()
	if err != nil {
		log.Printf("Failed to get file info for intervention %d: %v", interventionId, err)
		return
	}

	var attachmentService attachments.Attacher = attachments.NewAttachmentService(db, storageService)

	attachmentService.Attach(ctx, pdfFile, fileInfo.Name(), intervention.ID, "interventions", "report")

	// Send notification email in a separate goroutine
	go func() {
		notificationService := NewNotificationService(pdfService, emailService)
		if err := notificationService.SendInterventionReport(&intervention); err != nil {
			log.Printf("Failed to send notification email for intervention %d: %v", interventionId, err)
		} else {
			log.Printf("Successfully sent notification email for intervention %d", interventionId)
		}
	}()
}
