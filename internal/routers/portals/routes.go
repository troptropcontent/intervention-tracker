package portals

import (
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
)

func NewRouter(parentGroup echo.Group, dependencies *routers.Dependencies) {
	intervention_routes := parentGroup.Group("/portals")
	intervention_routes.GET("", GetPortals(dependencies))
}
