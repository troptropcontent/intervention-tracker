package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

// IsLoggedIn reports whether the request carries a valid session, without
// requiring RequireAuth to have run first. Useful for handlers on routes
// that must stay reachable by anonymous users but still branch on auth state.
func IsLoggedIn(c echo.Context) bool {
	sess, err := session.Get("session", c)
	if err != nil {
		return false
	}

	userID, ok := sess.Values["user_id"]
	return ok && userID != nil
}

// LoadSessionUser populates user_id/user_email in the request context when a
// valid session exists, without requiring one. Unlike RequireAuth, it never
// redirects, so it's safe to run on every route (including public ones) and
// lets the shared layout render the authenticated nav for logged-in users
// even on pages anonymous visitors can also reach.
func LoadSessionUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get("session", c)
			if err == nil {
				if userID, ok := sess.Values["user_id"]; ok && userID != nil {
					c.Set("user_id", userID)
					c.Set("user_email", sess.Values["user_email"])
				}
			}

			return next(c)
		}
	}
}

// RequireAuth rejects the request unless LoadSessionUser (registered globally
// in main.go, ahead of route matching) already found a valid session and
// populated user_id in the context.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("user_id") == nil {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

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
