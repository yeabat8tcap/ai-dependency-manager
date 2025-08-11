package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a monitored project
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Path        string    `gorm:"not null" json:"path"`
	Type        string    `gorm:"not null" json:"type"` // npm, pip, maven, gradle
	ConfigFile  string    `json:"config_file"`          // package.json, requirements.txt, etc.
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	LastScan    *time.Time `json:"last_scan,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relationships
	Dependencies []Dependency `gorm:"foreignKey:ProjectID" json:"dependencies,omitempty"`
	Settings     ProjectSettings `gorm:"foreignKey:ProjectID" json:"settings,omitempty"`
}

// ProjectSettings holds project-specific configuration
type ProjectSettings struct {
	ID                   uint   `gorm:"primaryKey" json:"id"`
	ProjectID            uint   `gorm:"not null;uniqueIndex" json:"project_id"`
	AutoUpdateLevel      string `gorm:"default:'none'" json:"auto_update_level"` // none, security, minor, major
	RequireConfirmation  bool   `gorm:"default:true" json:"require_confirmation"`
	IgnorePatterns       string `json:"ignore_patterns"` // JSON array of package patterns to ignore
	TrustedSources       string `json:"trusted_sources"` // JSON array of trusted registry URLs
	NotificationEnabled  bool   `gorm:"default:true" json:"notification_enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Dependency represents a package dependency
type Dependency struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ProjectID       uint      `gorm:"not null;index" json:"project_id"`
	Name            string    `gorm:"not null" json:"name"`
	CurrentVersion  string    `gorm:"not null" json:"current_version"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	RequiredVersion string    `json:"required_version,omitempty"` // Version constraint from config file
	Type            string    `gorm:"not null" json:"type"`       // direct, dev, peer, optional
	Registry        string    `json:"registry,omitempty"`
	Status          string    `gorm:"default:'up-to-date'" json:"status"` // up-to-date, outdated, vulnerable, unknown
	LastChecked     *time.Time `json:"last_checked,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	
	// Relationships
	Project     Project       `gorm:"foreignKey:ProjectID" json:"-"`
	Updates     []Update      `gorm:"foreignKey:DependencyID" json:"updates,omitempty"`
	Predictions []AIPrediction `gorm:"foreignKey:DependencyID" json:"predictions,omitempty"`
}

// Update represents an available update for a dependency
type Update struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DependencyID  uint      `gorm:"not null;index" json:"dependency_id"`
	FromVersion   string    `gorm:"not null" json:"from_version"`
	ToVersion     string    `gorm:"not null" json:"to_version"`
	UpdateType    string    `gorm:"not null" json:"update_type"` // major, minor, patch, prerelease
	Severity      string    `json:"severity,omitempty"`          // low, medium, high, critical
	ChangelogURL  string    `json:"changelog_url,omitempty"`
	ReleaseNotes  string    `gorm:"type:text" json:"release_notes,omitempty"`
	SecurityFix   bool      `gorm:"default:false" json:"security_fix"`
	BreakingChange bool     `gorm:"default:false" json:"breaking_change"`
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, applied, failed, skipped
	AppliedAt     *time.Time `json:"applied_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// Relationships
	Dependency  Dependency     `gorm:"foreignKey:DependencyID" json:"-"`
	Predictions []AIPrediction `gorm:"foreignKey:UpdateID" json:"predictions,omitempty"`
}

// AIPrediction represents AI model predictions for updates
type AIPrediction struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	DependencyID     uint      `gorm:"index" json:"dependency_id,omitempty"`
	UpdateID         uint      `gorm:"index" json:"update_id,omitempty"`
	ModelName        string    `gorm:"not null" json:"model_name"`
	ModelVersion     string    `json:"model_version,omitempty"`
	PredictionType   string    `gorm:"not null" json:"prediction_type"` // breaking_change, security_risk, compatibility
	Confidence       float64   `gorm:"not null" json:"confidence"`      // 0.0 to 1.0
	Result           string    `gorm:"not null" json:"result"`          // true, false, unknown
	Reasoning        string    `gorm:"type:text" json:"reasoning,omitempty"`
	InputData        string    `gorm:"type:text" json:"input_data,omitempty"` // JSON of input used for prediction
	CreatedAt        time.Time `json:"created_at"`
	
	// Relationships
	Dependency *Dependency `gorm:"foreignKey:DependencyID" json:"-"`
	Update     *Update     `gorm:"foreignKey:UpdateID" json:"-"`
}

// ScanResult represents the result of a dependency scan
type ScanResult struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ProjectID        uint      `gorm:"not null;index" json:"project_id"`
	ScanType         string    `gorm:"not null" json:"scan_type"` // full, incremental, security
	Status           string    `gorm:"not null" json:"status"`    // running, completed, failed
	DependenciesFound int      `json:"dependencies_found"`
	UpdatesFound     int       `json:"updates_found"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message,omitempty"`
	Duration         int64     `json:"duration"` // Duration in milliseconds
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	
	// Relationships
	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

// AuditLog represents audit trail for all operations
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id,omitempty"`    // For future multi-user support
	Action    string    `gorm:"not null" json:"action"` // scan, update, configure, etc.
	Resource  string    `json:"resource,omitempty"`   // project name, dependency name, etc.
	Details   string    `gorm:"type:text" json:"details,omitempty"` // JSON details of the action
	Success   bool      `gorm:"not null" json:"success"`
	ErrorMsg  string    `gorm:"type:text" json:"error_message,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// RollbackPlan represents a rollback plan for updates
type RollbackPlan struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"not null;index" json:"project_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Status      string    `gorm:"default:'pending'" json:"status"` // pending, executing, completed, failed
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	Project Project        `gorm:"foreignKey:ProjectID" json:"-"`
	Items   []RollbackItem `gorm:"foreignKey:RollbackPlanID" json:"items,omitempty"`
}

// RollbackItem represents an individual rollback action
type RollbackItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	RollbackPlanID uint      `gorm:"not null;index" json:"rollback_plan_id"`
	DependencyID   uint      `gorm:"not null;index" json:"dependency_id"`
	DependencyName string    `gorm:"not null" json:"dependency_name"`
	FromVersion    string    `gorm:"not null" json:"from_version"`
	ToVersion      string    `gorm:"not null" json:"to_version"`
	Command        string    `gorm:"type:text" json:"command,omitempty"`
	Status         string    `gorm:"default:'pending'" json:"status"` // pending, completed, failed
	ErrorMessage   string    `gorm:"type:text" json:"error_message,omitempty"`
	ExecutedAt     *time.Time `json:"executed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	
	// Relationships
	RollbackPlan RollbackPlan `gorm:"foreignKey:RollbackPlanID" json:"-"`
	Dependency   Dependency   `gorm:"foreignKey:DependencyID" json:"-"`
}

// SecurityCheck represents a security vulnerability check
type SecurityCheck struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DependencyID uint      `gorm:"not null;index" json:"dependency_id"`
	Type         string    `gorm:"not null" json:"type"`       // vulnerability, integrity, malware
	PackageName  string    `gorm:"not null;index" json:"package_name"`
	Version      string    `gorm:"not null" json:"version"`    // Package version checked
	CheckType    string    `gorm:"not null" json:"check_type"` // vulnerability, integrity, malware
	Status       string    `gorm:"not null" json:"status"`     // passed, failed, warning
	Severity     string    `json:"severity,omitempty"`         // low, medium, high, critical
	Details      string    `gorm:"type:text" json:"details,omitempty"`
	Source       string    `json:"source,omitempty"`           // nvd, snyk, github, etc.
	CheckedAt    time.Time `json:"checked_at"`                 // When the check was performed
	CreatedAt    time.Time `json:"created_at"`
	
	// Relationships
	Dependency Dependency `gorm:"foreignKey:DependencyID" json:"-"`
}

// SecurityRule represents a security policy rule
type SecurityRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	RuleType    string    `gorm:"not null" json:"rule_type"` // whitelist, blacklist, severity_threshold
	Pattern     string    `gorm:"not null" json:"pattern"`   // Package pattern or rule definition
	Action      string    `gorm:"not null" json:"action"`    // allow, block, warn
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Credential represents stored credentials for registries
type Credential struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Type        string    `gorm:"not null" json:"type"`     // token, basic_auth, api_key
	Registry    string    `gorm:"not null" json:"registry"` // Registry URL or identifier
	Username    string    `json:"username,omitempty"`
	Password    string    `json:"password,omitempty"`       // Encrypted
	Token       string    `json:"token,omitempty"`          // Encrypted
	PublicKey   string    `json:"public_key,omitempty"`     // For key-based authentication
	PrivateKey  string    `json:"private_key,omitempty"`    // Encrypted private key
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VulnerabilityEntry represents a known vulnerability
type VulnerabilityEntry struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CVE           string    `gorm:"uniqueIndex" json:"cve"`
	PackageName   string    `gorm:"not null;index" json:"package_name"`
	AffectedVersions string `gorm:"not null" json:"affected_versions"`
	FixedVersion  string    `json:"fixed_version,omitempty"`
	Severity      string    `gorm:"not null" json:"severity"` // low, medium, high, critical
	Description   string    `gorm:"type:text" json:"description,omitempty"`
	References    string    `gorm:"type:text" json:"references,omitempty"` // JSON array of URLs
	PublishedAt   time.Time `json:"published_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdatePolicy represents a custom update policy
type UpdatePolicy struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	ProjectID   *uint     `json:"project_id,omitempty"` // nil for global policies
	Priority    int       `json:"priority"`             // Higher number = higher priority
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Policy conditions (stored as JSON)
	Conditions string `gorm:"type:text" json:"conditions"`

	// Policy actions (stored as JSON)
	Actions string `gorm:"type:text" json:"actions"`

	// Metadata
	Author  string `json:"author"`
	Version string `gorm:"default:'1.0'" json:"version"`
	Tags    string `json:"tags"` // Comma-separated tags
}
