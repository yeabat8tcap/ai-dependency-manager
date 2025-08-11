package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// OpenAIProvider implements AI analysis using OpenAI's GPT models
type OpenAIProvider struct {
	client    *Client
	model     string
	maxTokens int
	timeout   time.Duration
}

// OpenAIConfig holds configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey      string        `json:"api_key" yaml:"api_key"`
	Model       string        `json:"model" yaml:"model"`
	MaxTokens   int           `json:"max_tokens" yaml:"max_tokens"`
	Temperature float64       `json:"temperature" yaml:"temperature"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	BaseURL     string        `json:"base_url" yaml:"base_url"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *OpenAIConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	// Set defaults
	if config.Model == "" {
		config.Model = "gpt-4-turbo-preview"
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
		config.BaseURL = "https://api.openai.com/v1"
	}

	client := NewClient(config.APIKey, config.BaseURL)
	client.SetTimeout(config.Timeout)

	return &OpenAIProvider{
		client:    client,
		model:     config.Model,
		maxTokens: config.MaxTokens,
		timeout:   config.Timeout,
	}, nil
}

// GetName returns the provider name
func (o *OpenAIProvider) GetName() string {
	return "openai"
}

// GetVersion returns the provider version
func (o *OpenAIProvider) GetVersion() string {
	return "1.0.0"
}

// IsAvailable checks if the provider is available
func (o *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	// Simple health check - try to make a minimal API call
	request := &ChatCompletionRequest{
		Model: o.model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := o.client.CreateChatCompletion(ctx, request)
	return err == nil
}

// AnalyzeChangelog analyzes changelog text using OpenAI's advanced language understanding
func (o *OpenAIProvider) AnalyzeChangelog(ctx context.Context, request *types.ChangelogAnalysisRequest) (*types.ChangelogAnalysisResponse, error) {
	logger.Debug("Analyzing changelog with OpenAI for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	prompt := o.buildChangelogAnalysisPrompt(request)

	chatRequest := &ChatCompletionRequest{
		Model: o.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: o.getSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   o.maxTokens,
		Temperature: 0.1,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	response, err := o.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON response
	var analysisResult ChangelogAnalysisResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &analysisResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to our response format
	result := o.convertToChangelogResponse(request, &analysisResult)

	logger.Debug("OpenAI changelog analysis complete for %s: risk=%s, confidence=%.2f", 
		request.PackageName, result.RiskLevel, result.Confidence)

	return result, nil
}

// AnalyzeVersionDiff analyzes version differences using OpenAI
func (o *OpenAIProvider) AnalyzeVersionDiff(ctx context.Context, request *types.VersionDiffAnalysisRequest) (*types.VersionDiffAnalysisResponse, error) {
	logger.Debug("Analyzing version diff with OpenAI for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	prompt := o.buildVersionDiffPrompt(request)

	chatRequest := &ChatCompletionRequest{
		Model: o.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: o.getVersionDiffSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   o.maxTokens,
		Temperature: 0.1,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	response, err := o.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON response
	var diffResult VersionDiffAnalysisResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &diffResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to our response format
	result := o.convertToVersionDiffResponse(request, &diffResult)

	logger.Debug("OpenAI version diff analysis complete for %s: risk=%s, confidence=%.2f", 
		request.PackageName, result.RiskLevel, result.Confidence)

	return result, nil
}

// PredictCompatibility predicts compatibility issues using OpenAI
func (o *OpenAIProvider) PredictCompatibility(ctx context.Context, request *types.CompatibilityPredictionRequest) (*types.CompatibilityPredictionResponse, error) {
	logger.Debug("Predicting compatibility with OpenAI for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	prompt := o.buildCompatibilityPrompt(request)

	chatRequest := &ChatCompletionRequest{
		Model: o.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: o.getCompatibilitySystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   o.maxTokens,
		Temperature: 0.1,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	response, err := o.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON response
	var compatResult CompatibilityPredictionResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &compatResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to our response format
	result := o.convertToCompatibilityResponse(request, &compatResult)

	logger.Debug("OpenAI compatibility prediction complete for %s: score=%.2f, confidence=%.2f", 
		request.PackageName, result.CompatibilityScore, result.Confidence)

	return result, nil
}

// ClassifyUpdate classifies update types using OpenAI
func (o *OpenAIProvider) ClassifyUpdate(ctx context.Context, request *types.UpdateClassificationRequest) (*types.UpdateClassificationResponse, error) {
	logger.Debug("Classifying update with OpenAI for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)

	prompt := o.buildUpdateClassificationPrompt(request)

	chatRequest := &ChatCompletionRequest{
		Model: o.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: o.getUpdateClassificationSystemPrompt(),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   o.maxTokens,
		Temperature: 0.1,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	response, err := o.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON response
	var classifyResult UpdateClassificationResult
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &classifyResult); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to our response format
	result := o.convertToUpdateClassificationResponse(request, &classifyResult)

	logger.Debug("OpenAI update classification complete for %s: type=%s, priority=%s", 
		request.PackageName, result.UpdateType, result.Priority)

	return result, nil
}

// getSystemPrompt returns the system prompt for changelog analysis
func (o *OpenAIProvider) getSystemPrompt() string {
	return `You are an expert software dependency analyst with deep knowledge of package management, semantic versioning, and software engineering best practices. Your task is to analyze changelog text and provide detailed insights about dependency updates.

You must respond with valid JSON only, following the exact schema provided. Analyze the changelog text carefully and provide accurate assessments of:
1. Breaking changes and their impact
2. New features and their significance
3. Bug fixes and security patches
4. Deprecations and migration requirements
5. Overall risk assessment and recommendations

Be thorough but concise. Focus on actionable insights that help developers make informed update decisions.`
}

// getVersionDiffSystemPrompt returns the system prompt for version diff analysis
func (o *OpenAIProvider) getVersionDiffSystemPrompt() string {
	return `You are an expert software engineer specializing in dependency management and version analysis. Analyze version differences between package versions and provide detailed technical insights.

Focus on:
1. Semantic versioning implications
2. API changes and their impact
3. Behavioral changes that might affect existing code
4. Performance and security implications
5. Migration complexity and effort required

Respond with valid JSON only, providing precise technical analysis.`
}

// getCompatibilitySystemPrompt returns the system prompt for compatibility prediction
func (o *OpenAIProvider) getCompatibilitySystemPrompt() string {
	return `You are a software compatibility expert with extensive knowledge of dependency management across different programming languages and ecosystems. Predict compatibility issues and provide migration guidance.

Analyze:
1. Potential breaking changes and their likelihood
2. Compatibility issues with existing dependencies
3. Migration steps and effort required
4. Testing recommendations
5. Risk mitigation strategies

Provide actionable compatibility predictions with confidence scores.`
}

// getUpdateClassificationSystemPrompt returns the system prompt for update classification
func (o *OpenAIProvider) getUpdateClassificationSystemPrompt() string {
	return `You are a software update classification specialist. Classify dependency updates based on their content, impact, and urgency.

Classify updates by:
1. Update type (major, minor, patch, security, hotfix)
2. Priority level (critical, high, medium, low)
3. Categories (security, feature, bugfix, maintenance, performance)
4. Urgency and recommended timeline
5. Business impact assessment

Provide precise classifications with reasoning.`
}
