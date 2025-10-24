package interventions

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
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
	if err := db.Preload("User").Preload("Portal").Preload("Controls").First(&intervention, interventionId).Error; err != nil {
		log.Printf("Failed to load intervention %d: %v", interventionId, err)
		return
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

	// Generate unique storage key
	// Format: interventions/{intervention_id}/{uuid}.pdf
	storageKey := fmt.Sprintf("interventions/%d/%s.pdf", interventionId, uuid.New().String())

	// Seek to beginning of file before uploading
	if _, err := pdfFile.Seek(0, io.SeekStart); err != nil {
		log.Printf("Failed to seek file for intervention %d: %v", interventionId, err)
		return
	}

	// Upload to S3
	_, err = storageService.UploadFile(ctx, storageKey, pdfFile, "application/pdf")
	if err != nil {
		log.Printf("Failed to upload PDF to S3 for intervention %d: %v", interventionId, err)
		return
	}

	// Create Attachment record
	now := sql.NullTime{Time: time.Now(), Valid: true}
	attachment := models.Attachment{
		HolderType:  "interventions",
		HolderID:    interventionId,
		StorageKey:  storageKey,
		FileName:    fmt.Sprintf("intervention_%d_report.pdf", interventionId),
		ContentType: "application/pdf",
		FileSize:    fileInfo.Size(),
		UploadedAt:  now,
	}

	if err := db.Create(&attachment).Error; err != nil {
		log.Printf("Failed to create attachment record for intervention %d: %v", interventionId, err)
		// Attempt to clean up uploaded file
		if cleanupErr := storageService.DeleteFile(ctx, storageKey); cleanupErr != nil {
			log.Printf("Failed to cleanup S3 file after attachment creation failure: %v", cleanupErr)
		}
		return
	}

	log.Printf("Successfully attached PDF report for intervention %d (attachment ID: %d)", interventionId, attachment.ID)

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
