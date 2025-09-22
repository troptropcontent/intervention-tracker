package main

import (
	"log"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/handlers"
	authmiddleware "github.com/troptropcontent/qr_code_maintenance/internal/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

func main() {
	// Connect to database with GORM
	db, err := database.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	emailService, err := email.NewSMTPServiceGmail()
	if err != nil {
		log.Fatalf("failed to instanciate email service: %v", err)
	}
	// Initialize handlers
	h := &handlers.Handlers{DB: db, EmailNotificationService: emailService}

	e := echo.New()

	// Session middleware
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(utils.MustGetEnv("GO_COOKIE_SECRET")))))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files
	e.Static("/static", "static")

	// Public routes
	e.GET("/login", h.GetLogin)
	e.POST("/login", h.PostLogin)
	e.GET("/register", h.GetRegister)
	e.POST("/register", h.PostRegister)
	e.POST("/logout", h.PostLogout)

	// Protected routes
	e.GET("/portals/:id", h.GetPortal, authmiddleware.RequireAuth())
	e.GET("/qr_codes/:uuid", h.QRRedirect)

	// Admin routes (require authentication)
	admin_routes := e.Group("/admin", authmiddleware.RequireAuth())
	admin_routes.GET("/portals", h.GetAdminPortals)
	admin_routes.GET("/portals/:id", h.GetAdminPortal)
	admin_routes.GET("/portals/:id/edit", h.GetAdminPortalEdit)
	admin_routes.POST("/portals/:id", h.UpdatePortal)
	admin_routes.POST("/portals/:id/qr-code/associate", h.AssociateQRCode)
	admin_routes.POST("/portals/:id/qr-code/remove", h.RemoveQRCode)
	admin_routes.GET("/portals/:id/interventions/new", h.GetNewIntervention)
	admin_routes.POST("/portals/:id/interventions", h.PostIntervention)
	admin_routes.GET("/interventions/:id/report", h.GetInterventionReport)
	admin_routes.GET("/portals/scan", h.GetAdminPortalsScan)

	// 404 handler
	e.RouteNotFound("/*", h.NotFound)

	// Start server
	log.Println("Server starting on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
