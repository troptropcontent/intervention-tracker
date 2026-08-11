package interventions

import (
	"context"
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/interventions"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
	"gorm.io/gorm"
)

type CreateNewInterventionFormData struct {
	Type           string    `form:"type"`
	PortalID       uint      `form:"portal_id"`
	Date           time.Time `form:"date"`
	Summary        string    `form:"summary"`
	TimeSpentHours *float64  `form:"time_spent_hours"`
	Signature      string    `form:"signature"`
	Photos         []struct {
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to parse form data: %v", err))
		}

		interventionType := models.InterventionType(formData.Type)
		if !interventionType.IsValid() {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid type: '%s'", formData.Type))
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

		var timeSpentMinutes *int
		if formData.TimeSpentHours != nil {
			timeSpentMinutes = utils.Ptr(int(math.Round(*formData.TimeSpentHours * 60)))
		}

		// Initialize the create service
		createService, err := interventions.NewCreateInterventionService(
			dependencies.DB,
			dependencies.StorageService,
			dependencies.EmailNotificationService,
			dependencies.BackGroundJobRunner,
		)
		if err != nil {
			return err
		}

		args := &interventions.CreateArgs{
			Type:      interventionType,
			Date:      formData.Date,
			Summary:   formData.Summary,
			Signature: formData.Signature,
			TimeSpent: timeSpentMinutes,
			UserID:    user.ID,
			UserName:  user.FullName(),
			PortalID:  formData.PortalID,
		}

		// Transform form data to service args
		for _, photo := range formData.Photos {
			if photo.File != nil {
				args.Photos = append(args.Photos, interventions.PhotoData{
					Name: photo.Name,
					File: photo.File,
				})
			}
		}

		// For maintenance interventions, validate all required controls are present
		if interventionType == models.InterventionTypeMaintenance {
			// Early length check - fail fast if wrong number of controls
			if len(formData.Controls) != len(models.ControlKinds) {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					fmt.Sprintf("maintenance requires exactly %d control checks, received %d",
						len(models.ControlKinds),
						len(formData.Controls)),
				)
			}

			// Track which control kinds we've seen (for O(1) lookups)
			seenControlKinds := make(map[models.ControlKind]bool, len(models.ControlKinds))

			for _, control := range formData.Controls {
				kind := models.ControlKind(control.Kind)

				// Validate this is a recognized control kind
				if !kind.IsValid() {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid control kind: '%s'", kind))
				}

				// Check for duplicates
				if seenControlKinds[kind] {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("duplicate control kind: '%s'", kind))
				}
				seenControlKinds[kind] = true

				result := models.ControlResult(control.Result)
				if !result.IsValid() {
					return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid control result: '%s'", result))
				}

				args.Controls = append(args.Controls, interventions.ControlData{
					Kind:   kind,
					Result: result,
				})
			}

			// Verify all required control kinds are present
			var missingControls []models.ControlKind
			for _, requiredKind := range models.ControlKinds {
				if !seenControlKinds[requiredKind] {
					missingControls = append(missingControls, requiredKind)
				}
			}

			if len(missingControls) > 0 {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing required controls: %v", missingControls))
			}
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

// GetInterventionReportPDF resolves the PDF report that was actually generated and
// emailed for this intervention (stored in S3 as an Attachment of kind "report") and
// redirects to a short-lived signed URL for it.
func GetInterventionReportPDF(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		interventionIdString := c.Param("intervention_id")
		interventionId, err := strconv.Atoi(interventionIdString)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not extract interventionId from request path")
		}

		var attachment models.Attachment
		result := dependencies.DB.
			Where("holder_type = ? AND holder_id = ? AND kind = ? AND upload_status = ?", "interventions", interventionId, "report", models.AttachmentUploadStatusCompleted).
			Order("created_at desc").
			First(&attachment)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Report not generated yet for this intervention")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		url, err := dependencies.StorageService.GetFileURL(
			c.Request().Context(),
			attachment.StorageKey,
			15*time.Minute, // URL valid for 15 minutes
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not retrieve signed url for %v", attachment.FileName)
		}

		return c.Redirect(http.StatusFound, url)
	}
}

func GetNewInterventionForm(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := routers.FindAuthenticatedUser(c, dependencies.DB)
		if err != nil {
			return err
		}

		portalId, err := strconv.Atoi(c.QueryParam("portal_id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "query param portal_id is invalid or missing")
		}

		interventionType := models.InterventionType(c.QueryParam("intervention_type"))
		if !interventionType.IsValid() {
			return echo.NewHTTPError(http.StatusBadRequest, "query param intervention_type is invalid or missing")
		}

		var portal models.Portal
		result := dependencies.DB.Where("id = ?", portalId).First(&portal)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Portal not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		return templates.AdminInterventionNew(c, dependencies.TranslationService, &portal, user, interventionType).Render(c.Request().Context(), c.Response().Writer)
	}
}
