package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"gorm.io/gorm"
)

func (h *Handlers) GetLogin(c echo.Context) error {
	return templates.Login("", "", "", c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) PostLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return templates.Login("", "", "Email and password are required", c).Render(c.Request().Context(), c.Response().Writer)
	}

	var user models.User
	result := h.DB.Where("email = ? AND is_active = ?", email, true).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return templates.Login(email, password, "Invalid email or password", c).Render(c.Request().Context(), c.Response().Writer)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if !user.CheckPassword(password) {
		return templates.Login(email, password, "Invalid email or password", c).Render(c.Request().Context(), c.Response().Writer)
	}

	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Session error")
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect(http.StatusSeeOther, "/admin/portals")
}

func (h *Handlers) GetRegister(c echo.Context) error {
	return templates.Register("", c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) PostRegister(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")

	if email == "" || password == "" || firstName == "" || lastName == "" {
		return templates.Register("All fields are required", c).Render(c.Request().Context(), c.Response().Writer)
	}

	if len(password) < 8 {
		return templates.Register("Password must be at least 8 characters", c).Render(c.Request().Context(), c.Response().Writer)
	}

	var existingUser models.User
	result := h.DB.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		return templates.Register("Email already exists", c).Render(c.Request().Context(), c.Response().Writer)
	}

	user := models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	if err := user.SetPassword(password); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	result = h.DB.Create(&user)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	sess, err := session.Get("session", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Session error")
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handlers) PostLogout(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear session")
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}
