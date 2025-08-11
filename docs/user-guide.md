# AI Dependency Manager - User Guide

This comprehensive guide covers all aspects of using the AI Dependency Manager to effectively manage your project dependencies.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Configuration](#configuration)
3. [Project Management](#project-management)
4. [Dependency Scanning](#dependency-scanning)
5. [Update Management](#update-management)
6. [Security Features](#security-features)
7. [Background Agent](#background-agent)
8. [Reporting and Analytics](#reporting-and-analytics)
9. [Advanced Features](#advanced-features)
10. [Troubleshooting](#troubleshooting)

## Getting Started

### Initial Setup

After installation, initialize the AI Dependency Manager:

```bash
# Initialize configuration
ai-dep-manager configure

# Check system status
ai-dep-manager status

# Show version information
ai-dep-manager version
```

### Adding Your First Project

```bash
# Add a project from current directory
ai-dep-manager configure add-project .

# Add a project from specific path
ai-dep-manager configure add-project /path/to/your/project

# Add project with custom name
ai-dep-manager configure add-project /path/to/project --name "My Project"

# List configured projects
ai-dep-manager configure list-projects
```

## Configuration

### Configuration File

The configuration file is located at `~/.ai-dep-manager/config.yaml`. Key sections include:

```yaml
# Database configuration
database:
  path: "~/.ai-dep-manager/data.db"
  max_connections: 10

# Logging configuration
logging:
  level: "info"
  format: "json"
  file: "~/.ai-dep-manager/logs/app.log"

# Agent configuration
agent:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  max_concurrent_scans: 3
  auto_update: false

# Security configuration
security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  whitelist_enabled: false
  master_key: "your-encryption-key"

# Notification configuration
notifications:
  email:
    enabled: false
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
  slack:
    enabled: false
    webhook_url: ""
```

### Environment Variables

Override configuration with environment variables:

```bash
export AI_DEP_MANAGER_DATA_DIR="/custom/data/path"
export AI_DEP_MANAGER_LOG_LEVEL="debug"
export AI_DEP_MANAGER_DATABASE_PATH="/custom/db.sqlite"
```

### Configuration Commands

```bash
# View current configuration
ai-dep-manager configure show

# Set configuration values
ai-dep-manager configure set logging.level debug
ai-dep-manager configure set agent.schedule "0 */6 * * *"

# Reset configuration to defaults
ai-dep-manager configure reset

# Validate configuration
ai-dep-manager configure validate
```

## Project Management

### Project Discovery

The AI Dependency Manager automatically detects supported project types:

- **npm**: `package.json` files
- **pip**: `requirements.txt`, `setup.py`, `pyproject.toml`
- **Maven**: `pom.xml` files
- **Gradle**: `build.gradle`, `build.gradle.kts`

```bash
# Discover projects in directory
ai-dep-manager configure discover /path/to/search

# Auto-discover in current directory
ai-dep-manager configure discover .

# Show project details
ai-dep-manager configure show-project 1
```

### Project Configuration

```bash
# Update project settings
ai-dep-manager configure update-project 1 --name "New Name"

# Set project-specific policies
ai-dep-manager configure update-project 1 --auto-update true

# Remove a project
ai-dep-manager configure remove-project 1
```

## Dependency Scanning

### Basic Scanning

```bash
# Scan all projects
ai-dep-manager scan --all

# Scan specific project
ai-dep-manager scan --project-id 1

# Scan with verbose output
ai-dep-manager scan --project-id 1 --verbose

# Force rescan (ignore cache)
ai-dep-manager scan --project-id 1 --force
```

### Scan Results

```bash
# View scan history
ai-dep-manager scan history --project-id 1

# Show latest scan results
ai-dep-manager scan results --project-id 1

# Export scan results
ai-dep-manager scan export --project-id 1 --format json --output scan-results.json
```

### Advanced Scanning Options

```bash
# Concurrent scanning
ai-dep-manager scan --all --max-concurrent 5

# Scan with timeout
ai-dep-manager scan --project-id 1 --timeout 300s

# Include development dependencies
ai-dep-manager scan --project-id 1 --include-dev

# Exclude specific packages
ai-dep-manager scan --project-id 1 --exclude "package1,package2"
```

## Update Management

### Checking for Updates

```bash
# Check updates for all projects
ai-dep-manager check --all

# Check updates for specific project
ai-dep-manager check --project-id 1

# Show detailed update information
ai-dep-manager check --project-id 1 --detailed
```

### Update Strategies

The AI Dependency Manager supports multiple update strategies:

- **conservative**: Only patch and minor updates
- **balanced**: Minor updates and safe major updates
- **aggressive**: All available updates
- **security**: Only security-related updates

```bash
# Preview updates with different strategies
ai-dep-manager update --project-id 1 --strategy conservative --preview
ai-dep-manager update --project-id 1 --strategy balanced --preview
ai-dep-manager update --project-id 1 --strategy aggressive --preview

# Apply updates with specific strategy
ai-dep-manager update --project-id 1 --strategy balanced
```

### Interactive Updates

```bash
# Interactive update mode
ai-dep-manager update --project-id 1 --interactive

# Batch update with confirmation
ai-dep-manager update --all --batch --confirm
```

### Update Filtering

```bash
# Update specific packages only
ai-dep-manager update --project-id 1 --packages "express,lodash"

# Exclude specific packages
ai-dep-manager update --project-id 1 --exclude "react,vue"

# Update by risk level
ai-dep-manager update --project-id 1 --max-risk medium

# Update by priority
ai-dep-manager update --project-id 1 --min-priority high
```

## Security Features

### Vulnerability Scanning

```bash
# Scan for vulnerabilities
ai-dep-manager security scan --project-id 1

# Scan all projects
ai-dep-manager security scan --all

# Show vulnerability details
ai-dep-manager security vulnerabilities --project-id 1

# Export vulnerability report
ai-dep-manager security export --project-id 1 --format json
```

### Package Integrity

```bash
# Verify package integrity
ai-dep-manager security verify --project-id 1

# Check specific package
ai-dep-manager security verify-package express@4.18.0

# Batch integrity verification
ai-dep-manager security verify --all
```

### Security Rules Management

```bash
# List security rules
ai-dep-manager security rules list

# Add whitelist rule
ai-dep-manager security rules add --type whitelist --pattern "express*"

# Add blacklist rule
ai-dep-manager security rules add --type blacklist --pattern "malicious-package"

# Remove rule
ai-dep-manager security rules remove 1

# Test rules against package
ai-dep-manager security rules test express@4.18.0
```

### Credential Management

```bash
# Add registry credentials
ai-dep-manager security credentials add npm --username myuser --password mypass

# List stored credentials
ai-dep-manager security credentials list

# Update credentials
ai-dep-manager security credentials update npm --password newpass

# Remove credentials
ai-dep-manager security credentials remove npm
```

## Background Agent

### Agent Management

```bash
# Start the background agent
ai-dep-manager agent start

# Start in foreground (for debugging)
ai-dep-manager agent start --foreground

# Stop the agent
ai-dep-manager agent stop

# Restart the agent
ai-dep-manager agent restart

# Check agent status
ai-dep-manager agent status
```

### Agent Configuration

```bash
# Configure scan schedule (cron format)
ai-dep-manager configure set agent.schedule "0 2 * * *"  # Daily at 2 AM
ai-dep-manager configure set agent.schedule "0 */6 * * *"  # Every 6 hours

# Set concurrent scan limit
ai-dep-manager configure set agent.max_concurrent_scans 3

# Enable auto-updates
ai-dep-manager configure set agent.auto_update true

# Configure update strategy for auto-updates
ai-dep-manager configure set agent.update_strategy "conservative"
```

### Agent Monitoring

```bash
# View agent logs
ai-dep-manager agent logs

# Show agent statistics
ai-dep-manager agent stats

# View recent agent activity
ai-dep-manager agent activity
```

## Reporting and Analytics

### Generating Reports

```bash
# Generate summary report
ai-dep-manager report generate summary

# Generate detailed dependency report
ai-dep-manager report generate dependencies --project-id 1

# Generate security report
ai-dep-manager report generate security --all

# Generate update history report
ai-dep-manager report generate updates --days 30
```

### Report Formats and Export

```bash
# Export to different formats
ai-dep-manager report generate summary --format json --output report.json
ai-dep-manager report generate summary --format csv --output report.csv
ai-dep-manager report generate summary --format html --output report.html

# Email report
ai-dep-manager report generate summary --email admin@company.com

# Upload to webhook
ai-dep-manager report generate summary --webhook https://api.company.com/reports
```

### Analytics and Insights

```bash
# View dependency lag analysis
ai-dep-manager lag analyze --project-id 1

# Show update trends
ai-dep-manager report analytics trends --days 90

# Dependency health score
ai-dep-manager report analytics health --project-id 1

# Risk assessment
ai-dep-manager report analytics risk --all
```

## Advanced Features

### Custom Update Policies

```bash
# List available policy templates
ai-dep-manager policy templates

# Create policy from template
ai-dep-manager policy create --template security --name "Security Policy"

# Create custom policy
ai-dep-manager policy create --name "Custom Policy" --file policy.yaml

# Apply policy to project
ai-dep-manager policy apply "Security Policy" --project-id 1

# Test policy against updates
ai-dep-manager policy test "Security Policy" --project-id 1
```

### Dependency Lag Resolution

```bash
# Analyze dependency lag
ai-dep-manager lag analyze --project-id 1

# Create resolution plan
ai-dep-manager lag plan --project-id 1 --strategy balanced

# Execute resolution plan
ai-dep-manager lag execute --project-id 1 --plan-id 1

# Show lag statistics
ai-dep-manager lag stats --all
```

### Notifications

```bash
# Configure email notifications
ai-dep-manager notify configure email \
  --smtp-host smtp.gmail.com \
  --smtp-port 587 \
  --username user@gmail.com \
  --password app-password

# Configure Slack notifications
ai-dep-manager notify configure slack \
  --webhook-url https://hooks.slack.com/services/...

# Test notifications
ai-dep-manager notify test email --to admin@company.com
ai-dep-manager notify test slack

# Send manual notification
ai-dep-manager notify send "System maintenance completed" --channel email
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database status
ai-dep-manager status --verbose

# Reset database
ai-dep-manager configure reset-db

# Migrate database
ai-dep-manager configure migrate-db
```

#### 2. Package Manager Detection Issues

```bash
# Force package manager detection
ai-dep-manager scan --project-id 1 --force-detect

# Check package manager availability
ai-dep-manager status --check-tools

# Update package manager paths
ai-dep-manager configure set packagemanagers.npm.path /usr/local/bin/npm
```

#### 3. Network and Registry Issues

```bash
# Test registry connectivity
ai-dep-manager test-registry npm
ai-dep-manager test-registry pypi

# Configure proxy settings
ai-dep-manager configure set network.proxy.http http://proxy:8080
ai-dep-manager configure set network.proxy.https https://proxy:8080

# Set registry timeouts
ai-dep-manager configure set network.timeout 60s
```

#### 4. Permission Issues

```bash
# Check file permissions
ls -la ~/.ai-dep-manager/

# Fix permissions
chmod 700 ~/.ai-dep-manager/
chmod 600 ~/.ai-dep-manager/config.yaml
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Enable debug mode
export AI_DEP_MANAGER_LOG_LEVEL=debug

# Or set in configuration
ai-dep-manager configure set logging.level debug

# View debug logs
ai-dep-manager logs --level debug --tail 100
```

### Getting Help

```bash
# Show help for any command
ai-dep-manager help
ai-dep-manager scan --help
ai-dep-manager update --help

# Show configuration options
ai-dep-manager configure --help

# Display version and build information
ai-dep-manager version --verbose
```

## Best Practices

### 1. Regular Maintenance

- Run scans regularly (daily or weekly)
- Review security reports monthly
- Update dependencies in development environments first
- Maintain rollback plans for critical updates

### 2. Security

- Enable vulnerability scanning
- Use package integrity verification
- Regularly update security rules
- Monitor security notifications

### 3. Team Collaboration

- Share configuration files in version control (excluding credentials)
- Use consistent update policies across projects
- Document custom policies and rules
- Set up team notifications for critical updates

### 4. Production Deployments

- Use conservative update strategies in production
- Test updates in staging environments
- Enable audit logging
- Set up monitoring and alerting

---

For more detailed information, see the [API Reference](api-reference.md) and [Configuration Guide](configuration.md).
