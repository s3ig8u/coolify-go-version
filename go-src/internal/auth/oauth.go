package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

// OAuthProviderType defines supported providers
type OAuthProviderType string

const (
	ProviderGitHub     OAuthProviderType = "github"
	ProviderGitLab     OAuthProviderType = "gitlab"
	ProviderGoogle     OAuthProviderType = "google"
	ProviderAzure      OAuthProviderType = "azure"
	ProviderAuthentik  OAuthProviderType = "authentik"
	ProviderInfomaniak OAuthProviderType = "infomaniak"
)

// OAuthConfig holds provider configuration
type OAuthConfig struct {
	Provider     OAuthProviderType `json:"provider"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	RedirectURI  string            `json:"redirect_uri"`
	Tenant       string            `json:"tenant,omitempty"`   // For Azure
	BaseURL      string            `json:"base_url,omitempty"` // For Authentik, custom GitLab
	Enabled      bool              `json:"enabled"`
	Scopes       []string          `json:"scopes,omitempty"`
}

// OAuthUser represents user info from OAuth provider
type OAuthUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Provider string `json:"provider"`
}

// OAuthService handles OAuth logic
type OAuthService struct {
	configs map[OAuthProviderType]*OAuthConfig
}

// NewOAuthService creates a new OAuth service
func NewOAuthService() *OAuthService {
	return &OAuthService{
		configs: make(map[OAuthProviderType]*OAuthConfig),
	}
}

// AddProvider adds or updates an OAuth provider configuration
func (s *OAuthService) AddProvider(config *OAuthConfig) {
	s.configs[config.Provider] = config
}

// GetProvider returns OAuth provider configuration
func (s *OAuthService) GetProvider(provider OAuthProviderType) (*OAuthConfig, error) {
	config, exists := s.configs[provider]
	if !exists {
		return nil, fmt.Errorf("OAuth provider %s not configured", provider)
	}
	if !config.Enabled {
		return nil, fmt.Errorf("OAuth provider %s is disabled", provider)
	}
	return config, nil
}

// GetEnabledProviders returns all enabled OAuth providers
func (s *OAuthService) GetEnabledProviders() []*OAuthConfig {
	var enabled []*OAuthConfig
	for _, config := range s.configs {
		if config.Enabled {
			enabled = append(enabled, config)
		}
	}
	return enabled
}

// GetAuthURL returns the OAuth authorization URL
func (s *OAuthService) GetAuthURL(provider OAuthProviderType, state string) (string, error) {
	config, err := s.GetProvider(provider)
	if err != nil {
		return "", err
	}

	oauthConfig := s.buildOAuthConfig(config)
	return oauthConfig.AuthCodeURL(state), nil
}

// ExchangeCode exchanges authorization code for access token and user info
func (s *OAuthService) ExchangeCode(ctx context.Context, provider OAuthProviderType, code string) (*OAuthUser, error) {
	config, err := s.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	oauthConfig := s.buildOAuthConfig(config)

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user info from provider
	return s.getUserInfo(ctx, provider, config, token.AccessToken)
}

// buildOAuthConfig builds OAuth2 config for provider
func (s *OAuthService) buildOAuthConfig(config *OAuthConfig) *oauth2.Config {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Scopes:       s.getDefaultScopes(config.Provider),
	}

	// Override scopes if provided
	if len(config.Scopes) > 0 {
		oauthConfig.Scopes = config.Scopes
	}

	// Set provider-specific endpoints
	switch config.Provider {
	case ProviderGitHub:
		oauthConfig.Endpoint = github.Endpoint
	case ProviderGoogle:
		oauthConfig.Endpoint = google.Endpoint
	case ProviderAzure:
		if config.Tenant != "" {
			oauthConfig.Endpoint = microsoft.AzureADEndpoint(config.Tenant)
		} else {
			oauthConfig.Endpoint = microsoft.AzureADEndpoint("common")
		}
	case ProviderGitLab:
		if config.BaseURL != "" {
			oauthConfig.Endpoint = oauth2.Endpoint{
				AuthURL:  config.BaseURL + "/oauth/authorize",
				TokenURL: config.BaseURL + "/oauth/token",
			}
		} else {
			// Default GitLab.com
			oauthConfig.Endpoint = oauth2.Endpoint{
				AuthURL:  "https://gitlab.com/oauth/authorize",
				TokenURL: "https://gitlab.com/oauth/token",
			}
		}
	case ProviderAuthentik:
		if config.BaseURL != "" {
			oauthConfig.Endpoint = oauth2.Endpoint{
				AuthURL:  config.BaseURL + "/application/o/authorize/",
				TokenURL: config.BaseURL + "/application/o/token/",
			}
		}
	case ProviderInfomaniak:
		oauthConfig.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://login.infomaniak.com/oauth2/auth",
			TokenURL: "https://login.infomaniak.com/oauth2/token",
		}
	}

	return oauthConfig
}

// getDefaultScopes returns default scopes for provider
func (s *OAuthService) getDefaultScopes(provider OAuthProviderType) []string {
	switch provider {
	case ProviderGitHub:
		return []string{"user:email", "read:user"}
	case ProviderGitLab:
		return []string{"read_user"}
	case ProviderGoogle:
		return []string{"openid", "profile", "email"}
	case ProviderAzure:
		return []string{"openid", "profile", "email"}
	case ProviderAuthentik:
		return []string{"openid", "profile", "email"}
	case ProviderInfomaniak:
		return []string{"profile"}
	default:
		return []string{"openid", "profile", "email"}
	}
}

// getUserInfo fetches user information from OAuth provider
func (s *OAuthService) getUserInfo(ctx context.Context, provider OAuthProviderType, config *OAuthConfig, accessToken string) (*OAuthUser, error) {
	switch provider {
	case ProviderGitHub:
		return s.getGitHubUser(ctx, accessToken)
	case ProviderGitLab:
		return s.getGitLabUser(ctx, config, accessToken)
	case ProviderGoogle:
		return s.getGoogleUser(ctx, accessToken)
	case ProviderAzure:
		return s.getAzureUser(ctx, accessToken)
	case ProviderAuthentik:
		return s.getAuthentikUser(ctx, config, accessToken)
	case ProviderInfomaniak:
		return s.getInfomaniakUser(ctx, accessToken)
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// getGitHubUser fetches user info from GitHub
func (s *OAuthService) getGitHubUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	// If primary email is private, fetch from emails endpoint
	if githubUser.Email == "" {
		email, _ := s.getGitHubPrimaryEmail(ctx, accessToken)
		githubUser.Email = email
	}

	return &OAuthUser{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Name:     githubUser.Name,
		Email:    githubUser.Email,
		Avatar:   githubUser.AvatarURL,
		Provider: string(ProviderGitHub),
	}, nil
}

// getGitHubPrimaryEmail fetches primary email from GitHub
func (s *OAuthService) getGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no primary email found")
}

// getGitLabUser fetches user info from GitLab
func (s *OAuthService) getGitLabUser(ctx context.Context, config *OAuthConfig, accessToken string) (*OAuthUser, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/v4/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gitlabUser struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gitlabUser); err != nil {
		return nil, err
	}

	return &OAuthUser{
		ID:       fmt.Sprintf("%d", gitlabUser.ID),
		Name:     gitlabUser.Name,
		Email:    gitlabUser.Email,
		Avatar:   gitlabUser.AvatarURL,
		Provider: string(ProviderGitLab),
	}, nil
}

// getGoogleUser fetches user info from Google
func (s *OAuthService) getGoogleUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	return &OAuthUser{
		ID:       googleUser.ID,
		Name:     googleUser.Name,
		Email:    googleUser.Email,
		Avatar:   googleUser.Picture,
		Provider: string(ProviderGoogle),
	}, nil
}

// getAzureUser fetches user info from Azure AD
func (s *OAuthService) getAzureUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var azureUser struct {
		ID                string `json:"id"`
		DisplayName       string `json:"displayName"`
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&azureUser); err != nil {
		return nil, err
	}

	email := azureUser.Mail
	if email == "" {
		email = azureUser.UserPrincipalName
	}

	return &OAuthUser{
		ID:       azureUser.ID,
		Name:     azureUser.DisplayName,
		Email:    email,
		Avatar:   "", // Azure doesn't provide avatar in basic profile
		Provider: string(ProviderAzure),
	}, nil
}

// getAuthentikUser fetches user info from Authentik
func (s *OAuthService) getAuthentikUser(ctx context.Context, config *OAuthConfig, accessToken string) (*OAuthUser, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL required for Authentik provider")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", config.BaseURL+"/application/o/userinfo/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authentikUser struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authentikUser); err != nil {
		return nil, err
	}

	return &OAuthUser{
		ID:       authentikUser.Sub,
		Name:     authentikUser.Name,
		Email:    authentikUser.Email,
		Avatar:   authentikUser.Picture,
		Provider: string(ProviderAuthentik),
	}, nil
}

// getInfomaniakUser fetches user info from Infomaniak
func (s *OAuthService) getInfomaniakUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.infomaniak.com/1/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var infomaniakResp struct {
		Data struct {
			ID        int    `json:"id"`
			Login     string `json:"login"`
			Email     string `json:"email"`
			FirstName string `json:"firstname"`
			LastName  string `json:"lastname"`
			Avatar    string `json:"avatar"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&infomaniakResp); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(infomaniakResp.Data.FirstName + " " + infomaniakResp.Data.LastName)
	if name == "" {
		name = infomaniakResp.Data.Login
	}

	return &OAuthUser{
		ID:       fmt.Sprintf("%d", infomaniakResp.Data.ID),
		Name:     name,
		Email:    infomaniakResp.Data.Email,
		Avatar:   infomaniakResp.Data.Avatar,
		Provider: string(ProviderInfomaniak),
	}, nil
}

// ValidateConfig validates OAuth configuration
func (config *OAuthConfig) ValidateConfig() error {
	if config.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if config.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}
	if config.RedirectURI == "" {
		return fmt.Errorf("redirect URI is required")
	}

	switch config.Provider {
	case ProviderAzure:
		if config.Tenant == "" {
			return fmt.Errorf("tenant is required for Azure provider")
		}
	case ProviderAuthentik:
		if config.BaseURL == "" {
			return fmt.Errorf("base URL is required for Authentik provider")
		}
		// Validate URL format
		if _, err := url.Parse(config.BaseURL); err != nil {
			return fmt.Errorf("invalid base URL for Authentik provider: %w", err)
		}
	case ProviderGitLab:
		// BaseURL is optional for GitLab (defaults to gitlab.com)
		if config.BaseURL != "" {
			if _, err := url.Parse(config.BaseURL); err != nil {
				return fmt.Errorf("invalid base URL for GitLab provider: %w", err)
			}
		}
	}

	return nil
}
