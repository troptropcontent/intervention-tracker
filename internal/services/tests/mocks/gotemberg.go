package mocks

import (
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/services/pdf"
)

// GotenbergService is a mock implementation of services.GotenbergService for testing
type GotenbergService struct {
	ConvertCalled  bool
	ConvertError   error
	PDFContent     []byte
	CapturedFiles  []pdf.HtmlToPdfConverterFiles
	CapturedPrefix string
}

func (m *GotenbergService) ConvertHTMLToPDF(files []pdf.HtmlToPdfConverterFiles, filenamePrefix string) (*os.File, error) {
	m.ConvertCalled = true
	m.CapturedFiles = files
	m.CapturedPrefix = filenamePrefix

	if m.ConvertError != nil {
		return nil, m.ConvertError
	}

	// Create a temporary PDF file
	tempFile, err := os.CreateTemp("", filenamePrefix+"_*.pdf")
	if err != nil {
		return nil, err
	}

	content := m.PDFContent
	if content == nil {
		content = []byte("%PDF-1.4 mock gotenberg pdf content")
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
