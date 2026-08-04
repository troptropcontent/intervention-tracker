package qrcodes

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/mocks"
)

func TestNewPDFService(t *testing.T) {
	service := NewPDFService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.gotenbergService)
}

func TestPDFService_GenerateSheetPDF_Success(t *testing.T) {
	mockPDF := []byte("%PDF-1.4 test pdf content")
	mockGotemberg := &mocks.GotenbergService{
		PDFContent: mockPDF,
	}
	service := NewPDFService(mockGotemberg)

	codes := []models.QRCode{
		{UUID: "uuid-1"},
		{UUID: "uuid-2"},
	}

	tempFile, err := service.GenerateSheetPDF(codes, "http://example.com")
	require.NoError(t, err)
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	assert.Contains(t, tempFile.Name(), "qr_code_sheet")
	assert.Contains(t, tempFile.Name(), ".pdf")

	assert.True(t, mockGotemberg.ConvertCalled)
	assert.Equal(t, "qr_code_sheet", mockGotemberg.CapturedPrefix)

	require.Len(t, mockGotemberg.CapturedFiles, 2)
	assert.Equal(t, "index.html", mockGotemberg.CapturedFiles[0].Name)
	assert.Equal(t, "output.css", mockGotemberg.CapturedFiles[1].Name)

	htmlContent := string(mockGotemberg.CapturedFiles[0].ContentBytes)
	assert.Contains(t, htmlContent, "uuid-1")
	assert.Contains(t, htmlContent, "uuid-2")
	assert.Contains(t, htmlContent, "data:image/png;base64,")

	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)
}

func TestPDFService_GenerateSheetPDF_GotenbergError(t *testing.T) {
	mockGotemberg := &mocks.GotenbergService{
		ConvertError: fmt.Errorf("OUPSY"),
	}

	service := NewPDFService(mockGotemberg)
	codes := []models.QRCode{{UUID: "uuid-1"}}

	tempFile, err := service.GenerateSheetPDF(codes, "http://example.com")

	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to generate PDF")
}
