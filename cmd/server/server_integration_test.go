package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/handlers"
)

// Integration tests for the server routes
func TestServerRoutes_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create Echo instance with routes (similar to main.go)
	e := echo.New()
	h := &handlers.Handlers{} // Note: DB will be nil for template-only tests

	// Routes
	e.GET("/admin/portals/scan", h.GetAdminPortalsScan)
	e.RouteNotFound("/*", h.NotFound)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "Admin portals scan page",
			method:         http.MethodGet,
			path:           "/admin/portals/scan",
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"Scanner un QR Code",
				"reader",
				"manual-input",
				"qr-scanner.js",
				"Initialisation de la caméra",
			},
		},
		{
			name:           "404 page for non-existent route",
			method:         http.MethodGet,
			path:           "/non-existent-page",
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"404 - Page non trouvée",
				"La page que vous recherchez n'existe pas",
			},
		},
		{
			name:           "404 page for invalid portal UUID",
			method:         http.MethodGet,
			path:           "/portals/invalid-uuid",
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"404 - Page non trouvée",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			
			e.ServeHTTP(rec, req)
			
			assert.Equal(t, tt.expectedStatus, rec.Code)
			
			body := rec.Body.String()
			for _, expectedContent := range tt.expectedBody {
				assert.Contains(t, body, expectedContent)
			}
		})
	}
}

func TestServerRoutes_StaticFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	e := echo.New()
	e.Static("/static", "../../static") // Adjust path relative to test location

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "CSS file",
			path:           "/static/css/output.css",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HTMX file",
			path:           "/static/htmx.min.js",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "QR Scanner JS file",
			path:           "/static/js/qr-scanner.js",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent static file",
			path:           "/static/non-existent.js",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			
			e.ServeHTTP(rec, req)
			
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestServerRoutes_QRScannerPage_Elements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	e := echo.New()
	h := &handlers.Handlers{}
	e.GET("/admin/portals/scan", h.GetAdminPortalsScan)

	req := httptest.NewRequest(http.MethodGet, "/admin/portals/scan", nil)
	rec := httptest.NewRecorder()
	
	e.ServeHTTP(rec, req)
	
	require.Equal(t, http.StatusOK, rec.Code)
	
	body := rec.Body.String()
	
	// Verify HTML structure
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, "<html")
	assert.Contains(t, body, "</html>")
	
	// Verify essential QR scanner elements
	assert.Contains(t, body, `id="reader"`)
	assert.Contains(t, body, `id="loading"`)
	assert.Contains(t, body, `id="status"`)
	assert.Contains(t, body, `id="error"`)
	assert.Contains(t, body, `id="success"`)
	assert.Contains(t, body, `id="manual-input"`)
	assert.Contains(t, body, `id="manual-submit"`)
	
	// Verify JavaScript includes
	assert.Contains(t, body, "html5-qrcode.min.js")
	assert.Contains(t, body, "qr-scanner.js")
	assert.Contains(t, body, "initQRScanner()")
	assert.Contains(t, body, "setupManualInput()")
	
	// Verify CSS and styling
	assert.Contains(t, body, "bg-gray-50")
	assert.Contains(t, body, "bg-blue-600")
	assert.Contains(t, body, "rounded-lg")
	
	// Verify French content
	assert.Contains(t, body, "Scanner un QR Code")
	assert.Contains(t, body, "Pointez votre caméra vers un QR code")
	assert.Contains(t, body, "Problème avec la caméra?")
	assert.Contains(t, body, "Instructions:")
}

func TestServerRoutes_ContentTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	e := echo.New()
	h := &handlers.Handlers{}
	
	e.GET("/admin/portals/scan", h.GetAdminPortalsScan)
	e.RouteNotFound("/*", h.NotFound)

	tests := []struct {
		name         string
		path         string
		expectedType string
	}{
		{
			name:         "QR Scanner page content type",
			path:         "/admin/portals/scan",
			expectedType: "text/html",
		},
		{
			name:         "404 page content type",
			path:         "/non-existent",
			expectedType: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			
			e.ServeHTTP(rec, req)
			
			contentType := rec.Header().Get("Content-Type")
			assert.Contains(t, contentType, tt.expectedType)
		})
	}
}