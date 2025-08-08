package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/troptropcontent/qr_code_maintenance/internal/handlers"
)

func TestServer_404Routes(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"root nonexistent", "/nonexistent"},
		{"deep path", "/some/deep/path/that/does/not/exist"},
		{"api-like path", "/api/v1/users"},
		{"with query params", "/invalid?foo=bar&baz=qux"},
		{"with trailing slash", "/invalid/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Echo with minimal configuration (no DB needed for 404 test)
			e := echo.New()
			h := &handlers.Handlers{DB: nil} // DB not needed for 404 handler
			
			// Configure routes like main server
			e.RouteNotFound("/*", h.NotFound)
			
			// Make request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			// Assert 404 page is served with 200 status (successful render)
			assert.Equal(t, http.StatusOK, rec.Code)
			
			body := rec.Body.String()
			assert.Contains(t, body, "404 - Page non trouvée")
			assert.Contains(t, body, "La page que vous recherchez n'existe pas")
		})
	}
}

func TestServer_StaticFilesStillWork(t *testing.T) {
	// This test ensures static files are served correctly and don't trigger 404
	e := echo.New()
	h := &handlers.Handlers{DB: nil}
	
	// Configure like main server
	e.Static("/static", "../../static") // Adjust path for test context
	e.RouteNotFound("/*", h.NotFound)
	
	// Test static file request (should NOT trigger 404 handler)
	req := httptest.NewRequest(http.MethodGet, "/static/htmx.min.js", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Should either serve the file (200) or not found (404), but NOT our custom 404 page
	if rec.Code == http.StatusOK {
		// File exists and was served
		assert.NotContains(t, rec.Body.String(), "Page non trouvée")
	} else {
		// File doesn't exist, but should be proper 404, not our template
		assert.Equal(t, http.StatusNotFound, rec.Code)
		// Echo's default 404, not our template
		assert.NotContains(t, rec.Body.String(), "Page non trouvée")
	}
}

func TestServer_ValidRoutesWork(t *testing.T) {
	// Test that valid routes still work and don't trigger 404
	e := echo.New()
	h := &handlers.Handlers{DB: nil}
	
	// Add a simple test route
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test route works")
	})
	e.RouteNotFound("/*", h.NotFound)
	
	// Test valid route
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test route works", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), "Page non trouvée")
}