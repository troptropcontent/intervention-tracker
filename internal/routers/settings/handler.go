package settings

import (
	"net/http"

	"github.com/labstack/echo/v4"
	authmiddleware "github.com/troptropcontent/qr_code_maintenance/internal/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
)

func GetSettings(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authmiddleware.GetCurrentUser(c, dependencies.DB)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
		}

		saved := c.QueryParam("saved") == "1"

		return templates.AdminSettings(*user, saved, "", c).Render(c.Request().Context(), c.Response().Writer)
	}
}

func UpdateSettings(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authmiddleware.GetCurrentUser(c, dependencies.DB)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
		}

		notificationEmail := c.FormValue("notification_email")

		if err := user.SetNotificationEmail(notificationEmail); err != nil {
			return templates.AdminSettings(*user, false, "Adresse email invalide", c).Render(c.Request().Context(), c.Response().Writer)
		}

		if result := dependencies.DB.Save(user); result.Error != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update settings")
		}

		return c.Redirect(http.StatusSeeOther, "/admin/settings?saved=1")
	}
}
