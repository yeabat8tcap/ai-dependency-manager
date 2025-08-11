package github

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// PersonalAccessTokenAuth implements authentication using GitHub Personal Access Tokens
type PersonalAccessTokenAuth struct {
	Token string
}

// NewPersonalAccessTokenAuth creates a new PAT authentication provider
func NewPersonalAccessTokenAuth(token string) *PersonalAccessTokenAuth {
	return &PersonalAccessTokenAuth{
		Token: strings.TrimSpace(token),
	}
}

// NewPersonalAccessTokenAuthFromEnv creates PAT auth from environment variable
func NewPersonalAccessTokenAuthFromEnv() *PersonalAccessTokenAuth {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_PAT")
	}
	
	if token == "" {
		return nil
	}
	
	return NewPersonalAccessTokenAuth(token)
}

// Authenticate adds the PAT to the request
func (p *PersonalAccessTokenAuth) Authenticate(req *http.Request) error {
	if p.Token == "" {
		return fmt.Errorf("GitHub personal access token is empty")
	}
	
	req.Header.Set("Authorization", "Bearer "+p.Token)
	return nil
}

// GetType returns the authentication type
func (p *PersonalAccessTokenAuth) GetType() string {
	return "Personal Access Token"
}

// IsValid checks if the PAT is valid
func (p *PersonalAccessTokenAuth) IsValid() bool {
	return p.Token != "" && len(p.Token) > 10
}

// GitHubAppAuth implements authentication using GitHub Apps
type GitHubAppAuth struct {
	AppID          int64
	InstallationID int64
	PrivateKey     *rsa.PrivateKey
	
	// Cached JWT token
	jwtToken   string
	jwtExpires time.Time
	
	// Cached installation token
	installationToken   string
	installationExpires time.Time
}

// GitHubAppConfig represents GitHub App configuration
type GitHubAppConfig struct {
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
	PrivateKeyPath string `json:"private_key_path"`
	PrivateKeyPEM  string `json:"private_key_pem"`
}

// NewGitHubAppAuth creates a new GitHub App authentication provider
func NewGitHubAppAuth(config *GitHubAppConfig) (*GitHubAppAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("GitHub App config is required")
	}
	
	if config.AppID == 0 {
		return nil, fmt.Errorf("GitHub App ID is required")
	}
	
	if config.InstallationID == 0 {
		return nil, fmt.Errorf("GitHub App Installation ID is required")
	}
	
	// Load private key
	var privateKeyData []byte
	var err error
	
	if config.PrivateKeyPEM != "" {
		privateKeyData = []byte(config.PrivateKeyPEM)
	} else if config.PrivateKeyPath != "" {
		privateKeyData, err = os.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either private_key_pem or private_key_path must be provided")
	}
	
	// Parse private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}
	
	return &GitHubAppAuth{
		AppID:          config.AppID,
		InstallationID: config.InstallationID,
		PrivateKey:     privateKey,
	}, nil
}

// NewGitHubAppAuthFromEnv creates GitHub App auth from environment variables
func NewGitHubAppAuthFromEnv() (*GitHubAppAuth, error) {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID environment variable is required")
	}
	
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_APP_ID: %w", err)
	}
	
	installationIDStr := os.Getenv("GITHUB_INSTALLATION_ID")
	if installationIDStr == "" {
		return nil, fmt.Errorf("GITHUB_INSTALLATION_ID environment variable is required")
	}
	
	installationID, err := strconv.ParseInt(installationIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_INSTALLATION_ID: %w", err)
	}
	
	config := &GitHubAppConfig{
		AppID:          appID,
		InstallationID: installationID,
		PrivateKeyPath: os.Getenv("GITHUB_PRIVATE_KEY_PATH"),
		PrivateKeyPEM:  os.Getenv("GITHUB_PRIVATE_KEY_PEM"),
	}
	
	return NewGitHubAppAuth(config)
}

// Authenticate adds the GitHub App token to the request
func (g *GitHubAppAuth) Authenticate(req *http.Request) error {
	token, err := g.getInstallationToken()
	if err != nil {
		return fmt.Errorf("failed to get installation token: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// GetType returns the authentication type
func (g *GitHubAppAuth) GetType() string {
	return "GitHub App"
}

// IsValid checks if the GitHub App auth is valid
func (g *GitHubAppAuth) IsValid() bool {
	return g.AppID > 0 && g.InstallationID > 0 && g.PrivateKey != nil
}

// generateJWT generates a JWT token for GitHub App authentication
func (g *GitHubAppAuth) generateJWT() (string, error) {
	// Check if we have a valid cached JWT
	if g.jwtToken != "" && time.Now().Before(g.jwtExpires) {
		return g.jwtToken, nil
	}
	
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(), // GitHub requires max 10 minutes
		"iss": g.AppID,
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(g.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}
	
	// Cache the JWT
	g.jwtToken = tokenString
	g.jwtExpires = now.Add(9 * time.Minute) // Expire 1 minute early for safety
	
	logger.Debug("Generated new GitHub App JWT token")
	return tokenString, nil
}

// getInstallationToken gets an installation access token
func (g *GitHubAppAuth) getInstallationToken() (string, error) {
	// Check if we have a valid cached installation token
	if g.installationToken != "" && time.Now().Before(g.installationExpires) {
		return g.installationToken, nil
	}
	
	// Generate JWT for app authentication
	jwtToken, err := g.generateJWT()
	if err != nil {
		return "", err
	}
	
	// Create HTTP client for token request
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Create request for installation token
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", g.InstallationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", APIVersion)
	req.Header.Set("User-Agent", UserAgent)
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request installation token: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("failed to get installation token: HTTP %d", resp.StatusCode)
	}
	
	// Parse response
	var tokenResponse struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}
	
	// Cache the installation token
	g.installationToken = tokenResponse.Token
	g.installationExpires = tokenResponse.ExpiresAt.Add(-5 * time.Minute) // Expire 5 minutes early
	
	logger.Debug("Generated new GitHub App installation token, expires at %v", tokenResponse.ExpiresAt)
	return g.installationToken, nil
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type string `json:"type"` // "pat" or "app"
	
	// Personal Access Token config
	Token string `json:"token,omitempty"`
	
	// GitHub App config
	AppID          int64  `json:"app_id,omitempty"`
	InstallationID int64  `json:"installation_id,omitempty"`
	PrivateKeyPath string `json:"private_key_path,omitempty"`
	PrivateKeyPEM  string `json:"private_key_pem,omitempty"`
}

// CreateAuthProvider creates an authentication provider from configuration
func CreateAuthProvider(config *AuthConfig) (AuthProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("authentication config is required")
	}
	
	switch strings.ToLower(config.Type) {
	case "pat", "token", "personal_access_token":
		if config.Token == "" {
			return nil, fmt.Errorf("token is required for PAT authentication")
		}
		return NewPersonalAccessTokenAuth(config.Token), nil
		
	case "app", "github_app":
		appConfig := &GitHubAppConfig{
			AppID:          config.AppID,
			InstallationID: config.InstallationID,
			PrivateKeyPath: config.PrivateKeyPath,
			PrivateKeyPEM:  config.PrivateKeyPEM,
		}
		return NewGitHubAppAuth(appConfig)
		
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", config.Type)
	}
}

// CreateAuthProviderFromEnv creates an authentication provider from environment variables
func CreateAuthProviderFromEnv() (AuthProvider, error) {
	// Try GitHub App first
	if os.Getenv("GITHUB_APP_ID") != "" {
		logger.Info("Attempting GitHub App authentication from environment")
		if auth, err := NewGitHubAppAuthFromEnv(); err == nil {
			return auth, nil
		} else {
			logger.Warn("GitHub App authentication failed: %v", err)
		}
	}
	
	// Fall back to Personal Access Token
	if auth := NewPersonalAccessTokenAuthFromEnv(); auth != nil {
		logger.Info("Using Personal Access Token authentication from environment")
		return auth, nil
	}
	
	return nil, fmt.Errorf("no valid GitHub authentication found in environment variables")
}
