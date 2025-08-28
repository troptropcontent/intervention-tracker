package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

func main() {
	// Connect to database
	db, err := database.ConnectGORM()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create a sample portal
	portal := models.Portal{
		UUID:              uuid.New().String(),
		Name:              "Main Entrance Portal",
		AddressStreet:     "123 Main Street",
		AddressZipcode:    "12345",
		AddressCity:       "New York",
		ContractorCompany: "Portal Solutions Inc.",
		ContactPhone:      "+1-555-0123",
		ContactEmail:      "contact@portalsolutions.com",
		InstallationDate:  time.Now().AddDate(0, -1, 0), // Installed 1 month ago
	}

	// Insert the portal into database
	result := db.Create(&portal)
	if result.Error != nil {
		log.Fatalf("Failed to create portal: %v", result.Error)
	}

	log.Printf("Successfully created portal with ID: %d, UUID: %s", portal.ID, portal.UUID)
}
