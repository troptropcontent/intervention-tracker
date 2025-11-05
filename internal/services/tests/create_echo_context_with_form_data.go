package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func CreateEchoContextWithFormData(formData map[string]string) echo.Context {
	e := echo.New()
	f := make(url.Values)
	for key, value := range formData {
		f.Set(key, value)
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func CreateEchoContextWithMultipartData(formData map[string]string, files map[string][]byte) echo.Context {
	e := echo.New()

	// Create a buffer to write our multipart form data to
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range formData {
		_ = writer.WriteField(key, value)
	}

	// Add files
	for fieldName, fileContent := range files {
		part, _ := writer.CreateFormFile(fieldName, fieldName)
		part.Write(fileContent)
	}

	// Close the writer to finalize the multipart message
	writer.Close()

	// Create the request
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

// CreateEchoContextWithInvalidMultipartForm creates a context with a malformed multipart form
// that will cause parsing errors
func CreateEchoContextWithInvalidMultipartForm() echo.Context {
	e := echo.New()
	// Create a malformed multipart body (missing boundary closing)
	body := strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"key\"\r\n\r\nvalue\r\n")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, "multipart/form-data; boundary=boundary")
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

// CreateEchoContextWithCorruptedForm creates a context with corrupted form data
// that will cause parsing errors
func CreateEchoContextWithCorruptedForm() echo.Context {
	e := echo.New()
	// Create a body that will fail parsing - invalid URL encoding
	body := io.NopCloser(bytes.NewReader([]byte("key=%ZZ"))) // %ZZ is invalid hex encoding
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}
