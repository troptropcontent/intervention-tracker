package mocks

import (
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

// PDFService is a mock implementation of interventions.PDFService for testing
type PDFService struct {
	GenerateCalled bool
	GenerateError  error
	PDFContent     []byte
}

func (m *PDFService) GenerateReportPDF(intervention *models.Intervention) (*os.File, error) {
	m.GenerateCalled = true

	if m.GenerateError != nil {
		return nil, m.GenerateError
	}

	// Create a temporary PDF file
	tempFile, err := os.CreateTemp("", "test_report_*.pdf")
	if err != nil {
		return nil, err
	}

	content := m.PDFContent
	if content == nil {
		content = []byte("%PDF-1.4 test pdf content")
	}

	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	// Reset file pointer to beginning
	if _, err := tempFile.Seek(0, 0); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	return tempFile, nil
}
