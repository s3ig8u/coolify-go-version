package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coolify-go/internal/config"
	"coolify-go/pkg/database"
	"coolify-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

var (
	Version   = "v1.4.0"
	BuildTime = "2025-06-20T11:38:00Z"
	GitCommit = "azure-registry-v1.4.0"
)

func main() {
	// Load config
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

	// Set up database
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.WithError(err).Warn("Database connection failed. Continuing without DB.")
	} else {
		logger.Info("Database connected successfully")
		if err := db.Migrate(); err != nil {
			logger.WithError(err).Warn("Database migration failed")
		}
	}

	// Set up Gin router
	r := gin.New()
	r.Use(gin.Recovery())

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

	// Routes
	r.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html", []byte(homeHTML(Version, BuildTime)))
	})

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

func homeHTML(version, buildTime string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Coolify Go</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
        .header { text-align: center; margin-bottom: 2rem; }
        .status { background: #f0f9ff; border: 1px solid #0ea5e9; border-radius: 8px; padding: 1rem; margin: 1rem 0; }
        .feature { background: #fefce8; border: 1px solid #eab308; border-radius: 8px; padding: 1rem; margin: 1rem 0; }
        .code { background: #f1f5f9; padding: 0.5rem; border-radius: 4px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Coolify Go</h1>
        <p>Open-source deployment platform (Go port)</p>
        <p>Version: %s | Build: %s</p>
    </div>
    <div class="status">
        <h3>📊 System Status</h3>
        <p><strong>Server:</strong> Running</p>
        <p><strong>Database:</strong> <span id="db-status">Checking...</span></p>
        <p><strong>API:</strong> <a href="/health">/health</a></p>
    </div>
    <div class="feature">
        <h3>⚠️ Development Status</h3>
        <p>This is a minimal Go port of Coolify. Currently implemented:</p>
        <ul>
            <li>✅ Basic HTTP server (Gin)</li>
            <li>✅ Health checks</li>
            <li>✅ Database connectivity (GORM)</li>
            <li>⏳ User authentication (coming soon)</li>
            <li>⏳ Application deployment (coming soon)</li>
            <li>⏳ Server management (coming soon)</li>
        </ul>
    </div>
    <div class="status">
        <h3>🛠️ API Endpoints</h3>
        <p><span class="code">GET /</span> - This page</p>
        <p><span class="code">GET /health</span> - Health check</p>
        <p><span class="code">GET /version</span> - Version info</p>
    </div>
    <script>
        fetch('/health')
            .then(r => r.json())
            .then(data => {
                document.getElementById('db-status').textContent = data.database;
            })
            .catch(() => {
                document.getElementById('db-status').textContent = 'error';
            });
    </script>
</body>
</html>`, version, buildTime)
}
