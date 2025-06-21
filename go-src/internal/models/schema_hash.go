package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SchemaHash represents the database schema version
type SchemaHash struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Hash      string         `json:"hash" gorm:"uniqueIndex;not null"`
	Version   string         `json:"version" gorm:"not null"`
	Models    string         `json:"models" gorm:"not null"` // JSON array of model names
	AppliedAt time.Time      `json:"applied_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for SchemaHash
func (SchemaHash) TableName() string {
	return "schema_hashes"
}

// GenerateSchemaHash creates a hash from all model definitions
func GenerateSchemaHash() (string, error) {
	// Get all model names in a consistent order
	modelNames := []string{
		"User",
		"Team",
		"Server",
		"Application",
		"SchemaHash",
	}

	sort.Strings(modelNames)

	// Create a hash from model names and their field definitions
	var hashInput strings.Builder

	for _, modelName := range modelNames {
		hashInput.WriteString(modelName)
		hashInput.WriteString(":")

		// Add field definitions for each model
		switch modelName {
		case "User":
			hashInput.WriteString("ID(uint),Email(string),Password(string),Name(string),Role(string),IsActive(bool),LastLogin(*time.Time),CreatedAt(time.Time),UpdatedAt(time.Time),DeletedAt(gorm.DeletedAt)")
		case "Team":
			hashInput.WriteString("ID(uint),Name(string),Description(*string),Slug(string),IsActive(bool),CreatedAt(time.Time),UpdatedAt(time.Time),DeletedAt(gorm.DeletedAt)")
		case "Server":
			hashInput.WriteString("ID(uint),Name(string),Description(*string),Host(string),Port(int),Username(string),SSHKey(string),Status(string),Type(string),Provider(*string),IsActive(bool),CreatedAt(time.Time),UpdatedAt(time.Time),DeletedAt(gorm.DeletedAt)")
		case "Application":
			hashInput.WriteString("ID(uint),Name(string),Description(*string),Slug(string),Type(string),Status(string),Port(int),Domain(*string),GitURL(*string),GitBranch(string),BuildPack(*string),IsActive(bool),UserID(uint),TeamID(uint),ServerID(uint),CreatedAt(time.Time),UpdatedAt(time.Time),DeletedAt(gorm.DeletedAt)")
		case "SchemaHash":
			hashInput.WriteString("ID(uint),Hash(string),Version(string),Models(string),AppliedAt(time.Time),CreatedAt(time.Time),UpdatedAt(time.Time),DeletedAt(gorm.DeletedAt)")
		}
		hashInput.WriteString(";")
	}

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(hashInput.String()))
	return hex.EncodeToString(hash[:]), nil
}

// GetCurrentSchemaHash retrieves the latest schema hash from database
func GetCurrentSchemaHash(db *gorm.DB) (*SchemaHash, error) {
	var schemaHash SchemaHash
	err := db.Order("applied_at DESC").First(&schemaHash).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No schema hash found
		}
		return nil, err
	}
	return &schemaHash, nil
}

// SaveSchemaHash saves a new schema hash to the database
func SaveSchemaHash(db *gorm.DB, hash, version string, models []string) error {
	modelsJSON := fmt.Sprintf(`["%s"]`, strings.Join(models, `","`))

	schemaHash := SchemaHash{
		Hash:      hash,
		Version:   version,
		Models:    modelsJSON,
		AppliedAt: time.Now(),
	}

	return db.Create(&schemaHash).Error
}

// ValidateSchemaHash checks if the current database schema matches the expected hash
func ValidateSchemaHash(db *gorm.DB) (bool, string, error) {
	expectedHash, err := GenerateSchemaHash()
	if err != nil {
		return false, "", fmt.Errorf("failed to generate expected hash: %w", err)
	}

	currentHash, err := GetCurrentSchemaHash(db)
	if err != nil {
		return false, "", fmt.Errorf("failed to get current hash: %w", err)
	}

	if currentHash == nil {
		return false, expectedHash, nil // No hash found, needs migration
	}

	return currentHash.Hash == expectedHash, expectedHash, nil
}
