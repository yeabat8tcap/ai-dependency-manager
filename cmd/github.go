package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/8tcapital/ai-dep-manager/internal/github"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub integration for automated dependency patching",
	Long: `GitHub integration commands for managing automated dependency patching.
This includes setting up webhooks, creating patch pull requests, and managing
GitHub repository integration for the AI Dependency Manager.`,
}

var githubSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup GitHub integration",
	Long: `Setup GitHub integration by configuring authentication, repositories,
and webhooks for automated dependency patching.`,
	RunE: runGitHubSetup,
}

var githubStatusCmd = &cobra.Command{
	Use:   "status [owner/repo]",
	Short: "Show GitHub integration status",
	Long: `Show the status of GitHub integration for all repositories or a specific repository.
Displays information about webhooks, patch PRs, and configuration.`,
	RunE: runGitHubStatus,
}

var githubWebhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage GitHub webhooks",
	Long:  "Manage GitHub webhooks for dependency update notifications.",
}

// Phase 5 Enterprise Features Commands
var githubBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Manage batch dependency updates",
	Long:  "Create and manage batch jobs for processing multiple dependency updates.",
}

var githubBatchCreateCmd = &cobra.Command{
	Use:   "create [config-file]",
	Short: "Create a new batch update job",
	Long:  "Create a new batch job for processing multiple dependency updates.",
	RunE:  runGitHubBatchCreate,
}

var githubBatchStatusCmd = &cobra.Command{
	Use:   "status [job-id]",
	Short: "Show batch job status",
	Long:  "Show the status of batch update jobs.",
	RunE:  runGitHubBatchStatus,
}

var githubAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View patch success analytics",
	Long:  "Generate and view analytics reports for patch success rates and performance.",
}

var githubAnalyticsReportCmd = &cobra.Command{
	Use:   "report [repository]",
	Short: "Generate analytics report",
	Long:  "Generate a comprehensive analytics report for patch operations.",
	RunE:  runGitHubAnalyticsReport,
}

var githubPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage organization policies",
	Long:  "Create and manage organization policies for dependency updates.",
}

var githubPolicyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all policies",
	Long:  "List all organization policies and patch rules.",
	RunE:  runGitHubPolicyList,
}

var githubApprovalCmd = &cobra.Command{
	Use:   "approval",
	Short: "Manage approval workflows",
	Long:  "Create and manage approval workflows for dependency updates.",
}

var githubApprovalStatusCmd = &cobra.Command{
	Use:   "status [workflow-id]",
	Short: "Show approval workflow status",
	Long:  "Show the status of approval workflows.",
	RunE:  runGitHubApprovalStatus,
}

var githubWebhookStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start webhook server",
	Long:  "Start the GitHub webhook server to receive dependency update notifications.",
	RunE:  runGitHubWebhookStart,
}

var githubWebhookStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop webhook server",
	Long:  "Stop the GitHub webhook server.",
	RunE:  runGitHubWebhookStop,
}

var githubWebhookListCmd = &cobra.Command{
	Use:   "list <owner/repo>",
	Short: "List webhooks for a repository",
	Long:  "List all webhooks configured for a specific repository.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGitHubWebhookList,
}

var githubPatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Manage dependency patches",
	Long:  "Manage dependency patches and pull requests.",
}

var githubPatchCreateCmd = &cobra.Command{
	Use:   "create <owner/repo>",
	Short: "Create a dependency patch",
	Long:  "Create a dependency patch pull request for a repository.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGitHubPatchCreate,
}

var githubPatchListCmd = &cobra.Command{
	Use:   "list [owner/repo]",
	Short: "List patch pull requests",
	Long:  "List all patch pull requests for all repositories or a specific repository.",
	RunE:  runGitHubPatchList,
}

var githubCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup old patch branches",
	Long:  "Cleanup old patch branches that no longer have open pull requests.",
	RunE:  runGitHubCleanup,
}

// Command flags
var (
	githubAuthType        string
	githubToken           string
	githubAppID           int64
	githubInstallationID  int64
	githubPrivateKeyPath  string
	githubWebhookURL      string
	githubWebhookSecret   string
	githubWebhookPort     int
	githubRepositories    []string
	githubReviewers       []string
	githubLabels          []string
	githubAutoMerge       bool
	githubBaseBranch      string
	githubUpdateType      string
	githubDependencies    []string
	githubPatchContent    string
	githubCommitMessage   string
	githubPRTitle         string
	githubPRDescription   string
	githubOutputFormat    string
)

func init() {
	rootCmd.AddCommand(githubCmd)
	
	// Add subcommands
	githubCmd.AddCommand(githubSetupCmd)
	githubCmd.AddCommand(githubStatusCmd)
	githubCmd.AddCommand(githubWebhookCmd)
	githubCmd.AddCommand(githubPatchCmd)
	githubCmd.AddCommand(githubCleanupCmd)
	
	// Phase 5 Enterprise Features
	githubCmd.AddCommand(githubBatchCmd)
	githubCmd.AddCommand(githubAnalyticsCmd)
	githubCmd.AddCommand(githubPolicyCmd)
	githubCmd.AddCommand(githubApprovalCmd)
	
	// Webhook subcommands
	githubWebhookCmd.AddCommand(githubWebhookStartCmd)
	githubWebhookCmd.AddCommand(githubWebhookStopCmd)
	githubWebhookCmd.AddCommand(githubWebhookListCmd)
	
	// Patch subcommands
	githubPatchCmd.AddCommand(githubPatchCreateCmd)
	githubPatchCmd.AddCommand(githubPatchListCmd)
	
	// Phase 5 Enterprise Subcommands
	githubBatchCmd.AddCommand(githubBatchCreateCmd)
	githubBatchCmd.AddCommand(githubBatchStatusCmd)
	
	githubAnalyticsCmd.AddCommand(githubAnalyticsReportCmd)
	
	githubPolicyCmd.AddCommand(githubPolicyListCmd)
	
	githubApprovalCmd.AddCommand(githubApprovalStatusCmd)
	
	// Setup command flags
	githubSetupCmd.Flags().StringVar(&githubAuthType, "auth-type", "pat", "Authentication type (pat, app)")
	githubSetupCmd.Flags().StringVar(&githubToken, "token", "", "GitHub personal access token")
	githubSetupCmd.Flags().Int64Var(&githubAppID, "app-id", 0, "GitHub App ID")
	githubSetupCmd.Flags().Int64Var(&githubInstallationID, "installation-id", 0, "GitHub App installation ID")
	githubSetupCmd.Flags().StringVar(&githubPrivateKeyPath, "private-key-path", "", "Path to GitHub App private key")
	githubSetupCmd.Flags().StringVar(&githubWebhookURL, "webhook-url", "", "Webhook URL for receiving events")
	githubSetupCmd.Flags().StringVar(&githubWebhookSecret, "webhook-secret", "", "Webhook secret for payload validation")
	githubSetupCmd.Flags().IntVar(&githubWebhookPort, "webhook-port", 8080, "Port for webhook server")
	githubSetupCmd.Flags().StringSliceVar(&githubRepositories, "repositories", []string{}, "Repositories to enable (owner/repo format)")
	githubSetupCmd.Flags().StringSliceVar(&githubReviewers, "reviewers", []string{}, "Default reviewers for patch PRs")
	githubSetupCmd.Flags().StringSliceVar(&githubLabels, "labels", []string{"dependencies", "automated"}, "Default labels for patch PRs")
	githubSetupCmd.Flags().BoolVar(&githubAutoMerge, "auto-merge", false, "Enable auto-merge for patch PRs")
	
	// Status command flags
	githubStatusCmd.Flags().StringVar(&githubOutputFormat, "output", "table", "Output format (table, json)")
	
	// Webhook start command flags
	githubWebhookStartCmd.Flags().IntVar(&githubWebhookPort, "port", 8080, "Port for webhook server")
	
	// Patch create command flags
	githubPatchCreateCmd.Flags().StringVar(&githubBaseBranch, "base-branch", "main", "Base branch for the patch")
	githubPatchCreateCmd.Flags().StringVar(&githubUpdateType, "update-type", "patch", "Type of update (patch, minor, major, security)")
	githubPatchCreateCmd.Flags().StringSliceVar(&githubDependencies, "dependencies", []string{}, "Dependencies to update (name:version format)")
	githubPatchCreateCmd.Flags().StringVar(&githubPatchContent, "patch-content", "", "Patch content or file path")
	githubPatchCreateCmd.Flags().StringVar(&githubCommitMessage, "commit-message", "", "Commit message for the patch")
	githubPatchCreateCmd.Flags().StringVar(&githubPRTitle, "pr-title", "", "Pull request title")
	githubPatchCreateCmd.Flags().StringVar(&githubPRDescription, "pr-description", "", "Pull request description")
	githubPatchCreateCmd.Flags().StringSliceVar(&githubReviewers, "reviewers", []string{}, "Reviewers for the pull request")
	githubPatchCreateCmd.Flags().StringSliceVar(&githubLabels, "labels", []string{}, "Labels for the pull request")
	githubPatchCreateCmd.Flags().BoolVar(&githubAutoMerge, "auto-merge", false, "Enable auto-merge for this patch")
	
	// Patch list command flags
	githubPatchListCmd.Flags().StringVar(&githubOutputFormat, "output", "table", "Output format (table, json)")
}

func runGitHubSetup(cmd *cobra.Command, args []string) error {
	logger.Info("Setting up GitHub integration")
	
	// Load existing configuration or create new one
	config := github.LoadConfigFromEnv()
	
	// Override with command line flags
	if cmd.Flags().Changed("auth-type") {
		config.AuthType = githubAuthType
	}
	if cmd.Flags().Changed("token") {
		config.PersonalAccessToken = githubToken
	}
	if cmd.Flags().Changed("app-id") {
		config.AppID = githubAppID
	}
	if cmd.Flags().Changed("installation-id") {
		config.InstallationID = githubInstallationID
	}
	if cmd.Flags().Changed("private-key-path") {
		config.PrivateKeyPath = githubPrivateKeyPath
	}
	if cmd.Flags().Changed("webhook-url") {
		config.WebhookURL = githubWebhookURL
	}
	if cmd.Flags().Changed("webhook-secret") {
		config.WebhookSecret = githubWebhookSecret
	}
	if cmd.Flags().Changed("webhook-port") {
		config.WebhookPort = githubWebhookPort
	}
	if cmd.Flags().Changed("reviewers") {
		config.DefaultReviewers = githubReviewers
	}
	if cmd.Flags().Changed("labels") {
		config.DefaultLabels = githubLabels
	}
	if cmd.Flags().Changed("auto-merge") {
		config.AutoMergePatch = githubAutoMerge
	}
	
	// Add repositories
	if cmd.Flags().Changed("repositories") {
		config.Repositories = []github.RepositoryConfig{}
		for _, repo := range githubRepositories {
			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
			}
			config.AddRepository(parts[0], parts[1], true)
		}
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Initialize GitHub integration
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	logger.Info("GitHub integration setup completed successfully")
	return nil
}

// Phase 5 Enterprise Feature Implementation Functions

func runGitHubBatchCreate(cmd *cobra.Command, args []string) error {
	logger.Info("Creating batch update job")
	
	// Load GitHub configuration
	config := github.LoadConfigFromEnv()
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Create batch processor
	batchConfig := &github.BatchConfig{
		MaxConcurrentJobs:    5,
		MaxUpdatesPerBatch:   50,
		BatchTimeout:         30 * time.Minute,
		RetryAttempts:        3,
		RetryDelay:           5 * time.Minute,
		GroupingStrategy:     "by_risk",
		ProcessingMode:       "adaptive",
		ConflictResolution:   "resolve",
		NotificationChannels: []string{"console"},
		ReportingEnabled:     true,
	}
	
	batchProcessor := github.NewBatchProcessor(
		manager.GetClient(),
		manager.GetPatchGenerator(),
		manager.GetApplicator(),
		manager.GetPRManager(),
		batchConfig,
	)
	
	// Create sample batch job
	updates := []*github.DependencyUpdate{
		{
			ID:              "update_1",
			Repository:      "example/repo",
			DependencyName:  "react",
			CurrentVersion:  "17.0.0",
			TargetVersion:   "18.2.0",
			UpdateType:      "major",
			Priority:        "high",
			RiskLevel:       "medium",
			BreakingChanges: true,
			SecurityFix:     false,
		},
		{
			ID:              "update_2",
			Repository:      "example/repo",
			DependencyName:  "lodash",
			CurrentVersion:  "4.17.20",
			TargetVersion:   "4.17.21",
			UpdateType:      "patch",
			Priority:        "medium",
			RiskLevel:       "low",
			BreakingChanges: false,
			SecurityFix:     true,
		},
	}
	
	jobConfig := &github.BatchJobConfig{
		GroupingStrategy:     "by_risk",
		ProcessingMode:       "adaptive",
		MaxConcurrency:       3,
		ConflictResolution:   "resolve",
		CreatePRs:            true,
		PRTemplate:           "default",
		AutoMerge:            false,
		NotifyOnCompletion:   true,
		TestingRequired:      true,
		ApprovalRequired:     true,
	}
	
	ctx := context.Background()
	job, err := batchProcessor.CreateBatchJob(ctx, updates, jobConfig)
	if err != nil {
		return fmt.Errorf("failed to create batch job: %w", err)
	}
	
	logger.Info("Batch job created successfully: %s", job.ID)
	
	// Start processing in background
	go func() {
		if err := batchProcessor.ProcessBatchJob(ctx, job.ID); err != nil {
			logger.Error("Batch job processing failed: %v", err)
		}
	}()
	
	// Output job details
	jobJSON, _ := json.MarshalIndent(job, "", "  ")
	fmt.Println(string(jobJSON))
	
	return nil
}

func runGitHubBatchStatus(cmd *cobra.Command, args []string) error {
	logger.Info("Checking batch job status")
	
	if len(args) == 0 {
		return fmt.Errorf("job ID is required")
	}
	
	jobID := args[0]
	
	// Load GitHub configuration and create batch processor
	config := github.LoadConfigFromEnv()
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	batchConfig := &github.BatchConfig{}
	batchProcessor := github.NewBatchProcessor(
		manager.GetClient(),
		manager.GetPatchGenerator(),
		manager.GetApplicator(),
		manager.GetPRManager(),
		batchConfig,
	)
	
	job, err := batchProcessor.GetBatchJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to get batch job: %w", err)
	}
	
	// Output job status
	statusJSON, _ := json.MarshalIndent(job, "", "  ")
	fmt.Println(string(statusJSON))
	
	return nil
}

func runGitHubAnalyticsReport(cmd *cobra.Command, args []string) error {
	logger.Info("Generating analytics report")
	
	repository := "all"
	if len(args) > 0 {
		repository = args[0]
	}
	
	// Load GitHub configuration
	config := github.LoadConfigFromEnv()
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Create analytics manager
	analyticsConfig := &github.AnalyticsConfig{
		RetentionPeriod:   30 * 24 * time.Hour,
		ReportingInterval: time.Hour,
		MetricsEnabled:    true,
		ExportFormats:     []string{"json", "csv"},
		DashboardEnabled:  true,
		AlertThresholds: &github.AlertThresholds{
			SuccessRateMin:    80.0,
			FailureRateMax:    20.0,
			ProcessingTimeMax: 10 * time.Minute,
			ConflictRateMax:   15.0,
		},
	}
	
	analyticsManager := github.NewAnalyticsManager(manager.GetClient(), analyticsConfig)
	
	// Generate report
	timeRange := &github.TimeRange{
		StartTime: time.Now().AddDate(0, 0, -30),
		EndTime:   time.Now(),
		Duration:  30 * 24 * time.Hour,
	}
	
	ctx := context.Background()
	report, err := analyticsManager.GenerateRepositoryReport(ctx, repository, timeRange)
	if err != nil {
		return fmt.Errorf("failed to generate analytics report: %w", err)
	}
	
	// Output report
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(reportJSON))
	
	return nil
}

func runGitHubPolicyList(cmd *cobra.Command, args []string) error {
	logger.Info("Listing organization policies")
	
	// Load GitHub configuration
	config := github.LoadConfigFromEnv()
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Create policy manager
	policyConfig := &github.PolicyConfig{
		EnablePolicyEnforcement: true,
		DefaultRiskTolerance:    "medium",
		RequireApprovalFor:      []string{"high", "critical"},
		BlockedDependencies:     []string{},
		AllowedLicenses:         []string{"MIT", "Apache-2.0", "BSD-3-Clause"},
		SecurityScanRequired:    true,
		ComplianceMode:          "strict",
		PolicyViolationAction:   "block",
	}
	
	policyManager := github.NewPolicyManager(manager.GetClient(), policyConfig)
	
	// Get policies and rules
	policies := policyManager.GetPolicies()
	rules := policyManager.GetPatchRules()
	
	// Output policies and rules
	output := map[string]interface{}{
		"policies": policies,
		"rules":    rules,
	}
	
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputJSON))
	
	return nil
}

func runGitHubApprovalStatus(cmd *cobra.Command, args []string) error {
	logger.Info("Checking approval workflow status")
	
	if len(args) == 0 {
		return fmt.Errorf("workflow ID is required")
	}
	
	workflowID := args[0]
	
	// Load GitHub configuration
	config := github.LoadConfigFromEnv()
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Create approval workflow manager
	approvalConfig := &github.ApprovalConfig{
		RequiredApprovals:     2,
		RequireOwnerApproval:  true,
		RequireSecurityReview: true,
		ApprovalRules:         []*github.ApprovalRule{},
		EscalationRules:       []*github.EscalationRule{},
		NotificationChannels:  []string{"email", "slack"},
		TimeoutDuration:       24 * time.Hour,
		AutoApprovalRules:     []*github.AutoApprovalRule{},
	}
	
	approvalManager := github.NewApprovalWorkflowManager(manager.GetClient(), approvalConfig)
	
	// Simulate workflow status (in real implementation, would query database)
	workflow := &github.ApprovalWorkflow{
		ID:                workflowID,
		PullRequestID:     123,
		WorkflowType:      "dependency_update",
		Status:            "pending",
		RequiredApprovals: 2,
		ReceivedApprovals: 1,
		StartedAt:         time.Now().Add(-2 * time.Hour),
		DeadlineAt:        time.Now().Add(22 * time.Hour),
	}
	
	// Output workflow status
	workflowJSON, _ := json.MarshalIndent(workflow, "", "  ")
	fmt.Println(string(workflowJSON))
	
	return nil
}

func runGitHubStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	config := github.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	ctx := context.Background()
	
	if len(args) > 0 {
		// Show status for specific repository
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
		}
		
		status, err := manager.GetRepositoryStatus(ctx, parts[0], parts[1])
		if err != nil {
			return fmt.Errorf("failed to get repository status: %w", err)
		}
		
		return displayRepositoryStatus(status, githubOutputFormat)
	}
	
	// Show status for all repositories
	repos := config.GetEnabledRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured for GitHub integration")
		return nil
	}
	
	fmt.Printf("GitHub Integration Status (%d repositories)\n\n", len(repos))
	
	for _, repo := range repos {
		status, err := manager.GetRepositoryStatus(ctx, repo.Owner, repo.Name)
		if err != nil {
			logger.Error("Failed to get status for %s/%s: %v", repo.Owner, repo.Name, err)
			continue
		}
		
		if err := displayRepositoryStatus(status, githubOutputFormat); err != nil {
			logger.Error("Failed to display status for %s/%s: %v", repo.Owner, repo.Name, err)
		}
		fmt.Println()
	}
	
	return nil
}

func runGitHubWebhookStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	config := github.LoadConfigFromEnv()
	if cmd.Flags().Changed("port") {
		config.WebhookPort = githubWebhookPort
	}
	
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	// Initialize and start webhook server
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	if err := manager.StartWebhookServer(); err != nil {
		return fmt.Errorf("failed to start webhook server: %w", err)
	}
	
	fmt.Printf("🚀 Webhook server started on port %d\n", config.WebhookPort)
	fmt.Println("Press Ctrl+C to stop the server")
	
	// Wait for interrupt signal
	select {}
}

func runGitHubWebhookStop(cmd *cobra.Command, args []string) error {
	// This would typically connect to a running webhook server to stop it
	// For now, we'll just show a message
	fmt.Println("Webhook server stop command - would stop running webhook server")
	return nil
}

func runGitHubWebhookList(cmd *cobra.Command, args []string) error {
	parts := strings.Split(args[0], "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
	}
	
	// Load configuration
	config := github.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	// List webhooks
	webhooks, err := manager.webhooks.List(ctx, parts[0], parts[1])
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}
	
	if len(webhooks) == 0 {
		fmt.Printf("No webhooks configured for %s\n", args[0])
		return nil
	}
	
	fmt.Printf("Webhooks for %s:\n\n", args[0])
	for _, webhook := range webhooks {
		fmt.Printf("ID: %d\n", webhook.ID)
		fmt.Printf("Name: %s\n", webhook.Name)
		fmt.Printf("URL: %s\n", webhook.Config.URL)
		fmt.Printf("Events: %v\n", webhook.Events)
		fmt.Printf("Active: %t\n", webhook.Active)
		fmt.Printf("Created: %s\n", webhook.CreatedAt.Format(time.RFC3339))
		fmt.Println()
	}
	
	return nil
}

func runGitHubPatchCreate(cmd *cobra.Command, args []string) error {
	parts := strings.Split(args[0], "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
	}
	
	// Parse dependencies
	var dependencies []*github.DependencyUpdate
	for _, dep := range githubDependencies {
		depParts := strings.Split(dep, ":")
		if len(depParts) != 2 {
			return fmt.Errorf("invalid dependency format: %s (expected name:version)", dep)
		}
		
		dependencies = append(dependencies, &github.DependencyUpdate{
			Name:           depParts[0],
			LatestVersion:  depParts[1],
			UpdateType:     githubUpdateType,
			BreakingChange: githubUpdateType == "major",
			SecurityFix:    githubUpdateType == "security",
		})
	}
	
	if len(dependencies) == 0 {
		return fmt.Errorf("no dependencies specified")
	}
	
	// Load configuration
	config := github.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	// Generate default values if not provided
	if githubCommitMessage == "" {
		if len(dependencies) == 1 {
			dep := dependencies[0]
			githubCommitMessage = fmt.Sprintf("Update %s to %s", dep.Name, dep.LatestVersion)
		} else {
			githubCommitMessage = fmt.Sprintf("Update %d dependencies", len(dependencies))
		}
	}
	
	if githubPRTitle == "" {
		githubPRTitle = githubCommitMessage
	}
	
	if githubPRDescription == "" {
		githubPRDescription = fmt.Sprintf("Automated dependency update created by AI Dependency Manager.\n\nDependencies updated:\n")
		for _, dep := range dependencies {
			githubPRDescription += fmt.Sprintf("- %s: %s\n", dep.Name, dep.LatestVersion)
		}
	}
	
	// Create patch request
	request := &github.PatchRequest{
		Repository:    args[0],
		Dependencies:  dependencies,
		UpdateType:    githubUpdateType,
		BaseBranch:    githubBaseBranch,
		PatchContent:  githubPatchContent,
		CommitMessage: githubCommitMessage,
		PRTitle:       githubPRTitle,
		PRDescription: githubPRDescription,
		Reviewers:     githubReviewers,
		Labels:        githubLabels,
		AutoMerge:     githubAutoMerge,
	}
	
	// Create patch
	result, err := manager.CreatePatch(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to create patch: %w", err)
	}
	
	if !result.Success {
		return fmt.Errorf("patch creation failed: %s", result.Error)
	}
	
	fmt.Println("✅ Patch created successfully!")
	fmt.Printf("   Repository: %s\n", result.Repository)
	fmt.Printf("   Branch: %s\n", result.BranchName)
	if result.PullRequest != nil {
		fmt.Printf("   Pull Request: #%d\n", result.PullRequest.Number)
		fmt.Printf("   URL: %s\n", result.PullRequest.HTMLURL)
	}
	
	return nil
}

func runGitHubPatchList(cmd *cobra.Command, args []string) error {
	// Load configuration
	config := github.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	if len(args) > 0 {
		// List patches for specific repository
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format: %s (expected owner/repo)", args[0])
		}
		
		prs, err := manager.pullRequests.ListPatchPRs(ctx, parts[0], parts[1], config.PatchBranchPrefix)
		if err != nil {
			return fmt.Errorf("failed to list patch PRs: %w", err)
		}
		
		return displayPatchPRs(args[0], prs, githubOutputFormat)
	}
	
	// List patches for all repositories
	repos := config.GetEnabledRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured for GitHub integration")
		return nil
	}
	
	for _, repo := range repos {
		repoName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
		prs, err := manager.pullRequests.ListPatchPRs(ctx, repo.Owner, repo.Name, config.PatchBranchPrefix)
		if err != nil {
			logger.Error("Failed to list patch PRs for %s: %v", repoName, err)
			continue
		}
		
		if err := displayPatchPRs(repoName, prs, githubOutputFormat); err != nil {
			logger.Error("Failed to display patch PRs for %s: %v", repoName, err)
		}
	}
	
	return nil
}

func runGitHubCleanup(cmd *cobra.Command, args []string) error {
	// Load configuration
	config := github.LoadConfigFromEnv()
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Create GitHub manager
	manager, err := github.NewManager(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub manager: %w", err)
	}
	
	ctx := context.Background()
	if err := manager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize GitHub integration: %w", err)
	}
	
	// Run cleanup
	if err := manager.CleanupOldBranches(ctx); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}
	
	fmt.Println("✅ Cleanup completed successfully!")
	
	return nil
}

func displayRepositoryStatus(status *github.RepositoryStatus, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		
	case "table":
		fmt.Printf("Repository: %s\n", status.Repository.FullName)
		fmt.Printf("Enabled: %t\n", status.Enabled)
		fmt.Printf("Webhook Active: %t\n", status.WebhookActive)
		fmt.Printf("Open Patch PRs: %d\n", len(status.PatchPRs))
		fmt.Printf("Patch Branches: %d\n", len(status.PatchBranches))
		fmt.Printf("Last Checked: %s\n", status.LastChecked.Format(time.RFC3339))
		
		if len(status.PatchPRs) > 0 {
			fmt.Println("\nOpen Patch PRs:")
			for _, pr := range status.PatchPRs {
				fmt.Printf("  #%d: %s (%s)\n", pr.Number, pr.Title, pr.State)
			}
		}
		
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	
	return nil
}

func displayPatchPRs(repository string, prs []*github.PullRequest, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(prs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		
	case "table":
		if len(prs) == 0 {
			fmt.Printf("No patch PRs found for %s\n", repository)
			return nil
		}
		
		fmt.Printf("Patch PRs for %s:\n\n", repository)
		for _, pr := range prs {
			fmt.Printf("#%d: %s\n", pr.Number, pr.Title)
			fmt.Printf("  State: %s\n", pr.State)
			fmt.Printf("  Branch: %s\n", pr.Head.Ref)
			fmt.Printf("  Created: %s\n", pr.CreatedAt.Format(time.RFC3339))
			fmt.Printf("  URL: %s\n", pr.HTMLURL)
			fmt.Println()
		}
		
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	
	return nil
}
