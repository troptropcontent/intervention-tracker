package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
)

type Handlers struct {
	DB *sqlx.DB
}

func (h *Handlers) GetPortal(c echo.Context) error {
	id := c.Param("id")
	var portal models.Portal
	err := h.DB.Get(&portal, "SELECT * FROM portals WHERE id = $1", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
	}

	return templates.PortalShow(portal).Render(c.Request().Context(), c.Response().Writer)
}