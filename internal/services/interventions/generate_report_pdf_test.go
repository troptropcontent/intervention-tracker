package interventions

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

func TestNewPDFService(t *testing.T) {
	gotenbergURL := "http://localhost:3000"
	service := NewPDFService(gotenbergURL)

	assert.NotNil(t, service)
	assert.NotNil(t, service.gotenbergService)
	assert.Equal(t, gotenbergURL, service.gotenbergService.BaseURL)
}

func TestPDFService_GenerateReportPDF_Success(t *testing.T) {
	// Create a mock server that returns PDF content
	mockPDF := []byte("%PDF-1.4 fake pdf content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/forms/chromium/convert/html", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse multipart form to verify contents
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		// Check HTML file content contains intervention data
		file, header, err := r.FormFile("files")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "index.html", header.Filename)
		content, err := io.ReadAll(file)
		require.NoError(t, err)

		htmlContent := string(content)
		assert.Contains(t, htmlContent, "Test Portal")
		assert.Contains(t, htmlContent, "John Doe")
		assert.Contains(t, htmlContent, "Test summary")

		// Return mock PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDF)
	}))
	defer server.Close()

	service := NewPDFService(server.URL)
	intervention := createTestIntervention()

	tempFile, err := service.GenerateReportPDF(intervention)
	require.NoError(t, err)
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify file was created and contains PDF content
	assert.Contains(t, tempFile.Name(), "intervention_report")
	assert.Contains(t, tempFile.Name(), ".pdf")

	// Read file content
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)
}

func TestPDFService_GenerateReportPDF_EmptyData(t *testing.T) {
	// Use invalid URL to simulate network error and test early failure
	service := NewPDFService("http://invalid-url:9999")
	intervention := &models.Intervention{}

	tempFile, err := service.GenerateReportPDF(intervention)

	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to generate PDF")
}

func TestPDFService_GenerateReportPDF_GotenbergError(t *testing.T) {
	// Create a mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewPDFService(server.URL)
	intervention := createTestIntervention()

	tempFile, err := service.GenerateReportPDF(intervention)

	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to generate PDF")
}

func TestPDFService_GenerateReportPDF_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	service := NewPDFService("http://invalid-gotenberg-url:9999")
	intervention := createTestIntervention()

	tempFile, err := service.GenerateReportPDF(intervention)

	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to generate PDF")
}

func TestPDFService_Integration_EndToEnd(t *testing.T) {
	// Create a mock server that simulates Gotenberg
	mockPDF := []byte("%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\nxref\n0 4\ntrailer<</Size 4/Root 1 0 R>>\n%%EOF")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request structure
		assert.Equal(t, "/forms/chromium/convert/html", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Parse and verify form data
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		// Verify PDF options are set
		assert.Equal(t, "8.27", r.FormValue("paperWidth"))
		assert.Equal(t, "11.7", r.FormValue("paperHeight"))
		assert.Equal(t, "true", r.FormValue("printBackground"))

		// Verify HTML file content
		file, header, err := r.FormFile("files")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "index.html", header.Filename)
		content, err := io.ReadAll(file)
		require.NoError(t, err)

		htmlContent := string(content)
		assert.Contains(t, htmlContent, "Test Portal")
		assert.Contains(t, htmlContent, "John Doe")
		assert.Contains(t, htmlContent, "Test summary")

		// Return mock PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDF)
	}))
	defer server.Close()

	service := NewPDFService(server.URL)
	intervention := createTestIntervention()

	tempFile, err := service.GenerateReportPDF(intervention)
	require.NoError(t, err, "Failed to generate PDF report")

	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify temp file properties
	assert.Contains(t, tempFile.Name(), "intervention_report")
	assert.Contains(t, tempFile.Name(), ".pdf")

	// Verify file content
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)

	// Verify file size
	fileInfo, err := tempFile.Stat()
	require.NoError(t, err)
	assert.Equal(t, int64(len(mockPDF)), fileInfo.Size())
}

// Helper function to create test intervention
func createTestIntervention() *models.Intervention {
	testDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	summary := "Test summary"

	return &models.Intervention{
		ID:       1,
		Date:     testDate,
		Summary:  &summary,
		UserID:   1,
		UserName: "John Doe",
		PortalID: 1,
		Portal: models.Portal{
			ID:   1,
			Name: "Test Portal",
		},
		User: models.User{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
		},
		Controls: []models.Control{},
	}
}
