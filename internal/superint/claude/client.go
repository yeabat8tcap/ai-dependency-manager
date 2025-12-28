package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the Anthropic Claude API
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessageRequest represents a request to create a message
type MessageRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	TopK        int       `json:"top_k,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

// ContentBlock represents a block of content in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MessageResponse represents the response from Claude API
type MessageResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        Usage          `json:"usage"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Claude API client
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// CreateMessage sends a message to Claude and returns the response
func (c *Client) CreateMessage(ctx context.Context, request *MessageRequest) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/v1/messages", c.baseURL)

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("Claude API error (%s): %s", errorResp.Error.Type, errorResp.Error.Message)
	}

	// Parse successful response
	var messageResp MessageResponse
	if err := json.Unmarshal(body, &messageResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &messageResp, nil
}

// TestConnection tests the connection to Claude API
func (c *Client) TestConnection(ctx context.Context) error {
	request := &MessageRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 10,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Temperature: 0.1,
	}

	_, err := c.CreateMessage(ctx, request)
	return err
}
