package interventions

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/mocks"
)

func TestNewPDFService(t *testing.T) {
	service := NewPDFService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.gotenbergService)
}

func TestPDFService_GenerateReportPDF_Success(t *testing.T) {
	// Setup mock with expected PDF content
	mockPDF := []byte("%PDF-1.4 test pdf content")
	mockGotemberg := &mocks.GotenbergService{
		PDFContent: mockPDF,
	}
	service := NewPDFService(mockGotemberg)

	db, _ := database.ConnectGORM()
	intervention := factories.NewIntervention().Create(db)

	tempFile, err := service.GenerateReportPDF(intervention)
	require.NoError(t, err)
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify file was created and contains PDF content
	assert.Contains(t, tempFile.Name(), "intervention_report")
	assert.Contains(t, tempFile.Name(), ".pdf")

	// Verify mock was called
	assert.True(t, mockGotemberg.ConvertCalled)
	assert.Equal(t, "intervention_report", mockGotemberg.CapturedPrefix)

	// Verify the files passed to the mock include HTML, CSS, and logo
	require.Len(t, mockGotemberg.CapturedFiles, 3)
	assert.Equal(t, "index.html", mockGotemberg.CapturedFiles[0].Name)
	assert.Equal(t, "output.css", mockGotemberg.CapturedFiles[1].Name)
	assert.Equal(t, "logo.png", mockGotemberg.CapturedFiles[2].Name)

	// Verify HTML content contains expected data
	htmlContent := string(mockGotemberg.CapturedFiles[0].ContentBytes)
	assert.NotEmpty(t, htmlContent)

	// Read file content
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)
}

func TestPDFService_GenerateReportPDF_GotenbergError(t *testing.T) {
	// Create a mock server that returns error
	mockGotemberg := &mocks.GotenbergService{
		ConvertError: fmt.Errorf("OUPSY"),
	}

	service := NewPDFService(mockGotemberg)
	db, _ := database.ConnectGORM()
	intervention := factories.NewIntervention().Create(db)

	tempFile, err := service.GenerateReportPDF(intervention)

	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to generate PDF")
}
