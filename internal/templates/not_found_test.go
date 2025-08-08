package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotFound_Template(t *testing.T) {
	// Setup
	ctx := context.Background()
	var buf strings.Builder

	// Execute template
	component := NotFound()
	err := component.Render(ctx, &buf)

	// Assert
	require.NoError(t, err)
	
	html := buf.String()
	
	// Check structure
	assert.Contains(t, html, "<!doctype html>")
	assert.Contains(t, html, "<html lang=\"fr\">")
	assert.Contains(t, html, "</html>")
	
	// Check title in head
	assert.Contains(t, html, "<title>Page non trouvée - Maintenance Portails</title>")
	
	// Check CSS and JS includes
	assert.Contains(t, html, `href="/static/css/output.css"`)
	assert.Contains(t, html, `src="/static/htmx.min.js"`)
	
	// Check navigation
	assert.Contains(t, html, "Maintenance Portails")
	assert.Contains(t, html, `href="/portals"`)
	assert.Contains(t, html, `href="/interventions"`)
}

func TestNotFound_Content(t *testing.T) {
	// Setup
	ctx := context.Background()
	var buf strings.Builder

	// Execute
	err := NotFound().Render(ctx, &buf)
	require.NoError(t, err)
	
	html := buf.String()
	
	// Check 404 specific content
	assert.Contains(t, html, "404 - Page non trouvée")
	assert.Contains(t, html, "La page que vous recherchez n'existe pas ou a été déplacée")
	assert.Contains(t, html, "Retour à l'accueil")
	
	// Check home link
	assert.Contains(t, html, `href="/"`)
	
	// Check CSS classes for styling
	assert.Contains(t, html, "text-center")
	assert.Contains(t, html, "text-3xl")
	assert.Contains(t, html, "bg-blue-600")
	assert.Contains(t, html, "hover:bg-blue-700")
}

func TestNotFound_Accessibility(t *testing.T) {
	// Setup
	ctx := context.Background()
	var buf strings.Builder

	// Execute
	err := NotFound().Render(ctx, &buf)
	require.NoError(t, err)
	
	html := buf.String()
	
	// Check accessibility features
	assert.Contains(t, html, `lang="fr"`) // Language specified
	assert.Contains(t, html, `charset="UTF-8"`) // Character encoding
	assert.Contains(t, html, `name="viewport"`) // Responsive viewport
	
	// Check semantic HTML structure
	assert.Contains(t, html, "<h1")
	assert.Contains(t, html, "<nav")
	assert.Contains(t, html, "<main")
	
	// Check that the home link has meaningful text (not just "click here")
	homeButtonStart := strings.Index(html, `href="/"`) 
	if homeButtonStart > 0 {
		// Find the second href="/" which should be our button, not the nav link
		secondHomeLink := strings.Index(html[homeButtonStart+10:], `href="/"`)
		if secondHomeLink > 0 {
			buttonStart := homeButtonStart + 10 + secondHomeLink
			buttonEnd := strings.Index(html[buttonStart:], "</a>") + buttonStart + 4
			buttonText := html[buttonStart:buttonEnd]
			assert.Contains(t, buttonText, "Retour à l'accueil")
		}
	}
}