package config

import (
	"fmt"
	"log"
	"time"

	"coolify-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase establishes connection to the database using the provided config
func ConnectDatabase(dbConfig DatabaseConfig) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(dbConfig.MaxConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
	return nil
}

// AutoMigrate runs database migrations with hash validation
func AutoMigrate() error {
	log.Println("Running database migrations...")

	// Run GORM auto-migration for all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.TeamInvitation{},
		&models.Server{},
		&models.Application{},
		&models.SchemaHash{},
	)

	if err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	// Validate schema hash
	isValid, expectedHash, err := models.ValidateSchemaHash(DB)
	if err != nil {
		return fmt.Errorf("failed to validate schema hash: %w", err)
	}

	if !isValid {
		log.Printf("Schema hash mismatch detected. Expected: %s", expectedHash)

		// Save new schema hash
		modelNames := []string{"User", "Team", "TeamMember", "TeamInvitation", "Server", "Application", "SchemaHash"}
		err = models.SaveSchemaHash(DB, expectedHash, "v1.4.0", modelNames)
		if err != nil {
			return fmt.Errorf("failed to save schema hash: %w", err)
		}

		log.Printf("New schema hash saved: %s", expectedHash)
	} else {
		log.Println("Schema hash validation passed - no changes needed")
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetSchemaInfo returns current schema information
func GetSchemaInfo() (map[string]interface{}, error) {
	currentHash, err := models.GetCurrentSchemaHash(DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get current schema hash: %w", err)
	}

	expectedHash, err := models.GenerateSchemaHash()
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected hash: %w", err)
	}

	info := map[string]interface{}{
		"expected_hash": expectedHash,
		"is_valid":      false,
	}

	if currentHash != nil {
		info["current_hash"] = currentHash.Hash
		info["current_version"] = currentHash.Version
		info["applied_at"] = currentHash.AppliedAt
		info["is_valid"] = currentHash.Hash == expectedHash
	}

	return info, nil
}
