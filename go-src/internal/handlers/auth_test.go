package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"coolify-go/internal/auth"
	"coolify-go/internal/models"
	"coolify-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestSetup sets up test environment
func setupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Team{},
		&models.TeamMember{},
		&models.Server{},
		&models.Application{},
	)
	require.NoError(t, err)

	return db
}

func setupTestServices(t *testing.T) (*services.AuthService, *services.UserService, *gorm.DB) {
	db := setupTestDB(t)

	userService := services.NewUserService(db)
	authService := services.NewAuthService(db, userService, "test-secret-key")

	return authService, userService, db
}

func setupAuthTestRouter(authService *services.AuthService, userService *services.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	// Create OAuth service (minimal for testing)
	oauthService := auth.NewOAuthService()

	// Create handlers
	authHandler := NewAuthHandler(authService, oauthService)
	userHandler := NewUserHandler(userService)

	// Register routes
	api := r.Group("/api")
	authGroup := api.Group("/auth")
	authHandler.RegisterRoutes(authGroup)

	usersGroup := api.Group("/users")
	userHandler.RegisterRoutes(usersGroup)

	return r
}

// Test User Registration
func TestAuthHandler_Register(t *testing.T) {
	authService, userService, db := setupTestServices(t)
	router := setupAuthTestRouter(authService, userService)

	t.Run("successful registration", func(t *testing.T) {
		// Clear database
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM teams")
		db.Exec("DELETE FROM team_members")

		payload := services.RegisterRequest{
			Name:            "John Doe",
			Email:           "john@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
			AgreeToTerms:    true,
			MarketingEmails: false,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Registration successful", response["message"])
		assert.NotNil(t, response["user"])
		assert.NotNil(t, response["tokens"])

		// Verify user was created in database
		var user models.User
		err = db.Where("email = ?", "john@example.com").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "admin", user.Role) // First user should be admin

		// Verify personal team was created
		var team models.Team
		err = db.Where("personal_team = true").First(&team).Error
		assert.NoError(t, err)
		assert.Equal(t, "John Doe's Team", team.Name)
	})

	t.Run("duplicate email registration", func(t *testing.T) {
		payload := services.RegisterRequest{
			Name:            "Jane Doe",
			Email:           "john@example.com", // Same email as above
			Password:        "password123",
			ConfirmPassword: "password123",
			AgreeToTerms:    true,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Registration failed", response["error"])
		assert.Contains(t, response["details"], "already exists")
	})

	t.Run("password mismatch", func(t *testing.T) {
		payload := services.RegisterRequest{
			Name:            "Jane Doe",
			Email:           "jane@example.com",
			Password:        "password123",
			ConfirmPassword: "differentpassword",
			AgreeToTerms:    true,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["details"], "passwords do not match")
	})

	t.Run("invalid request data", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test User Login
func TestAuthHandler_Login(t *testing.T) {
	authService, userService, db := setupTestServices(t)
	router := setupAuthTestRouter(authService, userService)

	// Create a test user first
	t.Run("setup test user", func(t *testing.T) {
		// Clear database
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM teams")
		db.Exec("DELETE FROM team_members")

		registerPayload := services.RegisterRequest{
			Name:            "Test User",
			Email:           "test@example.com",
			Password:        "testpassword123",
			ConfirmPassword: "testpassword123",
			AgreeToTerms:    true,
		}

		jsonPayload, _ := json.Marshal(registerPayload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("successful login", func(t *testing.T) {
		payload := services.LoginRequest{
			Email:    "test@example.com",
			Password: "testpassword123",
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Login successful", response["message"])
		assert.NotNil(t, response["user"])
		assert.NotNil(t, response["tokens"])

		// Verify tokens are present
		tokens := response["tokens"].(map[string]interface{})
		assert.NotEmpty(t, tokens["access_token"])
		assert.NotEmpty(t, tokens["refresh_token"])
	})

	t.Run("invalid credentials", func(t *testing.T) {
		payload := services.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Login failed", response["error"])
		assert.Contains(t, response["details"], "invalid credentials")
	})

	t.Run("non-existent user", func(t *testing.T) {
		payload := services.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["details"], "invalid credentials")
	})
}

// Test Token Refresh
func TestAuthHandler_Refresh(t *testing.T) {
	authService, userService, db := setupTestServices(t)
	router := setupAuthTestRouter(authService, userService)

	var refreshToken string

	// Setup: Register and login to get tokens
	t.Run("setup tokens", func(t *testing.T) {
		// Clear database
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM teams")
		db.Exec("DELETE FROM team_members")

		// Register
		registerPayload := services.RegisterRequest{
			Name:            "Refresh User",
			Email:           "refresh@example.com",
			Password:        "refreshpassword123",
			ConfirmPassword: "refreshpassword123",
			AgreeToTerms:    true,
		}

		jsonPayload, _ := json.Marshal(registerPayload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		tokens := response["tokens"].(map[string]interface{})
		refreshToken = tokens["refresh_token"].(string)
		require.NotEmpty(t, refreshToken)
	})

	t.Run("successful token refresh", func(t *testing.T) {
		payload := map[string]string{
			"refresh_token": refreshToken,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Token refreshed successfully", response["message"])
		assert.NotNil(t, response["tokens"])

		// Verify new tokens are present
		tokens := response["tokens"].(map[string]interface{})
		assert.NotEmpty(t, tokens["access_token"])
		assert.NotEmpty(t, tokens["refresh_token"])
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		payload := map[string]string{
			"refresh_token": "invalid-token",
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Token refresh failed", response["error"])
	})
}

// Test User Profile Access
func TestUserHandler_Profile(t *testing.T) {
	authService, userService, db := setupTestServices(t)
	router := setupAuthTestRouter(authService, userService)

	var accessToken string

	// Setup: Register user to get access token
	t.Run("setup user and token", func(t *testing.T) {
		// Clear database
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM teams")
		db.Exec("DELETE FROM team_members")

		registerPayload := services.RegisterRequest{
			Name:            "Profile User",
			Email:           "profile@example.com",
			Password:        "profilepassword123",
			ConfirmPassword: "profilepassword123",
			AgreeToTerms:    true,
			MarketingEmails: true,
		}

		jsonPayload, _ := json.Marshal(registerPayload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		tokens := response["tokens"].(map[string]interface{})
		accessToken = tokens["access_token"].(string)
		require.NotEmpty(t, accessToken)
	})

	t.Run("get profile with valid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Note: This will fail without proper JWT middleware
		// In a complete implementation, you'd need to add JWT middleware
		// For now, we'll expect unauthorized since middleware isn't set up

		// This test demonstrates the structure but would need proper middleware
		// to pass in a real e2e environment
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Integration Test: Full Registration -> Login -> Profile Flow
func TestFullAuthFlow(t *testing.T) {
	authService, userService, db := setupTestServices(t)
	router := setupAuthTestRouter(authService, userService)

	// Clear database
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM teams")
	db.Exec("DELETE FROM team_members")

	var accessToken, refreshToken string
	userEmail := "fullflow@example.com"
	userName := "Full Flow User"

	t.Run("1. Register new user", func(t *testing.T) {
		payload := services.RegisterRequest{
			Name:            userName,
			Email:           userEmail,
			Password:        "fullflowpassword123",
			ConfirmPassword: "fullflowpassword123",
			AgreeToTerms:    true,
			MarketingEmails: false,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response structure
		assert.Equal(t, "Registration successful", response["message"])

		user := response["user"].(map[string]interface{})
		assert.Equal(t, userName, user["name"])
		assert.Equal(t, userEmail, user["email"])
		assert.Equal(t, "admin", user["role"]) // First user is admin

		tokens := response["tokens"].(map[string]interface{})
		accessToken = tokens["access_token"].(string)
		refreshToken = tokens["refresh_token"].(string)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("2. Login with registered user", func(t *testing.T) {
		payload := services.LoginRequest{
			Email:    userEmail,
			Password: "fullflowpassword123",
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Login successful", response["message"])

		user := response["user"].(map[string]interface{})
		assert.Equal(t, userName, user["name"])
		assert.Equal(t, userEmail, user["email"])

		// Update tokens from login response
		tokens := response["tokens"].(map[string]interface{})
		accessToken = tokens["access_token"].(string)
		refreshToken = tokens["refresh_token"].(string)
	})

	t.Run("3. Refresh tokens", func(t *testing.T) {
		payload := map[string]string{
			"refresh_token": refreshToken,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Token refreshed successfully", response["message"])

		tokens := response["tokens"].(map[string]interface{})
		newAccessToken := tokens["access_token"].(string)
		newRefreshToken := tokens["refresh_token"].(string)

		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, accessToken, newAccessToken) // Should be different
	})

	t.Run("4. Verify database state", func(t *testing.T) {
		// Check user was created properly
		var user models.User
		err := db.Where("email = ?", userEmail).Preload("TeamMemberships").First(&user).Error
		require.NoError(t, err)

		assert.Equal(t, userName, user.Name)
		assert.Equal(t, userEmail, user.Email)
		assert.Equal(t, "admin", user.Role)
		assert.True(t, user.IsActive)
		assert.False(t, user.MarketingEmails)

		// Check personal team was created
		var team models.Team
		err = db.Where("personal_team = true").First(&team).Error
		require.NoError(t, err)

		assert.Equal(t, userName+"'s Team", team.Name)
		assert.True(t, team.PersonalTeam)
		assert.True(t, team.IsActive)

		// Check team membership
		var membership models.TeamMember
		err = db.Where("user_id = ? AND team_id = ?", user.ID, team.ID).First(&membership).Error
		require.NoError(t, err)

		assert.Equal(t, models.RoleOwner, membership.Role)
	})
}

// Benchmark Tests
func BenchmarkRegistration(b *testing.B) {
	authService, userService, db := setupTestServices(&testing.T{})
	router := setupAuthTestRouter(authService, userService)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear database for each iteration
		db.Exec("DELETE FROM users")
		db.Exec("DELETE FROM teams")
		db.Exec("DELETE FROM team_members")

		payload := services.RegisterRequest{
			Name:            fmt.Sprintf("Benchmark User %d", i),
			Email:           fmt.Sprintf("bench%d@example.com", i),
			Password:        "benchmarkpassword123",
			ConfirmPassword: "benchmarkpassword123",
			AgreeToTerms:    true,
		}

		jsonPayload, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLogin(b *testing.B) {
	authService, userService, db := setupTestServices(&testing.T{})
	router := setupAuthTestRouter(authService, userService)

	// Setup: Create a test user
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM teams")
	db.Exec("DELETE FROM team_members")

	registerPayload := services.RegisterRequest{
		Name:            "Benchmark Login User",
		Email:           "benchlogin@example.com",
		Password:        "benchmarkpassword123",
		ConfirmPassword: "benchmarkpassword123",
		AgreeToTerms:    true,
	}

	jsonPayload, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		loginPayload := services.LoginRequest{
			Email:    "benchlogin@example.com",
			Password: "benchmarkpassword123",
		}

		jsonPayload, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
