package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_NotFound(t *testing.T) {
	// Setup
	h := &Handlers{DB: &sqlx.DB{}} // Mock DB, not needed for 404 test
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
	h := &Handlers{DB: &sqlx.DB{}}
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