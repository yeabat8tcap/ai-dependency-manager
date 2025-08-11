package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/config"
	"github.com/8tcapital/ai-dep-manager/internal/logger"
	"github.com/8tcapital/ai-dep-manager/internal/models"
)

// NotificationService handles sending notifications
type NotificationService struct {
	config *config.Config
	client *http.Client
}

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeSecurityAlert   NotificationType = "security_alert"
	NotificationTypeUpdateAvailable NotificationType = "update_available"
	NotificationTypeUpdateApplied   NotificationType = "update_applied"
	NotificationTypeUpdateFailed    NotificationType = "update_failed"
	NotificationTypeScanCompleted   NotificationType = "scan_completed"
	NotificationTypeAgentError      NotificationType = "agent_error"
	NotificationTypeCredentialExpiry NotificationType = "credential_expiry"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityMedium   NotificationPriority = "medium"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// Notification represents a notification to be sent
type Notification struct {
	Type        NotificationType     `json:"type"`
	Priority    NotificationPriority `json:"priority"`
	Title       string               `json:"title"`
	Message     string               `json:"message"`
	ProjectName string               `json:"project_name,omitempty"`
	PackageName string               `json:"package_name,omitempty"`
	Version     string               `json:"version,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time            `json:"timestamp"`
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	Enabled    bool     `json:"enabled"`
	SMTPHost   string   `json:"smtp_host"`
	SMTPPort   int      `json:"smtp_port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	From       string   `json:"from"`
	Recipients []string `json:"recipients"`
	TLS        bool     `json:"tls"`
}

// SlackConfig holds Slack notification configuration
type SlackConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
}

// WebhookConfig holds generic webhook notification configuration
type WebhookConfig struct {
	Enabled bool              `json:"enabled"`
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(cfg *config.Config) *NotificationService {
	return &NotificationService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendNotification sends a notification through all configured channels
func (ns *NotificationService) SendNotification(ctx context.Context, notification *Notification) error {
	logger.Info("Sending notification: %s - %s", notification.Type, notification.Title)
	
	notification.Timestamp = time.Now()
	
	var errors []string
	
	// Send email notification
	if err := ns.sendEmailNotification(ctx, notification); err != nil {
		logger.Warn("Failed to send email notification: %v", err)
		errors = append(errors, fmt.Sprintf("email: %v", err))
	}
	
	// Send Slack notification
	if err := ns.sendSlackNotification(ctx, notification); err != nil {
		logger.Warn("Failed to send Slack notification: %v", err)
		errors = append(errors, fmt.Sprintf("slack: %v", err))
	}
	
	// Send webhook notification
	if err := ns.sendWebhookNotification(ctx, notification); err != nil {
		logger.Warn("Failed to send webhook notification: %v", err)
		errors = append(errors, fmt.Sprintf("webhook: %v", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("notification failures: %s", strings.Join(errors, ", "))
	}
	
	logger.Info("Notification sent successfully: %s", notification.Title)
	return nil
}

// SendSecurityAlert sends a security-related notification
func (ns *NotificationService) SendSecurityAlert(ctx context.Context, packageName, version, severity, description string) error {
	priority := ns.mapSeverityToPriority(severity)
	
	notification := &Notification{
		Type:        NotificationTypeSecurityAlert,
		Priority:    priority,
		Title:       fmt.Sprintf("🚨 Security Alert: %s@%s", packageName, version),
		Message:     fmt.Sprintf("Security issue detected in %s@%s: %s", packageName, version, description),
		PackageName: packageName,
		Version:     version,
		Metadata: map[string]interface{}{
			"severity": severity,
		},
	}
	
	return ns.SendNotification(ctx, notification)
}

// SendUpdateNotification sends an update-related notification
func (ns *NotificationService) SendUpdateNotification(ctx context.Context, notificationType NotificationType, projectName string, updates []models.Update) error {
	var title, message string
	var priority NotificationPriority = PriorityMedium
	
	switch notificationType {
	case NotificationTypeUpdateAvailable:
		title = fmt.Sprintf("📦 Updates Available: %s", projectName)
		message = fmt.Sprintf("%d update(s) available for project %s", len(updates), projectName)
		priority = PriorityLow
		
	case NotificationTypeUpdateApplied:
		title = fmt.Sprintf("✅ Updates Applied: %s", projectName)
		message = fmt.Sprintf("%d update(s) successfully applied to project %s", len(updates), projectName)
		priority = PriorityMedium
		
	case NotificationTypeUpdateFailed:
		title = fmt.Sprintf("❌ Update Failed: %s", projectName)
		message = fmt.Sprintf("Failed to apply %d update(s) to project %s", len(updates), projectName)
		priority = PriorityHigh
	}
	
	// Count security updates for priority adjustment
	securityUpdates := 0
	for _, update := range updates {
		if update.SecurityFix {
			securityUpdates++
		}
	}
	
	if securityUpdates > 0 {
		priority = PriorityHigh
		message += fmt.Sprintf(" (%d security update(s))", securityUpdates)
	}
	
	notification := &Notification{
		Type:        notificationType,
		Priority:    priority,
		Title:       title,
		Message:     message,
		ProjectName: projectName,
		Metadata: map[string]interface{}{
			"total_updates":    len(updates),
			"security_updates": securityUpdates,
		},
	}
	
	return ns.SendNotification(ctx, notification)
}

// SendScanCompletedNotification sends a scan completion notification
func (ns *NotificationService) SendScanCompletedNotification(ctx context.Context, projectName string, totalDeps, updatesAvailable, securityIssues int) error {
	priority := PriorityLow
	if securityIssues > 0 {
		priority = PriorityHigh
	} else if updatesAvailable > 5 {
		priority = PriorityMedium
	}
	
	notification := &Notification{
		Type:        NotificationTypeScanCompleted,
		Priority:    priority,
		Title:       fmt.Sprintf("🔍 Scan Completed: %s", projectName),
		Message:     fmt.Sprintf("Scanned %d dependencies. Found %d update(s) and %d security issue(s)", totalDeps, updatesAvailable, securityIssues),
		ProjectName: projectName,
		Metadata: map[string]interface{}{
			"total_dependencies": totalDeps,
			"updates_available":  updatesAvailable,
			"security_issues":    securityIssues,
		},
	}
	
	return ns.SendNotification(ctx, notification)
}

// SendAgentErrorNotification sends an agent error notification
func (ns *NotificationService) SendAgentErrorNotification(ctx context.Context, errorMessage string) error {
	notification := &Notification{
		Type:     NotificationTypeAgentError,
		Priority: PriorityHigh,
		Title:    "🚨 Agent Error",
		Message:  fmt.Sprintf("AI Dependency Manager agent encountered an error: %s", errorMessage),
		Metadata: map[string]interface{}{
			"error": errorMessage,
		},
	}
	
	return ns.SendNotification(ctx, notification)
}

// SendCredentialExpiryNotification sends a credential expiry notification
func (ns *NotificationService) SendCredentialExpiryNotification(ctx context.Context, credentialName string, expiresAt time.Time) error {
	daysUntilExpiry := int(time.Until(expiresAt).Hours() / 24)
	
	priority := PriorityMedium
	if daysUntilExpiry <= 1 {
		priority = PriorityHigh
	}
	
	notification := &Notification{
		Type:     NotificationTypeCredentialExpiry,
		Priority: priority,
		Title:    "🔐 Credential Expiring",
		Message:  fmt.Sprintf("Credential '%s' expires in %d day(s) (%s)", credentialName, daysUntilExpiry, expiresAt.Format("2006-01-02")),
		Metadata: map[string]interface{}{
			"credential_name":   credentialName,
			"expires_at":        expiresAt,
			"days_until_expiry": daysUntilExpiry,
		},
	}
	
	return ns.SendNotification(ctx, notification)
}

// Private methods for different notification channels

func (ns *NotificationService) sendEmailNotification(ctx context.Context, notification *Notification) error {
	emailConfig := ns.getEmailConfig()
	if !emailConfig.Enabled {
		return nil // Email notifications disabled
	}
	
	// Create email content
	subject := fmt.Sprintf("[AI Dep Manager] %s", notification.Title)
	body := ns.formatEmailBody(notification)
	
	// Set up authentication
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.SMTPHost)
	
	// Create message
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		strings.Join(emailConfig.Recipients, ","), subject, body)
	
	// Send email
	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	err := smtp.SendMail(addr, auth, emailConfig.From, emailConfig.Recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	logger.Debug("Email notification sent successfully")
	return nil
}

func (ns *NotificationService) sendSlackNotification(ctx context.Context, notification *Notification) error {
	slackConfig := ns.getSlackConfig()
	if !slackConfig.Enabled {
		return nil // Slack notifications disabled
	}
	
	// Create Slack payload
	payload := map[string]interface{}{
		"channel":   slackConfig.Channel,
		"username":  slackConfig.Username,
		"icon_emoji": slackConfig.IconEmoji,
		"text":      notification.Title,
		"attachments": []map[string]interface{}{
			{
				"color":     ns.getPriorityColor(notification.Priority),
				"title":     notification.Title,
				"text":      notification.Message,
				"timestamp": notification.Timestamp.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Priority",
						"value": string(notification.Priority),
						"short": true,
					},
					{
						"title": "Type",
						"value": string(notification.Type),
						"short": true,
					},
				},
			},
		},
	}
	
	// Add project and package information if available
	if notification.ProjectName != "" {
		attachment := payload["attachments"].([]map[string]interface{})[0]
		fields := attachment["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"title": "Project",
			"value": notification.ProjectName,
			"short": true,
		})
		attachment["fields"] = fields
	}
	
	if notification.PackageName != "" {
		attachment := payload["attachments"].([]map[string]interface{})[0]
		fields := attachment["fields"].([]map[string]interface{})
		fields = append(fields, map[string]interface{}{
			"title": "Package",
			"value": fmt.Sprintf("%s@%s", notification.PackageName, notification.Version),
			"short": true,
		})
		attachment["fields"] = fields
	}
	
	// Send to Slack
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", slackConfig.WebhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}
	
	logger.Debug("Slack notification sent successfully")
	return nil
}

func (ns *NotificationService) sendWebhookNotification(ctx context.Context, notification *Notification) error {
	webhookConfig := ns.getWebhookConfig()
	if !webhookConfig.Enabled {
		return nil // Webhook notifications disabled
	}
	
	// Create webhook payload
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	// Create request
	method := webhookConfig.Method
	if method == "" {
		method = "POST"
	}
	
	req, err := http.NewRequestWithContext(ctx, method, webhookConfig.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range webhookConfig.Headers {
		req.Header.Set(key, value)
	}
	
	// Send webhook
	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook notification: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	logger.Debug("Webhook notification sent successfully")
	return nil
}

// Helper methods

func (ns *NotificationService) getEmailConfig() *EmailConfig {
	// In a real implementation, this would read from ns.config
	return &EmailConfig{
		Enabled:    false, // Default disabled
		SMTPHost:   "smtp.gmail.com",
		SMTPPort:   587,
		TLS:        true,
		Recipients: []string{},
	}
}

func (ns *NotificationService) getSlackConfig() *SlackConfig {
	// In a real implementation, this would read from ns.config
	return &SlackConfig{
		Enabled:   false, // Default disabled
		Username:  "AI Dependency Manager",
		IconEmoji: ":robot_face:",
	}
}

func (ns *NotificationService) getWebhookConfig() *WebhookConfig {
	// In a real implementation, this would read from ns.config
	return &WebhookConfig{
		Enabled: false, // Default disabled
		Method:  "POST",
		Headers: map[string]string{
			"User-Agent": "AI-Dependency-Manager/1.0",
		},
	}
}

func (ns *NotificationService) mapSeverityToPriority(severity string) NotificationPriority {
	switch strings.ToLower(severity) {
	case "critical":
		return PriorityCritical
	case "high":
		return PriorityHigh
	case "medium":
		return PriorityMedium
	default:
		return PriorityLow
	}
}

func (ns *NotificationService) getPriorityColor(priority NotificationPriority) string {
	switch priority {
	case PriorityCritical:
		return "danger"
	case PriorityHigh:
		return "warning"
	case PriorityMedium:
		return "good"
	default:
		return "#36a64f"
	}
}

func (ns *NotificationService) formatEmailBody(notification *Notification) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
        .priority-%s { border-left: 5px solid %s; padding-left: 15px; }
        .metadata { background-color: #f8f9fa; padding: 10px; margin-top: 20px; border-radius: 3px; }
        .timestamp { color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="header priority-%s">
        <h2>%s</h2>
        <p><strong>Priority:</strong> %s</p>
        <p><strong>Type:</strong> %s</p>
    </div>
    
    <div style="margin: 20px 0;">
        <p>%s</p>
    </div>
    
    %s
    
    <div class="timestamp">
        <p>Sent at: %s</p>
    </div>
</body>
</html>`,
		notification.Priority,
		ns.getPriorityColor(notification.Priority),
		notification.Priority,
		notification.Title,
		notification.Priority,
		notification.Type,
		notification.Message,
		ns.formatMetadata(notification),
		notification.Timestamp.Format("2006-01-02 15:04:05 UTC"),
	)
}

func (ns *NotificationService) formatMetadata(notification *Notification) string {
	if len(notification.Metadata) == 0 && notification.ProjectName == "" && notification.PackageName == "" {
		return ""
	}
	
	var metadata strings.Builder
	metadata.WriteString(`<div class="metadata"><h4>Details:</h4><ul>`)
	
	if notification.ProjectName != "" {
		metadata.WriteString(fmt.Sprintf("<li><strong>Project:</strong> %s</li>", notification.ProjectName))
	}
	
	if notification.PackageName != "" {
		metadata.WriteString(fmt.Sprintf("<li><strong>Package:</strong> %s@%s</li>", notification.PackageName, notification.Version))
	}
	
	for key, value := range notification.Metadata {
		metadata.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %v</li>", key, value))
	}
	
	metadata.WriteString("</ul></div>")
	return metadata.String()
}
