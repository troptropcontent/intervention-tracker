package interventions

import (
	"context"
	"fmt"
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/translation"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

// PDFService handles PDF generation for interventions
type PDFService struct {
	gotenbergService services.HtmlToPdfConverter
}

// NewPDFService creates a new intervention PDF service
func NewPDFService(gotembergService ...services.HtmlToPdfConverter) *PDFService {
	var service services.HtmlToPdfConverter
	if len(gotembergService) > 0 {
		service = gotembergService[0]
	} else {
		service = services.NewGotenbergService()
	}

	return &PDFService{
		gotenbergService: service,
	}
}

// GenerateReport generates a PDF report for an intervention
func (s *PDFService) GenerateReportPDF(intervention *models.Intervention) (*os.File, error) {
	// Render intervention template
	html_string, err := s.renderInterventionHTML(intervention)
	if err != nil {
		return nil, fmt.Errorf("failed to render intervention HTML: %w", err)
	}

	cssPath, err := utils.GetStaticFilePath("css/output.css")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve CSS file path: %w", err)
	}

	css_bytes, err := os.ReadFile(cssPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSS file: %w", err)
	}

	logoPath, err := utils.GetStaticFilePath("images/logo.png")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve logo file path: %w", err)
	}

	logo_bytes, err := os.ReadFile(logoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read logo file: %w", err)
	}

	var files []services.ConvertHtmlToPdfFiles
	files = append(files, services.ConvertHtmlToPdfFiles{
		Name:         "index.html",
		ContentBytes: []byte(html_string),
	}, services.ConvertHtmlToPdfFiles{
		Name:         "output.css",
		ContentBytes: css_bytes,
	}, services.ConvertHtmlToPdfFiles{
		Name:         "logo.png",
		ContentBytes: logo_bytes,
	})

	// Generate PDF using Gotenberg service
	tempFile, err := s.gotenbergService.ConvertHTMLToPDF(files, "intervention_report")
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return tempFile, nil
}

// renderInterventionHTML renders the intervention template to HTML string
func (s *PDFService) renderInterventionHTML(intervention *models.Intervention) (string, error) {
	translator := translation.NewTranslator()
	var buf []byte
	htmlBuffer := &htmlWriter{buf: buf}

	templateConfig := templates.InterventionReportConfig{
		Intervention:   intervention,
		StylesheetPath: "output.css",
		LogoPath:       "logo.png",
		Translator:     translator,
	}
	if err := templates.InterventionReport(templateConfig).Render(context.Background(), htmlBuffer); err != nil {
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
