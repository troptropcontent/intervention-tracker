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
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

type EchoContextBuilder struct {
	Context echo.Context
}

func NewContext(args ...string) *EchoContextBuilder {
	method := http.MethodGet
	path := "/"

	if len(args) > 0 {
		method = args[0]
	}
	if len(args) > 1 {
		path = args[1]
	}

	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(""))
	rec := httptest.NewRecorder()
	return &EchoContextBuilder{
		Context: e.NewContext(req, rec),
	}
}

func (b *EchoContextBuilder) WithFormData(formData map[string]string) *EchoContextBuilder {
	req := b.Context.Request()
	f := make(url.Values)
	for key, value := range formData {
		f.Set(key, value)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Body = io.NopCloser(strings.NewReader(f.Encode()))
	return b
}

func (b *EchoContextBuilder) WithMultiPartData(formData map[string]string, files map[string][]byte) *EchoContextBuilder {
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

	req := b.Context.Request()
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	req.Body = io.NopCloser(body)
	return b
}

func (b *EchoContextBuilder) WithQueryParams(params map[string]string) *EchoContextBuilder {
	q := make(url.Values)
	for key, value := range params {
		q.Set(key, value)
	}
	req := b.Context.Request()
	req.URL.RawQuery = q.Encode()
	return b
}

func (b *EchoContextBuilder) WithAuthenticatedUser(user *models.User) *EchoContextBuilder {
	b.Context.Set("user_id", user.ID)
	b.Context.Set("user_email", user.Email)

	return b
}

func (b *EchoContextBuilder) Build() echo.Context {
	return b.Context
}
