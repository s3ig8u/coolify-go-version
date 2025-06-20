package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Port   string
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	BuildTime string    `json:"buildTime"`
	Commit    string    `json:"commit"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
}

var db *sql.DB
var config Config

func loadConfig() Config {
	return Config{
		Port:   getEnv("APP_PORT", "8080"),
		DBHost: getEnv("DB_HOST", "postgres"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBName: getEnv("DB_NAME", "coolify_go"),
		DBUser: getEnv("DB_USER", "coolify_go"),
		DBPass: getEnv("DB_PASSWORD", "changeme"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDatabase() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPass, config.DBName)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	return db.Ping()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "connected"
	if db == nil {
		dbStatus = "not_initialized"
	} else if err := db.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	response := HealthResponse{
		Status:    "healthy",
		Version:   Version,
		BuildTime: BuildTime,
		Commit:    GitCommit,
		Timestamp: time.Now(),
		Database:  dbStatus,
	}

	json.NewEncoder(w).Encode(response)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
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
            <li>✅ Basic HTTP server</li>
            <li>✅ Health checks</li>
            <li>✅ Database connectivity</li>
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
</html>`, Version, BuildTime)

	fmt.Fprint(w, html)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"version":   Version,
		"buildTime": BuildTime,
		"commit":    GitCommit,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	config = loadConfig()

	// Try to connect to database
	if err := initDatabase(); err != nil {
		log.Printf("Database connection failed: %v", err)
		log.Printf("Continuing without database...")
	} else {
		log.Printf("Database connected successfully")
	}

	// Setup routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/version", versionHandler)

	port := ":" + config.Port
	fmt.Printf("🚀 Coolify Go v%s server running at http://localhost%s\n", Version, port)
	fmt.Printf("📊 Health check: http://localhost%s/health\n", port)

	log.Fatal(http.ListenAndServe(port, nil))
}
