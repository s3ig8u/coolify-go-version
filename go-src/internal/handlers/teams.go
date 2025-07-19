package handlers

import (
	"net/http"
	"strconv"

	"coolify-go/internal/services"

	"github.com/gin-gonic/gin"
)

// TeamHandler handles team-related HTTP requests
type TeamHandler struct {
	teamService       *services.TeamService
	invitationService *services.InvitationService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamService *services.TeamService, invitationService *services.InvitationService) *TeamHandler {
	return &TeamHandler{
		teamService:       teamService,
		invitationService: invitationService,
	}
}

// RegisterRoutes registers team endpoints
func (h *TeamHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Team management
	teams := r.Group("/teams")
	{
		teams.GET("", h.GetTeams)
		teams.POST("", h.CreateTeam)
		teams.GET("/:id", h.GetTeam)
		teams.PUT("/:id", h.UpdateTeam)
		teams.DELETE("/:id", h.DeleteTeam)

		// Team members
		teams.POST("/:id/members", h.AddMember)
		teams.PUT("/:id/members/:memberid", h.UpdateMember)
		teams.DELETE("/:id/members/:memberid", h.RemoveMember)

		// Team invitations
		teams.POST("/:id/invitations", h.CreateInvitation)
		teams.GET("/:id/invitations", h.GetTeamInvitations)
		teams.DELETE("/:id/invitations/:invitationid", h.CancelInvitation)
	}

	// User invitations
	invitations := r.Group("/invitations")
	{
		invitations.GET("", h.GetUserInvitations)
		invitations.POST("/:uuid/accept", h.AcceptInvitation)
		invitations.POST("/:uuid/reject", h.RejectInvitation)
	}
}

// CreateTeam handles POST /api/teams
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req services.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID := getUserID(c)
	team, err := h.teamService.CreateTeam(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": team})
}

// GetTeams handles GET /api/teams
func (h *TeamHandler) GetTeams(c *gin.Context) {
	userID := getUserID(c)
	teams, err := h.teamService.GetTeamsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get teams", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": teams})
}

// GetTeam handles GET /api/teams/:id
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	userID := getUserID(c)
	team, err := h.teamService.GetTeam(c.Request.Context(), teamID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": team})
}

// UpdateTeam handles PUT /api/teams/:id
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req services.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID := getUserID(c)
	team, err := h.teamService.UpdateTeam(c.Request.Context(), teamID, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": team})
}

// DeleteTeam handles DELETE /api/teams/:id
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	userID := getUserID(c)
	if err := h.teamService.DeleteTeam(c.Request.Context(), teamID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// AddMember handles POST /api/teams/:id/members
func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req services.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID := getUserID(c)
	member, err := h.teamService.AddMember(c.Request.Context(), teamID, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": member})
}

// UpdateMember handles PUT /api/teams/:id/members/:memberid
func (h *TeamHandler) UpdateMember(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := parseUintParam(c, "memberid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	var req services.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID := getUserID(c)
	member, err := h.teamService.UpdateMember(c.Request.Context(), teamID, memberID, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": member})
}

// RemoveMember handles DELETE /api/teams/:id/members/:memberid
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := parseUintParam(c, "memberid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	userID := getUserID(c)
	if err := h.teamService.RemoveMember(c.Request.Context(), teamID, memberID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// CreateInvitation handles POST /api/teams/:id/invitations
func (h *TeamHandler) CreateInvitation(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req services.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	userID := getUserID(c)
	invitation, err := h.invitationService.CreateInvitation(c.Request.Context(), teamID, &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invitation})
}

// GetTeamInvitations handles GET /api/teams/:id/invitations
func (h *TeamHandler) GetTeamInvitations(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	userID := getUserID(c)
	invitations, err := h.invitationService.GetTeamInvitations(c.Request.Context(), teamID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invitations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invitations})
}

// CancelInvitation handles DELETE /api/teams/:id/invitations/:invitationid
func (h *TeamHandler) CancelInvitation(c *gin.Context) {
	teamID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	invitationID, err := parseUintParam(c, "invitationid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	userID := getUserID(c)
	if err := h.invitationService.CancelInvitation(c.Request.Context(), invitationID, teamID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel invitation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation cancelled successfully"})
}

// GetUserInvitations handles GET /api/invitations
func (h *TeamHandler) GetUserInvitations(c *gin.Context) {
	userEmail := getUserEmail(c)
	invitations, err := h.invitationService.GetUserInvitations(c.Request.Context(), userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invitations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invitations})
}

// AcceptInvitation handles POST /api/invitations/:uuid/accept
func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation UUID"})
		return
	}

	userID := getUserID(c)
	member, err := h.invitationService.AcceptInvitation(c.Request.Context(), uuid, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": member, "message": "Invitation accepted successfully"})
}

// RejectInvitation handles POST /api/invitations/:uuid/reject
func (h *TeamHandler) RejectInvitation(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation UUID"})
		return
	}

	userID := getUserID(c)
	if err := h.invitationService.RejectInvitation(c.Request.Context(), uuid, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject invitation", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation rejected successfully"})
}

// Helper functions

// getUserID extracts user ID from context (set by auth middleware)
func getUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// getUserEmail extracts user email from context (set by auth middleware)
func getUserEmail(c *gin.Context) string {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		return ""
	}
	if email, ok := userEmail.(string); ok {
		return email
	}
	return ""
}

// parseUintParam parses a URL parameter as uint
func parseUintParam(c *gin.Context, param string) (uint, error) {
	str := c.Param(param)
	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}
