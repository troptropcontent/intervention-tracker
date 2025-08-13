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
	uuid := c.Param("uuid")

	var portal models.Portal
	err := h.DB.Get(&portal, "SELECT * FROM portals WHERE uuid = $1", uuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
	}

	return templates.PortalShow(portal).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) NotFound(c echo.Context) error {
	return templates.NotFound().Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortalsScan(c echo.Context) error {
	return templates.AdminPortalsScan().Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortals(c echo.Context) error {
	var portals []models.Portal
	err := h.DB.Select(&portals, "SELECT * FROM portals ORDER BY name")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch portals")
	}

	return templates.AdminPortals(portals).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortal(c echo.Context) error {
	id := c.Param("id")

	var portal models.Portal
	err := h.DB.Get(&portal, "SELECT * FROM portals WHERE id = $1", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
	}

	return templates.AdminPortal(portal).Render(c.Request().Context(), c.Response().Writer)
}
