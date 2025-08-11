package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// OllamaConfig represents configuration for Ollama provider
type OllamaConfig struct {
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	TopK        int     `json:"top_k"`
	NumPredict  int     `json:"num_predict"`
}

// OllamaProvider implements the AIProvider interface for Ollama
type OllamaProvider struct {
	client *Client
	config *OllamaConfig
}

// PopularModels defines configurations for popular Ollama models
var PopularModels = map[string]*OllamaConfig{
	"llama2": {
		Model:       "llama2",
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		NumPredict:  2048,
	},
	"codellama": {
		Model:       "codellama",
		Temperature: 0.3, // Lower temperature for code analysis
		TopP:        0.8,
		TopK:        30,
		NumPredict:  2048,
	},
	"mistral": {
		Model:       "mistral",
		Temperature: 0.5,
		TopP:        0.9,
		TopK:        40,
		NumPredict:  2048,
	},
	"phi": {
		Model:       "phi",
		Temperature: 0.6,
		TopP:        0.85,
		TopK:        35,
		NumPredict:  1024, // Smaller context for lightweight model
	},
	"llama2:13b": {
		Model:       "llama2:13b",
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
		NumPredict:  4096, // Larger context for bigger model
	},
	"codellama:13b": {
		Model:       "codellama:13b",
		Temperature: 0.3,
		TopP:        0.8,
		TopK:        30,
		NumPredict:  4096,
	},
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config *OllamaConfig) (*OllamaProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("ollama config is required")
	}
	
	if config.Model == "" {
		config.Model = "llama2" // Default model
	}
	
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}
	
	// Apply popular model defaults if available
	if modelConfig, exists := PopularModels[config.Model]; exists {
		if config.Temperature == 0 {
			config.Temperature = modelConfig.Temperature
		}
		if config.TopP == 0 {
			config.TopP = modelConfig.TopP
		}
		if config.TopK == 0 {
			config.TopK = modelConfig.TopK
		}
		if config.NumPredict == 0 {
			config.NumPredict = modelConfig.NumPredict
		}
	} else {
		// Set default generation parameters for unknown models
		if config.Temperature == 0 {
			config.Temperature = 0.7
		}
	}
	if config.TopP == 0 {
		config.TopP = 0.9
	}
	if config.TopK == 0 {
		config.TopK = 40
	}
	if config.NumPredict == 0 {
		config.NumPredict = 2048
	}
	
	client := NewClient(config.BaseURL, config.Model)
	
	return &OllamaProvider{
		client: client,
		config: config,
	}, nil
}

// SwitchModel switches to a different Ollama model with optimized parameters
func (o *OllamaProvider) SwitchModel(modelName string) error {
	logger.Info("Switching Ollama model from %s to %s", o.config.Model, modelName)
	
	// Check if model is available
	ctx := context.Background()
	models, err := o.client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list available models: %w", err)
	}
	
	modelAvailable := false
	for _, model := range models.Models {
		if model.Name == modelName {
			modelAvailable = true
			break
		}
	}
	
	if !modelAvailable {
		return fmt.Errorf("model %s is not available. Please run 'ollama pull %s' first", modelName, modelName)
	}
	
	// Update configuration with new model
	oldModel := o.config.Model
	o.config.Model = modelName
	
	// Apply optimized parameters for the new model
	if modelConfig, exists := PopularModels[modelName]; exists {
		o.config.Temperature = modelConfig.Temperature
		o.config.TopP = modelConfig.TopP
		o.config.TopK = modelConfig.TopK
		o.config.NumPredict = modelConfig.NumPredict
		logger.Debug("Applied optimized parameters for model %s", modelName)
	}
	
	logger.Info("Successfully switched from %s to %s", oldModel, modelName)
	return nil
}

// GetAvailableModels returns a list of available models with their configurations
func (o *OllamaProvider) GetAvailableModels() (map[string]*OllamaConfig, error) {
	ctx := context.Background()
	models, err := o.client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	
	availableModels := make(map[string]*OllamaConfig)
	
	for _, model := range models.Models {
		if config, exists := PopularModels[model.Name]; exists {
			// Use predefined configuration for popular models
			availableModels[model.Name] = config
		} else {
			// Create default configuration for unknown models
			availableModels[model.Name] = &OllamaConfig{
				Model:       model.Name,
				Temperature: 0.7,
				TopP:        0.9,
				TopK:        40,
				NumPredict:  2048,
			}
		}
	}
	
	return availableModels, nil
}

// GetCurrentModel returns the currently configured model name
func (o *OllamaProvider) GetCurrentModel() string {
	return o.config.Model
}

// GetName returns the provider name
func (o *OllamaProvider) GetName() string {
	return "ollama"
}

// GetVersion returns the provider version
func (o *OllamaProvider) GetVersion() string {
	return "1.0.0"
}

// IsAvailable checks if the Ollama provider is available
func (o *OllamaProvider) IsAvailable(ctx context.Context) bool {
	return o.client.IsAvailable(ctx)
}

// AnalyzeChangelog analyzes changelog text using Ollama
func (o *OllamaProvider) AnalyzeChangelog(ctx context.Context, request *types.ChangelogAnalysisRequest) (*types.ChangelogAnalysisResponse, error) {
	logger.Debug("Ollama analyzing changelog for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	prompt := o.generateChangelogAnalysisPrompt(request)
	
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	options := &Options{
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		TopK:        o.config.TopK,
		NumPredict:  o.config.NumPredict,
	}
	
	response, err := o.client.Chat(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}
	
	// Extract JSON from response
	jsonContent := o.extractJSON(response.Message.Content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no valid JSON found in ollama response")
	}
	
	var result OllamaChangelogAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		logger.Warn("Failed to parse Ollama JSON response: %v", err)
		logger.Debug("Raw response: %s", response.Message.Content)
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}
	
	return o.convertToChangelogResponse(request, &result), nil
}

// AnalyzeVersionDiff analyzes version differences using Ollama
func (o *OllamaProvider) AnalyzeVersionDiff(ctx context.Context, request *types.VersionDiffAnalysisRequest) (*types.VersionDiffAnalysisResponse, error) {
	logger.Debug("Ollama analyzing version diff for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	prompt := o.generateVersionDiffAnalysisPrompt(request)
	
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	options := &Options{
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		TopK:        o.config.TopK,
		NumPredict:  o.config.NumPredict,
	}
	
	response, err := o.client.Chat(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}
	
	// Extract JSON from response
	jsonContent := o.extractJSON(response.Message.Content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no valid JSON found in ollama response")
	}
	
	var result OllamaVersionDiffAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		logger.Warn("Failed to parse Ollama JSON response: %v", err)
		logger.Debug("Raw response: %s", response.Message.Content)
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}
	
	return o.convertToVersionDiffResponse(request, &result), nil
}

// PredictCompatibility predicts compatibility using Ollama
func (o *OllamaProvider) PredictCompatibility(ctx context.Context, request *types.CompatibilityPredictionRequest) (*types.CompatibilityPredictionResponse, error) {
	logger.Debug("Ollama predicting compatibility for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	prompt := o.generateCompatibilityPredictionPrompt(request)
	
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	options := &Options{
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		TopK:        o.config.TopK,
		NumPredict:  o.config.NumPredict,
	}
	
	response, err := o.client.Chat(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}
	
	// Extract JSON from response
	jsonContent := o.extractJSON(response.Message.Content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no valid JSON found in ollama response")
	}
	
	var result OllamaCompatibilityPredictionResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		logger.Warn("Failed to parse Ollama JSON response: %v", err)
		logger.Debug("Raw response: %s", response.Message.Content)
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}
	
	return o.convertToCompatibilityResponse(request, &result), nil
}

// ClassifyUpdate classifies updates using Ollama
func (o *OllamaProvider) ClassifyUpdate(ctx context.Context, request *types.UpdateClassificationRequest) (*types.UpdateClassificationResponse, error) {
	logger.Debug("Ollama classifying update for %s: %s -> %s", request.PackageName, request.FromVersion, request.ToVersion)
	
	prompt := o.generateUpdateClassificationPrompt(request)
	
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	options := &Options{
		Temperature: o.config.Temperature,
		TopP:        o.config.TopP,
		TopK:        o.config.TopK,
		NumPredict:  o.config.NumPredict,
	}
	
	response, err := o.client.Chat(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}
	
	// Extract JSON from response
	jsonContent := o.extractJSON(response.Message.Content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no valid JSON found in ollama response")
	}
	
	var result OllamaUpdateClassificationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		logger.Warn("Failed to parse Ollama JSON response: %v", err)
		logger.Debug("Raw response: %s", response.Message.Content)
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}
	
	return o.convertToUpdateClassificationResponse(request, &result), nil
}

// extractJSON extracts JSON content from a response that might contain additional text
func (o *OllamaProvider) extractJSON(content string) string {
	// Look for JSON block markers
	if start := strings.Index(content, "```json"); start != -1 {
		start += 7 // Skip "```json"
		if end := strings.Index(content[start:], "```"); end != -1 {
			return strings.TrimSpace(content[start : start+end])
		}
	}
	
	// Look for JSON object boundaries
	start := strings.Index(content, "{")
	if start == -1 {
		return ""
	}
	
	// Find the matching closing brace
	braceCount := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				return strings.TrimSpace(content[start : i+1])
			}
		}
	}
	
	return ""
}

// GetModel returns the current model name
func (o *OllamaProvider) GetModel() string {
	return o.config.Model
}

// GetBaseURL returns the current base URL
func (o *OllamaProvider) GetBaseURL() string {
	return o.config.BaseURL
}

// ListAvailableModels returns a list of available models
func (o *OllamaProvider) ListAvailableModels(ctx context.Context) ([]string, error) {
	modelsResp, err := o.client.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	
	var models []string
	for _, model := range modelsResp.Models {
		models = append(models, model.Name)
	}
	
	return models, nil
}

// TestConnection tests the connection to Ollama
func (o *OllamaProvider) TestConnection(ctx context.Context) error {
	if !o.IsAvailable(ctx) {
		return fmt.Errorf("ollama is not available at %s or model %s is not installed", o.config.BaseURL, o.config.Model)
	}
	
	// Test with a simple request
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello! Please respond with just 'OK' to confirm you're working.",
		},
	}
	
	options := &Options{
		Temperature: 0.1,
		NumPredict:  10,
	}
	
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	_, err := o.client.Chat(ctx, messages, options)
	return err
}
