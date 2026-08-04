package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	echosession "github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

// sessionStoreContextKey mirrors the unexported key echo-contrib/session.Middleware
// stores its session.Store under, so tests can seed a session without wiring up
// the real middleware chain.
const sessionStoreContextKey = "_session_store"

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

// WithInvalidFormData sets up a request that will fail when ParseFormData is called.
// This creates a scenario with malformed URL-encoded data that cannot be parsed.
// Useful for testing error handling in handlers that call routers.ParseFormData.
func (b *EchoContextBuilder) WithInvalidFormData() *EchoContextBuilder {
	req := b.Context.Request()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	// Invalid percent-encoding that will cause ParseForm() to fail
	req.Body = io.NopCloser(strings.NewReader("key=%ZZ"))
	return b
}

func (b *EchoContextBuilder) WithAuthenticatedUser(user *models.User) *EchoContextBuilder {
	b.Context.Set("user_id", user.ID)
	b.Context.Set("user_email", user.Email)

	return b
}

// WithLoggedInSession simulates a real, cookie-backed logged-in session, for
// handlers (like QRRedirect) that read the session directly instead of
// relying on RequireAuth middleware to have populated the echo context.
func (b *EchoContextBuilder) WithLoggedInSession(userID uint, userEmail string) *EchoContextBuilder {
	b.withSessionStore()

	sess, _ := echosession.Get("session", b.Context)
	sess.Values["user_id"] = userID
	sess.Values["user_email"] = userEmail
	_ = sess.Save(b.Context.Request(), b.Context.Response())

	// Round-trip the Set-Cookie header back onto the request, the way a
	// browser would on the next request, so a fresh session.Get() call
	// inside the handler decodes the same session.
	resp := &http.Response{Header: b.Context.Response().Header()}
	for _, cookie := range resp.Cookies() {
		b.Context.Request().AddCookie(cookie)
	}

	return b
}

// WithLoggedOutSession simulates an anonymous request under the real session
// middleware: a store is present but no session cookie was sent.
func (b *EchoContextBuilder) WithLoggedOutSession() *EchoContextBuilder {
	b.withSessionStore()
	return b
}

func (b *EchoContextBuilder) withSessionStore() {
	store := sessions.NewCookieStore([]byte("test-session-secret"))
	b.Context.Set(sessionStoreContextKey, store)
}

func (b *EchoContextBuilder) Build() echo.Context {
	return b.Context
}
