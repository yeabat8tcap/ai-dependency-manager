package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/8tcapital/superint-dep-manager/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitHubClient tests the GitHub API client functionality
func TestGitHubClient(t *testing.T) {
	t.Run("NewClient", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: "test-token"}
		client, err := github.NewClient(auth)
		
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Repositories)
		assert.NotNil(t, client.PullRequests)
		assert.NotNil(t, client.Branches)
		assert.NotNil(t, client.Webhooks)
	})

	t.Run("NewClient_InvalidAuth", func(t *testing.T) {
		client, err := github.NewClient(nil)
		
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "authentication provider is required")
	})

	t.Run("RateLimit", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: "test-token"}
		client, err := github.NewClient(auth)
		require.NoError(t, err)

		remaining, reset := client.GetRateLimit()
		assert.GreaterOrEqual(t, remaining, 0)
		assert.True(t, reset.IsZero() || reset.After(time.Now().Add(-time.Hour)))
	})

	t.Run("SetBaseURL", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: "test-token"}
		client, err := github.NewClient(auth)
		require.NoError(t, err)

		err = client.SetBaseURL("https://api.github.enterprise.com")
		assert.NoError(t, err)

		// Test invalid URL
		err = client.SetBaseURL("invalid-url")
		assert.Error(t, err)
	})
}

// TestPersonalAccessToken tests Personal Access Token authentication
func TestPersonalAccessToken(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: "ghp_test_token_123"}
		
		assert.True(t, auth.IsValid())
		assert.Equal(t, "personal_access_token", auth.GetType())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: ""}
		
		assert.False(t, auth.IsValid())
		assert.Equal(t, "personal_access_token", auth.GetType())
	})

	t.Run("Authenticate", func(t *testing.T) {
		auth := &github.PersonalAccessToken{Token: "test-token"}
		req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		
		err := auth.Authenticate(req)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
	})
}

// TestGitHubAppAuth tests GitHub App authentication
func TestGitHubAppAuth(t *testing.T) {
	t.Run("ValidApp", func(t *testing.T) {
		// Mock private key for testing
		privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef...
-----END RSA PRIVATE KEY-----`

		auth := &github.GitHubApp{
			AppID:      12345,
			PrivateKey: privateKey,
		}

		assert.Equal(t, "github_app", auth.GetType())
		// Note: IsValid() would require a real private key, so we skip that test
	})

	t.Run("InvalidApp", func(t *testing.T) {
		auth := &github.GitHubApp{
			AppID:      0,
			PrivateKey: "",
		}

		assert.False(t, auth.IsValid())
		assert.Equal(t, "github_app", auth.GetType())
	})
}

// TestRepositoriesService tests the repositories service
func TestRepositoriesService(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": 123,
				"name": "repo",
				"full_name": "owner/repo",
				"private": false,
				"html_url": "https://github.com/owner/repo",
				"description": "Test repository",
				"default_branch": "main"
			}`))
		case "/repos/owner/repo/contents/package.json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"type": "file",
				"name": "package.json",
				"path": "package.json",
				"content": "ewogICJuYW1lIjogInRlc3QiCn0K",
				"encoding": "base64",
				"sha": "abc123"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with test server
	auth := &github.PersonalAccessToken{Token: "test-token"}
	client, err := github.NewClient(auth)
	require.NoError(t, err)
	
	err = client.SetBaseURL(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("GetRepository", func(t *testing.T) {
		repo, err := client.Repositories.Get(ctx, "owner", "repo")
		
		assert.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, "repo", repo.Name)
		assert.Equal(t, "owner/repo", repo.FullName)
		assert.Equal(t, "main", repo.DefaultBranch)
	})

	t.Run("GetContent", func(t *testing.T) {
		content, err := client.Repositories.GetContent(ctx, "owner", "repo", "package.json", nil)
		
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.Equal(t, "file", content.Type)
		assert.Equal(t, "package.json", content.Name)
		assert.Equal(t, "package.json", content.Path)
		assert.NotEmpty(t, content.Content)
	})

	t.Run("GetRepository_NotFound", func(t *testing.T) {
		_, err := client.Repositories.Get(ctx, "owner", "nonexistent")
		
		assert.Error(t, err)
		assert.True(t, github.IsNotFoundError(err))
	})
}

// TestErrorHandling tests error handling functionality
func TestErrorHandling(t *testing.T) {
	t.Run("IsRateLimitError", func(t *testing.T) {
		errorResp := &github.ErrorResponse{
			Response: &http.Response{StatusCode: 403},
			Message:  "API rate limit exceeded",
		}
		
		assert.True(t, github.IsRateLimitError(errorResp))
	})

	t.Run("IsNotFoundError", func(t *testing.T) {
		errorResp := &github.ErrorResponse{
			Response: &http.Response{StatusCode: 404},
			Message:  "Not Found",
		}
		
		assert.True(t, github.IsNotFoundError(errorResp))
	})

	t.Run("IsUnauthorizedError", func(t *testing.T) {
		errorResp := &github.ErrorResponse{
			Response: &http.Response{StatusCode: 401},
			Message:  "Bad credentials",
		}
		
		assert.True(t, github.IsUnauthorizedError(errorResp))
	})

	t.Run("ErrorResponse_Error", func(t *testing.T) {
		errorResp := &github.ErrorResponse{
			Response: &http.Response{StatusCode: 422},
			Message:  "Validation Failed",
		}
		
		assert.Contains(t, errorResp.Error(), "422")
		assert.Contains(t, errorResp.Error(), "Validation Failed")
	})
}

// TestWebhooksService tests the webhooks service
func TestWebhooksService(t *testing.T) {
	// Create a test server for webhook validation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/webhook" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	auth := &github.PersonalAccessToken{Token: "test-token"}
	client, err := github.NewClient(auth)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("ValidateWebhookPayload", func(t *testing.T) {
		payload := []byte(`{"action": "opened", "pull_request": {"id": 123}}`)
		secret := "webhook-secret"
		signature := "sha256=test-signature"

		// This would normally validate the HMAC signature
		// For testing, we'll just check that the function exists and can be called
		isValid := client.Webhooks.ValidatePayload(payload, secret, signature)
		// In a real implementation, this would validate the HMAC
		assert.False(t, isValid) // Expected to fail with test signature
	})

	t.Run("ParseWebhookPayload", func(t *testing.T) {
		payload := []byte(`{
			"action": "opened",
			"pull_request": {
				"id": 123,
				"title": "Test PR",
				"state": "open"
			}
		}`)

		event, err := client.Webhooks.ParsePayload("pull_request", payload)
		assert.NoError(t, err)
		assert.NotNil(t, event)
	})
}

// BenchmarkGitHubClient benchmarks GitHub client operations
func BenchmarkGitHubClient(b *testing.B) {
	auth := &github.PersonalAccessToken{Token: "test-token"}
	client, err := github.NewClient(auth)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("ClientCreation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := github.NewClient(auth)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RateLimitCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client.GetRateLimit()
		}
	})

	b.Run("AuthenticationValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			auth.IsValid()
		}
	})
}
