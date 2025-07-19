package database

import (
	"log"

	"coolify-go/internal/config"
	"coolify-go/internal/models"
)

// SeedDatabase creates initial data for development
func SeedDatabase() error {
	if config.DB == nil {
		return nil
	}

	// Create a mock user for development
	var user models.User
	result := config.DB.Where("email = ?", "admin@coolify.local").First(&user)
	if result.Error != nil {
		// User doesn't exist, create it
		user = models.User{
			Email:    "admin@coolify.local",
			Name:     "Admin User",
			Password: "hashed_password", // In production, this would be properly hashed
			Role:     "admin",
			IsActive: true,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			return err
		}
		log.Printf("✅ Created mock user: %s", user.Email)
	}

	return nil
}
