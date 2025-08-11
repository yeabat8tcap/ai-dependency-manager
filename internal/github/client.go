package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

const (
	// GitHub API base URL
	DefaultBaseURL = "https://api.github.com"
	
	// API version header
	APIVersion = "2022-11-28"
	
	// User agent for API requests
	UserAgent = "AI-Dependency-Manager/1.0"
	
	// Rate limit headers
	RateLimitRemaining = "X-RateLimit-Remaining"
	RateLimitReset     = "X-RateLimit-Reset"
)

// Client represents a GitHub API client
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       AuthProvider
	
	// Rate limiting
	rateLimitRemaining int
	rateLimitReset     time.Time
	
	// Services
	Repositories *RepositoriesService
	PullRequests *PullRequestsService
	Branches     *BranchesService
	Webhooks     *WebhooksService
}

// AuthProvider interface for different authentication methods
type AuthProvider interface {
	// Authenticate adds authentication to the HTTP request
	Authenticate(req *http.Request) error
	// GetType returns the authentication type
	GetType() string
	// IsValid checks if the authentication is valid
	IsValid() bool
}

// NewClient creates a new GitHub API client
func NewClient(auth AuthProvider) (*Client, error) {
	if auth == nil {
		return nil, fmt.Errorf("authentication provider is required")
	}
	
	if !auth.IsValid() {
		return nil, fmt.Errorf("invalid authentication credentials")
	}
	
	baseURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			MaxIdleConnsPerHost: 10,
		},
	}
	
	client := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		auth:       auth,
	}
	
	// Initialize services
	client.Repositories = &RepositoriesService{client: client}
	client.PullRequests = &PullRequestsService{client: client}
	client.Branches = &BranchesService{client: client}
	client.Webhooks = &WebhooksService{client: client}
	
	logger.Info("GitHub client initialized with %s authentication", auth.GetType())
	
	return client, nil
}

// NewRequest creates an authenticated HTTP request
func (c *Client) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// Build URL
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL path %s: %w", path, err)
	}
	
	// Prepare request body
	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", APIVersion)
	req.Header.Set("User-Agent", UserAgent)
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Add authentication
	if err := c.auth.Authenticate(req); err != nil {
		return nil, fmt.Errorf("failed to authenticate request: %w", err)
	}
	
	return req, nil
}

// Do executes an HTTP request and handles the response
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// Check rate limits before making request
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}
	
	logger.Debug("GitHub API request: %s %s", req.Method, req.URL.String())
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Update rate limit information
	c.updateRateLimit(resp)
	
	// Check for API errors
	if err := c.checkResponse(resp); err != nil {
		return resp, err
	}
	
	// Decode response if needed
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			err = decErr
		}
	}
	
	return resp, err
}

// checkRateLimit checks if we're hitting rate limits
func (c *Client) checkRateLimit() error {
	if c.rateLimitRemaining <= 0 && time.Now().Before(c.rateLimitReset) {
		waitTime := time.Until(c.rateLimitReset)
		logger.Warn("GitHub API rate limit exceeded, waiting %v", waitTime)
		time.Sleep(waitTime)
	}
	return nil
}

// updateRateLimit updates rate limit information from response headers
func (c *Client) updateRateLimit(resp *http.Response) {
	if remaining := resp.Header.Get(RateLimitRemaining); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			c.rateLimitRemaining = r
		}
	}
	
	if reset := resp.Header.Get(RateLimitReset); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			c.rateLimitReset = time.Unix(r, 0)
		}
	}
	
	logger.Debug("GitHub API rate limit: %d remaining, resets at %v", 
		c.rateLimitRemaining, c.rateLimitReset)
}

// checkResponse checks the API response for errors
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	
	errorResponse := &ErrorResponse{Response: resp}
	data, err := io.ReadAll(resp.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	
	return errorResponse
}

// GetRateLimit returns current rate limit information
func (c *Client) GetRateLimit() (remaining int, reset time.Time) {
	return c.rateLimitRemaining, c.rateLimitReset
}

// SetBaseURL sets a custom base URL (useful for GitHub Enterprise)
func (c *Client) SetBaseURL(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	
	c.baseURL = u
	logger.Info("GitHub client base URL updated to %s", baseURL)
	return nil
}

// ErrorResponse represents a GitHub API error response
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Message  string         `json:"message"`
	Errors   []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors"`
	DocumentationURL string `json:"documentation_url"`
}

func (r *ErrorResponse) Error() string {
	if r.Message != "" {
		return fmt.Sprintf("GitHub API error (%d): %s", r.Response.StatusCode, r.Message)
	}
	return fmt.Sprintf("GitHub API error (%d)", r.Response.StatusCode)
}

// IsRateLimitError checks if the error is due to rate limiting
func IsRateLimitError(err error) bool {
	if errorResponse, ok := err.(*ErrorResponse); ok {
		return errorResponse.Response.StatusCode == 403 && 
			   strings.Contains(strings.ToLower(errorResponse.Message), "rate limit")
	}
	return false
}

// IsNotFoundError checks if the error is a 404 not found
func IsNotFoundError(err error) bool {
	if errorResponse, ok := err.(*ErrorResponse); ok {
		return errorResponse.Response.StatusCode == 404
	}
	return false
}

// IsUnauthorizedError checks if the error is due to authentication
func IsUnauthorizedError(err error) bool {
	if errorResponse, ok := err.(*ErrorResponse); ok {
		return errorResponse.Response.StatusCode == 401
	}
	return false
}
