package handlers

import (
	"net/http"

	"coolify-go/internal/templates"

	"github.com/gin-gonic/gin"
)

// WebHandler handles web page requests
type WebHandler struct {
	renderer *templates.Renderer
}

// NewWebHandler creates a new web handler
func NewWebHandler(renderer *templates.Renderer) *WebHandler {
	return &WebHandler{
		renderer: renderer,
	}
}

// Dashboard renders the dashboard page
func (h *WebHandler) Dashboard(c *gin.Context) {
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

// Applications renders the applications page
func (h *WebHandler) Applications(c *gin.Context) {
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

// Login renders the login page
func (h *WebHandler) Login(c *gin.Context) {
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

// Register renders the register page
func (h *WebHandler) Register(c *gin.Context) {
	data := templates.LoginData{
		PageData: templates.PageData{
			Title: "Register",
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

// RegisterPost handles POST /auth/register
func (h *WebHandler) RegisterPost(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirm := c.PostForm("confirm-password")
	agree := c.PostForm("agree-terms")

	errors := make(map[string]string)
	if name == "" {
		errors["name"] = "Full name is required."
	}
	if email == "" {
		errors["email"] = "Email is required."
	}
	if password == "" {
		errors["password"] = "Password is required."
	}
	if confirm == "" {
		errors["confirm-password"] = "Please confirm your password."
	}
	if password != confirm {
		errors["confirm-password"] = "Passwords do not match."
	}
	if agree != "on" {
		errors["agree-terms"] = "You must agree to the terms."
	}

	if len(errors) > 0 {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Title":      "Register",
			"ActivePage": "register",
			"Errors":     errors,
			"Form": gin.H{
				"name":  name,
				"email": email,
			},
		})
		return
	}

	// Simulate user creation (in-memory or log)
	c.HTML(http.StatusOK, "base.html", gin.H{
		"Title":      "Register Success",
		"ActivePage": "register",
		"Success":    true,
		"Form": gin.H{
			"name":  name,
			"email": email,
		},
	})
}
