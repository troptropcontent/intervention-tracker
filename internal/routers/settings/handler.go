package settings

import (
	"errors"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	authmiddleware "github.com/troptropcontent/qr_code_maintenance/internal/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"gorm.io/gorm"
)

var errEmailTaken = errors.New("cette adresse email est déjà utilisée")

const minPasswordLength = 8

func GetSettings(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authmiddleware.GetCurrentUser(c, dependencies.DB)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
		}

		saved := c.QueryParam("saved") == "1"

		return templates.AdminSettings(*user, saved, "", "", c).Render(c.Request().Context(), c.Response().Writer)
	}
}

func UpdateSettings(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authmiddleware.GetCurrentUser(c, dependencies.DB)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
		}

		firstName := c.FormValue("first_name")
		lastName := c.FormValue("last_name")
		email := c.FormValue("email")
		notificationEmail := c.FormValue("notification_email")

		// Echo back whatever was submitted (not the stale DB values) if we
		// need to re-render with an error, so the user doesn't lose their edits.
		formUser := *user
		formUser.FirstName = firstName
		formUser.LastName = lastName
		formUser.Email = email
		formUser.NotificationEmail = notificationEmail

		if firstName == "" || lastName == "" {
			return templates.AdminSettings(formUser, false, "Le prénom et le nom sont requis", "", c).Render(c.Request().Context(), c.Response().Writer)
		}

		// Check availability before validating format: an invalid-format
		// address simply won't match any row, so this is harmless to run
		// first and keeps the "taken" vs "invalid" messages distinct.
		if err := emailAvailable(dependencies.DB, email, user.ID); err != nil {
			if errors.Is(err, errEmailTaken) {
				return templates.AdminSettings(formUser, false, "Cette adresse email est déjà utilisée", "", c).Render(c.Request().Context(), c.Response().Writer)
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		if err := user.SetEmail(email); err != nil {
			return templates.AdminSettings(formUser, false, "Adresse email invalide", "", c).Render(c.Request().Context(), c.Response().Writer)
		}

		if err := user.SetNotificationEmail(notificationEmail); err != nil {
			return templates.AdminSettings(formUser, false, "Adresse email de notification invalide", "", c).Render(c.Request().Context(), c.Response().Writer)
		}

		user.FirstName = firstName
		user.LastName = lastName

		if result := dependencies.DB.Save(user); result.Error != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update settings")
		}

		// Keep the session's cached display email in sync so the nav shows
		// the new address immediately, without requiring a fresh login.
		if sess, err := session.Get("session", c); err == nil {
			sess.Values["user_email"] = user.Email
			_ = sess.Save(c.Request(), c.Response())
		}

		return c.Redirect(http.StatusSeeOther, "/admin/settings?saved=1")
	}
}

// UpdatePassword changes the authenticated user's password after
// re-verifying their current one. It's a separate form/endpoint from
// UpdateSettings so a rejected password attempt never touches or re-renders
// stale edits to the identity fields, and vice versa.
func UpdatePassword(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authmiddleware.GetCurrentUser(c, dependencies.DB)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
		}

		currentPassword := c.FormValue("current_password")
		newPassword := c.FormValue("new_password")
		confirmPassword := c.FormValue("confirm_password")

		if !user.CheckPassword(currentPassword) {
			return templates.AdminSettings(*user, false, "", "Mot de passe actuel incorrect", c).Render(c.Request().Context(), c.Response().Writer)
		}

		if len(newPassword) < minPasswordLength {
			return templates.AdminSettings(*user, false, "", "Le nouveau mot de passe doit contenir au moins 8 caractères", c).Render(c.Request().Context(), c.Response().Writer)
		}

		if newPassword != confirmPassword {
			return templates.AdminSettings(*user, false, "", "Les mots de passe ne correspondent pas", c).Render(c.Request().Context(), c.Response().Writer)
		}

		if err := user.SetPassword(newPassword); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
		}

		if result := dependencies.DB.Save(user); result.Error != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update password")
		}

		return c.Redirect(http.StatusSeeOther, "/admin/settings?saved=1")
	}
}

// emailAvailable returns errEmailTaken if email already belongs to a
// different user, nil if it's free or already belongs to selfID.
func emailAvailable(db *gorm.DB, email string, selfID uint) error {
	var existing models.User
	result := db.Where("email = ?", email).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	if existing.ID != selfID {
		return errEmailTaken
	}
	return nil
}
