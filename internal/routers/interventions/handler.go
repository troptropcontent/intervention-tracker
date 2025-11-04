package interventions

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/interventions"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
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

		utils.PP(args)

		// Create the intervention
		intervention, err := createService.Create(args)
		if err != nil {
			return err
		}

		// Redirect to the intervention detail page or success page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/interventions/%d", intervention.ID))
	}
}
