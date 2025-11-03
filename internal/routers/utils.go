package routers

import (
	"fmt"
	"strings"

	"github.com/go-playground/form"
	"github.com/labstack/echo/v4"
)

func ParseFormData(c echo.Context, target any) error {
	decoder := form.NewDecoder()

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
