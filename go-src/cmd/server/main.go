package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"coolify-go/internal/config"
	"coolify-go/internal/handlers"
	"coolify-go/internal/templates"
	"coolify-go/pkg/database"
	"coolify-go/pkg/logging"

	"html/template"

	"github.com/gin-gonic/gin"
)

var (
	Version   = "v1.4.1"
	BuildTime = "2025-06-21T18:30:00Z"
	GitCommit = "a2624e7"
)

func main() {
	// Load config first (from current directory)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Set up logger
	logger := logging.New(cfg.Logging.Level, cfg.Logging.Format)
	logger.Infof("Starting Coolify Go v%s (commit %s, built %s)", Version, GitCommit, BuildTime)

	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Set up database (optional for development)
	var db *database.Database
	if cfg.Database.Host != "localhost" || isDatabaseAvailable(&cfg.Database) {
		db, err = database.New(&cfg.Database)
		if err != nil {
			logger.WithError(err).Warn("Database connection failed. Continuing without DB.")
		} else {
			logger.Info("Database connected successfully")
			if err := db.Migrate(); err != nil {
				logger.WithError(err).Warn("Database migration failed")
			}
		}
	} else {
		logger.Info("Database not available or configured for localhost. Continuing without DB.")
	}

	// Set up template renderer (with fallback for missing templates)
	var renderer *templates.Renderer
	if templatesExist() {
		renderer = templates.NewRenderer()
	} else {
		logger.Warn("Templates not found. Using basic renderer.")
		renderer = templates.NewBasicRenderer()
	}

	// Set up Gin router
	r := gin.New()
	r.Use(gin.Recovery())

	// Load HTML templates only if they exist
	if templatesExist() {
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

		// Set template functions and load templates
		r.SetFuncMap(funcMap)
		r.LoadHTMLGlob("internal/templates/**/*.html")
	}

	// Logging middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.WithFields(map[string]interface{}{
			"status":  c.Writer.Status(),
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"latency": latency.String(),
		}).Info("request completed")
	})

	// Set up web handlers
	webHandler := handlers.NewWebHandler(renderer)

	// Web routes
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})
	r.GET("/dashboard", webHandler.Dashboard)
	r.GET("/applications", webHandler.Applications)
	r.GET("/login", webHandler.Login)
	r.GET("/register", webHandler.Register)

	// Registration POST handler
	r.POST("/auth/register", webHandler.RegisterPost)

	// API routes
	r.GET("/health", func(c *gin.Context) {
		dbStatus := "not_initialized"
		if db != nil {
			if err := db.Ping(); err == nil {
				dbStatus = "connected"
			} else {
				dbStatus = "disconnected"
			}
		}
		c.JSON(200, gin.H{
			"status":    "healthy",
			"version":   Version,
			"buildTime": BuildTime,
			"commit":    GitCommit,
			"timestamp": time.Now(),
			"database":  dbStatus,
		})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":   Version,
			"buildTime": BuildTime,
			"commit":    GitCommit,
		})
	})

	// TODO: Add authentication, team, app, server, and deployment endpoints

	// Start server with graceful shutdown
	srv := &gin.Engine{}
	*srv = *r
	server := &httpServer{
		Engine: srv,
		Addr:   fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Logger: logger,
	}
	server.RunWithGracefulShutdown()
}

type httpServer struct {
	Engine *gin.Engine
	Addr   string
	Logger *logging.Logger
}

func (s *httpServer) RunWithGracefulShutdown() {
	srv := &http.Server{
		Addr:    s.Addr,
		Handler: s.Engine,
	}

	// Listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.Logger.Infof("🚀 Server running at http://%s", s.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.WithError(err).Fatal("Server error")
		}
	}()

	<-quit
	s.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.Logger.WithError(err).Error("Server forced to shutdown")
	}

	s.Logger.Info("Server exited cleanly")
}

// Helper functions

// isDatabaseAvailable checks if database is available for connection
func isDatabaseAvailable(cfg *config.DatabaseConfig) bool {
	// For now, just return true if it's not localhost
	// In a real implementation, you might want to ping the database
	return cfg.Host != "localhost"
}

// templatesExist checks if template files exist
func templatesExist() bool {
	// Check if the templates directory exists
	if _, err := os.Stat("internal/templates"); os.IsNotExist(err) {
		return false
	}

	// Check if at least one of the expected template files exists
	expectedFiles := []string{
		"internal/templates/layouts/base.html",
		"internal/templates/pages/dashboard.html",
		"internal/templates/pages/applications.html",
		"internal/templates/pages/login.html",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	return false
}
