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

	if qrCodeUUID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "QR Code UUID is required")
	}

	// Convert portal ID to uint
	portalID, err := strconv.ParseUint(portalIDStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid portal ID")
	}

	// Use GORM transaction
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Check if QR code exists and is available
		var qrCode models.QRCode
		result := tx.Where("uuid = ?", qrCodeUUID).First(&qrCode)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "QR Code not found")
			}
			return result.Error
		}

		if qrCode.Status != models.QRCodeStatusAvailable {
			return echo.NewHTTPError(http.StatusBadRequest, "QR Code is not available")
		}

		// Check if portal already has a QR code
		var existingCount int64
		result = tx.Model(&models.QRCode{}).Where("portal_id = ? AND status = ?", portalID, models.QRCodeStatusAssociated).Count(&existingCount)
		if result.Error != nil {
			return result.Error
		}
		if existingCount > 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Portal already has an associated QR code")
		}

		// Associate the QR code with the portal
		portalIDUint := uint(portalID)
		now := time.Now()
		result = tx.Model(&qrCode).Updates(models.QRCode{
			PortalID:     &portalIDUint,
			Status:       models.QRCodeStatusAssociated,
			AssociatedAt: &now,
		})
		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		if echoErr, ok := err.(*echo.HTTPError); ok {
			return echoErr
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to associate QR code")
	}

	// Redirect back to portal page
	return c.Redirect(http.StatusSeeOther, "/admin/portals/"+portalIDStr)
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
