package app

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"coolify-go/internal/config"
	"coolify-go/internal/docker"
	"coolify-go/internal/handlers"
	"coolify-go/internal/middleware"
	"coolify-go/internal/services"
	"coolify-go/internal/templates"

	"github.com/gin-gonic/gin"
)

// App represents the main application
type App struct {
	Config            *config.Config
	Router            *gin.Engine
	TeamService       *services.TeamService
	InvitationService *services.InvitationService
	DockerClient      *docker.Client
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	app := &App{
		Config: cfg,
	}

	// Initialize services
	app.initServices()

	// Initialize Docker client
	app.initDockerClient()

	// Initialize router
	app.initRouter()

	return app
}

// initServices initializes all application services
func (a *App) initServices() {
	// Initialize team services
	a.TeamService = services.NewTeamService(config.DB)
	a.InvitationService = services.NewInvitationService(config.DB)
}

// initDockerClient initializes the Docker client
func (a *App) initDockerClient() {
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Printf("⚠️  Docker client initialization failed: %v", err)
		log.Printf("   Docker features will be disabled")
		a.DockerClient = nil
	} else {
		log.Printf("✅ Docker client initialized successfully")
		a.DockerClient = dockerClient
	}
}

// initRouter initializes the Gin router with all routes and middleware
func (a *App) initRouter() {
	// Set Gin mode based on environment
	// Default to debug mode for development
	gin.SetMode(gin.DebugMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Load HTML templates
	a.loadTemplates(router)

	// Serve static files (for CSS, JS, images)
	router.Static("/static", "./web/static")

	// Health check endpoints
	router.GET("/health", a.healthHandler)
	router.GET("/version", a.versionHandler)

	// Home page
	router.GET("/", a.homeHandler)

	// API routes
	api := router.Group("/api")
	{
		// Apply authentication middleware to API routes
		api.Use(middleware.MockAuth()) // Replace with real auth when available

		// Team API endpoints
		teamHandler := handlers.NewTeamHandler(a.TeamService, a.InvitationService)
		teamHandler.RegisterRoutes(api)

		// Docker API endpoints (if Docker client is available)
		if a.DockerClient != nil {
			dockerHandler := handlers.NewDockerHandler(a.DockerClient)
			dockerHandler.RegisterRoutes(api)
		}
	}

	// Web routes
	web := router.Group("")
	{
		// Apply authentication middleware to web routes
		web.Use(middleware.MockAuth()) // Replace with real auth when available

		// Main web pages - use Gin's template system directly
		web.GET("/dashboard", a.dashboardHandler)
		web.GET("/applications", a.applicationsHandler)
		web.GET("/login", a.loginHandler)

		// Team web pages
		webTeamHandler := handlers.NewWebTeamHandler(a.TeamService, a.InvitationService)
		webTeamHandler.RegisterWebRoutes(web)
	}

	a.Router = router
}

// loadTemplates loads HTML templates
func (a *App) loadTemplates(router *gin.Engine) {
	// Define custom template functions
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(string(s[0])) + s[1:]
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}

	// Load templates with custom functions
	router.SetFuncMap(funcMap)
	router.LoadHTMLGlob("internal/templates/**/*")
}

// Run starts the HTTP server
func (a *App) Run() error {
	port := ":" + a.Config.Server.Port
	log.Printf("🚀 Coolify Go server running at http://localhost%s", port)
	log.Printf("📊 Health check: http://localhost%s/health", port)
	log.Printf("👥 Teams: http://localhost%s/teams", port)
	log.Printf("🔧 API: http://localhost%s/api", port)

	if a.DockerClient != nil {
		log.Printf("🐳 Docker API: http://localhost%s/api/docker", port)
	}

	return a.Router.Run(port)
}

// Health check handler
func (a *App) healthHandler(c *gin.Context) {
	dbStatus := "disconnected"
	if config.DB != nil {
		sqlDB, err := config.DB.DB()
		if err == nil && sqlDB.Ping() == nil {
			dbStatus = "connected"
		}
	}

	dockerStatus := "disabled"
	if a.DockerClient != nil {
		if err := a.DockerClient.Ping(); err == nil {
			dockerStatus = "connected"
		} else {
			dockerStatus = "error"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   config.Version,
		"buildTime": config.BuildTime,
		"commit":    config.GitCommit,
		"database":  dbStatus,
		"docker":    dockerStatus,
		"features": gin.H{
			"teams":       "enabled",
			"invitations": "enabled",
			"api":         "enabled",
			"docker":      a.DockerClient != nil,
		},
	})
}

// Version handler
func (a *App) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":   config.Version,
		"buildTime": config.BuildTime,
		"commit":    config.GitCommit,
	})
}

// Home page handler
func (a *App) homeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title":     "Coolify Go",
		"version":   config.Version,
		"buildTime": config.BuildTime,
	})
}

// Dashboard handler
func (a *App) dashboardHandler(c *gin.Context) {
	data := templates.DashboardData{
		PageData: templates.PageData{
			Title:      "Dashboard",
			ActivePage: "dashboard",
			User: &templates.User{
				ID:    "1",
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  "admin",
			},
		},
		Stats: templates.DashboardStats{
			Applications: 5,
			Servers:      3,
			Deployments:  12,
			TeamMembers:  8,
		},
		RecentDeployments: []templates.Deployment{
			{
				ID:              "1",
				ApplicationName: "my-app",
				Status:          "success",
				CommitHash:      "abc123",
				CreatedAt:       "2 hours ago",
			},
			{
				ID:              "2",
				ApplicationName: "api-service",
				Status:          "success",
				CommitHash:      "def456",
				CreatedAt:       "4 hours ago",
			},
		},
	}

	c.HTML(http.StatusOK, "base.html", data)
}

// Applications handler
func (a *App) applicationsHandler(c *gin.Context) {
	data := templates.ApplicationData{
		PageData: templates.PageData{
			Title:      "Applications",
			ActivePage: "applications",
			User: &templates.User{
				ID:    "1",
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  "admin",
			},
		},
		Applications: []templates.Application{
			{
				ID:         "1",
				Name:       "my-app",
				Status:     "running",
				Domain:     "my-app.example.com",
				LastDeploy: "2 hours ago",
				Repository: "github.com/user/my-app",
			},
			{
				ID:         "2",
				Name:       "api-service",
				Status:     "running",
				Domain:     "api.example.com",
				LastDeploy: "4 hours ago",
				Repository: "github.com/user/api-service",
			},
		},
	}

	c.HTML(http.StatusOK, "base.html", data)
}

// Login handler
func (a *App) loginHandler(c *gin.Context) {
	data := templates.LoginData{
		PageData: templates.PageData{
			Title: "Login",
		},
		Providers: []templates.OAuthProvider{
			{
				Name: "GitHub",
				URL:  "/auth/github",
				Icon: "github",
			},
			{
				Name: "GitLab",
				URL:  "/auth/gitlab",
				Icon: "gitlab",
			},
		},
	}

	c.HTML(http.StatusOK, "base.html", data)
}
