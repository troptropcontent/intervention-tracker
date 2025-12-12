package portals

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/templates"
)

func GetPortals(dependencies *routers.Dependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get query parameters
		page, _ := strconv.Atoi(c.QueryParam("page"))
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		if limit < 1 {
			limit = 20
		}

		search := c.QueryParam("search")

		// Build query
		query := dependencies.DB.Model(&models.Portal{})

		// Apply search filter if provided
		if search != "" {
			searchPattern := "%" + search + "%"
			query = query.Where(
				"internal_id ILIKE ? OR address_street ILIKE ? OR address_city ILIKE ? OR contractor_company ILIKE ? OR contact_phone ILIKE ? OR contact_email ILIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Get total count for pagination
		var totalCount int64
		if err := query.Count(&totalCount).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to count portals")
		}

		// Calculate pagination
		offset := (page - 1) * limit
		totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

		// Fetch portals with pagination
		var portals []models.Portal
		result := query.Order("name").Limit(limit).Offset(offset).Find(&portals)
		if result.Error != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch portals")
		}

		// Prepare pagination data
		paginationData := templates.PaginationData{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  int(totalCount),
			Limit:       limit,
			HasPrev:     page > 1,
			HasNext:     page < totalPages,
			Search:      search,
		}

		return templates.AdminPortals(portals, paginationData, c).Render(c.Request().Context(), c.Response().Writer)
	}
}
