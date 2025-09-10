package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// GotenbergService handles PDF generation via Gotenberg
type GotenbergService struct {
	BaseURL string
	Client  *http.Client
}

// NewGotenbergService creates a new Gotenberg service instance
func NewGotenbergService(baseURL string) *GotenbergService {
	return &GotenbergService{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ConvertHTMLToPDF converts HTML string to PDF and returns a temporary file
func (s *GotenbergService) ConvertHTMLToPDF(htmlContent, filenamePrefix string) (*os.File, error) {
	// Create temporary file for PDF
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s_*.pdf", filenamePrefix))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Convert HTML to PDF
	if err := s.convertHTML(htmlContent, tempFile); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	// Seek to beginning for reading
	if _, err := tempFile.Seek(0, 0); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to seek temp file: %w", err)
	}

	return tempFile, nil
}

// convertHTML sends HTML to Gotenberg and writes PDF response to writer
func (s *GotenbergService) convertHTML(htmlContent string, writer io.Writer) error {
	// Create multipart form
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// Add HTML file
	htmlPart, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := htmlPart.Write([]byte(htmlContent)); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}

	// Set PDF options (A4, margins, print background)
	options := map[string]string{
		"paperWidth":      "8.27",
		"paperHeight":     "11.7",
		"marginTop":       "0.4",
		"marginBottom":    "0.4",
		"marginLeft":      "0.4",
		"marginRight":     "0.4",
		"printBackground": "true",
	}

	for key, value := range options {
		if err := w.WriteField(key, value); err != nil {
			return fmt.Errorf("failed to set option %s: %w", key, err)
		}
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Send request to Gotenberg
	url := fmt.Sprintf("%s/forms/chromium/convert/html", s.BaseURL)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Gotenberg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gotenberg returned status %d", resp.StatusCode)
	}

	// Copy PDF to writer
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("failed to copy PDF response: %w", err)
	}

	return nil
}