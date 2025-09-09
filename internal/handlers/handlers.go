package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/middleware"
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

	return templates.PortalShow(portal, c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) NotFound(c echo.Context) error {
	return templates.NotFound(c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortalsScan(c echo.Context) error {
	return templates.AdminPortalsScan(c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortals(c echo.Context) error {
	var portals []models.Portal
	result := h.DB.Order("name").Find(&portals)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch portals")
	}

	return templates.AdminPortals(portals, c).Render(c.Request().Context(), c.Response().Writer)
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

	// Fetch associated QR code if exists
	var interventions []models.Intervention

	result = h.DB.Preload("Controls").Find(&interventions, "portal_id = ?", portal.ID)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Failed to fetch interventions")
	}

	return templates.AdminPortal(portal, qrCodePtr, interventions, c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) GetAdminPortalEdit(c echo.Context) error {
	id := c.Param("id")

	var portal models.Portal
	result := h.DB.First(&portal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return templates.AdminPortalEdit(portal, c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) AssociateQRCode(c echo.Context) error {
	portalIDStr := c.Param("id")

	// Parse JSON body instead of form values
	var requestBody struct {
		QRCodeUUID string `json:"qr_code_uuid"`
	}
	if err := c.Bind(&requestBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON body")
	}
	qrCodeUUID := requestBody.QRCodeUUID

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

	var portal models.Portal
	result := h.DB.First(&portal, portalID)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find portal")
	}

	var qrCode models.QRCode
	result = h.DB.Where("portal_id = ?", portal.ID).First(&qrCode)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find qr code")
	}

	return templates.AdminQrCodeAssociated(&portal, &qrCode).Render(c.Request().Context(), c.Response().Writer)
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

	var portal models.Portal
	result := h.DB.First(&portal, portalIDUint)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find portal")
	}

	var qrCode models.QRCode
	result = h.DB.Where("portal_id = ?", portal.ID).First(&qrCode)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to find qr code")
	}

	qrCode.PortalID = nil
	qrCode.Status = models.QRCodeStatusLost

	result = h.DB.Save(&qrCode)
	if result.Error != nil {
		fmt.Printf("DEBUG: Database error: %v\n", result.Error)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to updte QR code")
	}

	return templates.AdminQrCodeUnassociated(&portal).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) UpdatePortal(c echo.Context) error {
	id := c.Param("id")

	var portal models.Portal
	result := h.DB.First(&portal, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	var updateData struct {
		Name              string `json:"name" form:"name"`
		AddressStreet     string `json:"address_street" form:"address_street"`
		AddressZipcode    string `json:"address_zipcode" form:"address_zipcode"`
		AddressCity       string `json:"address_city" form:"address_city"`
		ContractorCompany string `json:"contractor_company" form:"contractor_company"`
		ContactPhone      string `json:"contact_phone" form:"contact_phone"`
		ContactEmail      string `json:"contact_email" form:"contact_email"`
		InstallationDate  string `json:"installation_date" form:"installation_date"`
	}

	if err := c.Bind(&updateData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Parse installation date if provided
	if updateData.InstallationDate != "" {
		installationDate, err := time.Parse("2006-01-02", updateData.InstallationDate)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid installation date format")
		}
		portal.InstallationDate = installationDate
	}

	// Update portal fields
	if updateData.Name != "" {
		portal.Name = updateData.Name
	}
	if updateData.AddressStreet != "" {
		portal.AddressStreet = updateData.AddressStreet
	}
	if updateData.AddressZipcode != "" {
		portal.AddressZipcode = updateData.AddressZipcode
	}
	if updateData.AddressCity != "" {
		portal.AddressCity = updateData.AddressCity
	}
	if updateData.ContractorCompany != "" {
		portal.ContractorCompany = updateData.ContractorCompany
	}
	if updateData.ContactPhone != "" {
		portal.ContactPhone = updateData.ContactPhone
	}
	portal.ContactEmail = updateData.ContactEmail

	result = h.DB.Save(&portal)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update portal")
	}

	// Return to admin portal view
	return c.Redirect(http.StatusSeeOther, "/admin/portals/"+id)
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
	return c.Redirect(http.StatusSeeOther, "/portals/"+strconv.Itoa(int(qrCode.Portal.ID)))
}

func (h *Handlers) GetNewIntervention(c echo.Context) error {
	user, err := middleware.GetCurrentUser(c, h.DB)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}

	id := c.Param("id")

	var portal models.Portal
	result := h.DB.Where("id = ?", id).First(&portal)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return templates.AdminInterventionNew(portal, *user, c).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) PostIntervention(c echo.Context) error {
	user, err := middleware.GetCurrentUser(c, h.DB)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user")
	}

	id := c.Param("id")
	portalID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid portal ID")
	}

	var portal models.Portal
	result := h.DB.Where("id = ?", id).First(&portal)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// Parse form data
	var formData struct {
		Date    string `form:"date"`
		Summary string `form:"summary"`
	}

	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Parse intervention date
	interventionDate, err := time.Parse("2006-01-02", formData.Date)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format")
	}

	// Create intervention
	intervention := models.Intervention{
		Date:     interventionDate,
		UserID:   user.ID,
		UserName: user.FullName(),
		PortalID: uint(portalID),
	}

	// Set summary if provided
	if formData.Summary != "" {
		intervention.Summary = &formData.Summary
	}

	// Start transaction
	tx := h.DB.Begin()
	if tx.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	// Save intervention
	if result := tx.Create(&intervention); result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create intervention")
	}

	// Process control results
	allControlTypes := append(models.ControlTypesByKind[models.ControlKindSecurity], models.ControlTypesByKind[models.ControlKindOther]...)

	for _, controlType := range allControlTypes {
		controlValue := c.FormValue("control_" + controlType)

		// Only create control records for non-empty values (user made a selection)
		if controlValue != "" {
			var result models.ControlResult
			switch controlValue {
			case "true":
				trueVal := true
				result = &trueVal
			case "false":
				falseVal := false
				result = &falseVal
			}
			// If controlValue is empty string, result stays nil (not controlled)

			control := models.Control{
				Kind:           controlType,
				Result:         result,
				InterventionID: intervention.ID,
			}

			if res := tx.Create(&control); res.Error != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create control")
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save intervention")
	}

	// Redirect to portal admin page
	return c.Redirect(http.StatusSeeOther, "/admin/portals/"+id)
}

func (h *Handlers) GetInterventionReport(c echo.Context) error {
	id := c.Param("id")

	var intervention models.Intervention
	result := h.DB.Preload("Portal").Preload("Controls").First(&intervention, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Intervention not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return templates.InterventionReport(intervention).Render(c.Request().Context(), c.Response().Writer)
}
