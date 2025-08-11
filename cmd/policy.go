package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/services"
	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage custom update policies",
	Long:  `Create, manage, and apply custom update policies to control how dependencies are updated across projects.`,
}

var createPolicyCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new update policy",
	Long:  `Create a new custom update policy with specified conditions and actions.`,
	Args:  cobra.ExactArgs(1),
	Run:   runCreatePolicy,
}

var listPoliciesCmd = &cobra.Command{
	Use:   "list",
	Short: "List update policies",
	Long:  `List all configured update policies, optionally filtered by project.`,
	Run:   runListPolicies,
}

var showPolicyCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show policy details",
	Long:  `Show detailed information about a specific update policy.`,
	Args:  cobra.ExactArgs(1),
	Run:   runShowPolicy,
}

var updatePolicyCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an existing policy",
	Long:  `Update an existing update policy's configuration.`,
	Args:  cobra.ExactArgs(1),
	Run:   runUpdatePolicy,
}

var deletePolicyCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a policy",
	Long:  `Delete an existing update policy.`,
	Args:  cobra.ExactArgs(1),
	Run:   runDeletePolicy,
}

var testPolicyCmd = &cobra.Command{
	Use:   "test [id] [package-name]",
	Short: "Test a policy against a package",
	Long:  `Test how a policy would evaluate against a specific package update.`,
	Args:  cobra.ExactArgs(2),
	Run:   runTestPolicy,
}

var templatePolicyCmd = &cobra.Command{
	Use:   "template [type]",
	Short: "Generate policy templates",
	Long: `Generate policy templates for common scenarios. Available types:
- security: Block updates with security risks
- conservative: Require approval for major updates
- aggressive: Auto-update everything
- business-hours: Only update during business hours
- staged: Staged rollout policy`,
	Args: cobra.ExactArgs(1),
	Run:  runTemplatePolicyCmd,
}

// Global variables for policy command flags
var (
	policyProject     string
	policyFormat      string
	policyEnabled     bool
	policyPriority    int
	policyDescription string
	policyAuthor      string
	policyTags        string
	policyForce       bool
)

func init() {
	rootCmd.AddCommand(policyCmd)
	policyCmd.AddCommand(createPolicyCmd)
	policyCmd.AddCommand(listPoliciesCmd)
	policyCmd.AddCommand(showPolicyCmd)
	policyCmd.AddCommand(updatePolicyCmd)
	policyCmd.AddCommand(deletePolicyCmd)
	policyCmd.AddCommand(testPolicyCmd)
	policyCmd.AddCommand(templatePolicyCmd)

	// Create policy flags
	createPolicyCmd.Flags().StringVar(&policyProject, "project", "", "Project ID or name to apply policy to")
	createPolicyCmd.Flags().StringVar(&policyDescription, "description", "", "Policy description")
	createPolicyCmd.Flags().IntVar(&policyPriority, "priority", 50, "Policy priority (0-100)")
	createPolicyCmd.Flags().BoolVar(&policyEnabled, "enabled", true, "Enable policy immediately")
	createPolicyCmd.Flags().StringVar(&policyAuthor, "author", "", "Policy author")
	createPolicyCmd.Flags().StringVar(&policyTags, "tags", "", "Comma-separated tags")

	// List policies flags
	listPoliciesCmd.Flags().StringVar(&policyProject, "project", "", "Filter by project")
	listPoliciesCmd.Flags().StringVar(&policyFormat, "format", "table", "Output format (table, json)")

	// Show policy flags
	showPolicyCmd.Flags().StringVar(&policyFormat, "format", "yaml", "Output format (yaml, json)")

	// Delete policy flags
	deletePolicyCmd.Flags().BoolVar(&policyForce, "force", false, "Force deletion without confirmation")
}

func runCreatePolicy(cmd *cobra.Command, args []string) {
	policyName := args[0]
	
	fmt.Printf("🔧 Creating Update Policy: %s\n", policyName)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	// Interactive policy creation
	policy := &services.UpdatePolicy{
		Name:        policyName,
		Description: policyDescription,
		Priority:    policyPriority,
		Enabled:     policyEnabled,
		Author:      policyAuthor,
		Tags:        policyTags,
		Version:     "1.0",
	}
	
	// Parse project if provided
	if policyProject != "" {
		// In a real implementation, we'd look up the project ID
		// For now, we'll assume it's a numeric ID
		if projectID, err := strconv.ParseUint(policyProject, 10, 32); err == nil {
			projectIDVal := uint(projectID)
			policy.ProjectID = &projectIDVal
		}
	}
	
	// Interactive condition setup
	fmt.Println("📋 Policy Conditions Setup")
	fmt.Println("Configure when this policy should be applied:")
	fmt.Println()
	
	conditions := services.PolicyConditions{}
	
	// Package matching
	fmt.Print("Package names (comma-separated, or press Enter to skip): ")
	var packageNames string
	fmt.Scanln(&packageNames)
	if packageNames != "" {
		conditions.PackageNames = strings.Split(packageNames, ",")
		for i := range conditions.PackageNames {
			conditions.PackageNames[i] = strings.TrimSpace(conditions.PackageNames[i])
		}
	}
	
	// Update types
	fmt.Print("Update types (security,major,minor,patch - comma-separated, or press Enter to skip): ")
	var updateTypes string
	fmt.Scanln(&updateTypes)
	if updateTypes != "" {
		conditions.UpdateTypes = strings.Split(updateTypes, ",")
		for i := range conditions.UpdateTypes {
			conditions.UpdateTypes[i] = strings.TrimSpace(conditions.UpdateTypes[i])
		}
	}
	
	// Security risk
	fmt.Print("Security risk filter (true/false/skip): ")
	var securityRisk string
	fmt.Scanln(&securityRisk)
	if securityRisk == "true" {
		securityRiskBool := true
		conditions.SecurityRisk = &securityRiskBool
	} else if securityRisk == "false" {
		securityRiskBool := false
		conditions.SecurityRisk = &securityRiskBool
	}
	
	policy.Conditions = conditions
	
	fmt.Println()
	fmt.Println("⚙️ Policy Actions Setup")
	fmt.Println("Configure what actions to take when conditions are met:")
	fmt.Println()
	
	actions := services.PolicyActions{}
	
	// Auto update
	fmt.Print("Auto-update (true/false/skip): ")
	var autoUpdate string
	fmt.Scanln(&autoUpdate)
	if autoUpdate == "true" {
		autoUpdateBool := true
		actions.AutoUpdate = &autoUpdateBool
	} else if autoUpdate == "false" {
		autoUpdateBool := false
		actions.AutoUpdate = &autoUpdateBool
	}
	
	// Require approval
	fmt.Print("Require approval (true/false/skip): ")
	var requireApproval string
	fmt.Scanln(&requireApproval)
	if requireApproval == "true" {
		requireApprovalBool := true
		actions.RequireApproval = &requireApprovalBool
	} else if requireApproval == "false" {
		requireApprovalBool := false
		actions.RequireApproval = &requireApprovalBool
	}
	
	// Block update
	fmt.Print("Block update (true/false/skip): ")
	var blockUpdate string
	fmt.Scanln(&blockUpdate)
	if blockUpdate == "true" {
		blockUpdateBool := true
		actions.BlockUpdate = &blockUpdateBool
	} else if blockUpdate == "false" {
		blockUpdateBool := false
		actions.BlockUpdate = &blockUpdateBool
	}
	
	// Notification channels
	fmt.Print("Notification channels (email,slack,webhook - comma-separated, or press Enter to skip): ")
	var notifyChannels string
	fmt.Scanln(&notifyChannels)
	if notifyChannels != "" {
		actions.NotifyChannels = strings.Split(notifyChannels, ",")
		for i := range actions.NotifyChannels {
			actions.NotifyChannels[i] = strings.TrimSpace(actions.NotifyChannels[i])
		}
	}
	
	policy.Actions = actions
	
	// Create the policy
	policyService := services.NewPolicyService()
	if err := policyService.CreatePolicy(context.Background(), policy); err != nil {
		logger.Error("Failed to create policy: %v", err)
		os.Exit(1)
	}
	
	fmt.Println()
	fmt.Printf("✅ Policy '%s' created successfully with ID %d\n", policy.Name, policy.ID)
	fmt.Printf("Priority: %d | Enabled: %v\n", policy.Priority, policy.Enabled)
	
	if len(conditions.PackageNames) > 0 {
		fmt.Printf("Applies to packages: %s\n", strings.Join(conditions.PackageNames, ", "))
	}
	
	if len(conditions.UpdateTypes) > 0 {
		fmt.Printf("Update types: %s\n", strings.Join(conditions.UpdateTypes, ", "))
	}
}

func runListPolicies(cmd *cobra.Command, args []string) {
	var projectID *uint
	
	if policyProject != "" {
		if id, err := strconv.ParseUint(policyProject, 10, 32); err == nil {
			projectIDVal := uint(id)
			projectID = &projectIDVal
		}
	}
	
	policyService := services.NewPolicyService()
	policies, err := policyService.ListPolicies(context.Background(), projectID)
	if err != nil {
		logger.Error("Failed to list policies: %v", err)
		os.Exit(1)
	}
	
	if policyFormat == "json" {
		printJSON(policies)
		return
	}
	
	// Display in table format
	fmt.Println("📋 Update Policies")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	
	if len(policies) == 0 {
		fmt.Println("No policies configured.")
		fmt.Println()
		fmt.Println("💡 Use 'ai-dep-manager policy create [name]' to create a policy")
		fmt.Println("💡 Use 'ai-dep-manager policy template [type]' to see policy templates")
		return
	}
	
	fmt.Printf("%-4s %-20s %-8s %-8s %-10s %s\n", 
		"ID", "Name", "Priority", "Enabled", "Project", "Description")
	fmt.Println(strings.Repeat("-", 70))
	
	for _, policy := range policies {
		projectStr := "Global"
		if policy.ProjectID != nil {
			projectStr = fmt.Sprintf("%d", *policy.ProjectID)
		}
		
		enabledStr := "✅"
		if !policy.Enabled {
			enabledStr = "❌"
		}
		
		fmt.Printf("%-4d %-20s %-8d %-8s %-10s %s\n",
			policy.ID,
			truncateString(policy.Name, 20),
			policy.Priority,
			enabledStr,
			projectStr,
			truncateString(policy.Description, 30),
		)
	}
	
	fmt.Println()
	fmt.Printf("Total: %d policies\n", len(policies))
	fmt.Println()
	fmt.Println("💡 Use 'ai-dep-manager policy show [id]' to view policy details")
}

func runShowPolicy(cmd *cobra.Command, args []string) {
	policyID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid policy ID: %v", err)
		os.Exit(1)
	}
	
	policyService := services.NewPolicyService()
	policy, err := policyService.GetPolicy(context.Background(), uint(policyID))
	if err != nil {
		logger.Error("Failed to get policy: %v", err)
		os.Exit(1)
	}
	
	if policyFormat == "json" {
		printJSON(policy)
		return
	}
	
	// Display policy details
	fmt.Printf("📋 Policy Details: %s\n", policy.Name)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("ID: %d\n", policy.ID)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Priority: %d\n", policy.Priority)
	fmt.Printf("Enabled: %v\n", policy.Enabled)
	fmt.Printf("Author: %s\n", policy.Author)
	fmt.Printf("Version: %s\n", policy.Version)
	fmt.Printf("Created: %s\n", policy.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", policy.UpdatedAt.Format("2006-01-02 15:04:05"))
	
	if policy.ProjectID != nil {
		fmt.Printf("Project: %d\n", *policy.ProjectID)
	} else {
		fmt.Printf("Project: Global\n")
	}
	
	if policy.Tags != "" {
		fmt.Printf("Tags: %s\n", policy.Tags)
	}
	
	fmt.Println()
	
	// Display conditions
	fmt.Println("📋 Conditions:")
	conditions := policy.Conditions
	
	if len(conditions.PackageNames) > 0 {
		fmt.Printf("  Package Names: %s\n", strings.Join(conditions.PackageNames, ", "))
	}
	
	if len(conditions.PackagePatterns) > 0 {
		fmt.Printf("  Package Patterns: %s\n", strings.Join(conditions.PackagePatterns, ", "))
	}
	
	if len(conditions.UpdateTypes) > 0 {
		fmt.Printf("  Update Types: %s\n", strings.Join(conditions.UpdateTypes, ", "))
	}
	
	if conditions.SecurityRisk != nil {
		fmt.Printf("  Security Risk: %v\n", *conditions.SecurityRisk)
	}
	
	if conditions.BreakingChange != nil {
		fmt.Printf("  Breaking Change: %v\n", *conditions.BreakingChange)
	}
	
	if conditions.RiskScoreMin != nil || conditions.RiskScoreMax != nil {
		if conditions.RiskScoreMin != nil && conditions.RiskScoreMax != nil {
			fmt.Printf("  Risk Score Range: %.1f - %.1f\n", *conditions.RiskScoreMin, *conditions.RiskScoreMax)
		} else if conditions.RiskScoreMin != nil {
			fmt.Printf("  Risk Score Min: %.1f\n", *conditions.RiskScoreMin)
		} else {
			fmt.Printf("  Risk Score Max: %.1f\n", *conditions.RiskScoreMax)
		}
	}
	
	fmt.Println()
	
	// Display actions
	fmt.Println("⚙️ Actions:")
	actions := policy.Actions
	
	if actions.AutoUpdate != nil {
		fmt.Printf("  Auto Update: %v\n", *actions.AutoUpdate)
	}
	
	if actions.RequireApproval != nil {
		fmt.Printf("  Require Approval: %v\n", *actions.RequireApproval)
	}
	
	if actions.BlockUpdate != nil {
		fmt.Printf("  Block Update: %v\n", *actions.BlockUpdate)
	}
	
	if actions.UpdateStrategy != "" {
		fmt.Printf("  Update Strategy: %s\n", actions.UpdateStrategy)
	}
	
	if actions.DelayDays > 0 {
		fmt.Printf("  Delay Days: %d\n", actions.DelayDays)
	}
	
	if len(actions.NotifyChannels) > 0 {
		fmt.Printf("  Notification Channels: %s\n", strings.Join(actions.NotifyChannels, ", "))
	}
	
	if actions.NotifyLevel != "" {
		fmt.Printf("  Notification Level: %s\n", actions.NotifyLevel)
	}
	
	if actions.RunTests != nil {
		fmt.Printf("  Run Tests: %v\n", *actions.RunTests)
	}
	
	if actions.RollbackOnFail != nil {
		fmt.Printf("  Rollback On Fail: %v\n", *actions.RollbackOnFail)
	}
}

func runUpdatePolicy(cmd *cobra.Command, args []string) {
	policyID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid policy ID: %v", err)
		os.Exit(1)
	}
	
	policyService := services.NewPolicyService()
	
	// Get existing policy
	policy, err := policyService.GetPolicy(context.Background(), uint(policyID))
	if err != nil {
		logger.Error("Failed to get policy: %v", err)
		os.Exit(1)
	}
	
	fmt.Printf("🔧 Updating Policy: %s (ID: %d)\n", policy.Name, policy.ID)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Press Enter to keep current value, or enter new value:")
	fmt.Println()
	
	// Update basic fields
	fmt.Printf("Description [%s]: ", policy.Description)
	var newDescription string
	fmt.Scanln(&newDescription)
	if newDescription != "" {
		policy.Description = newDescription
	}
	
	fmt.Printf("Priority [%d]: ", policy.Priority)
	var newPriority string
	fmt.Scanln(&newPriority)
	if newPriority != "" {
		if priority, err := strconv.Atoi(newPriority); err == nil {
			policy.Priority = priority
		}
	}
	
	fmt.Printf("Enabled [%v]: ", policy.Enabled)
	var newEnabled string
	fmt.Scanln(&newEnabled)
	if newEnabled != "" {
		policy.Enabled = strings.ToLower(newEnabled) == "true"
	}
	
	// Update the policy
	if err := policyService.UpdatePolicy(context.Background(), policy); err != nil {
		logger.Error("Failed to update policy: %v", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Policy '%s' updated successfully\n", policy.Name)
}

func runDeletePolicy(cmd *cobra.Command, args []string) {
	policyID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid policy ID: %v", err)
		os.Exit(1)
	}
	
	if !policyForce {
		fmt.Printf("Are you sure you want to delete policy %d? (y/N): ", policyID)
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled.")
			return
		}
	}
	
	policyService := services.NewPolicyService()
	if err := policyService.DeletePolicy(context.Background(), uint(policyID)); err != nil {
		logger.Error("Failed to delete policy: %v", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Policy %d deleted successfully\n", policyID)
}

func runTestPolicy(cmd *cobra.Command, args []string) {
	policyID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		logger.Error("Invalid policy ID: %v", err)
		os.Exit(1)
	}
	
	packageName := args[1]
	
	fmt.Printf("🧪 Testing Policy %d against package: %s\n", policyID, packageName)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	// This would require implementing the test logic
	// For now, we'll show a placeholder
	fmt.Println("Policy testing functionality would:")
	fmt.Println("1. Load the specified policy")
	fmt.Println("2. Create a mock dependency and update for the package")
	fmt.Println("3. Evaluate the policy conditions")
	fmt.Println("4. Show the resulting actions and decision")
	fmt.Println()
	fmt.Println("📋 Mock Results:")
	fmt.Printf("  Package: %s\n", packageName)
	fmt.Printf("  Policy: %d\n", policyID)
	fmt.Printf("  Decision: allow\n")
	fmt.Printf("  Explanation: Policy conditions not met\n")
}

func runTemplatePolicyCmd(cmd *cobra.Command, args []string) {
	templateType := args[0]
	
	fmt.Printf("📄 Policy Template: %s\n", templateType)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	
	switch templateType {
	case "security":
		showSecurityPolicyTemplate()
	case "conservative":
		showConservativePolicyTemplate()
	case "aggressive":
		showAggressivePolicyTemplate()
	case "business-hours":
		showBusinessHoursPolicyTemplate()
	case "staged":
		showStagedPolicyTemplate()
	default:
		logger.Error("Unknown template type: %s", templateType)
		fmt.Println("Available templates: security, conservative, aggressive, business-hours, staged")
		os.Exit(1)
	}
}

// Template display functions

func showSecurityPolicyTemplate() {
	fmt.Println("🚨 Security Policy Template")
	fmt.Println("Automatically applies security updates while blocking risky changes")
	fmt.Println()
	fmt.Println("Conditions:")
	fmt.Println("  - Update types: security")
	fmt.Println("  - Security risk: true")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  - Auto update: true (for security updates)")
	fmt.Println("  - Block update: true (for non-security with security risks)")
	fmt.Println("  - Notify channels: email, slack")
	fmt.Println("  - Run tests: true")
	fmt.Println("  - Rollback on fail: true")
	fmt.Println()
	fmt.Println("💡 Use: ai-dep-manager policy create security-policy --description 'Auto-apply security updates'")
}

func showConservativePolicyTemplate() {
	fmt.Println("🛡️ Conservative Policy Template")
	fmt.Println("Requires approval for major updates, auto-applies patches")
	fmt.Println()
	fmt.Println("Conditions:")
	fmt.Println("  - Update types: major, minor")
	fmt.Println("  - Risk score max: 5.0")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  - Require approval: true (for major/minor)")
	fmt.Println("  - Auto update: true (for patches)")
	fmt.Println("  - Update strategy: conservative")
	fmt.Println("  - Delay days: 7")
	fmt.Println("  - Run tests: true")
	fmt.Println()
	fmt.Println("💡 Use: ai-dep-manager policy create conservative-policy --description 'Conservative update approach'")
}

func showAggressivePolicyTemplate() {
	fmt.Println("🚀 Aggressive Policy Template")
	fmt.Println("Auto-updates everything with minimal restrictions")
	fmt.Println()
	fmt.Println("Conditions:")
	fmt.Println("  - All update types")
	fmt.Println("  - Risk score max: 8.0")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  - Auto update: true")
	fmt.Println("  - Update strategy: aggressive")
	fmt.Println("  - Batch size: 10")
	fmt.Println("  - Run tests: true")
	fmt.Println("  - Rollback on fail: true")
	fmt.Println()
	fmt.Println("💡 Use: ai-dep-manager policy create aggressive-policy --description 'Aggressive auto-updates'")
}

func showBusinessHoursPolicyTemplate() {
	fmt.Println("🕒 Business Hours Policy Template")
	fmt.Println("Only applies updates during business hours")
	fmt.Println()
	fmt.Println("Conditions:")
	fmt.Println("  - Time window: Monday-Friday, 09:00-17:00")
	fmt.Println("  - All update types")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  - Auto update: true")
	fmt.Println("  - Schedule: 0 9-17 * * 1-5")
	fmt.Println("  - Notify channels: slack")
	fmt.Println("  - Run tests: true")
	fmt.Println()
	fmt.Println("💡 Use: ai-dep-manager policy create business-hours --description 'Updates during business hours only'")
}

func showStagedPolicyTemplate() {
	fmt.Println("📊 Staged Rollout Policy Template")
	fmt.Println("Gradually rolls out updates in batches")
	fmt.Println()
	fmt.Println("Conditions:")
	fmt.Println("  - Update types: minor, major")
	fmt.Println("  - Risk score max: 6.0")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  - Auto update: true")
	fmt.Println("  - Batch size: 3")
	fmt.Println("  - Batch interval: 2h")
	fmt.Println("  - Delay days: 1")
	fmt.Println("  - Run tests: true")
	fmt.Println("  - Rollback on fail: true")
	fmt.Println()
	fmt.Println("💡 Use: ai-dep-manager policy create staged-rollout --description 'Staged rollout of updates'")
}
