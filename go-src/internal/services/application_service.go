package services

import (
	"context"
	"fmt"
	"time"

	"coolify-go/internal/docker"
	"coolify-go/internal/models"
	"coolify-go/internal/utils"

	"gorm.io/gorm"
)

// ApplicationService handles application management
type ApplicationService struct {
	db           *gorm.DB
	dockerClient *docker.Client
}

// NewApplicationService creates a new application service
func NewApplicationService(db *gorm.DB, dockerClient *docker.Client) *ApplicationService {
	return &ApplicationService{
		db:           db,
		dockerClient: dockerClient,
	}
}

// CreateApplicationRequest represents application creation request
type CreateApplicationRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Type        string  `json:"type" validate:"required,oneof=docker static nodejs python php"`
	Port        int     `json:"port" validate:"omitempty,min=1,max=65535"`
	Domain      *string `json:"domain" validate:"omitempty"`
	GitURL      *string `json:"git_url" validate:"omitempty,url"`
	GitBranch   string  `json:"git_branch" validate:"omitempty"`
	BuildPack   *string `json:"build_pack" validate:"omitempty"`
	TeamID      uint    `json:"team_id" validate:"required"`
	ServerID    uint    `json:"server_id" validate:"required"`
}

// CreateApplication creates a new application
func (s *ApplicationService) CreateApplication(ctx context.Context, req *CreateApplicationRequest, userID uint) (*models.Application, error) {
	// Verify user has access to team
	if !s.hasTeamAccess(userID, req.TeamID) {
		return nil, fmt.Errorf("access denied to team")
	}

	// Verify server exists and belongs to team
	var server models.Server
	if err := s.db.Where("id = ? AND team_id = ?", req.ServerID, req.TeamID).First(&server).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("server not found or access denied")
		}
		return nil, fmt.Errorf("failed to verify server: %w", err)
	}

	// Generate slug from name
	slug := s.generateSlug(req.Name, req.TeamID)

	// Set defaults
	gitBranch := req.GitBranch
	if gitBranch == "" {
		gitBranch = "main"
	}

	port := req.Port
	if port == 0 {
		port = 3000
	}

	// Create application
	application := &models.Application{
		Name:        req.Name,
		Description: req.Description,
		Slug:        slug,
		Type:        req.Type,
		Status:      "stopped",
		Port:        port,
		Domain:      req.Domain,
		GitURL:      req.GitURL,
		GitBranch:   gitBranch,
		BuildPack:   req.BuildPack,
		IsActive:    true,
		UserID:      userID,
		TeamID:      req.TeamID,
		ServerID:    req.ServerID,
	}

	if err := s.db.Create(application).Error; err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("User").Preload("Team").Preload("Server").First(application, application.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load application: %w", err)
	}

	return application, nil
}

// GetApplicationsByTeam returns applications for a specific team
func (s *ApplicationService) GetApplicationsByTeam(ctx context.Context, teamID, userID uint, offset, limit int) ([]*models.Application, int64, error) {
	// Verify user has access to team
	if !s.hasTeamAccess(userID, teamID) {
		return nil, 0, fmt.Errorf("access denied to team")
	}

	var applications []*models.Application
	var total int64

	// Get total count
	if err := s.db.Model(&models.Application{}).Where("team_id = ? AND is_active = true", teamID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count applications: %w", err)
	}

	// Get applications with pagination
	err := s.db.Where("team_id = ? AND is_active = true", teamID).
		Offset(offset).Limit(limit).
		Preload("User").Preload("Team").Preload("Server").
		Order("created_at DESC").
		Find(&applications).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list applications: %w", err)
	}

	return applications, total, nil
}

// GetApplication returns an application by ID
func (s *ApplicationService) GetApplication(ctx context.Context, applicationID, userID uint) (*models.Application, error) {
	var application models.Application

	// Get application with team membership check
	err := s.db.
		Joins("JOIN team_members ON applications.team_id = team_members.team_id").
		Where("applications.id = ? AND team_members.user_id = ? AND applications.is_active = true", applicationID, userID).
		Preload("User").Preload("Team").Preload("Server").
		First(&application).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("application not found or access denied")
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return &application, nil
}

// UpdateApplicationRequest represents application update request
type UpdateApplicationRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Port        *int    `json:"port" validate:"omitempty,min=1,max=65535"`
	Domain      *string `json:"domain" validate:"omitempty"`
	GitURL      *string `json:"git_url" validate:"omitempty,url"`
	GitBranch   *string `json:"git_branch" validate:"omitempty"`
	BuildPack   *string `json:"build_pack" validate:"omitempty"`
}

// UpdateApplication updates an application
func (s *ApplicationService) UpdateApplication(ctx context.Context, applicationID uint, req *UpdateApplicationRequest, userID uint) (*models.Application, error) {
	// Get existing application with access check
	application, err := s.GetApplication(ctx, applicationID, userID)
	if err != nil {
		return nil, err
	}

	// Check if user has manage permission
	if !s.hasApplicationManagePermission(userID, application.TeamID) {
		return nil, fmt.Errorf("permission denied")
	}

	// Update fields if provided
	if req.Name != nil {
		// Check if name change requires slug update
		if *req.Name != application.Name {
			application.Name = *req.Name
			application.Slug = s.generateSlug(*req.Name, application.TeamID)
		}
	}

	if req.Description != nil {
		application.Description = req.Description
	}

	if req.Port != nil {
		application.Port = *req.Port
	}

	if req.Domain != nil {
		application.Domain = req.Domain
	}

	if req.GitURL != nil {
		application.GitURL = req.GitURL
	}

	if req.GitBranch != nil {
		application.GitBranch = *req.GitBranch
	}

	if req.BuildPack != nil {
		application.BuildPack = req.BuildPack
	}

	// Save application
	if err := s.db.Save(application).Error; err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return application, nil
}

// DeleteApplication soft deletes an application
func (s *ApplicationService) DeleteApplication(ctx context.Context, applicationID, userID uint) error {
	// Get existing application with access check
	application, err := s.GetApplication(ctx, applicationID, userID)
	if err != nil {
		return err
	}

	// Check if user has manage permission
	if !s.hasApplicationManagePermission(userID, application.TeamID) {
		return fmt.Errorf("permission denied")
	}

	// Stop application if running
	if application.Status == "running" {
		if err := s.StopApplication(ctx, applicationID, userID); err != nil {
			// Log error but continue with deletion
			fmt.Printf("Warning: Failed to stop application before deletion: %v\n", err)
		}
	}

	// Soft delete application
	if err := s.db.Delete(application).Error; err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	return nil
}

// DeployApplication deploys an application
func (s *ApplicationService) DeployApplication(ctx context.Context, applicationID, userID uint) error {
	// Get application
	application, err := s.GetApplication(ctx, applicationID, userID)
	if err != nil {
		return err
	}

	// Check if user has deploy permission
	if !s.hasApplicationDeployPermission(userID, application.TeamID) {
		return fmt.Errorf("permission denied")
	}

	// Check if Docker client is available
	if s.dockerClient == nil {
		return fmt.Errorf("Docker not available")
	}

	// Update status to building
	application.Status = "building"
	if err := s.db.Save(application).Error; err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// TODO: Implement actual deployment logic
	// This would involve:
	// 1. Cloning repository if GitURL is provided
	// 2. Building Docker image or using buildpacks
	// 3. Creating/updating Docker container
	// 4. Starting the container
	// 5. Setting up networking/proxy

	// For now, simulate deployment
	go s.simulateDeployment(application)

	return nil
}

// StopApplication stops a running application
func (s *ApplicationService) StopApplication(ctx context.Context, applicationID, userID uint) error {
	// Get application
	application, err := s.GetApplication(ctx, applicationID, userID)
	if err != nil {
		return err
	}

	// Check if user has manage permission
	if !s.hasApplicationManagePermission(userID, application.TeamID) {
		return fmt.Errorf("permission denied")
	}

	// Check if Docker client is available
	if s.dockerClient == nil {
		return fmt.Errorf("Docker not available")
	}

	// Try to stop Docker container
	containerName := fmt.Sprintf("coolify-%s", application.Slug)
	if err := s.dockerClient.StopContainer(containerName, nil); err != nil {
		// Log warning but continue
		fmt.Printf("Warning: Failed to stop container %s: %v\n", containerName, err)
	}

	// Update status
	application.Status = "stopped"
	if err := s.db.Save(application).Error; err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// RestartApplication restarts an application
func (s *ApplicationService) RestartApplication(ctx context.Context, applicationID, userID uint) error {
	// Stop first
	if err := s.StopApplication(ctx, applicationID, userID); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}

	// Wait a moment
	time.Sleep(2 * time.Second)

	// Deploy again
	return s.DeployApplication(ctx, applicationID, userID)
}

// GetApplicationLogs returns application logs
func (s *ApplicationService) GetApplicationLogs(ctx context.Context, applicationID, userID uint, tail string) ([]string, error) {
	// Get application
	application, err := s.GetApplication(ctx, applicationID, userID)
	if err != nil {
		return nil, err
	}

	// Check if Docker client is available
	if s.dockerClient == nil {
		return []string{"Docker not available"}, nil
	}

	// Get container logs
	containerName := fmt.Sprintf("coolify-%s", application.Slug)
	logs, err := s.dockerClient.GetContainerLogs(containerName, tail, false)
	if err != nil {
		return []string{fmt.Sprintf("Failed to get logs: %v", err)}, nil
	}
	defer logs.Close()

	// TODO: Parse logs and return as string slice
	// For now, return a placeholder
	return []string{
		"Application logs would be displayed here",
		"Container: " + containerName,
		"Status: " + application.Status,
	}, nil
}

// Helper methods

// hasTeamAccess checks if user has access to team
func (s *ApplicationService) hasTeamAccess(userID, teamID uint) bool {
	var membership models.TeamMember
	err := s.db.Where("user_id = ? AND team_id = ?", userID, teamID).First(&membership).Error
	return err == nil
}

// hasApplicationManagePermission checks if user can manage applications
func (s *ApplicationService) hasApplicationManagePermission(userID, teamID uint) bool {
	var membership models.TeamMember
	err := s.db.Where("user_id = ? AND team_id = ?", userID, teamID).First(&membership).Error
	if err != nil {
		return false
	}
	return membership.HasPermission(models.PermissionAppManage)
}

// hasApplicationDeployPermission checks if user can deploy applications
func (s *ApplicationService) hasApplicationDeployPermission(userID, teamID uint) bool {
	var membership models.TeamMember
	err := s.db.Where("user_id = ? AND team_id = ?", userID, teamID).First(&membership).Error
	if err != nil {
		return false
	}
	return membership.HasPermission(models.PermissionAppDeploy) || membership.HasPermission(models.PermissionAppManage)
}

// generateSlug generates a unique slug for application
func (s *ApplicationService) generateSlug(name string, teamID uint) string {
	baseSlug := utils.Slugify(name)
	slug := baseSlug

	for i := 2; i <= 100; i++ {
		var app models.Application
		if err := s.db.Where("slug = ? AND team_id = ?", slug, teamID).First(&app).Error; err != nil {
			break // Slug is available
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, i)
	}

	// Fallback to timestamp if all numbered variants are taken
	if i := 100; i > 0 {
		var app models.Application
		if err := s.db.Where("slug = ? AND team_id = ?", slug, teamID).First(&app).Error; err == nil {
			slug = fmt.Sprintf("%s-%d", baseSlug, time.Now().Unix())
		}
	}

	return slug
}

// simulateDeployment simulates application deployment (placeholder)
func (s *ApplicationService) simulateDeployment(application *models.Application) {
	// Simulate build time
	time.Sleep(5 * time.Second)

	// Update status to running (or error)
	application.Status = "running"
	if err := s.db.Save(application).Error; err != nil {
		fmt.Printf("Failed to update application status: %v\n", err)
		return
	}

	fmt.Printf("Application %s deployed successfully\n", application.Name)
}
