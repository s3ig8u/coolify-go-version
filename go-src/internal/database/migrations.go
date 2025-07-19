package database

import (
	"fmt"
	"log"
	"time"

	"coolify-go/internal/models"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex"`
	Version     string    `json:"version" gorm:"not null"`
	AppliedAt   time.Time `json:"applied_at" gorm:"default:CURRENT_TIMESTAMP"`
	Description string    `json:"description"`
}

// TableName specifies the table name for Migration
func (Migration) TableName() string {
	return "migrations"
}

// RunMigrations executes all pending database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Create migrations table if it doesn't exist
	err := db.AutoMigrate(&Migration{})
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Define migrations in order
	migrations := []struct {
		name        string
		version     string
		description string
		execute     func(*gorm.DB) error
	}{
		{
			name:        "001_create_team_members_table",
			version:     "v1.4.0",
			description: "Create team_members table for role-based team membership",
			execute:     createTeamMembersTable,
		},
		{
			name:        "002_create_team_invitations_table",
			version:     "v1.4.0",
			description: "Create team_invitations table for team invitation system",
			execute:     createTeamInvitationsTable,
		},
		{
			name:        "003_update_teams_table",
			version:     "v1.4.0",
			description: "Add personal_team and custom_server_limit columns to teams table",
			execute:     updateTeamsTable,
		},
		{
			name:        "004_update_users_table",
			version:     "v1.4.0",
			description: "Add current_team_id column to users table",
			execute:     updateUsersTable,
		},
	}

	// Execute each migration
	for _, migration := range migrations {
		// Check if migration already applied
		var existingMigration Migration
		result := db.Where("name = ?", migration.name).First(&existingMigration)

		if result.Error == gorm.ErrRecordNotFound {
			// Migration not applied, execute it
			log.Printf("Applying migration: %s", migration.name)

			err := migration.execute(db)
			if err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.name, err)
			}

			// Record migration as applied
			record := Migration{
				Name:        migration.name,
				Version:     migration.version,
				Description: migration.description,
				AppliedAt:   time.Now(),
			}

			err = db.Create(&record).Error
			if err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.name, err)
			}

			log.Printf("Migration %s applied successfully", migration.name)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check migration %s: %w", migration.name, result.Error)
		} else {
			log.Printf("Migration %s already applied", migration.name)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// createTeamMembersTable creates the team_members table
func createTeamMembersTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.TeamMember{})
}

// createTeamInvitationsTable creates the team_invitations table
func createTeamInvitationsTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.TeamInvitation{})
}

// updateTeamsTable adds new columns to the teams table
func updateTeamsTable(db *gorm.DB) error {
	// Add personal_team column if it doesn't exist
	if !db.Migrator().HasColumn(&models.Team{}, "personal_team") {
		err := db.Exec("ALTER TABLE teams ADD COLUMN personal_team BOOLEAN DEFAULT FALSE").Error
		if err != nil {
			return fmt.Errorf("failed to add personal_team column: %w", err)
		}
	}

	// Add custom_server_limit column if it doesn't exist
	if !db.Migrator().HasColumn(&models.Team{}, "custom_server_limit") {
		err := db.Exec("ALTER TABLE teams ADD COLUMN custom_server_limit INTEGER").Error
		if err != nil {
			return fmt.Errorf("failed to add custom_server_limit column: %w", err)
		}
	}

	// Add show_boarding column if it doesn't exist
	if !db.Migrator().HasColumn(&models.Team{}, "show_boarding") {
		err := db.Exec("ALTER TABLE teams ADD COLUMN show_boarding BOOLEAN DEFAULT TRUE").Error
		if err != nil {
			return fmt.Errorf("failed to add show_boarding column: %w", err)
		}
	}

	return nil
}

// updateUsersTable adds new columns to the users table
func updateUsersTable(db *gorm.DB) error {
	// Add current_team_id column if it doesn't exist
	if !db.Migrator().HasColumn(&models.User{}, "current_team_id") {
		err := db.Exec("ALTER TABLE users ADD COLUMN current_team_id INTEGER REFERENCES teams(id)").Error
		if err != nil {
			return fmt.Errorf("failed to add current_team_id column: %w", err)
		}
	}

	return nil
}
