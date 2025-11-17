package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/troptropcontent/qr_code_maintenance/internal/jobs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: background_jobs <subcommand>")
		fmt.Println("Available subcommands: migrate, start")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	ctx := context.Background()

	switch subcommand {
	case "migrate":
		// For migrations, we only need the DB pool
		cfg, err := jobs.NewConfig(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize configuration: %v", err)
		}
		defer cfg.Close()

		migrator, err := rivermigrate.New(riverpgxv5.New(cfg.DBPool), nil)
		if err != nil {
			log.Fatalf("Failed to create migrator: %v", err)
		}

		res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
		if err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		printVersions := func(res *rivermigrate.MigrateResult) {
			for _, version := range res.Versions {
				fmt.Printf("Migrated [%s] version %d\n", strings.ToUpper(string(res.Direction)), version.Version)
			}
		}

		if len(res.Versions) > 0 {
			printVersions(res)
		} else {
			fmt.Println("Database already up to date")
		}

	case "start":
		// Create shared configuration
		cfg, err := jobs.NewConfig(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize configuration: %v", err)
		}
		defer cfg.Close()

		// Create the job runner
		runner, err := jobs.NewRunner(ctx, cfg)
		if err != nil {
			log.Fatalf("Failed to create job runner: %v", err)
		}
		defer func() {
			stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			runner.Stop(stopCtx)
		}()

		// Start the runner
		if err := runner.Start(ctx); err != nil {
			log.Fatalf("Failed to start job runner: %v", err)
		}

		fmt.Println("Background job runner started. Press Ctrl+C to stop.")

		// Simple graceful shutdown: wait for SIGINT/SIGTERM, then stop
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down gracefully...")
		stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := runner.Stop(stopCtx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		} else {
			fmt.Println("Shutdown complete")
		}

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
