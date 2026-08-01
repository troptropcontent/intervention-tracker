package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/troptropcontent/qr_code_maintenance/internal/database"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

func main() {
	var (
		count   = flag.Int("count", 50, "Number of QR codes to generate")
		baseURL = flag.String("url", "http://localhost:8080", "Base URL for QR codes")
		output  = flag.String("output", "qr_codes", "Output directory for QR code images")
		size    = flag.Int("size", 256, "QR code image size in pixels")
		help    = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		fmt.Println("QR Code Generator for Portal Maintenance System")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Printf("  %s [options]\n", os.Args[0])
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Printf("  %s -count=100 -url=https://portals.example.com\n", os.Args[0])
		fmt.Printf("  %s -count=25 -output=batch1 -size=512\n", os.Args[0])
		return
	}

	if *count <= 0 {
		log.Fatal("Count must be greater than 0")
	}

	fmt.Printf("🚀 Generating %d QR codes...\n", *count)
	fmt.Printf("📁 Output directory: %s\n", *output)
	fmt.Printf("🌐 Base URL: %s\n", *baseURL)
	fmt.Printf("📏 Size: %dx%d pixels\n", *size, *size)
	fmt.Println()

	// Connect to database with GORM
	db, err := database.ConnectGORM()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get underlying sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Create output directory if it doesn't exist
	err = os.MkdirAll(*output, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate QR codes
	var generatedCodes []models.QRCode

	for i := 0; i < *count; i++ {
		// Generate UUID
		qrUUID := uuid.New().String()

		// Create QR code URL - pointing to the redirect endpoint
		qrURL := fmt.Sprintf("%s/qr_codes/%s", *baseURL, qrUUID)

		// Generate QR code image
		filename := fmt.Sprintf("%s/qr_%s.png", *output, qrUUID)
		err = qrcode.WriteFile(qrURL, qrcode.Medium, *size, filename)
		if err != nil {
			log.Printf("Failed to generate QR code %s: %v", qrUUID, err)
			continue
		}

		// Create database record
		qrCode := models.QRCode{
			UUID:   qrUUID,
			Status: models.QRCodeStatusAvailable,
		}

		generatedCodes = append(generatedCodes, qrCode)

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == *count {
			fmt.Printf("✅ Generated %d/%d QR codes\n", i+1, *count)
		}
	}

	fmt.Println()
	fmt.Printf("💾 Saving %d QR codes to database...\n", len(generatedCodes))

	// Insert QR codes into database using GORM batch insert
	result := db.Create(&generatedCodes)
	if result.Error != nil {
		log.Fatalf("Failed to insert QR codes: %v", result.Error)
	}

	fmt.Println("🎉 QR code generation completed successfully!")
	fmt.Println()
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   • Generated: %d QR codes\n", len(generatedCodes))
	fmt.Printf("   • Images saved to: %s/\n", *output)
	fmt.Printf("   • Database records: %d\n", len(generatedCodes))
	fmt.Printf("   • Status: Available for association\n")
	fmt.Println()
	fmt.Printf("💡 Next steps:\n")
	fmt.Printf("   1. Print the QR code images from %s/\n", *output)
	fmt.Printf("   2. Stick them on portals as needed\n")
	fmt.Printf("   3. Associate them via the admin interface\n")
}
