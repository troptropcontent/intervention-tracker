package interventions

import (
	"fmt"
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
)

// NotificationService handles sending intervention notifications
type NotificationService struct {
	pdfService   *PDFService
	emailService email.EmailService
}

// NewNotificationService creates a new notification service
func NewNotificationService(pdfService *PDFService, emailService email.EmailService) *NotificationService {
	return &NotificationService{
		pdfService:   pdfService,
		emailService: emailService,
	}
}

// SendInterventionReport generates a PDF report and sends it via email
func (s *NotificationService) SendInterventionReport(intervention *models.Intervention) error {
	// Generate PDF report
	pdfFile, err := s.pdfService.GenerateReportPDF(intervention)
	if err != nil {
		return fmt.Errorf("failed to generate PDF report: %w", err)
	}
	defer pdfFile.Close()
	defer os.Remove(pdfFile.Name()) // Clean up temporary file

	// Prepare email content
	subject := fmt.Sprintf("Rapport d'Intervention #%d - %s", intervention.ID, intervention.Portal.Name)
	body := s.buildEmailBody(intervention)

	if err := s.emailService.Send(intervention.User.Email, subject, body, pdfFile.Name()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildEmailBody creates the email body content
func (s *NotificationService) buildEmailBody(intervention *models.Intervention) string {
	body := fmt.Sprintf(`Cher/Chère %s %s,

Veuillez trouver en pièce jointe le rapport d'intervention pour :

Portail : %s
ID d'intervention : %d
Date : %s
Technicien : %s`,
		intervention.User.FirstName,
		intervention.User.LastName,
		intervention.Portal.Name,
		intervention.ID,
		intervention.Date.Format("2006-01-02"),
		intervention.UserName)

	if intervention.Summary != nil && *intervention.Summary != "" {
		body += fmt.Sprintf(`
Résumé : %s`, *intervention.Summary)
	}

	body += `

Cordialement,
Système de Maintenance QR Code`

	return body
}
