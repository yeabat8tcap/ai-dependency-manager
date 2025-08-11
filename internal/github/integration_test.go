package github

import (
	"context"
	"testing"
	"time"
)

func TestGitHubIntegration(t *testing.T) {
	// Skip if no GitHub token available
	config := LoadConfigFromEnv()
	if config.PersonalAccessToken == "" && config.AppID == 0 {
		t.Skip("No GitHub authentication configured")
	}

	// Test manager creation
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test initialization
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Test configuration validation
	if err := config.Validate(); err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}

	t.Log("GitHub integration test passed")
}

func TestWebhookPayloadValidation(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"test": "payload"}`)
	
	// Test without signature (should pass with empty secret)
	if !ValidatePayload(payload, "", "") {
		t.Error("Validation should pass with empty secret")
	}
	
	// Test with valid signature
	signature := "sha256=test-signature"
	if ValidatePayload(payload, signature, secret) {
		t.Error("Validation should fail with invalid signature")
	}
	
	t.Log("Webhook payload validation test passed")
}

func TestConfigurationLoading(t *testing.T) {
	config := DefaultConfig()
	
	// Test default values
	if config.AuthType != "pat" {
		t.Errorf("Expected default auth type 'pat', got '%s'", config.AuthType)
	}
	
	if config.WebhookPort != 8080 {
		t.Errorf("Expected default webhook port 8080, got %d", config.WebhookPort)
	}
	
	if config.BranchMaxAge != 7*24*time.Hour {
		t.Errorf("Expected default branch max age 7 days, got %v", config.BranchMaxAge)
	}
	
	t.Log("Configuration loading test passed")
}
