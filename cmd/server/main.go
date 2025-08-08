package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/handlers"
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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
	
	// 404 handler
	e.RouteNotFound("/*", h.NotFound)
	
	// Start server
	log.Println("Server starting on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}