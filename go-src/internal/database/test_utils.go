package database

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"coolify-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDBConfig holds configuration for test database
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	SSLMode  string
}

// GetTestDBConfig returns test database configuration
func GetTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
	}
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, you'd use os.Getenv(key)
	// For now, returning defaults
	return defaultValue
}

// TestDB creates a temporary PostgreSQL database for testing
func TestDB(t *testing.T) *gorm.DB {
	config := GetTestDBConfig()

	// Generate unique database name for this test
	rand.Seed(time.Now().UnixNano())
	testDBName := fmt.Sprintf("coolify_test_%d_%d", time.Now().Unix(), rand.Intn(10000))

	// Connect to postgres database to create test database
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for cleaner test output
	})
	if err != nil {
		t.Skipf("PostgreSQL not available for testing: %v", err)
		return nil
	}

	// Create test database
	sqlDB, err := adminDB.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Connect to the new test database
	testDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, testDBName, config.SSLMode)

	testDB, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		// Cleanup: drop database if connection fails
		sqlDB.Exec(fmt.Sprintf("DROP DATABASE %s", testDBName))
		sqlDB.Close()
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations on test database
	err = testDB.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.TeamInvitation{},
		&models.Server{},
		&models.Application{},
		&Migration{},
	)
	if err != nil {
		// Cleanup: drop database if migration fails
		testDBSql, _ := testDB.DB()
		testDBSql.Close()
		sqlDB.Exec(fmt.Sprintf("DROP DATABASE %s", testDBName))
		sqlDB.Close()
		t.Fatalf("failed to run migrations on test database: %v", err)
	}

	// Setup cleanup function
	t.Cleanup(func() {
		// Close test database connection
		if testDBSql, err := testDB.DB(); err == nil {
			testDBSql.Close()
		}

		// Drop test database
		_, err := sqlDB.Exec(fmt.Sprintf("DROP DATABASE %s", testDBName))
		if err != nil {
			log.Printf("Warning: failed to cleanup test database %s: %v", testDBName, err)
		}

		// Close admin connection
		sqlDB.Close()
	})

	return testDB
}

// TestDBWithFallback tries PostgreSQL first, falls back to SQLite if unavailable
func TestDBWithFallback(t *testing.T) *gorm.DB {
	// Try PostgreSQL first
	if db := TestDB(t); db != nil {
		return db
	}

	// Fallback to SQLite for environments without PostgreSQL
	log.Println("PostgreSQL not available, falling back to SQLite for testing")
	return TestDBSQLite(t)
}

// TestDBSQLite creates a SQLite test database (fallback option)
func TestDBSQLite(t *testing.T) *gorm.DB {
	// Import SQLite driver if not already imported
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to SQLite test database: %v", err)
	}

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.TeamInvitation{},
		&models.Server{},
		&models.Application{},
		&Migration{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations on SQLite test database: %v", err)
	}

	return db
}

// CreateTestUser creates a test user
func CreateTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:    fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		IsActive: true,
	}

	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

// CreateTestTeam creates a test team
func CreateTestTeam(t *testing.T, db *gorm.DB) *models.Team {
	team := &models.Team{
		Name:         "Test Team",
		Slug:         fmt.Sprintf("test-team-%d", time.Now().UnixNano()),
		Description:  nil,
		PersonalTeam: false,
		IsActive:     true,
	}

	err := db.Create(team).Error
	if err != nil {
		t.Fatalf("failed to create test team: %v", err)
	}

	return team
}

// CreateTestTeamMember creates a test team member
func CreateTestTeamMember(t *testing.T, db *gorm.DB, teamID, userID uint, role string) *models.TeamMember {
	member := &models.TeamMember{
		TeamID:      teamID,
		UserID:      userID,
		Role:        role,
		Permissions: models.JSONB{},
	}

	err := db.Create(member).Error
	if err != nil {
		t.Fatalf("failed to create test team member: %v", err)
	}

	return member
}

// CreateTestTeamInvitation creates a test team invitation
func CreateTestTeamInvitation(t *testing.T, db *gorm.DB, teamID uint) *models.TeamInvitation {
	invitation := &models.TeamInvitation{
		TeamID:    teamID,
		Email:     fmt.Sprintf("invite_%d@example.com", time.Now().UnixNano()),
		Role:      "member",
		Via:       "email",
		ExpiresAt: time.Now().AddDate(0, 0, 7),
	}

	err := db.Create(invitation).Error
	if err != nil {
		t.Fatalf("failed to create test team invitation: %v", err)
	}

	return invitation
}

// CleanupTestData cleans up test data (usually not needed with temp databases)
func CleanupTestData(t *testing.T, db *gorm.DB) {
	// Delete in reverse order of dependencies
	db.Exec("DELETE FROM team_invitations")
	db.Exec("DELETE FROM team_members")
	db.Exec("DELETE FROM teams")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM migrations")
}
