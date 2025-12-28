# Superintelligence Dependency Manager - User Guide

This comprehensive guide covers all aspects of using the Superintelligence Dependency Manager to effectively manage your project dependencies.

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

After installation, initialize the Superintelligence Dependency Manager:

```bash
# Initialize configuration
superint-dep-manager configure

# Check system status
superint-dep-manager status

# Show version information
superint-dep-manager version
```

### Adding Your First Project

```bash
# Add a project from current directory
superint-dep-manager configure add-project .

# Add a project from specific path
superint-dep-manager configure add-project /path/to/your/project

# Add project with custom name
superint-dep-manager configure add-project /path/to/project --name "My Project"

# List configured projects
superint-dep-manager configure list-projects
```

## Configuration

### Configuration File

The configuration file is located at `~/.superint-dep-manager/config.yaml`. Key sections include:

```yaml
# Database configuration
database:
  path: "~/.superint-dep-manager/data.db"
  max_connections: 10

# Logging configuration
logging:
  level: "info"
  format: "json"
  file: "~/.superint-dep-manager/logs/app.log"

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
superint-dep-manager configure show

# Set configuration values
superint-dep-manager configure set logging.level debug
superint-dep-manager configure set agent.schedule "0 */6 * * *"

# Reset configuration to defaults
superint-dep-manager configure reset

# Validate configuration
superint-dep-manager configure validate
```

## Project Management

### Project Discovery

The Superintelligence Dependency Manager automatically detects supported project types:

- **npm**: `package.json` files
- **pip**: `requirements.txt`, `setup.py`, `pyproject.toml`
- **Maven**: `pom.xml` files
- **Gradle**: `build.gradle`, `build.gradle.kts`

```bash
# Discover projects in directory
superint-dep-manager configure discover /path/to/search

# Auto-discover in current directory
superint-dep-manager configure discover .

# Show project details
superint-dep-manager configure show-project 1
```

### Project Configuration

```bash
# Update project settings
superint-dep-manager configure update-project 1 --name "New Name"

# Set project-specific policies
superint-dep-manager configure update-project 1 --auto-update true

# Remove a project
superint-dep-manager configure remove-project 1
```

## Dependency Scanning

### Basic Scanning

```bash
# Scan all projects
superint-dep-manager scan --all

# Scan specific project
superint-dep-manager scan --project-id 1

# Scan with verbose output
superint-dep-manager scan --project-id 1 --verbose

# Force rescan (ignore cache)
superint-dep-manager scan --project-id 1 --force
```

### Scan Results

```bash
# View scan history
superint-dep-manager scan history --project-id 1

# Show latest scan results
superint-dep-manager scan results --project-id 1

# Export scan results
superint-dep-manager scan export --project-id 1 --format json --output scan-results.json
```

### Advanced Scanning Options

```bash
# Concurrent scanning
superint-dep-manager scan --all --max-concurrent 5

# Scan with timeout
superint-dep-manager scan --project-id 1 --timeout 300s

# Include development dependencies
superint-dep-manager scan --project-id 1 --include-dev

# Exclude specific packages
superint-dep-manager scan --project-id 1 --exclude "package1,package2"
```

## Update Management

### Checking for Updates

```bash
# Check updates for all projects
superint-dep-manager check --all

# Check updates for specific project
superint-dep-manager check --project-id 1

# Show detailed update information
superint-dep-manager check --project-id 1 --detailed
```

### Update Strategies

The Superintelligence Dependency Manager supports multiple update strategies:

- **conservative**: Only patch and minor updates
- **balanced**: Minor updates and safe major updates
- **aggressive**: All available updates
- **security**: Only security-related updates

```bash
# Preview updates with different strategies
superint-dep-manager update --project-id 1 --strategy conservative --preview
superint-dep-manager update --project-id 1 --strategy balanced --preview
superint-dep-manager update --project-id 1 --strategy aggressive --preview

# Apply updates with specific strategy
superint-dep-manager update --project-id 1 --strategy balanced
```

### Interactive Updates

```bash
# Interactive update mode
superint-dep-manager update --project-id 1 --interactive

# Batch update with confirmation
superint-dep-manager update --all --batch --confirm
```

### Update Filtering

```bash
# Update specific packages only
superint-dep-manager update --project-id 1 --packages "express,lodash"

# Exclude specific packages
superint-dep-manager update --project-id 1 --exclude "react,vue"

# Update by risk level
superint-dep-manager update --project-id 1 --max-risk medium

# Update by priority
superint-dep-manager update --project-id 1 --min-priority high
```

## Security Features

### Vulnerability Scanning

```bash
# Scan for vulnerabilities
superint-dep-manager security scan --project-id 1

# Scan all projects
superint-dep-manager security scan --all

# Show vulnerability details
superint-dep-manager security vulnerabilities --project-id 1

# Export vulnerability report
superint-dep-manager security export --project-id 1 --format json
```

### Package Integrity

```bash
# Verify package integrity
superint-dep-manager security verify --project-id 1

# Check specific package
superint-dep-manager security verify-package express@4.18.0

# Batch integrity verification
superint-dep-manager security verify --all
```

### Security Rules Management

```bash
# List security rules
superint-dep-manager security rules list

# Add whitelist rule
superint-dep-manager security rules add --type whitelist --pattern "express*"

# Add blacklist rule
superint-dep-manager security rules add --type blacklist --pattern "malicious-package"

# Remove rule
superint-dep-manager security rules remove 1

# Test rules against package
superint-dep-manager security rules test express@4.18.0
```

### Credential Management

```bash
# Add registry credentials
superint-dep-manager security credentials add npm --username myuser --password mypass

# List stored credentials
superint-dep-manager security credentials list

# Update credentials
superint-dep-manager security credentials update npm --password newpass

# Remove credentials
superint-dep-manager security credentials remove npm
```

## Background Agent

### Agent Management

```bash
# Start the background agent
superint-dep-manager agent start

# Start in foreground (for debugging)
superint-dep-manager agent start --foreground

# Stop the agent
superint-dep-manager agent stop

# Restart the agent
superint-dep-manager agent restart

# Check agent status
superint-dep-manager agent status
```

### Agent Configuration

```bash
# Configure scan schedule (cron format)
superint-dep-manager configure set agent.schedule "0 2 * * *"  # Daily at 2 AM
superint-dep-manager configure set agent.schedule "0 */6 * * *"  # Every 6 hours

# Set concurrent scan limit
superint-dep-manager configure set agent.max_concurrent_scans 3

# Enable auto-updates
superint-dep-manager configure set agent.auto_update true

# Configure update strategy for auto-updates
superint-dep-manager configure set agent.update_strategy "conservative"
```

### Agent Monitoring

```bash
# View agent logs
superint-dep-manager agent logs

# Show agent statistics
superint-dep-manager agent stats

# View recent agent activity
superint-dep-manager agent activity
```

## Reporting and Analytics

### Generating Reports

```bash
# Generate summary report
superint-dep-manager report generate summary

# Generate detailed dependency report
superint-dep-manager report generate dependencies --project-id 1

# Generate security report
superint-dep-manager report generate security --all

# Generate update history report
superint-dep-manager report generate updates --days 30
```

### Report Formats and Export

```bash
# Export to different formats
superint-dep-manager report generate summary --format json --output report.json
superint-dep-manager report generate summary --format csv --output report.csv
superint-dep-manager report generate summary --format html --output report.html

# Email report
superint-dep-manager report generate summary --email admin@company.com

# Upload to webhook
superint-dep-manager report generate summary --webhook https://api.company.com/reports
```

### Analytics and Insights

```bash
# View dependency lag analysis
superint-dep-manager lag analyze --project-id 1

# Show update trends
superint-dep-manager report analytics trends --days 90

# Dependency health score
superint-dep-manager report analytics health --project-id 1

# Risk assessment
superint-dep-manager report analytics risk --all
```

## Advanced Features

### Custom Update Policies

```bash
# List available policy templates
superint-dep-manager policy templates

# Create policy from template
superint-dep-manager policy create --template security --name "Security Policy"

# Create custom policy
superint-dep-manager policy create --name "Custom Policy" --file policy.yaml

# Apply policy to project
superint-dep-manager policy apply "Security Policy" --project-id 1

# Test policy against updates
superint-dep-manager policy test "Security Policy" --project-id 1
```

### Dependency Lag Resolution

```bash
# Analyze dependency lag
superint-dep-manager lag analyze --project-id 1

# Create resolution plan
superint-dep-manager lag plan --project-id 1 --strategy balanced

# Execute resolution plan
superint-dep-manager lag execute --project-id 1 --plan-id 1

# Show lag statistics
superint-dep-manager lag stats --all
```

### Notifications

```bash
# Configure email notifications
superint-dep-manager notify configure email \
  --smtp-host smtp.gmail.com \
  --smtp-port 587 \
  --username user@gmail.com \
  --password app-password

# Configure Slack notifications
superint-dep-manager notify configure slack \
  --webhook-url https://hooks.slack.com/services/...

# Test notifications
superint-dep-manager notify test email --to admin@company.com
superint-dep-manager notify test slack

# Send manual notification
superint-dep-manager notify send "System maintenance completed" --channel email
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Issues

```bash
# Check database status
superint-dep-manager status --verbose

# Reset database
superint-dep-manager configure reset-db

# Migrate database
superint-dep-manager configure migrate-db
```

#### 2. Package Manager Detection Issues

```bash
# Force package manager detection
superint-dep-manager scan --project-id 1 --force-detect

# Check package manager availability
superint-dep-manager status --check-tools

# Update package manager paths
superint-dep-manager configure set packagemanagers.npm.path /usr/local/bin/npm
```

#### 3. Network and Registry Issues

```bash
# Test registry connectivity
superint-dep-manager test-registry npm
superint-dep-manager test-registry pypi

# Configure proxy settings
superint-dep-manager configure set network.proxy.http http://proxy:8080
superint-dep-manager configure set network.proxy.https https://proxy:8080

# Set registry timeouts
superint-dep-manager configure set network.timeout 60s
```

#### 4. Permission Issues

```bash
# Check file permissions
ls -la ~/.superint-dep-manager/

# Fix permissions
chmod 700 ~/.superint-dep-manager/
chmod 600 ~/.superint-dep-manager/config.yaml
```

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# Enable debug mode
export AI_DEP_MANAGER_LOG_LEVEL=debug

# Or set in configuration
superint-dep-manager configure set logging.level debug

# View debug logs
superint-dep-manager logs --level debug --tail 100
```

### Getting Help

```bash
# Show help for any command
superint-dep-manager help
superint-dep-manager scan --help
superint-dep-manager update --help

# Show configuration options
superint-dep-manager configure --help

# Display version and build information
superint-dep-manager version --verbose
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
