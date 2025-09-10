package interventions

import (
	"context"
	"fmt"
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
)

// PDFService handles PDF generation for interventions
type PDFService struct {
	gotenbergService *services.GotenbergService
}

// NewPDFService creates a new intervention PDF service
func NewPDFService(gotenbergURL string) *PDFService {
	return &PDFService{
		gotenbergService: services.NewGotenbergService(gotenbergURL),
	}
}

// GenerateReport generates a PDF report for an intervention
func (s *PDFService) GenerateReportPDF(intervention *models.Intervention) (*os.File, error) {
	// Render intervention template
	html_string, err := s.renderInterventionHTML(intervention)
	if err != nil {
		return nil, fmt.Errorf("failed to render intervention HTML: %w", err)
	}

	// Generate PDF using Gotenberg service
	tempFile, err := s.gotenbergService.ConvertHTMLToPDF(html_string, "intervention_report")
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return tempFile, nil
}

// renderInterventionHTML renders the intervention template to HTML string
func (s *PDFService) renderInterventionHTML(intervention *models.Intervention) (string, error) {
	var buf []byte
	htmlBuffer := &htmlWriter{buf: buf}

	if err := templates.InterventionReport(intervention).Render(context.Background(), htmlBuffer); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return string(htmlBuffer.buf), nil
}

// htmlWriter implements io.Writer to capture template output as string
type htmlWriter struct {
	buf []byte
}

func (w *htmlWriter) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}
