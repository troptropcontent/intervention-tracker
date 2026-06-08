package routers

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/form"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

func ParseFormData(c echo.Context, target any) error {
	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (any, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			return fmt.Errorf("could not parse multipart form: %w", err)
		}
	} else {
		if err := c.Request().ParseForm(); err != nil {
			return fmt.Errorf("could not parse form data: %w", err)
		}
	}

	if err := decoder.Decode(target, c.Request().Form); err != nil {
		return fmt.Errorf("could not decode form data: %w", err)
	}

	return nil
}

func FindAuthenticatedUser(c echo.Context, db *gorm.DB) (*models.User, error) {
	userID := c.Get("user_id")
	if userID == nil {
		return nil, fmt.Errorf("no authenticated user in context")
	}

	var user models.User
	result := db.First(&user, userID)

	if result.Error != nil {
		return nil, fmt.Errorf("could not find authenticated user: %w", result.Error)
	}

	return &user, nil
}
