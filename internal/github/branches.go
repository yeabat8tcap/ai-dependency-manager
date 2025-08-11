package github

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BranchesService handles branch-related GitHub API operations
type BranchesService struct {
	client *Client
}

// Branch represents a GitHub branch
type Branch struct {
	Name      string    `json:"name"`
	Commit    *Commit   `json:"commit"`
	Protected bool      `json:"protected"`
	Links     *BranchLinks `json:"_links,omitempty"`
}

// BranchLinks represents links associated with a branch
type BranchLinks struct {
	Self string `json:"self"`
	HTML string `json:"html"`
}

// Reference represents a Git reference
type Reference struct {
	Ref    string    `json:"ref"`
	NodeID string    `json:"node_id"`
	URL    string    `json:"url"`
	Object *GitObject `json:"object"`
}

// GitObject represents a Git object
type GitObject struct {
	Type string `json:"type"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
}

// BranchProtection represents branch protection settings
type BranchProtection struct {
	RequiredStatusChecks       *RequiredStatusChecks       `json:"required_status_checks"`
	EnforceAdmins              *EnforceAdmins              `json:"enforce_admins"`
	RequiredPullRequestReviews *RequiredPullRequestReviews `json:"required_pull_request_reviews"`
	Restrictions               *BranchRestrictions         `json:"restrictions"`
	RequiredLinearHistory      *RequiredLinearHistory      `json:"required_linear_history"`
	AllowForcePushes           *AllowForcePushes           `json:"allow_force_pushes"`
	AllowDeletions             *AllowDeletions             `json:"allow_deletions"`
}

// RequiredStatusChecks represents required status checks
type RequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

// EnforceAdmins represents admin enforcement settings
type EnforceAdmins struct {
	Enabled bool `json:"enabled"`
}

// RequiredPullRequestReviews represents required PR review settings
type RequiredPullRequestReviews struct {
	RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
	DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
}

// BranchRestrictions represents branch access restrictions
type BranchRestrictions struct {
	Users []User `json:"users"`
	Teams []Team `json:"teams"`
	Apps  []App  `json:"apps"`
}

// Team represents a GitHub team
type Team struct {
	ID          int64  `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
	URL         string `json:"url"`
}

// App represents a GitHub App
type App struct {
	ID     int64  `json:"id"`
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
}

// RequiredLinearHistory represents linear history requirement
type RequiredLinearHistory struct {
	Enabled bool `json:"enabled"`
}

// AllowForcePushes represents force push settings
type AllowForcePushes struct {
	Enabled bool `json:"enabled"`
}

// AllowDeletions represents deletion settings
type AllowDeletions struct {
	Enabled bool `json:"enabled"`
}

// PatchBranchConfig represents configuration for creating patch branches
type PatchBranchConfig struct {
	BaseBranch    string
	BranchPrefix  string
	PackageName   string
	FromVersion   string
	ToVersion     string
	PatchType     string // "breaking", "security", "minor", etc.
	Timestamp     time.Time
}

// GeneratePatchBranchName generates a standardized patch branch name
func GeneratePatchBranchName(config *PatchBranchConfig) string {
	if config.Timestamp.IsZero() {
		config.Timestamp = time.Now()
	}
	
	// Sanitize package name for branch naming
	packageName := strings.ReplaceAll(config.PackageName, "/", "-")
	packageName = strings.ReplaceAll(packageName, "@", "")
	packageName = strings.ToLower(packageName)
	
	// Create branch name with format: prefix/package-name/from-to/type/timestamp
	timestamp := config.Timestamp.Format("20060102-150405")
	
	branchName := fmt.Sprintf("%s/%s/%s-%s/%s/%s",
		config.BranchPrefix,
		packageName,
		config.FromVersion,
		config.ToVersion,
		config.PatchType,
		timestamp,
	)
	
	// Ensure branch name is valid (no spaces, special characters)
	branchName = strings.ReplaceAll(branchName, " ", "-")
	branchName = strings.ReplaceAll(branchName, "_", "-")
	
	return branchName
}

// Get retrieves a specific branch
func (b *BranchesService) Get(ctx context.Context, owner, repo, branch string) (*Branch, error) {
	path := fmt.Sprintf("repos/%s/%s/branches/%s", owner, repo, branch)
	
	req, err := b.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	branchInfo := new(Branch)
	_, err = b.client.Do(req, branchInfo)
	if err != nil {
		return nil, err
	}
	
	return branchInfo, nil
}

// List lists all branches in a repository
func (b *BranchesService) List(ctx context.Context, owner, repo string) ([]*Branch, error) {
	path := fmt.Sprintf("repos/%s/%s/branches", owner, repo)
	
	req, err := b.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	var branches []*Branch
	_, err = b.client.Do(req, &branches)
	if err != nil {
		return nil, err
	}
	
	return branches, nil
}

// Create creates a new branch from a base branch
func (b *BranchesService) Create(ctx context.Context, owner, repo, branchName, baseBranch string) (*Reference, error) {
	// First, get the SHA of the base branch
	baseBranchInfo, err := b.Get(ctx, owner, repo, baseBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get base branch %s: %w", baseBranch, err)
	}
	
	// Create the new branch reference
	path := fmt.Sprintf("repos/%s/%s/git/refs", owner, repo)
	
	createRef := struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	}{
		Ref: "refs/heads/" + branchName,
		SHA: baseBranchInfo.Commit.SHA,
	}
	
	req, err := b.client.NewRequest(ctx, "POST", path, createRef)
	if err != nil {
		return nil, err
	}
	
	reference := new(Reference)
	_, err = b.client.Do(req, reference)
	if err != nil {
		return nil, err
	}
	
	return reference, nil
}

// CreatePatchBranch creates a new patch branch with standardized naming
func (b *BranchesService) CreatePatchBranch(ctx context.Context, owner, repo string, config *PatchBranchConfig) (*Reference, error) {
	if config.BranchPrefix == "" {
		config.BranchPrefix = "patch"
	}
	
	if config.BaseBranch == "" {
		// Get the default branch
		repoInfo, err := b.client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository info: %w", err)
		}
		config.BaseBranch = repoInfo.DefaultBranch
	}
	
	branchName := GeneratePatchBranchName(config)
	
	// Check if branch already exists
	if _, err := b.Get(ctx, owner, repo, branchName); err == nil {
		return nil, fmt.Errorf("patch branch %s already exists", branchName)
	}
	
	return b.Create(ctx, owner, repo, branchName, config.BaseBranch)
}

// Delete deletes a branch
func (b *BranchesService) Delete(ctx context.Context, owner, repo, branch string) error {
	path := fmt.Sprintf("repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	
	req, err := b.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	
	_, err = b.client.Do(req, nil)
	return err
}

// Exists checks if a branch exists
func (b *BranchesService) Exists(ctx context.Context, owner, repo, branch string) (bool, error) {
	_, err := b.Get(ctx, owner, repo, branch)
	if err != nil {
		if IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetProtection retrieves branch protection settings
func (b *BranchesService) GetProtection(ctx context.Context, owner, repo, branch string) (*BranchProtection, error) {
	path := fmt.Sprintf("repos/%s/%s/branches/%s/protection", owner, repo, branch)
	
	req, err := b.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	protection := new(BranchProtection)
	_, err = b.client.Do(req, protection)
	if err != nil {
		return nil, err
	}
	
	return protection, nil
}

// IsProtected checks if a branch is protected
func (b *BranchesService) IsProtected(ctx context.Context, owner, repo, branch string) (bool, error) {
	branchInfo, err := b.Get(ctx, owner, repo, branch)
	if err != nil {
		return false, err
	}
	
	return branchInfo.Protected, nil
}

// GetDefaultBranch gets the default branch for a repository
func (b *BranchesService) GetDefaultBranch(ctx context.Context, owner, repo string) (string, error) {
	repoInfo, err := b.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	
	return repoInfo.DefaultBranch, nil
}

// ListPatchBranches lists all patch branches created by the AI Dependency Manager
func (b *BranchesService) ListPatchBranches(ctx context.Context, owner, repo string, prefix string) ([]*Branch, error) {
	if prefix == "" {
		prefix = "patch"
	}
	
	branches, err := b.List(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	
	var patchBranches []*Branch
	for _, branch := range branches {
		if strings.HasPrefix(branch.Name, prefix+"/") {
			patchBranches = append(patchBranches, branch)
		}
	}
	
	return patchBranches, nil
}

// CleanupOldPatchBranches removes old patch branches based on age
func (b *BranchesService) CleanupOldPatchBranches(ctx context.Context, owner, repo string, maxAge time.Duration, prefix string) ([]string, error) {
	if prefix == "" {
		prefix = "patch"
	}
	
	patchBranches, err := b.ListPatchBranches(ctx, owner, repo, prefix)
	if err != nil {
		return nil, err
	}
	
	var deletedBranches []string
	cutoffTime := time.Now().Add(-maxAge)
	
	for _, branch := range patchBranches {
		// Extract timestamp from branch name
		parts := strings.Split(branch.Name, "/")
		if len(parts) < 5 {
			continue // Invalid patch branch format
		}
		
		timestampStr := parts[len(parts)-1]
		branchTime, err := time.Parse("20060102-150405", timestampStr)
		if err != nil {
			continue // Invalid timestamp format
		}
		
		if branchTime.Before(cutoffTime) {
			if err := b.Delete(ctx, owner, repo, branch.Name); err != nil {
				logger.Warn("Failed to delete old patch branch %s: %v", branch.Name, err)
				continue
			}
			deletedBranches = append(deletedBranches, branch.Name)
			logger.Info("Deleted old patch branch: %s", branch.Name)
		}
	}
	
	return deletedBranches, nil
}

// GetBranchCommits gets the commits for a specific branch
func (b *BranchesService) GetBranchCommits(ctx context.Context, owner, repo, branch string, limit int) ([]*Commit, error) {
	path := fmt.Sprintf("repos/%s/%s/commits", owner, repo)
	
	req, err := b.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	// Add query parameters
	q := req.URL.Query()
	q.Add("sha", branch)
	if limit > 0 {
		q.Add("per_page", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()
	
	var commits []*Commit
	_, err = b.client.Do(req, &commits)
	if err != nil {
		return nil, err
	}
	
	return commits, nil
}
