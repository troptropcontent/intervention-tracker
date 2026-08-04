package qrcodes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/pdf"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/qrcodes"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
	"gorm.io/gorm"
)

const PAGE_LIMIT = 20

type GetQRCodesQueryParams struct {
	Page   string `query:"page"`
	Status string `query:"status"`
	Search string `query:"search"`
}

// ParseAndValidate parses and validates query parameters
func (q *GetQRCodesQueryParams) ParseAndValidate() (page int, status string, search string) {
	page, _ = strconv.Atoi(q.Page)
	if page < 1 {
		page = 1
	}
	status = q.Status
	if !models.QRCodeStatus(status).IsValid() {
		status = ""
	}
	search = q.Search
	return
}

// BuildCurrentURL constructs the current URL with query parameters
func (q *GetQRCodesQueryParams) BuildCurrentURL(page int, status string, search string) string {
	currentUrlQueryParams := make(map[string]string)
	if q.Page != "" {
		currentUrlQueryParams["page"] = strconv.Itoa(page)
	}
	if status != "" {
		currentUrlQueryParams["status"] = status
	}
	if search != "" {
		currentUrlQueryParams["search"] = search
	}
	return templates.MustAddQueryParamsToURL("/admin/qr-codes", currentUrlQueryParams)
}

func applyStatusFilter(query *gorm.DB, status string) *gorm.DB {
	if status == "" {
		return query
	}
	return query.Where("status = ?", status)
}

func applyUUIDSearchFilter(query *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return query
	}
	return query.Where("uuid ILIKE ?", "%"+search+"%")
}

func fetchQRCodesWithPagination(query *gorm.DB, page int) (qrCodes []models.QRCode, totalCount int64, totalPages int, err error) {
	if err = query.Count(&totalCount).Error; err != nil {
		return
	}

	offset := (page - 1) * PAGE_LIMIT
	totalPages = int((totalCount + int64(PAGE_LIMIT) - 1) / int64(PAGE_LIMIT))

	err = query.Preload("Portal").Order("generated_at desc").Limit(PAGE_LIMIT).Offset(offset).Find(&qrCodes).Error
	return
}

func GetQRCodes(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		var queryParams GetQRCodesQueryParams
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &queryParams); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse query params")
		}

		page, status, search := queryParams.ParseAndValidate()

		query := dependencies.DB.Model(&models.QRCode{})
		query = applyStatusFilter(query, status)
		query = applyUUIDSearchFilter(query, search)

		qrCodes, totalCount, totalPages, err := fetchQRCodesWithPagination(query, page)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch qr codes")
		}

		currentUrl := queryParams.BuildCurrentURL(page, status, search)
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Replace-Url", currentUrl)
		}

		paginationData := templates.PaginationData{
			BaseUrl:     currentUrl,
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  int(totalCount),
			HasPrev:     page > 1,
			HasNext:     page < totalPages,
		}

		return templates.AdminQRCodes(c, paginationData, status, search, qrCodes).Render(c.Request().Context(), c.Response().Writer)
	}
}

func GetNewBatchForm(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		return templates.AdminQRCodeBatchNew(c).Render(c.Request().Context(), c.Response().Writer)
	}
}

type CreateBatchFormData struct {
	Count int    `form:"count"`
	Label string `form:"label"`
}

func CreateBatch(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		var formData CreateBatchFormData
		if err := routers.ParseFormData(c, &formData); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to parse form data: %v", err))
		}

		generateService, err := qrcodes.NewGenerateService(dependencies.DB)
		if err != nil {
			return err
		}

		var label *string
		if formData.Label != "" {
			label = &formData.Label
		}

		batch, err := generateService.GenerateBatch(formData.Count, label)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/qr-codes/batches/%d", batch.ID))
	}
}

func GetBatch(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		batchID := c.Param("batch_id")

		var batch models.QRCodeBatch
		result := dependencies.DB.Preload("QRCodes", func(db *gorm.DB) *gorm.DB {
			return db.Order("uuid")
		}).First(&batch, batchID)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Batch not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		return templates.AdminQRCodeBatchShow(c, batch).Render(c.Request().Context(), c.Response().Writer)
	}
}

func DownloadBatchPDF(dependencies *routers.Dependencies, pdfConverter pdf.HtmlToPdfConverter) echo.HandlerFunc {
	return func(c echo.Context) error {
		batchID := c.Param("batch_id")

		var batch models.QRCodeBatch
		result := dependencies.DB.First(&batch, batchID)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Batch not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		var codes []models.QRCode
		if err := dependencies.DB.Where("batch_id = ?", batch.ID).Order("uuid").Find(&codes).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch qr codes")
		}

		baseURL := utils.GetEnv("APP_BASE_URL", "http://localhost:8080")
		pdfService := qrcodes.NewPDFService(pdfConverter)

		pdfFile, err := pdfService.GenerateSheetPDF(codes, baseURL)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate PDF")
		}
		defer func() {
			pdfFile.Close()
			os.Remove(pdfFile.Name())
		}()

		return c.Attachment(pdfFile.Name(), fmt.Sprintf("qr-codes-batch-%d.pdf", batch.ID))
	}
}

type UpdateStatusFormData struct {
	Status string `form:"status"`
}

var manuallyAssignableQRCodeStatuses = map[models.QRCodeStatus]bool{
	models.QRCodeStatusAvailable: true,
	models.QRCodeStatusDamaged:   true,
	models.QRCodeStatusLost:      true,
}

func UpdateStatus(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		uuid := c.Param("uuid")

		var formData UpdateStatusFormData
		if err := routers.ParseFormData(c, &formData); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to parse form data: %v", err))
		}

		newStatus := models.QRCodeStatus(formData.Status)
		if !manuallyAssignableQRCodeStatuses[newStatus] {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid status: '%s'", formData.Status))
		}

		var qrCode models.QRCode
		result := dependencies.DB.Where("uuid = ?", uuid).First(&qrCode)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "QR code not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}

		if qrCode.Status == models.QRCodeStatusAssociated {
			return echo.NewHTTPError(http.StatusConflict, "Cannot change the status of a QR code associated to a portal")
		}

		result = dependencies.DB.Model(&qrCode).Update("status", newStatus)
		if result.Error != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update qr code status")
		}

		referer := c.Request().Header.Get("Referer")
		if referer == "" {
			referer = "/admin/qr-codes"
		}
		return c.Redirect(http.StatusSeeOther, referer)
	}
}
