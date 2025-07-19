package main

import (
	"flag"
	"fmt"
	"log"

	"coolify-go/internal/app"
	"coolify-go/internal/config"
	"coolify-go/internal/database"
)

// Version information - these would normally be set at build time
var (
	Version   = "1.0.0-dev"
	BuildTime = "development"
	GitCommit = "unknown"
)

func printVersion() {
	fmt.Printf("Coolify Go v%s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Git commit: %s\n", GitCommit)
}

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var demoMode = flag.Bool("demo", false, "Run in demo mode without database")
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	// Set version information in config
	config.Version = Version
	config.BuildTime = BuildTime
	config.GitCommit = GitCommit

	// Load configuration
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database (skip if demo mode)
	if !*demoMode {
		if err := config.ConnectDatabase(appConfig.Database); err != nil {
			log.Printf("Database connection failed: %v", err)
			log.Printf("Continuing without database...")
		} else {
			log.Printf("✅ Database connected successfully")
			// Run database migrations
			if err := config.AutoMigrate(); err != nil {
				log.Printf("Database migration failed: %v", err)
			} else {
				log.Printf("✅ Database migrations completed")
				// Seed database with initial data for development
				if err := database.SeedDatabase(); err != nil {
					log.Printf("Database seeding failed: %v", err)
				} else {
					log.Printf("✅ Database seeded successfully")
				}
			}
		}
	} else {
		log.Printf("🎭 Running in DEMO mode - no database required")
	}

	// Create and run application
	application := app.NewApp(appConfig)
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
