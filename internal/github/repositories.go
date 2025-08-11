package github

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// RepositoriesService handles repository-related GitHub API operations
type RepositoriesService struct {
	client *Client
}

// Repository represents a GitHub repository
type Repository struct {
	ID               int64     `json:"id"`
	NodeID           string    `json:"node_id"`
	Name             string    `json:"name"`
	FullName         string    `json:"full_name"`
	Owner            *User     `json:"owner"`
	Private          bool      `json:"private"`
	HTMLURL          string    `json:"html_url"`
	Description      *string   `json:"description"`
	Fork             bool      `json:"fork"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         *time.Time `json:"pushed_at"`
	GitURL           string    `json:"git_url"`
	SSHURL           string    `json:"ssh_url"`
	CloneURL         string    `json:"clone_url"`
	SVNURL           string    `json:"svn_url"`
	Homepage         *string   `json:"homepage"`
	Size             int       `json:"size"`
	StargazersCount  int       `json:"stargazers_count"`
	WatchersCount    int       `json:"watchers_count"`
	Language         *string   `json:"language"`
	HasIssues        bool      `json:"has_issues"`
	HasProjects      bool      `json:"has_projects"`
	HasWiki          bool      `json:"has_wiki"`
	HasPages         bool      `json:"has_pages"`
	ForksCount       int       `json:"forks_count"`
	Archived         bool      `json:"archived"`
	Disabled         bool      `json:"disabled"`
	OpenIssuesCount  int       `json:"open_issues_count"`
	License          *License  `json:"license"`
	AllowForking     bool      `json:"allow_forking"`
	IsTemplate       bool      `json:"is_template"`
	Topics           []string  `json:"topics"`
	Visibility       string    `json:"visibility"`
	DefaultBranch    string    `json:"default_branch"`
	Permissions      *RepoPermissions `json:"permissions,omitempty"`
}

// User represents a GitHub user
type User struct {
	Login             string `json:"login"`
	ID                int64  `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

// License represents a repository license
type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
	URL    *string `json:"url"`
	NodeID string `json:"node_id"`
}

// RepoPermissions represents repository permissions
type RepoPermissions struct {
	Admin    bool `json:"admin"`
	Maintain bool `json:"maintain"`
	Push     bool `json:"push"`
	Triage   bool `json:"triage"`
	Pull     bool `json:"pull"`
}

// RepositoryContent represents a file or directory in a repository
type RepositoryContent struct {
	Type        string `json:"type"`
	Encoding    string `json:"encoding"`
	Size        int    `json:"size"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Content     string `json:"content"`
	SHA         string `json:"sha"`
	URL         string `json:"url"`
	GitURL      string `json:"git_url"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url"`
}

// RepositoryListOptions specifies options for listing repositories
type RepositoryListOptions struct {
	Visibility  string `url:"visibility,omitempty"`  // all, public, private
	Affiliation string `url:"affiliation,omitempty"` // owner, collaborator, organization_member
	Type        string `url:"type,omitempty"`        // all, owner, public, private, member
	Sort        string `url:"sort,omitempty"`        // created, updated, pushed, full_name
	Direction   string `url:"direction,omitempty"`   // asc, desc
	PerPage     int    `url:"per_page,omitempty"`
	Page        int    `url:"page,omitempty"`
}

// Get retrieves a repository by owner and name
func (r *RepositoriesService) Get(ctx context.Context, owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("repos/%s/%s", owner, repo)
	
	req, err := r.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	repository := new(Repository)
	_, err = r.client.Do(req, repository)
	if err != nil {
		return nil, err
	}
	
	return repository, nil
}

// List lists repositories for the authenticated user
func (r *RepositoriesService) List(ctx context.Context, opts *RepositoryListOptions) ([]*Repository, error) {
	path := "user/repos"
	
	req, err := r.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.Visibility != "" {
			q.Add("visibility", opts.Visibility)
		}
		if opts.Affiliation != "" {
			q.Add("affiliation", opts.Affiliation)
		}
		if opts.Type != "" {
			q.Add("type", opts.Type)
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
		}
		if opts.Direction != "" {
			q.Add("direction", opts.Direction)
		}
		if opts.PerPage > 0 {
			q.Add("per_page", fmt.Sprintf("%d", opts.PerPage))
		}
		if opts.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", opts.Page))
		}
		req.URL.RawQuery = q.Encode()
	}
	
	var repositories []*Repository
	_, err = r.client.Do(req, &repositories)
	if err != nil {
		return nil, err
	}
	
	return repositories, nil
}

// ListByOrg lists repositories for an organization
func (r *RepositoriesService) ListByOrg(ctx context.Context, org string, opts *RepositoryListOptions) ([]*Repository, error) {
	path := fmt.Sprintf("orgs/%s/repos", org)
	
	req, err := r.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	
	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.Type != "" {
			q.Add("type", opts.Type)
		}
		if opts.Sort != "" {
			q.Add("sort", opts.Sort)
		}
		if opts.Direction != "" {
			q.Add("direction", opts.Direction)
		}
		if opts.PerPage > 0 {
			q.Add("per_page", fmt.Sprintf("%d", opts.PerPage))
		}
		if opts.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", opts.Page))
		}
		req.URL.RawQuery = q.Encode()
	}
	
	var repositories []*Repository
	_, err = r.client.Do(req, &repositories)
	if err != nil {
		return nil, err
	}
	
	return repositories, nil
}

// GetContent retrieves the contents of a file or directory
func (r *RepositoriesService) GetContent(ctx context.Context, owner, repo, path string, opts *RepositoryContentGetOptions) (*RepositoryContent, error) {
	urlPath := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	
	req, err := r.client.NewRequest(ctx, "GET", urlPath, nil)
	if err != nil {
		return nil, err
	}
	
	// Add query parameters
	if opts != nil {
		q := req.URL.Query()
		if opts.Ref != "" {
			q.Add("ref", opts.Ref)
		}
		req.URL.RawQuery = q.Encode()
	}
	
	content := new(RepositoryContent)
	_, err = r.client.Do(req, content)
	if err != nil {
		return nil, err
	}
	
	return content, nil
}

// RepositoryContentGetOptions specifies options for getting repository content
type RepositoryContentGetOptions struct {
	Ref string `url:"ref,omitempty"` // The name of the commit/branch/tag
}

// CreateFile creates a new file in a repository
func (r *RepositoriesService) CreateFile(ctx context.Context, owner, repo, path string, opts *RepositoryContentFileOptions) (*RepositoryContentResponse, error) {
	urlPath := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	
	req, err := r.client.NewRequest(ctx, "PUT", urlPath, opts)
	if err != nil {
		return nil, err
	}
	
	response := new(RepositoryContentResponse)
	_, err = r.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// UpdateFile updates an existing file in a repository
func (r *RepositoriesService) UpdateFile(ctx context.Context, owner, repo, path string, opts *RepositoryContentFileOptions) (*RepositoryContentResponse, error) {
	return r.CreateFile(ctx, owner, repo, path, opts)
}

// DeleteFile deletes a file from a repository
func (r *RepositoriesService) DeleteFile(ctx context.Context, owner, repo, path string, opts *RepositoryContentFileOptions) (*RepositoryContentResponse, error) {
	urlPath := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	
	req, err := r.client.NewRequest(ctx, "DELETE", urlPath, opts)
	if err != nil {
		return nil, err
	}
	
	response := new(RepositoryContentResponse)
	_, err = r.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// RepositoryContentFileOptions represents options for creating, updating, or deleting files
type RepositoryContentFileOptions struct {
	Message   string     `json:"message"`
	Content   string     `json:"content,omitempty"`   // Base64 encoded content
	SHA       string     `json:"sha,omitempty"`       // Required for updates and deletes
	Branch    string     `json:"branch,omitempty"`
	Author    *CommitAuthor `json:"author,omitempty"`
	Committer *CommitAuthor `json:"committer,omitempty"`
}

// CommitAuthor represents the author of a commit
type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  *time.Time `json:"date,omitempty"`
}

// RepositoryContentResponse represents the response from creating, updating, or deleting files
type RepositoryContentResponse struct {
	Content *RepositoryContent `json:"content"`
	Commit  *Commit           `json:"commit"`
}

// Commit represents a Git commit
type Commit struct {
	SHA       string       `json:"sha"`
	NodeID    string       `json:"node_id"`
	URL       string       `json:"url"`
	HTMLURL   string       `json:"html_url"`
	Author    *CommitAuthor `json:"author"`
	Committer *CommitAuthor `json:"committer"`
	Message   string       `json:"message"`
	Tree      *Tree        `json:"tree"`
	Parents   []*Commit    `json:"parents"`
}

// Tree represents a Git tree
type Tree struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// GetPermissions checks the permissions for a repository
func (r *RepositoriesService) GetPermissions(ctx context.Context, owner, repo string) (*RepoPermissions, error) {
	repository, err := r.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	
	if repository.Permissions == nil {
		return &RepoPermissions{}, nil
	}
	
	return repository.Permissions, nil
}

// HasWriteAccess checks if the authenticated user has write access to the repository
func (r *RepositoriesService) HasWriteAccess(ctx context.Context, owner, repo string) (bool, error) {
	permissions, err := r.GetPermissions(ctx, owner, repo)
	if err != nil {
		return false, err
	}
	
	return permissions.Push || permissions.Admin || permissions.Maintain, nil
}

// HasAdminAccess checks if the authenticated user has admin access to the repository
func (r *RepositoriesService) HasAdminAccess(ctx context.Context, owner, repo string) (bool, error) {
	permissions, err := r.GetPermissions(ctx, owner, repo)
	if err != nil {
		return false, err
	}
	
	return permissions.Admin, nil
}
