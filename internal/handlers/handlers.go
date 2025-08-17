package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"gorm.io/gorm"
)

type Handlers struct {
	DB *gorm.DB
}

func (h *Handlers) GetPortal(c echo.Context) error {
	id := c.Param("id")

	var portal models.Portal
	result := h.DB.Where("id = ?", id).First(&portal)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
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
	result := h.DB.Order("name").Find(&portals)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch portals")
	}

	return templates.AdminPortals(portals).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortal(c echo.Context) error {
	id := c.Param("id")

	var portal models.Portal
	result := h.DB.First(&portal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Fetch associated QR code if exists
	var qrCode models.QRCode
	result = h.DB.Where("portal_id = ? AND status = ?", portal.ID, models.QRCodeStatusAssociated).First(&qrCode)
	var qrCodePtr *models.QRCode
	if result.Error == nil {
		qrCodePtr = &qrCode
	}

	return templates.AdminPortal(portal, qrCodePtr).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) AssociateQRCode(c echo.Context) error {
	portalIDStr := c.Param("id")
	qrCodeUUID := c.FormValue("qr_code_uuid")

	portalID, err := h.parseAndValidateInput(portalIDStr, qrCodeUUID)
	if err != nil {
		return err
	}

	if err := h.validateAssociation(portalID, qrCodeUUID); err != nil {
		return err
	}

	if err := h.performAssociation(portalID, qrCodeUUID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to associate QR code")
	}

	return c.Redirect(http.StatusSeeOther, "/admin/portals/"+portalIDStr)
}

func (h *Handlers) parseAndValidateInput(portalIDStr, qrCodeUUID string) (uint, error) {
	if qrCodeUUID == "" {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "QR Code UUID is required")
	}

	portalID, err := strconv.ParseUint(portalIDStr, 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid portal ID")
	}

	return uint(portalID), nil
}

func (h *Handlers) validateAssociation(portalID uint, qrCodeUUID string) error {
	// Check QR code exists and is available
	var qrCode models.QRCode
	result := h.DB.Where("uuid = ? AND status = ?", qrCodeUUID, models.QRCodeStatusAvailable).First(&qrCode)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusBadRequest, "QR Code not found or not available")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Check portal doesn't already have a QR code
	var count int64
	result = h.DB.Model(&models.QRCode{}).Where("portal_id = ? AND status = ?", portalID, models.QRCodeStatusAssociated).Count(&count)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	if count > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Portal already has an associated QR code")
	}

	return nil
}

func (h *Handlers) performAssociation(portalID uint, qrCodeUUID string) error {
	now := time.Now()
	result := h.DB.Model(&models.QRCode{}).Where("uuid = ?", qrCodeUUID).Updates(models.QRCode{
		PortalID:     &portalID,
		Status:       models.QRCodeStatusAssociated,
		AssociatedAt: &now,
	})
	return result.Error
}

func (h *Handlers) RemoveQRCode(c echo.Context) error {
	portalID := c.Param("id")

	// Convert portal ID to uint
	portalIDUint, err := strconv.ParseUint(portalID, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid portal ID")
	}

	// Update QR code to available status
	result := h.DB.Model(&models.QRCode{}).Where("portal_id = ? AND status = ?", portalIDUint, models.QRCodeStatusAssociated).Updates(models.QRCode{
		PortalID:     nil,
		Status:       models.QRCodeStatusAvailable,
		AssociatedAt: nil,
	})
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to remove QR code association")
	}

	// Redirect back to portal page
	return c.Redirect(http.StatusSeeOther, "/admin/portals/"+portalID)
}

func (h *Handlers) QRRedirect(c echo.Context) error {
	qrUUID := c.Param("uuid")

	// Find the associated portal using GORM joins
	var qrCode models.QRCode
	result := h.DB.Preload("Portal").Where("uuid = ? AND status = ?", qrUUID, models.QRCodeStatusAssociated).First(&qrCode)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "QR Code not found or not associated")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Check if portal is loaded
	if qrCode.Portal == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Associated portal not found")
	}

	// Redirect to the portal page using its UUID
	return c.Redirect(http.StatusSeeOther, "/portals/"+qrCode.Portal.UUID)
}
