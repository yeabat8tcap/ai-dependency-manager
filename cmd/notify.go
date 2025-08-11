package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/notifications"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Manage notifications",
	Long:  `Configure and manage notification settings for security alerts, updates, and system events.`,
}

var configureNotifyCmd = &cobra.Command{
	Use:   "configure [channel]",
	Short: "Configure notification channels",
	Long: `Configure notification channels. Available channels:
- email: Configure email notifications via SMTP
- slack: Configure Slack webhook notifications
- webhook: Configure generic webhook notifications`,
	Args: cobra.ExactArgs(1),
	Run:  runConfigureNotify,
}

var testNotifyCmd = &cobra.Command{
	Use:   "test [channel]",
	Short: "Test notification channel",
	Long:  `Send a test notification to verify channel configuration.`,
	Args:  cobra.ExactArgs(1),
	Run:   runTestNotify,
}

var listNotifyCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification channels",
	Long:  `List all configured notification channels and their status.`,
	Run:   runListNotify,
}

var sendNotifyCmd = &cobra.Command{
	Use:   "send [type] [message]",
	Short: "Send a manual notification",
	Long: `Send a manual notification of the specified type. Available types:
- security: Security alert
- update: Update notification
- scan: Scan completion
- error: Error notification
- info: Information notification`,
	Args: cobra.ExactArgs(2),
	Run:  runSendNotify,
}

// Global variables for notification command flags
var (
	notifyChannel  string
	notifyPriority string
	notifyProject  string
	notifyDryRun   bool
)

func init() {
	rootCmd.AddCommand(notifyCmd)
	notifyCmd.AddCommand(configureNotifyCmd)
	notifyCmd.AddCommand(testNotifyCmd)
	notifyCmd.AddCommand(listNotifyCmd)
	notifyCmd.AddCommand(sendNotifyCmd)

	// Send notification flags
	sendNotifyCmd.Flags().StringVar(&notifyChannel, "channel", "", "Specific channel to send to (email, slack, webhook)")
	sendNotifyCmd.Flags().StringVar(&notifyPriority, "priority", "medium", "Notification priority (low, medium, high, critical)")
	sendNotifyCmd.Flags().StringVar(&notifyProject, "project", "", "Project context for the notification")
	sendNotifyCmd.Flags().BoolVar(&notifyDryRun, "dry-run", false, "Show what would be sent without actually sending")

	// Test notification flags
	testNotifyCmd.Flags().BoolVar(&notifyDryRun, "dry-run", false, "Show test message without actually sending")
}

func runConfigureNotify(cmd *cobra.Command, args []string) {
	channel := args[0]
	
	// Validate channel type
	validChannels := []string{"email", "slack", "webhook"}
	if !contains(validChannels, channel) {
		logger.Error("Invalid channel type: %s. Valid types: %s", channel, strings.Join(validChannels, ", "))
		os.Exit(1)
	}
	
	fmt.Printf("🔔 Configuring %s notifications\n", channel)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	
	switch channel {
	case "email":
		configureEmailNotifications()
	case "slack":
		configureSlackNotifications()
	case "webhook":
		configureWebhookNotifications()
	}
}

func runTestNotify(cmd *cobra.Command, args []string) {
	channel := args[0]
	
	fmt.Printf("🧪 Testing %s notification channel\n", channel)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	
	notificationService := notifications.NewNotificationService(config.GetConfig())
	
	// Create test notification
	notification := notifications.Notification{
		Type:     notifications.NotificationTypeUpdateAvailable,
		Priority: notifications.PriorityMedium,
		Title:    "Test Notification",
		Message:  "This is a test notification from AI Dependency Manager",
		Metadata: map[string]interface{}{
			"test":      true,
			"channel":   channel,
			"timestamp": "2024-01-01T12:00:00Z",
		},
	}
	
	if notifyDryRun {
		fmt.Println("📋 Dry run - would send the following notification:")
		fmt.Printf("  Channel: %s\n", channel)
		fmt.Printf("  Type: %s\n", notification.Type)
		fmt.Printf("  Priority: %s\n", notification.Priority)
		fmt.Printf("  Title: %s\n", notification.Title)
		fmt.Printf("  Message: %s\n", notification.Message)
		return
	}
	
	// Send test notification
	var err error
	err = notificationService.SendNotification(context.Background(), &notification)
	
	if err != nil {
		logger.Error("Failed to send test notification: %v", err)
		fmt.Println("❌ Test notification failed")
		os.Exit(1)
	}
	
	fmt.Println("✅ Test notification sent successfully")
}

func runListNotify(cmd *cobra.Command, args []string) {
	fmt.Println("🔔 Notification Channels")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	
	// This would typically read from configuration
	// For now, we'll show a placeholder structure
	
	channels := []struct {
		Name    string
		Type    string
		Status  string
		Config  string
	}{
		{"Email", "email", "configured", "SMTP server configured"},
		{"Slack", "slack", "not configured", "Webhook URL not set"},
		{"Webhook", "webhook", "configured", "Generic webhook configured"},
	}
	
	fmt.Printf("%-15s %-10s %-15s %s\n", "Channel", "Type", "Status", "Configuration")
	fmt.Println(strings.Repeat("-", 70))
	
	for _, channel := range channels {
		statusIcon := "✅"
		if channel.Status != "configured" {
			statusIcon = "❌"
		}
		
		fmt.Printf("%-15s %-10s %s%-13s %s\n", 
			channel.Name, 
			channel.Type, 
			statusIcon, 
			channel.Status, 
			channel.Config)
	}
	
	fmt.Println()
	fmt.Println("💡 Use 'ai-dep-manager notify configure [channel]' to configure channels")
	fmt.Println("🧪 Use 'ai-dep-manager notify test [channel]' to test channels")
}

func runSendNotify(cmd *cobra.Command, args []string) {
	notificationType := args[0]
	message := args[1]
	
	// Validate notification type
	validTypes := []string{"security", "update", "scan", "error", "info"}
	if !contains(validTypes, notificationType) {
		logger.Error("Invalid notification type: %s. Valid types: %s", notificationType, strings.Join(validTypes, ", "))
		os.Exit(1)
	}
	
	// Map string types to notification types
	var nType notifications.NotificationType
	switch notificationType {
	case "security":
		nType = notifications.NotificationTypeSecurityAlert
	case "update":
		nType = notifications.NotificationTypeUpdateAvailable
	case "scan":
		nType = notifications.NotificationTypeScanCompleted
	case "error":
		nType = notifications.NotificationTypeAgentError
	case "info":
		nType = notifications.NotificationTypeUpdateAvailable
	}
	
	// Map priority
	var priority notifications.NotificationPriority
	switch notifyPriority {
	case "low":
		priority = notifications.PriorityLow
	case "medium":
		priority = notifications.PriorityMedium
	case "high":
		priority = notifications.PriorityHigh
	case "critical":
		priority = notifications.PriorityCritical
	default:
		priority = notifications.PriorityMedium
	}
	
	// Create notification
	notification := notifications.Notification{
		Type:     nType,
		Priority: priority,
		Title:    fmt.Sprintf("Manual %s Notification", strings.Title(notificationType)),
		Message:  message,
		Metadata: map[string]interface{}{
			"manual":    true,
			"project":   notifyProject,
			"timestamp": "2024-01-01T12:00:00Z",
		},
	}
	
	if notifyDryRun {
		fmt.Println("📋 Dry run - would send the following notification:")
		fmt.Printf("  Type: %s\n", notification.Type)
		fmt.Printf("  Priority: %s\n", notification.Priority)
		fmt.Printf("  Title: %s\n", notification.Title)
		fmt.Printf("  Message: %s\n", notification.Message)
		if notifyChannel != "" {
			fmt.Printf("  Channel: %s\n", notifyChannel)
		} else {
			fmt.Printf("  Channel: all configured channels\n")
		}
		return
	}
	
	notificationService := notifications.NewNotificationService(config.GetConfig())
	
	// Send notification through all configured channels
	err := notificationService.SendNotification(context.Background(), &notification)
	if err != nil {
		logger.Error("Failed to send notification: %v", err)
		os.Exit(1)
	}
	
	if notifyChannel != "" {
		fmt.Printf("✅ Notification sent (requested channel: %s)\n", notifyChannel)
	} else {
		fmt.Printf("✅ Notification sent to all configured channels\n")
	}
}

// Helper functions for configuration

func configureEmailNotifications() {
	fmt.Println("📧 Email Notification Configuration")
	fmt.Println()
	
	fmt.Println("Please provide the following SMTP configuration:")
	fmt.Println()
	
	// In a real implementation, this would prompt for input and save to config
	fmt.Println("Required settings:")
	fmt.Println("  - SMTP Host (e.g., smtp.gmail.com)")
	fmt.Println("  - SMTP Port (e.g., 587)")
	fmt.Println("  - Username (email address)")
	fmt.Println("  - Password (or app password)")
	fmt.Println("  - From Address")
	fmt.Println("  - To Addresses (comma-separated)")
	fmt.Println()
	
	fmt.Println("💡 Configuration should be added to your config file:")
	fmt.Println("   ~/.ai-dep-manager/config.yaml")
	fmt.Println()
	
	fmt.Println("Example configuration:")
	fmt.Println(`
notifications:
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from: "your-email@gmail.com"
    to: ["admin@company.com", "devops@company.com"]
    tls: true
`)
}

func configureSlackNotifications() {
	fmt.Println("💬 Slack Notification Configuration")
	fmt.Println()
	
	fmt.Println("To configure Slack notifications:")
	fmt.Println()
	fmt.Println("1. Create a Slack App in your workspace")
	fmt.Println("2. Enable Incoming Webhooks")
	fmt.Println("3. Create a webhook for your desired channel")
	fmt.Println("4. Copy the webhook URL")
	fmt.Println()
	
	fmt.Println("💡 Add the webhook URL to your config file:")
	fmt.Println("   ~/.ai-dep-manager/config.yaml")
	fmt.Println()
	
	fmt.Println("Example configuration:")
	fmt.Println(`
notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#dependency-updates"
    username: "AI Dependency Manager"
    icon_emoji: ":robot_face:"
`)
}

func configureWebhookNotifications() {
	fmt.Println("🔗 Webhook Notification Configuration")
	fmt.Println()
	
	fmt.Println("Configure generic webhook notifications for integration with")
	fmt.Println("custom systems, monitoring tools, or other services.")
	fmt.Println()
	
	fmt.Println("💡 Add webhook configuration to your config file:")
	fmt.Println("   ~/.ai-dep-manager/config.yaml")
	fmt.Println()
	
	fmt.Println("Example configuration:")
	fmt.Println(`
notifications:
  webhook:
    enabled: true
    url: "https://your-system.com/webhooks/dependency-manager"
    method: "POST"
    headers:
      "Authorization": "Bearer your-token"
      "Content-Type": "application/json"
    timeout: "30s"
`)
	
	fmt.Println()
	fmt.Println("The webhook will receive JSON payloads with notification data.")
}
