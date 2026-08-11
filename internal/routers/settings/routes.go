package settings

import (
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
)

func NewRouter(parentGroup echo.Group, dependencies *routers.Dependencies) {
	settings_routes := parentGroup.Group("/settings")
	settings_routes.GET("", GetSettings(dependencies))
	settings_routes.POST("", UpdateSettings(dependencies))
	settings_routes.POST("/password", UpdatePassword(dependencies))
}
