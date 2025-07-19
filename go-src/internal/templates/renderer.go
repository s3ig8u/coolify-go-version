package templates

import (
	"html/template"
	"net/http"
	"strings"
)

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
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

	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("internal/templates/**/*.html"))
	return &Renderer{
		templates: templates,
	}
}

// NewBasicRenderer creates a basic template renderer without loading templates
func NewBasicRenderer() *Renderer {
	return &Renderer{
		templates: template.New("basic"),
	}
}

// Render renders a template with data
func (r *Renderer) Render(w http.ResponseWriter, templateName string, data interface{}) error {
	return r.templates.ExecuteTemplate(w, templateName, data)
}

// PageData represents common data for all pages
type PageData struct {
	Title      string
	ActivePage string
	User       *User
	Error      string
	Success    string
	Data       interface{}
}

// User represents the current user
type User struct {
	ID    string
	Name  string
	Email string
	Role  string
}

// DashboardData represents dashboard page data
type DashboardData struct {
	PageData
	Stats             DashboardStats
	RecentDeployments []Deployment
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	Applications int
	Servers      int
	Deployments  int
	TeamMembers  int
}

// Deployment represents a deployment
type Deployment struct {
	ID              string
	ApplicationName string
	Status          string
	CommitHash      string
	CreatedAt       string
}

// ApplicationData represents applications page data
type ApplicationData struct {
	PageData
	Applications []Application
}

// Application represents an application
type Application struct {
	ID         string
	Name       string
	Status     string
	Domain     string
	LastDeploy string
	Repository string
}

// LoginData represents login page data
type LoginData struct {
	PageData
	Providers []OAuthProvider
}

// OAuthProvider represents an OAuth provider
type OAuthProvider struct {
	Name string
	URL  string
	Icon string
}
