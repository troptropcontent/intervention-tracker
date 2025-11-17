package main

import (
	"context"
	"fmt"
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
		fmt.Println("Usage: database <subcommand>")
		fmt.Println("Available subcommands: migrate")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	ctx := context.Background()

	switch subcommand {
	case "migrate":
		dbPool := jobs.GetDbPool(ctx)
		migrator, err := rivermigrate.New(riverpgxv5.New(dbPool), nil)
		if err != nil {
			panic(err)
		}
		defer dbPool.Close()

		res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
		if err != nil {
			panic(err)
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
		runner := jobs.NewRunner(ctx)
		if err := runner.RiverClient.Start(ctx); err != nil {
			panic(err)
		}

		fmt.Println("Background job runner started. Press Ctrl+C to stop.")

		// Simple graceful shutdown: wait for SIGINT/SIGTERM, then stop
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nShutting down gracefully...")
		stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := runner.RiverClient.Stop(stopCtx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		} else {
			fmt.Println("Shutdown complete")
		}
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
