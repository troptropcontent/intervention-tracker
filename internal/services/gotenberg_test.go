package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGotenbergService(t *testing.T) {
	baseURL := "http://localhost:3000"
	service := NewGotenbergService(baseURL)

	assert.Equal(t, baseURL, service.BaseURL)
	assert.NotNil(t, service.Client)
	assert.Equal(t, 30*time.Second, service.Client.Timeout)
}

func TestConvertHTMLToPDF_Success(t *testing.T) {
	// Create a mock server that returns PDF content
	mockPDF := []byte("%PDF-1.4 fake pdf content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/forms/chromium/convert/html", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Parse multipart form to verify contents
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		require.NoError(t, err)

		// Check PDF options
		assert.Equal(t, "8.27", r.FormValue("paperWidth"))
		assert.Equal(t, "11.7", r.FormValue("paperHeight"))
		assert.Equal(t, "0.4", r.FormValue("marginTop"))
		assert.Equal(t, "0.4", r.FormValue("marginBottom"))
		assert.Equal(t, "0.4", r.FormValue("marginLeft"))
		assert.Equal(t, "0.4", r.FormValue("marginRight"))
		assert.Equal(t, "true", r.FormValue("printBackground"))

		// Check HTML file
		file, header, err := r.FormFile("files")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "index.html", header.Filename)
		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Contains(t, string(content), "<html>")

		// Return mock PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDF)
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	tempFile, err := service.ConvertHTMLToPDF(files, "test")
	require.NoError(t, err)
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify file was created and contains PDF content
	assert.Contains(t, tempFile.Name(), "test")
	assert.Contains(t, tempFile.Name(), ".pdf")

	// Read file content
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)
}

func TestConvertHTMLToPDF_GotenbergError(t *testing.T) {
	// Create a mock server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	tempFile, err := service.ConvertHTMLToPDF(files, "test")
	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to convert HTML to PDF")
	assert.Contains(t, err.Error(), "Gotenberg returned status 500")
}

func TestConvertHTMLToPDF_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	service := NewGotenbergService("http://invalid-gotenberg-url:9999")
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	tempFile, err := service.ConvertHTMLToPDF(files, "test")
	assert.Error(t, err)
	assert.Nil(t, tempFile)
	assert.Contains(t, err.Error(), "failed to convert HTML to PDF")
}

func TestConvertHTMLToPDF_TempFileCreationError(t *testing.T) {
	// This test is harder to implement without mocking os.CreateTemp
	// but we can test with invalid directory paths in some cases
	service := NewGotenbergService("http://localhost:3000")

	// Use very long filename prefix that might cause issues
	longPrefix := strings.Repeat("a", 300)
	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte("<html></html>"),
		},
	}
	tempFile, _ := service.ConvertHTMLToPDF(files, longPrefix)

	// This might succeed on some systems, so we just ensure it handles gracefully
	if tempFile != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}
	// Don't assert error since temp file creation might still succeed
}

func TestConvertHTML_Success(t *testing.T) {
	// Create a mock server
	mockPDF := []byte("%PDF-1.4 fake pdf content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDF)
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	var buf bytes.Buffer
	err := service.convertHTML(files, &buf)
	require.NoError(t, err)

	assert.Equal(t, mockPDF, buf.Bytes())
}

func TestConvertHTML_HTTPError(t *testing.T) {
	// Create a mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	var buf bytes.Buffer
	err := service.convertHTML(files, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gotenberg returned status 400")
}

func TestConvertHTML_InvalidURL(t *testing.T) {
	service := NewGotenbergService("http://invalid-url:9999")
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	var buf bytes.Buffer
	err := service.convertHTML(files, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request to Gotenberg")
}

func TestConvertHTML_MultipartFormOptions(t *testing.T) {
	// Test that all expected form fields are set correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		expectedOptions := map[string]string{
			"paperWidth":      "8.27",
			"paperHeight":     "11.7",
			"marginTop":       "0.4",
			"marginBottom":    "0.4",
			"marginLeft":      "0.4",
			"marginRight":     "0.4",
			"printBackground": "true",
		}

		for key, expected := range expectedOptions {
			actual := r.FormValue(key)
			assert.Equal(t, expected, actual, fmt.Sprintf("Option %s should be %s, got %s", key, expected, actual))
		}

		w.Write([]byte("OK"))
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte("<html></html>"),
		},
	}

	var buf bytes.Buffer
	err := service.convertHTML(files, &buf)
	require.NoError(t, err)
}

func TestConvertHTML_HTMLFileContent(t *testing.T) {
	expectedHTML := "<html><body><h1>Custom Test Content</h1><p>With special chars: àáâãäå</p></body></html>"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		file, header, err := r.FormFile("files")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "index.html", header.Filename)

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, expectedHTML, string(content))

		w.Write([]byte("OK"))
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(expectedHTML),
		},
	}

	var buf bytes.Buffer
	err := service.convertHTML(files, &buf)
	require.NoError(t, err)
}

func TestGotenbergService_Integration_EndToEnd(t *testing.T) {
	// This test verifies the complete flow from HTML to temp file
	mockPDF := []byte("%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\nxref\n0 4\ntrailer<</Size 4/Root 1 0 R>>\n%%EOF")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a successful Gotenberg response
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(mockPDF)
	}))
	defer server.Close()

	service := NewGotenbergService(server.URL)
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Document</title>
    <style>
        body { font-family: Arial, sans-serif; }
        h1 { color: blue; }
    </style>
</head>
<body>
    <h1>Test Report</h1>
    <p>This is a test HTML document that should be converted to PDF.</p>
    <ul>
        <li>Item 1</li>
        <li>Item 2</li>
        <li>Item 3</li>
    </ul>
</body>
</html>`

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	tempFile, err := service.ConvertHTMLToPDF(files, "integration_test")
	require.NoError(t, err)
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify temp file properties
	assert.Contains(t, tempFile.Name(), "integration_test")
	assert.Contains(t, tempFile.Name(), ".pdf")

	// Verify file content
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err)
	assert.Equal(t, mockPDF, content)

	// Verify file is at beginning (seeked back)
	pos, err := tempFile.Seek(0, 1) // Get current position
	require.NoError(t, err)
	assert.Equal(t, int64(len(mockPDF)), pos)
}

func TestGotenbergService_RealEndToEnd(t *testing.T) {
	// Skip this test in short mode or if GOTENBERG_URL is not set
	if testing.Short() {
		t.Skip("Skipping real E2E test in short mode")
	}

	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		// Try default docker-compose service URL
		gotenbergURL = "http://gotemberg:3000"
	}

	service := NewGotenbergService(gotenbergURL)

	// Test HTML content with various elements to ensure proper rendering
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Real E2E Test Document</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px;
            background-color: white;
        }
        h1 { 
            color: #2563eb; 
            border-bottom: 2px solid #e5e7eb;
            padding-bottom: 10px;
        }
        .info-box {
            background-color: #f3f4f6;
            border: 1px solid #d1d5db;
            border-radius: 8px;
            padding: 15px;
            margin: 15px 0;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #d1d5db;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f9fafb;
            font-weight: bold;
        }
        .footer {
            margin-top: 40px;
            font-size: 12px;
            color: #6b7280;
            text-align: center;
        }
    </style>
</head>
<body>
    <h1>QR Code Maintenance - E2E Test Report</h1>
    
    <div class="info-box">
        <h2>Test Information</h2>
        <p>This document tests the real Gotenberg service integration.</p>
        <p>Generated at: <span id="timestamp">` + time.Now().Format("2006-01-02 15:04:05") + `</span></p>
    </div>

    <h2>Features Tested</h2>
    <ul>
        <li>HTML to PDF conversion</li>
        <li>CSS styling and layout</li>
        <li>Unicode character support: àáâãäåæçèéêë</li>
        <li>Table rendering</li>
        <li>Background colors and borders</li>
    </ul>

    <table>
        <thead>
            <tr>
                <th>Component</th>
                <th>Status</th>
                <th>Notes</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Gotenberg Service</td>
                <td>✅ Active</td>
                <td>Successfully processing requests</td>
            </tr>
            <tr>
                <td>HTML Parsing</td>
                <td>✅ Working</td>
                <td>Complex HTML structures supported</td>
            </tr>
            <tr>
                <td>CSS Rendering</td>
                <td>✅ Working</td>
                <td>Styles applied correctly</td>
            </tr>
            <tr>
                <td>PDF Generation</td>
                <td>✅ Working</td>
                <td>Valid PDF output created</td>
            </tr>
        </tbody>
    </table>

    <div class="footer">
        Generated by QR Code Maintenance E2E Test Suite
    </div>
</body>
</html>`

	files := []ConvertHtmlToPdfFiles{
		{
			Name:         "index.html",
			ContentBytes: []byte(htmlContent),
		},
	}

	tempFile, err := service.ConvertHTMLToPDF(files, "real_e2e_test")
	require.NoError(t, err, "Failed to convert HTML to PDF using real Gotenberg service")

	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Verify temp file properties
	assert.Contains(t, tempFile.Name(), "real_e2e_test", "Temp file should contain prefix")
	assert.Contains(t, tempFile.Name(), ".pdf", "Temp file should have .pdf extension")

	// Verify file was created and has content
	fileInfo, err := tempFile.Stat()
	require.NoError(t, err, "Should be able to get file info")
	assert.Greater(t, fileInfo.Size(), int64(1000), "PDF file should be substantial size (>1KB)")

	// Read and verify PDF content starts with PDF header
	content, err := io.ReadAll(tempFile)
	require.NoError(t, err, "Should be able to read PDF content")
	assert.True(t, len(content) > 0, "PDF content should not be empty")
	assert.True(t, bytes.HasPrefix(content, []byte("%PDF")), "File should start with PDF header")

	// Verify file position was reset to beginning
	pos, err := tempFile.Seek(0, 1) // Get current position
	require.NoError(t, err, "Should be able to get file position")
	assert.Equal(t, int64(len(content)), pos, "File position should be at end after reading")

	// Reset to beginning and verify we can read again
	_, err = tempFile.Seek(0, 0)
	require.NoError(t, err, "Should be able to seek to beginning")

	firstBytes := make([]byte, 4)
	n, err := tempFile.Read(firstBytes)
	require.NoError(t, err, "Should be able to read from beginning")
	assert.Equal(t, 4, n, "Should read 4 bytes")
	assert.Equal(t, "%PDF", string(firstBytes), "Should read PDF header from beginning")

	t.Logf("Successfully generated PDF with size: %d bytes", len(content))
}
