package middleware

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
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
