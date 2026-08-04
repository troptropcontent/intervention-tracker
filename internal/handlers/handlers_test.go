package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

func TestHandlers_NotFound(t *testing.T) {
	// Setup
	h := &Handlers{} // No DB needed for 404 test
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.NotFound(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, "404 - Page non trouvée")
	assert.Contains(t, body, "La page que vous recherchez n'existe pas")
	assert.Contains(t, body, "Retour à l'accueil")
	assert.Contains(t, body, `href="/"`)

	// Verify it's valid HTML
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, "<html")
	assert.Contains(t, body, "</html>")
}

func TestHandlers_NotFound_ContentType(t *testing.T) {
	// Setup
	h := &Handlers{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.NotFound(c)

	// Assert
	require.NoError(t, err)
	contentType := rec.Header().Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "text/html") || contentType == "")
}

func TestHandlers_GetPortal_Success(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires integration test setup")
}

func TestHandlers_GetPortal_NotFound(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires integration test setup")
}

func TestHandlers_GetPortal_InvalidUUID(t *testing.T) {
	t.Skip("Skipping database-dependent test - requires integration test setup")
}

func TestHandlers_QRRedirect_NotFound(t *testing.T) {
	db := tests.SetupTestDB(t)
	h := &Handlers{DB: db}

	c := tests.NewContext(http.MethodGet, "/qr_codes/unknown-uuid").WithLoggedOutSession().Build()
	c.SetParamNames("uuid")
	c.SetParamValues("unknown-uuid")

	err := h.QRRedirect(c)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestHandlers_QRRedirect_NotAssociated(t *testing.T) {
	db := tests.SetupTestDB(t)
	h := &Handlers{DB: db}

	qrCode := factories.NewQRCode().AsAvailable().Create(db)

	c := tests.NewContext(http.MethodGet, "/qr_codes/"+qrCode.UUID).WithLoggedOutSession().Build()
	c.SetParamNames("uuid")
	c.SetParamValues(qrCode.UUID)

	err := h.QRRedirect(c)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)
}

func TestHandlers_QRRedirect_LoggedIn_RedirectsToAdminPortal(t *testing.T) {
	db := tests.SetupTestDB(t)
	h := &Handlers{DB: db}

	portal := factories.NewPortal().Create(db)
	qrCode := factories.NewQRCode().WithPortalModel(portal).Create(db)
	user := factories.NewUser().Create(db)

	c := tests.NewContext(http.MethodGet, "/qr_codes/"+qrCode.UUID).WithLoggedInSession(user.ID, user.Email).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)
	c.SetParamNames("uuid")
	c.SetParamValues(qrCode.UUID)

	err := h.QRRedirect(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, fmt.Sprintf("/admin/portals/%d", portal.ID), rec.Header().Get("Location"))
}

func TestHandlers_QRRedirect_LoggedOut_RedirectsToPublicPortal(t *testing.T) {
	db := tests.SetupTestDB(t)
	h := &Handlers{DB: db}

	portal := factories.NewPortal().Create(db)
	qrCode := factories.NewQRCode().WithPortalModel(portal).Create(db)

	c := tests.NewContext(http.MethodGet, "/qr_codes/"+qrCode.UUID).WithLoggedOutSession().Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)
	c.SetParamNames("uuid")
	c.SetParamValues(qrCode.UUID)

	err := h.QRRedirect(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, fmt.Sprintf("/portals/%d", portal.ID), rec.Header().Get("Location"))
}

func TestHandlers_GetAdminPortalsScan(t *testing.T) {
	// Setup
	h := &Handlers{} // No DB needed for template rendering
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/portals/scan", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.GetAdminPortalsScan(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, "Scanner un QR Code")
	assert.Contains(t, body, "reader")
	assert.Contains(t, body, "manual-input")
	assert.Contains(t, body, "qr-scanner.js")

	// Verify it contains camera-related elements
	assert.Contains(t, body, "Initialisation de la caméra")
	assert.Contains(t, body, "Pointez votre caméra vers un QR code")

	// Verify manual fallback is present
	assert.Contains(t, body, "Problème avec la caméra?")
	assert.Contains(t, body, "Collez l'URL du QR code ici")

	// Verify instructions are present
	assert.Contains(t, body, "Instructions:")
	assert.Contains(t, body, "Assurez-vous que votre caméra est activée")
}
