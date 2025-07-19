package handlers

import (
	"net/http"
	"strconv"

	"coolify-go/internal/services"

	"github.com/gin-gonic/gin"
)

// WebTeamHandler handles web-based team management
type WebTeamHandler struct {
	teamService       *services.TeamService
	invitationService *services.InvitationService
}

// NewWebTeamHandler creates a new web team handler
func NewWebTeamHandler(teamService *services.TeamService, invitationService *services.InvitationService) *WebTeamHandler {
	return &WebTeamHandler{
		teamService:       teamService,
		invitationService: invitationService,
	}
}

// RegisterWebRoutes registers web team endpoints
func (h *WebTeamHandler) RegisterWebRoutes(r *gin.RouterGroup) {
	teams := r.Group("/teams")
	{
		teams.GET("", h.TeamsIndex)
		teams.GET("/:id", h.TeamDetail)
	}
}

// TeamsIndex displays the teams listing page
func (h *WebTeamHandler) TeamsIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "teams/index.html", gin.H{
		"title": "Teams",
	})
}

// TeamDetail displays the team detail page
func (h *WebTeamHandler) TeamDetail(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid team ID",
		})
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), uint(teamID), userID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Team not found or access denied",
		})
		return
	}

	// Check if user can manage the team
	canManage := false
	for _, member := range team.Members {
		if member.UserID == userID {
			canManage = member.HasPermission("team:manage") || member.Role == "owner" || member.Role == "admin"
			break
		}
	}

	c.HTML(http.StatusOK, "teams/detail.html", gin.H{
		"Team":      team,
		"CanManage": canManage,
		"title":     team.Name + " - Team Details",
	})
}
