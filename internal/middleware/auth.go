package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get("session", c)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			userID, ok := sess.Values["user_id"]
			if !ok || userID == nil {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			c.Set("user_id", userID)
			c.Set("user_email", sess.Values["user_email"])

			return next(c)
		}
	}
}

func GetCurrentUser(c echo.Context, db *gorm.DB) (*models.User, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return nil, errors.New("user not authenticated")
	}

	var user models.User
	result := db.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
