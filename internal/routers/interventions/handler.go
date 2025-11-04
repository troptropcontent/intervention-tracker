package interventions

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/interventions"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"gorm.io/gorm"
)

type CreateNewInterventionFormData struct {
	PortalID  uint   `form:"portal_id"`
	Date      string `form:"date"`
	Summary   string `form:"summary"`
	Signature string `form:"signature"`
	Photos    []struct {
		Name string `form:"name"`
		File *multipart.FileHeader
	} `form:"photos"`
	Controls []struct {
		Kind   string `form:"kind"`
		Result string `form:"result"`
	} `form:"controls"`
}

func CreateNewIntervention(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {

		user, err := routers.FindAuthenticatedUser(c, dependencies.DB)
		if err != nil {
			return err
		}

		formData := CreateNewInterventionFormData{}

		err = routers.ParseFormData(c, &formData)
		if err != nil {
			return err
		}

		for i := range formData.Photos {
			file, err := c.FormFile(fmt.Sprintf("photos[%v].file", i))
			if err != nil && err != http.ErrMissingFile {
				return err
			}
			if file != nil {
				formData.Photos[i].File = file
			}
		}

		// Initialize the create service
		createService := &interventions.CreateInterventionService{
			DB:                       dependencies.DB,
			StorageService:           dependencies.StorageService,
			EmailNotificationService: dependencies.EmailNotificationService,
		}

		// Transform form data to service args
		photos := make([]interventions.PhotoData, 0, len(formData.Photos))
		for _, photo := range formData.Photos {
			if photo.File != nil {
				photos = append(photos, interventions.PhotoData{
					Name: photo.Name,
					File: photo.File,
				})
			}
		}

		controls := make([]struct {
			Kind   string
			Result string
		}, len(formData.Controls))
		for i, ctrl := range formData.Controls {
			controls[i] = struct {
				Kind   string
				Result string
			}{
				Kind:   ctrl.Kind,
				Result: ctrl.Result,
			}
		}

		args := &interventions.CreateArgs{
			Date:      formData.Date,
			Summary:   formData.Summary,
			Signature: formData.Signature,
			Photos:    photos,
			Controls:  controls,
			UserID:    user.ID,
			UserName:  user.FullName(),
			PortalID:  formData.PortalID,
		}

		// Create the intervention
		_, err = createService.Create(args)
		if err != nil {
			return err
		}

		// Redirect to the intervention detail page or success page
		// Use portalUri := c.Echo().Reverse("admin-get-portal", formData.PortalID) when all routes will be migrated to the server
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/portals/%d", formData.PortalID))
	}
}

func GetInterventionReport(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		interventionIdString := c.Param("intervention_id")
		interventionId, err := strconv.Atoi(interventionIdString)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not extract interventionId from request path")
		}

		var intervention models.Intervention
		result := dependencies.DB.Preload("Attachments", "kind = ?", "photo").Preload("Portal").Preload("Controls").First(&intervention, interventionId)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Intervention not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}
		ctx := context.Background()
		for i, photo := range intervention.Attachments {
			url, err := dependencies.StorageService.GetFileURL(
				ctx,
				photo.StorageKey,
				15*time.Minute, // URL valid for 15 minutes
			)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve signed url for %v", photo.FileName)
			}
			intervention.Attachments[i].SignedUrl = url
		}
		return templates.InterventionReport(templates.InterventionReportConfig{Intervention: &intervention, Translator: dependencies.TranslationService}).Render(c.Request().Context(), c.Response().Writer)
	}
}
