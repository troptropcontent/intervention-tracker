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

func TestGetSettings_RendersCurrentEmails(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("login@example.com").WithNotificationEmail("reports@example.com").Create(db)

	c := tests.NewContext(http.MethodGet, "/").WithAuthenticatedUser(user).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "login@example.com")
	assert.Contains(t, body, "reports@example.com")
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

func TestUpdateSettings_ValidNotificationEmail_UpdatesAndRedirects(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("login@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              user.Email,
			"notification_email": "new-notifications@example.com",
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
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

func TestUpdateSettings_InvalidNotificationEmail_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithNotificationEmail("original@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              user.Email,
			"notification_email": "not-an-email",
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Adresse email de notification invalide")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.Equal(t, "original@example.com", unchanged.NotificationEmail)
}

func TestUpdateSettings_ChangeLoginEmail_UpdatesWithoutTouchingNotificationEmail(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().
		WithEmail("old-login@example.com").
		WithNotificationEmail("reports@example.com").
		Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              "new-login@example.com",
			"notification_email": user.NotificationEmail,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	var updated models.User
	require.NoError(t, db.First(&updated, user.ID).Error)
	assert.Equal(t, "new-login@example.com", updated.Email)
	assert.Equal(t, "reports@example.com", updated.NotificationEmail, "notification email must stay untouched")
}

func TestUpdateSettings_InvalidLoginEmail_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("original@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              "not-an-email",
			"notification_email": user.NotificationEmail,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Adresse email invalide")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.Equal(t, "original@example.com", unchanged.Email)
}

func TestUpdateSettings_DuplicateLoginEmail_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	factories.NewUser().WithEmail("taken@example.com").Create(db)
	user := factories.NewUser().WithEmail("mine@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              "taken@example.com",
			"notification_email": user.NotificationEmail,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "déjà utilisée")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.Equal(t, "mine@example.com", unchanged.Email)
}

func TestUpdateSettings_SameEmailResubmitted_DoesNotTriggerDuplicateError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithEmail("mine@example.com").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              "mine@example.com",
			"notification_email": user.NotificationEmail,
			"first_name":         user.FirstName,
			"last_name":          user.LastName,
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestUpdateSettings_ValidName_UpdatesAndRedirects(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithName("Jane", "Smith").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              user.Email,
			"notification_email": user.NotificationEmail,
			"first_name":         "Janet",
			"last_name":          "Smithson",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	var updated models.User
	require.NoError(t, db.First(&updated, user.ID).Error)
	assert.Equal(t, "Janet", updated.FirstName)
	assert.Equal(t, "Smithson", updated.LastName)
}

func TestUpdateSettings_MissingName_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithName("Jane", "Smith").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"email":              user.Email,
			"notification_email": user.NotificationEmail,
			"first_name":         "",
			"last_name":          "Smith",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateSettings(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Le prénom et le nom sont requis")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.Equal(t, "Jane", unchanged.FirstName, "name must stay untouched")
}

func TestUpdatePassword_ValidCurrentAndMatchingNew_UpdatesAndRedirects(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithPassword("originalpass").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"current_password": "originalpass",
			"new_password":     "newpassword123",
			"confirm_password": "newpassword123",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdatePassword(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/settings?saved=1", rec.Header().Get("Location"))

	var updated models.User
	require.NoError(t, db.First(&updated, user.ID).Error)
	assert.True(t, updated.CheckPassword("newpassword123"))
	assert.False(t, updated.CheckPassword("originalpass"))
}

func TestUpdatePassword_WrongCurrentPassword_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithPassword("originalpass").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"current_password": "wrongpass",
			"new_password":     "newpassword123",
			"confirm_password": "newpassword123",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdatePassword(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Mot de passe actuel incorrect")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.True(t, unchanged.CheckPassword("originalpass"), "password must stay untouched")
}

func TestUpdatePassword_TooShortNewPassword_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithPassword("originalpass").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"current_password": "originalpass",
			"new_password":     "short",
			"confirm_password": "short",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdatePassword(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "au moins 8 caractères")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.True(t, unchanged.CheckPassword("originalpass"), "password must stay untouched")
}

func TestUpdatePassword_MismatchedConfirmation_ReRendersWithError(t *testing.T) {
	deps, db := setup(t)

	user := factories.NewUser().WithPassword("originalpass").Create(db)

	c := tests.NewContext(http.MethodPost, "/").
		WithAuthenticatedUser(user).
		WithFormData(map[string]string{
			"current_password": "originalpass",
			"new_password":     "newpassword123",
			"confirm_password": "somethingelse123",
		}).
		Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdatePassword(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ne correspondent pas")

	var unchanged models.User
	require.NoError(t, db.First(&unchanged, user.ID).Error)
	assert.True(t, unchanged.CheckPassword("originalpass"), "password must stay untouched")
}

func TestUpdatePassword_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	deps, _ := setup(t)

	c := tests.NewContext(http.MethodPost, "/").
		WithFormData(map[string]string{
			"current_password": "originalpass",
			"new_password":     "newpassword123",
			"confirm_password": "newpassword123",
		}).
		Build()

	handler := UpdatePassword(deps)
	err := handler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}
