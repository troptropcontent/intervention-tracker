package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

var (
	cities = []string{
		"Paris", "Lyon", "Marseille", "Toulouse", "Nice",
		"Nantes", "Strasbourg", "Montpellier", "Bordeaux", "Lille",
		"Rennes", "Reims", "Le Havre", "Saint-Étienne", "Toulon",
		"Grenoble", "Dijon", "Angers", "Nîmes", "Villeurbanne",
	}

	streets = []string{
		"Rue de la République", "Avenue des Champs-Élysées", "Boulevard Saint-Michel",
		"Rue du Commerce", "Avenue Victor Hugo", "Rue de Rivoli",
		"Boulevard Haussmann", "Rue Lafayette", "Avenue Montaigne",
		"Rue Saint-Honoré", "Boulevard Voltaire", "Avenue Foch",
	}

	companies = []string{
		"Portal Solutions Inc.", "Maintenance Pro SAS", "SecureGate SARL",
		"TechPortal France", "Access Control Systems", "GateKeeper SA",
		"Portail Automatique Services", "Smart Entry Solutions", "AutoGate France",
		"Maintenance Automatisme", "Portails & Services", "EntryTech Solutions",
	}
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

	// Create a user
	user := models.User{
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		IsActive:  true,
	}

	if err := user.SetPassword("password123"); err != nil {
		log.Fatalf("Failed to set user password: %v", err)
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Fatalf("Failed to create user: %v", result.Error)
	}

	log.Printf("Successfully created user with ID: %d, Email: %s", user.ID, user.Email)

	// Create 900 portals
	log.Println("Creating 900 portals...")

	portals := make([]models.Portal, 900)
	for i := 0; i < 900; i++ {
		streetNum := (i % 200) + 1
		city := cities[i%len(cities)]
		street := streets[i%len(streets)]
		company := companies[i%len(companies)]

		// Vary installation dates over the past 5 years
		daysAgo := i % 1825 // 5 years in days

		portals[i] = models.Portal{
			UUID:              uuid.New().String(),
			Name:              fmt.Sprintf("Portal %s - %d", city, i+1),
			AddressStreet:     fmt.Sprintf("%d %s", streetNum, street),
			AddressZipcode:    fmt.Sprintf("%05d", 75000+(i%25000)),
			AddressCity:       city,
			ContractorCompany: company,
			ContactPhone:      fmt.Sprintf("+33-%d-%02d-%02d-%02d-%02d", (i%9)+1, i%100, (i*7)%100, (i*13)%100, (i*17)%100),
			ContactEmail:      fmt.Sprintf("contact%d@%s.com", i+1, company[:10]),
			InstallationDate:  time.Now().AddDate(0, 0, -daysAgo),
		}

		// Log progress every 100 portals
		if (i+1)%100 == 0 {
			log.Printf("Generated %d portals...", i+1)
		}
	}

	// Batch insert portals
	log.Println("Inserting portals into database...")
	result = db.CreateInBatches(portals, 100)
	if result.Error != nil {
		log.Fatalf("Failed to create portals: %v", result.Error)
	}

	log.Printf("Successfully created %d portals", result.RowsAffected)
	log.Println("Seeding completed successfully!")
}
