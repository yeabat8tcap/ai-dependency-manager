package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// ClaudeProvider implements AI analysis using Anthropic's Claude models
type ClaudeProvider struct {
	client    *Client
	model     string
	maxTokens int
	timeout   time.Duration
}

// ClaudeConfig holds configuration for Claude provider
type ClaudeConfig struct {
	APIKey      string        `json:"api_key" yaml:"api_key"`
	Model       string        `json:"model" yaml:"model"`
	MaxTokens   int           `json:"max_tokens" yaml:"max_tokens"`
	Temperature float64       `json:"temperature" yaml:"temperature"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	BaseURL     string        `json:"base_url" yaml:"base_url"`
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config *ClaudeConfig) (*ClaudeProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	// Set defaults
	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Temperature == 0 {
		config.Temperature = 0.1 // Low temperature for consistent analysis
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}

	client := NewClient(config.APIKey, config.BaseURL)
	client.SetTimeout(config.Timeout)

	return &ClaudeProvider{
		client:    client,
		model:     config.Model,
		maxTokens: config.MaxTokens,
		timeout:   config.Timeout,
	}, nil
}

// GetName returns the provider name
func (c *ClaudeProvider) GetName() string {
	return "claude"
}

// GetVersion returns the provider version
func (c *ClaudeProvider) GetVersion() string {
	return "1.0.0"
}

// IsAvailable checks if the provider is available
func (c *ClaudeProvider) IsAvailable(ctx context.Context) bool {
	// Simple health check - try to make a minimal API call
	request := &MessageRequest{
		Model:     c.model,
		MaxTokens: 10,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Temperature: 0.1,
	}

	_, err := c.client.CreateMessage(ctx, request)
	return err == nil
}

// AnalyzeChangelog analyzes changelog text using Claude's advanced language understanding
func (c *ClaudeProvider) AnalyzeChangelog(ctx context.Context, request *types.ChangelogAnalysisRequest) (*types.ChangelogAnalysisResponse, error) {
	logger.Debug("Analyzing changelog with Claude for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	systemPrompt := c.getChangelogSystemPrompt()
	userPrompt := c.buildChangelogAnalysisPrompt(request)

	messageRequest := &MessageRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		System:      systemPrompt,
		Temperature: 0.1,
	}

	response, err := c.client.CreateMessage(ctx, messageRequest)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	// Extract JSON from response content
	jsonContent := c.extractJSONFromResponse(response.Content[0].Text)

	// Parse the JSON response
	var analysisResult ChangelogAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &analysisResult); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Convert to our response format
	result := c.convertToChangelogResponse(request, &analysisResult)

	logger.Debug("Claude changelog analysis complete for %s: risk=%s, confidence=%.2f", 
		request.PackageName, result.RiskLevel, result.Confidence)

	return result, nil
}

// AnalyzeVersionDiff analyzes version differences using Claude
func (c *ClaudeProvider) AnalyzeVersionDiff(ctx context.Context, request *types.VersionDiffAnalysisRequest) (*types.VersionDiffAnalysisResponse, error) {
	logger.Debug("Analyzing version diff with Claude for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	systemPrompt := c.getVersionDiffSystemPrompt()
	userPrompt := c.buildVersionDiffPrompt(request)

	messageRequest := &MessageRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		System:      systemPrompt,
		Temperature: 0.1,
	}

	response, err := c.client.CreateMessage(ctx, messageRequest)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	// Extract JSON from response content
	jsonContent := c.extractJSONFromResponse(response.Content[0].Text)

	// Parse the JSON response
	var diffResult VersionDiffAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &diffResult); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Convert to our response format
	result := c.convertToVersionDiffResponse(request, &diffResult)

	logger.Debug("Claude version diff analysis complete for %s: risk=%s, confidence=%.2f", 
		request.PackageName, result.RiskLevel, result.Confidence)

	return result, nil
}

// PredictCompatibility predicts compatibility issues using Claude
func (c *ClaudeProvider) PredictCompatibility(ctx context.Context, request *types.CompatibilityPredictionRequest) (*types.CompatibilityPredictionResponse, error) {
	logger.Debug("Predicting compatibility with Claude for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	systemPrompt := c.getCompatibilitySystemPrompt()
	userPrompt := c.buildCompatibilityPrompt(request)

	messageRequest := &MessageRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		System:      systemPrompt,
		Temperature: 0.1,
	}

	response, err := c.client.CreateMessage(ctx, messageRequest)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	// Extract JSON from response content
	jsonContent := c.extractJSONFromResponse(response.Content[0].Text)

	// Parse the JSON response
	var compatResult CompatibilityPredictionResult
	if err := json.Unmarshal([]byte(jsonContent), &compatResult); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Convert to our response format
	result := c.convertToCompatibilityResponse(request, &compatResult)

	logger.Debug("Claude compatibility prediction complete for %s: score=%.2f, confidence=%.2f", 
		request.PackageName, result.CompatibilityScore, result.Confidence)

	return result, nil
}

// ClassifyUpdate classifies update types using Claude
func (c *ClaudeProvider) ClassifyUpdate(ctx context.Context, request *types.UpdateClassificationRequest) (*types.UpdateClassificationResponse, error) {
	logger.Debug("Classifying update with Claude for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	systemPrompt := c.getUpdateClassificationSystemPrompt()
	userPrompt := c.buildUpdateClassificationPrompt(request)

	messageRequest := &MessageRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		System:      systemPrompt,
		Temperature: 0.1,
	}

	response, err := c.client.CreateMessage(ctx, messageRequest)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	// Extract JSON from response content
	jsonContent := c.extractJSONFromResponse(response.Content[0].Text)

	// Parse the JSON response
	var classifyResult UpdateClassificationResult
	if err := json.Unmarshal([]byte(jsonContent), &classifyResult); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Convert to our response format
	result := c.convertToUpdateClassificationResponse(request, &classifyResult)

	logger.Debug("Claude update classification complete for %s: type=%s, priority=%s", 
		request.PackageName, result.UpdateType, result.Priority)

	return result, nil
}

// extractJSONFromResponse extracts JSON content from Claude's response
func (c *ClaudeProvider) extractJSONFromResponse(content string) string {
	// Claude sometimes wraps JSON in markdown code blocks
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.Index(content[start:], "```")
		if end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}

	// Look for JSON object boundaries
	start := strings.Index(content, "{")
	if start == -1 {
		return content
	}

	// Find the matching closing brace
	braceCount := 0
	for i := start; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return content[start : i+1]
			}
		}
	}

	return content
}

// System prompts for different analysis types
func (c *ClaudeProvider) getChangelogSystemPrompt() string {
	return `You are an expert software dependency analyst with deep expertise in package management, semantic versioning, and software engineering best practices. Your role is to analyze changelog text and provide detailed, actionable insights about dependency updates.

Your analysis should be thorough, accurate, and focused on helping developers make informed decisions about updates. Pay special attention to:

1. Breaking changes and their real-world impact
2. Security implications and urgency
3. Migration complexity and effort required
4. Business and technical risk assessment
5. Actionable recommendations with clear timelines

Always respond with valid JSON only, following the exact schema provided. Base your confidence scores on the quality and completeness of the available information. Be conservative with confidence when information is limited or ambiguous.`
}

func (c *ClaudeProvider) getVersionDiffSystemPrompt() string {
	return `You are a senior software engineer specializing in dependency management and version analysis. Your expertise covers semantic versioning, API design, and software architecture across multiple programming languages and ecosystems.

Analyze version differences with precision, focusing on:

1. Semantic versioning compliance and implications
2. API surface changes and backward compatibility
3. Behavioral changes that might affect existing implementations
4. Performance, security, and reliability implications
5. Migration complexity and testing requirements

Provide technical analysis that helps developers understand the true impact of version changes. Always respond with valid JSON following the provided schema.`
}

func (c *ClaudeProvider) getCompatibilitySystemPrompt() string {
	return `You are a software compatibility expert with extensive knowledge of dependency management, software ecosystems, and integration challenges. Your role is to predict potential compatibility issues and provide actionable migration guidance.

Focus your analysis on:

1. Dependency conflicts and version constraints
2. API compatibility across different versions
3. Framework and runtime compatibility
4. Integration complexity and potential issues
5. Risk mitigation and testing strategies

Provide practical compatibility predictions with realistic confidence scores and actionable mitigation strategies. Respond with valid JSON following the provided schema.`
}

func (c *ClaudeProvider) getUpdateClassificationSystemPrompt() string {
	return `You are a software update classification specialist with expertise in release management, risk assessment, and software lifecycle management. Your role is to classify updates based on their content, impact, and business urgency.

Classify updates by analyzing:

1. Update type and semantic versioning implications
2. Priority level based on content and business impact
3. Categories (security, feature, bugfix, maintenance, performance)
4. Urgency and recommended implementation timeline
5. Risk assessment and mitigation strategies

Provide precise classifications with clear reasoning and actionable recommendations. Always respond with valid JSON following the provided schema.`
}
