package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"coolify-go/internal/config"
)

func main() {
	var (
		action = flag.String("action", "migrate", "Migration action: migrate, status, rollback, schema-info")
		step   = flag.Int("step", 1, "Number of steps to rollback (for rollback action)")
	)
	flag.Parse()

	// Load configuration
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := config.ConnectDatabase(appConfig.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	switch *action {
	case "migrate":
		fmt.Println("🔄 Running database migrations...")
		if err := config.AutoMigrate(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Database migrations completed successfully")

	case "status":
		fmt.Println("📊 Checking database migration status...")
		if err := showMigrationStatus(); err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}

	case "schema-info":
		fmt.Println("🔍 Checking schema hash information...")
		if err := showSchemaInfo(); err != nil {
			log.Fatalf("Failed to get schema info: %v", err)
		}

	case "rollback":
		fmt.Printf("⏪ Rolling back %d migration(s)...\n", *step)
		if err := rollbackMigrations(*step); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully")

	default:
		fmt.Printf("❌ Unknown action: %s\n", *action)
		fmt.Println("Available actions: migrate, status, schema-info, rollback")
		os.Exit(1)
	}
}

func showMigrationStatus() error {
	// Check if tables exist
	tables := []string{"users", "teams", "servers", "applications", "schema_hashes"}

	for _, table := range tables {
		if config.DB.Migrator().HasTable(table) {
			fmt.Printf("✅ Table '%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%s' missing\n", table)
		}
	}

	// Show schema hash info
	fmt.Println("\n🔍 Schema Hash Information:")
	schemaInfo, err := config.GetSchemaInfo()
	if err != nil {
		fmt.Printf("❌ Failed to get schema info: %v\n", err)
		return nil
	}

	if schemaInfo["is_valid"].(bool) {
		fmt.Printf("✅ Schema hash is valid: %s\n", schemaInfo["current_hash"])
	} else {
		fmt.Printf("⚠️  Schema hash mismatch detected\n")
		fmt.Printf("   Expected: %s\n", schemaInfo["expected_hash"])
		if schemaInfo["current_hash"] != nil {
			fmt.Printf("   Current:  %s\n", schemaInfo["current_hash"])
		} else {
			fmt.Printf("   Current:  None (needs migration)\n")
		}
	}

	return nil
}

func showSchemaInfo() error {
	schemaInfo, err := config.GetSchemaInfo()
	if err != nil {
		return fmt.Errorf("failed to get schema info: %v", err)
	}

	// Pretty print JSON
	jsonData, err := json.MarshalIndent(schemaInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema info: %v", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func rollbackMigrations(steps int) error {
	// For now, this is a placeholder
	// In a production system, you'd want to implement proper migration versioning
	fmt.Println("⚠️  Rollback functionality not yet implemented")
	fmt.Println("   Use 'docker-compose down' and remove volumes to reset database")
	return nil
}
