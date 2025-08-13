package main

import (
	"log"
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/database"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Insert sample portal
	query := `
		INSERT INTO portals (name, address_street, address_zipcode, address_city, contractor_company, contact_phone, contact_email, installation_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING`

	installDate, _ := time.Parse("2006-01-02", "2024-01-15")

	_, err = db.Exec(query,
		"Portail Principal",
		"123 Rue de la Paix",
		"75001",
		"Paris",
		"TechnoPorte SARL",
		"01.23.45.67.89",
		"contact@technoporte.fr",
		installDate)

	if err != nil {
		log.Fatalf("Failed to insert sample data: %v", err)
	}

	log.Println("Sample data inserted successfully")
}
