package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/troptropcontent/qr_code_maintenance/internal/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: database <subcommand>")
		fmt.Println("Available subcommands: migrate")
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		if err := migrateCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Error parsing migrate flags: %v", err)
		}

		log.Println("Starting database migration...")
		db := database.MustConnectGORM()

		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

		log.Println("Migration completed successfully!")

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
