package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/handlers"
)

func main() {
	// Connect to database with GORM
	db, err := database.ConnectGORM()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Get underlying sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Initialize handlers
	h := &handlers.Handlers{DB: db}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/portals/:id", h.GetPortal)
	e.GET("/qr_codes/:uuid", h.QRRedirect)
	admin_routes := e.Group("/admin")
	admin_routes.GET("/portals", h.GetAdminPortals)
	admin_routes.GET("/portals/:id", h.GetAdminPortal)
	admin_routes.POST("/portals/:id/qr-code/associate", h.AssociateQRCode)
	admin_routes.POST("/portals/:id/qr-code/remove", h.RemoveQRCode)
	admin_routes.GET("/portals/scan", h.GetAdminPortalsScan)

	// 404 handler
	e.RouteNotFound("/*", h.NotFound)

	// Start server
	log.Println("Server starting on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
