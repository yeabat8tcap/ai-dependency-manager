package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebhooksService handles webhook-related GitHub API operations
type WebhooksService struct {
	client *Client
}

// Webhook represents a GitHub webhook
type Webhook struct {
	ID        int64           `json:"id"`
	URL       string          `json:"url"`
	TestURL   string          `json:"test_url"`
	PingURL   string          `json:"ping_url"`
	Name      string          `json:"name"`
	Events    []string        `json:"events"`
	Active    bool            `json:"active"`
	Config    *WebhookConfig  `json:"config"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedAt time.Time       `json:"created_at"`
	AppID     *int64          `json:"app_id,omitempty"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL string `json:"insecure_ssl,omitempty"`
	Type        string `json:"type"`
}

// WebhookRequest represents a request to create or update a webhook
type WebhookRequest struct {
	Name   string         `json:"name"`
	Active bool           `json:"active"`
	Events []string       `json:"events"`
	Config *WebhookConfig `json:"config"`
}

// WebhookPayload represents the base structure of a webhook payload
type WebhookPayload struct {
	Action     string      `json:"action,omitempty"`
	Number     int         `json:"number,omitempty"`
	Repository *Repository `json:"repository"`
	Sender     *User       `json:"sender"`
}

// PushPayload represents a push webhook payload
type PushPayload struct {
	WebhookPayload
	Ref        string    `json:"ref"`
	Before     string    `json:"before"`
	After      string    `json:"after"`
	Created    bool      `json:"created"`
	Deleted    bool      `json:"deleted"`
	Forced     bool      `json:"forced"`
	BaseRef    *string   `json:"base_ref"`
	Compare    string    `json:"compare"`
	Commits    []*Commit `json:"commits"`
	HeadCommit *Commit   `json:"head_commit"`
	Pusher     *User     `json:"pusher"`
}

// PullRequestPayload represents a pull request webhook payload
type PullRequestPayload struct {
	WebhookPayload
	PullRequest *PullRequest `json:"pull_request"`
	Changes     *PRChanges   `json:"changes,omitempty"`
}

// PRChanges represents changes in a pull request
type PRChanges struct {
	Title *Change `json:"title,omitempty"`
	Body  *Change `json:"body,omitempty"`
	Base  *Change `json:"base,omitempty"`
}

// IssuesPayload represents an issues webhook payload
type IssuesPayload struct {
	WebhookPayload
	Issue   *Issue  `json:"issue"`
	Changes *Change `json:"changes,omitempty"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID        int64      `json:"id"`
	NodeID    string     `json:"node_id"`
	URL       string     `json:"url"`
	HTMLURL   string     `json:"html_url"`
	Number    int        `json:"number"`
	State     string     `json:"state"`
	Title     string     `json:"title"`
	Body      *string    `json:"body"`
	User      *User      `json:"user"`
	Labels    []*Label   `json:"labels"`
	Assignee  *User      `json:"assignee"`
	Assignees []*User    `json:"assignees"`
	Milestone *Milestone `json:"milestone"`
	Locked    bool       `json:"locked"`
	Comments  int        `json:"comments"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

// DependencyUpdatePayload represents a custom payload for dependency updates
type DependencyUpdatePayload struct {
	WebhookPayload
	Dependencies []*DependencyUpdate `json:"dependencies"`
}

// DependencyUpdate is now defined in shared_types.go

// Create creates a new webhook
func (w *WebhooksService) Create(ctx context.Context, owner, repo string, webhook *WebhookRequest) (*Webhook, error) {
	path := fmt.Sprintf("repos/%s/%s/hooks", owner, repo)
	
	req, err := w.client.NewRequest(ctx, "POST", path, webhook)
	if err != nil {
		return nil, err
	}
	
	hook := new(Webhook)
	_, err = w.client.Do(req, hook)
	if err != nil {
		return nil, err
	}
	
	return hook, nil
}

// CreateDependencyWebhook creates a webhook specifically for dependency updates
func (w *WebhooksService) CreateDependencyWebhook(ctx context.Context, owner, repo, webhookURL, secret string) (*Webhook, error) {
	webhook := &WebhookRequest{
		Name:   "web",
		Active: true,
		Events: []string{
			"push",
			"pull_request",
			"issues",
			"repository_vulnerability_alert",
			"dependabot_alert",
			"security_advisory",
		},
		Config: &WebhookConfig{
			URL:         webhookURL,
			ContentType: "json",
			Secret:      secret,
			InsecureSSL: "0",
		},
	}
	
	return w.Create(ctx, owner, repo, webhook)
}

// List lists webhooks for a repository
func (w *WebhooksService) List(ctx context.Context, owner, repo string) ([]*Webhook, error) {
	path := fmt.Sprintf("repos/%s/%s/hooks", owner, repo)
	
	req, err := w.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	var webhooks []*Webhook
	_, err = w.client.Do(req, &webhooks)
	if err != nil {
		return nil, err
	}
	
	return webhooks, nil
}

// Get retrieves a specific webhook
func (w *WebhooksService) Get(ctx context.Context, owner, repo string, id int64) (*Webhook, error) {
	path := fmt.Sprintf("repos/%s/%s/hooks/%d", owner, repo, id)
	
	req, err := w.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	webhook := new(Webhook)
	_, err = w.client.Do(req, webhook)
	if err != nil {
		return nil, err
	}
	
	return webhook, nil
}

// Update updates a webhook
func (w *WebhooksService) Update(ctx context.Context, owner, repo string, id int64, webhook *WebhookRequest) (*Webhook, error) {
	path := fmt.Sprintf("repos/%s/%s/hooks/%d", owner, repo, id)
	
	req, err := w.client.NewRequest(ctx, "PATCH", path, webhook)
	if err != nil {
		return nil, err
	}
	
	hook := new(Webhook)
	_, err = w.client.Do(req, hook)
	if err != nil {
		return nil, err
	}
	
	return hook, nil
}

// Delete deletes a webhook
func (w *WebhooksService) Delete(ctx context.Context, owner, repo string, id int64) error {
	path := fmt.Sprintf("repos/%s/%s/hooks/%d", owner, repo, id)
	
	req, err := w.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	
	_, err = w.client.Do(req, nil)
	return err
}

// Ping pings a webhook
func (w *WebhooksService) Ping(ctx context.Context, owner, repo string, id int64) error {
	path := fmt.Sprintf("repos/%s/%s/hooks/%d/pings", owner, repo, id)
	
	req, err := w.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	
	_, err = w.client.Do(req, nil)
	return err
}

// Test tests a webhook
func (w *WebhooksService) Test(ctx context.Context, owner, repo string, id int64) error {
	path := fmt.Sprintf("repos/%s/%s/hooks/%d/tests", owner, repo, id)
	
	req, err := w.client.NewRequest(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	
	_, err = w.client.Do(req, nil)
	return err
}

// ValidatePayload validates a webhook payload signature
func ValidatePayload(payload []byte, signature string, secret string) bool {
	if secret == "" {
		return true // No secret configured, skip validation
	}
	
	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")
	
	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ParseWebhookPayload parses a webhook payload based on the event type
func ParseWebhookPayload(eventType string, payload []byte) (interface{}, error) {
	switch eventType {
	case "push":
		var pushPayload PushPayload
		if err := json.Unmarshal(payload, &pushPayload); err != nil {
			return nil, fmt.Errorf("failed to parse push payload: %w", err)
		}
		return &pushPayload, nil
		
	case "pull_request":
		var prPayload PullRequestPayload
		if err := json.Unmarshal(payload, &prPayload); err != nil {
			return nil, fmt.Errorf("failed to parse pull request payload: %w", err)
		}
		return &prPayload, nil
		
	case "issues":
		var issuesPayload IssuesPayload
		if err := json.Unmarshal(payload, &issuesPayload); err != nil {
			return nil, fmt.Errorf("failed to parse issues payload: %w", err)
		}
		return &issuesPayload, nil
		
	case "repository_vulnerability_alert", "dependabot_alert", "security_advisory":
		var dependencyPayload DependencyUpdatePayload
		if err := json.Unmarshal(payload, &dependencyPayload); err != nil {
			return nil, fmt.Errorf("failed to parse dependency payload: %w", err)
		}
		return &dependencyPayload, nil
		
	default:
		// Generic payload for unknown event types
		var genericPayload WebhookPayload
		if err := json.Unmarshal(payload, &genericPayload); err != nil {
			return nil, fmt.Errorf("failed to parse generic payload: %w", err)
		}
		return &genericPayload, nil
	}
}

// WebhookHandler represents a function that handles webhook events
type WebhookHandler func(eventType string, payload interface{}) error

// WebhookServer represents a webhook server
type WebhookServer struct {
	secret   string
	handlers map[string]WebhookHandler
	server   *http.Server
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(secret string) *WebhookServer {
	return &WebhookServer{
		secret:   secret,
		handlers: make(map[string]WebhookHandler),
	}
}

// RegisterHandler registers a handler for a specific event type
func (ws *WebhookServer) RegisterHandler(eventType string, handler WebhookHandler) {
	ws.handlers[eventType] = handler
}

// HandleWebhook handles incoming webhook requests
func (ws *WebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Validate signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !ValidatePayload(payload, signature, ws.secret) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}
	
	// Get event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "Missing event type", http.StatusBadRequest)
		return
	}
	
	// Parse payload
	parsedPayload, err := ParseWebhookPayload(eventType, payload)
	if err != nil {
		logger.Error("Failed to parse webhook payload: %v", err)
		http.Error(w, "Failed to parse payload", http.StatusBadRequest)
		return
	}
	
	// Find and execute handler
	handler, exists := ws.handlers[eventType]
	if !exists {
		// No specific handler, try generic handler
		if genericHandler, hasGeneric := ws.handlers["*"]; hasGeneric {
			handler = genericHandler
		} else {
			logger.Debug("No handler registered for event type: %s", eventType)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	
	// Execute handler
	if err := handler(eventType, parsedPayload); err != nil {
		logger.Error("Webhook handler failed for event %s: %v", eventType, err)
		http.Error(w, "Handler failed", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Successfully processed webhook event: %s", eventType)
	w.WriteHeader(http.StatusOK)
}

// Start starts the webhook server
func (ws *WebhookServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", ws.HandleWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	ws.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	logger.Info("Starting webhook server on %s", addr)
	return ws.server.ListenAndServe()
}

// Stop stops the webhook server
func (ws *WebhookServer) Stop(ctx context.Context) error {
	if ws.server == nil {
		return nil
	}
	
	logger.Info("Stopping webhook server")
	return ws.server.Shutdown(ctx)
}

// FindDependencyWebhook finds the AI Dependency Manager webhook in a repository
func (w *WebhooksService) FindDependencyWebhook(ctx context.Context, owner, repo, webhookURL string) (*Webhook, error) {
	webhooks, err := w.List(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	
	for _, webhook := range webhooks {
		if webhook.Config != nil && webhook.Config.URL == webhookURL {
			return webhook, nil
		}
	}
	
	return nil, nil // Not found
}

// EnsureDependencyWebhook ensures a dependency webhook exists, creating it if necessary
func (w *WebhooksService) EnsureDependencyWebhook(ctx context.Context, owner, repo, webhookURL, secret string) (*Webhook, error) {
	// Check if webhook already exists
	existing, err := w.FindDependencyWebhook(ctx, owner, repo, webhookURL)
	if err != nil {
		return nil, err
	}
	
	if existing != nil {
		logger.Info("Dependency webhook already exists for %s/%s", owner, repo)
		return existing, nil
	}
	
	// Create new webhook
	logger.Info("Creating dependency webhook for %s/%s", owner, repo)
	return w.CreateDependencyWebhook(ctx, owner, repo, webhookURL, secret)
}
