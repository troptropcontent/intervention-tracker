package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

func TestGetSettings_RendersCurrentNotificationEmail(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("login@example.com").WithNotificationEmail("reports@example.com").Create(db)

	c := tests.NewContext(http.MethodGet, "/").WithAuthenticatedUser(user).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "reports@example.com")
}

func TestGetSettings_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	deps, _ := setup(t)

	c := tests.NewContext(http.MethodGet, "/").Build()

	handler := GetSettings(deps)
	err := handler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestUpdateSettings_ValidEmail_UpdatesAndRedirects(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("login@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{"notification_email": "new-notifications@example.com"}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/settings?saved=1", rec.Header().Get("Location"))

	var updated models.User
	require.NoError(t, db.First(&updated, user.ID).Error)
	assert.Equal(t, "new-notifications@example.com", updated.NotificationEmail)
	assert.Equal(t, "login@example.com", updated.Email, "login email must stay untouched")
}

func TestUpdateSettings_InvalidEmail_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithNotificationEmail("original@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{"notification_email": "not-an-email"}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Adresse email invalide")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.Equal(t, "original@example.com", unchanged.NotificationEmail)
}
