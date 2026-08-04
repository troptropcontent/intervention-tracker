package qrcodes

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/pdf"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

const qrImagePixelSize = 512

// PDFService generates a print-ready PDF sheet of QR codes
type PDFService struct {
	gotenbergService pdf.HtmlToPdfConverter
}

// NewPDFService creates a new QR code sheet PDF service
func NewPDFService(gotembergService ...pdf.HtmlToPdfConverter) *PDFService {
	var service pdf.HtmlToPdfConverter
	if len(gotembergService) > 0 {
		service = gotembergService[0]
	} else {
		service = pdf.NewGotenbergService()
	}

	return &PDFService{
		gotenbergService: service,
	}
}

// GenerateSheetPDF renders a printable grid of QR codes (image + UUID caption)
// pointing at baseURL + "/qr_codes/:uuid", and converts it to a PDF.
func (s *PDFService) GenerateSheetPDF(codes []models.QRCode, baseURL string) (*os.File, error) {
	htmlString, err := s.renderSheetHTML(codes, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to render qr code sheet HTML: %w", err)
	}

	cssPath, err := utils.GetStaticFilePath("css/output.css")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve CSS file path: %w", err)
	}

	cssBytes, err := os.ReadFile(cssPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSS file: %w", err)
	}

	files := []pdf.HtmlToPdfConverterFiles{
		{Name: "index.html", ContentBytes: []byte(htmlString)},
		{Name: "output.css", ContentBytes: cssBytes},
	}

	tempFile, err := s.gotenbergService.ConvertHTMLToPDF(files, "qr_code_sheet")
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return tempFile, nil
}

func (s *PDFService) renderSheetHTML(codes []models.QRCode, baseURL string) (string, error) {
	items := make([]templates.QRCodeSheetItem, len(codes))
	for i, code := range codes {
		dataURI, err := s.qrCodeDataURI(code.UUID, baseURL)
		if err != nil {
			return "", fmt.Errorf("failed to encode qr code %s: %w", code.UUID, err)
		}
		items[i] = templates.QRCodeSheetItem{UUID: code.UUID, ImageDataURI: dataURI}
	}

	var buf []byte
	htmlBuffer := &htmlWriter{buf: buf}

	if err := templates.QRCodeSheet(items, "output.css").Render(context.Background(), htmlBuffer); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return string(htmlBuffer.buf), nil
}

func (s *PDFService) qrCodeDataURI(codeUUID string, baseURL string) (string, error) {
	return QRCodeDataURI(codeUUID, baseURL)
}

// QRCodeDataURI encodes the URL pointing at baseURL + "/qr_codes/:uuid" as a
// base64 PNG data URI, suitable for embedding directly in HTML.
func QRCodeDataURI(codeUUID string, baseURL string) (string, error) {
	url := fmt.Sprintf("%s/qr_codes/%s", baseURL, codeUUID)

	png, err := qrcode.Encode(url, qrcode.Medium, qrImagePixelSize)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// htmlWriter implements io.Writer to capture template output as string
type htmlWriter struct {
	buf []byte
}

func (w *htmlWriter) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}
