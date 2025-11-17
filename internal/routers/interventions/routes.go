package interventions

import (
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
)

func NewRouter(parentGroup echo.Group, dependencies *routers.Dependencies) {
	intervention_routes := parentGroup.Group("/interventions")
	intervention_routes.POST("", CreateNewIntervention(dependencies))
	intervention_routes.GET("/:intervention_id/report", GetInterventionReport(dependencies))
	intervention_routes.GET("/new", GetNewInterventionForm(dependencies))
}
