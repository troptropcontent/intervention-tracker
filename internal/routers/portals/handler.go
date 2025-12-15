package portals

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
	"gorm.io/gorm"
)

const PAGE_LIMIT = 20

type GetPortalsQueryParams struct {
	Page   string `query:"page"`
	Search string `query:"search"`
}

// ParseAndValidate parses and validates query parameters
func (q *GetPortalsQueryParams) ParseAndValidate() (page int, search string) {
	page, _ = strconv.Atoi(q.Page)
	if page < 1 {
		page = 1
	}
	search = q.Search
	return
}

// BuildCurrentURL constructs the current URL with query parameters
func (q *GetPortalsQueryParams) BuildCurrentURL(page int, search string) string {
	currentUrlQueryParams := make(map[string]string)
	if q.Page != "" {
		currentUrlQueryParams["page"] = strconv.Itoa(page)
	}
	if q.Search != "" {
		currentUrlQueryParams["search"] = search
	}
	return templates.MustAddQueryParamsToURL("/admin/portals", currentUrlQueryParams)
}

// applySearchFilter applies search conditions to the query
func applySearchFilter(query *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return query
	}
	searchPattern := "%" + search + "%"
	return query.Where(
		"internal_id ILIKE ? OR address_street ILIKE ? OR address_city ILIKE ? OR contractor_company ILIKE ? OR contact_phone ILIKE ? OR contact_email ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)
}

// fetchPortalsWithPagination retrieves portals from the database with pagination
func fetchPortalsWithPagination(query *gorm.DB, page int) (portals []models.Portal, totalCount int64, totalPages int, err error) {
	// Get total count
	if err = query.Count(&totalCount).Error; err != nil {
		return
	}

	// Calculate pagination
	offset := (page - 1) * PAGE_LIMIT
	totalPages = int((totalCount + int64(PAGE_LIMIT) - 1) / int64(PAGE_LIMIT))

	// Fetch portals
	err = query.Order("name").Limit(PAGE_LIMIT).Offset(offset).Find(&portals).Error
	return
}

func GetPortals(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse query parameters
		var queryParams GetPortalsQueryParams
		if err := (&echo.DefaultBinder{}).BindQueryParams(c, &queryParams); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to parse query params")
		}

		page, search := queryParams.ParseAndValidate()

		// Build and execute query
		query := dependencies.DB.Model(&models.Portal{})
		query = applySearchFilter(query, search)

		portals, totalCount, totalPages, err := fetchPortalsWithPagination(query, page)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch portals")
		}

		// Build current URL and handle HTMX
		currentUrl := queryParams.BuildCurrentURL(page, search)
		if c.Request().Header.Get("HX-Request") == "true" {
			c.Response().Header().Set("HX-Replace-Url", currentUrl)
		}

		// Prepare pagination data
		paginationData := templates.PaginationData{
			BaseUrl:     currentUrl,
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  int(totalCount),
			HasPrev:     page > 1,
			HasNext:     page < totalPages,
		}

		return templates.AdminPortals(c, dependencies.TranslationService, paginationData, search, portals).Render(c.Request().Context(), c.Response().Writer)
	}
}
