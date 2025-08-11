# AI Dependency Manager - API Reference

Complete command-line interface reference for the AI Dependency Manager.

## Table of Contents

1. [Global Options](#global-options)
2. [Configuration Commands](#configuration-commands)
3. [Project Management](#project-management)
4. [Scanning Commands](#scanning-commands)
5. [Update Commands](#update-commands)
6. [Security Commands](#security-commands)
7. [Agent Commands](#agent-commands)
8. [Policy Commands](#policy-commands)
9. [Reporting Commands](#reporting-commands)
10. [Notification Commands](#notification-commands)
11. [Utility Commands](#utility-commands)

## Global Options

These options are available for all commands:

```
--config string      Configuration file path (default: ~/.ai-dep-manager/config.yaml)
--data-dir string    Data directory path (default: ~/.ai-dep-manager)
--log-level string   Log level: debug, info, warn, error (default: info)
--verbose, -v        Enable verbose output
--quiet, -q          Suppress non-error output
--help, -h           Show help information
--version            Show version information
```

## Configuration Commands

### `ai-dep-manager configure`

Initialize and manage system configuration.

#### Subcommands

##### `configure init`
Initialize the AI Dependency Manager configuration.

```bash
ai-dep-manager configure init [flags]
```

**Flags:**
- `--force`: Overwrite existing configuration
- `--interactive`: Interactive configuration setup

**Example:**
```bash
ai-dep-manager configure init --interactive
```

##### `configure show`
Display current configuration.

```bash
ai-dep-manager configure show [flags]
```

**Flags:**
- `--format string`: Output format (yaml, json, table) (default: yaml)
- `--section string`: Show specific configuration section

**Example:**
```bash
ai-dep-manager configure show --format json --section database
```

##### `configure set`
Set configuration values.

```bash
ai-dep-manager configure set <key> <value> [flags]
```

**Flags:**
- `--type string`: Value type (string, int, bool, duration) (default: string)

**Examples:**
```bash
ai-dep-manager configure set logging.level debug
ai-dep-manager configure set agent.enabled true --type bool
ai-dep-manager configure set agent.schedule "0 2 * * *"
```

##### `configure get`
Get configuration values.

```bash
ai-dep-manager configure get <key> [flags]
```

**Example:**
```bash
ai-dep-manager configure get logging.level
```

##### `configure validate`
Validate configuration file.

```bash
ai-dep-manager configure validate [flags]
```

**Flags:**
- `--config string`: Configuration file to validate

##### `configure reset`
Reset configuration to defaults.

```bash
ai-dep-manager configure reset [flags]
```

**Flags:**
- `--confirm`: Skip confirmation prompt

## Project Management

### `ai-dep-manager configure add-project`
Add a project to be managed.

```bash
ai-dep-manager configure add-project <path> [flags]
```

**Flags:**
- `--name string`: Project name
- `--type string`: Project type (auto, npm, pip, maven, gradle)
- `--auto-update`: Enable automatic updates
- `--priority string`: Project priority (low, medium, high)

**Example:**
```bash
ai-dep-manager configure add-project /path/to/project --name "My App" --type npm
```

### `ai-dep-manager configure list-projects`
List all configured projects.

```bash
ai-dep-manager configure list-projects [flags]
```

**Flags:**
- `--format string`: Output format (table, json, yaml) (default: table)
- `--filter string`: Filter projects by name or type

### `ai-dep-manager configure show-project`
Show detailed project information.

```bash
ai-dep-manager configure show-project <id> [flags]
```

**Flags:**
- `--format string`: Output format (table, json, yaml) (default: table)

### `ai-dep-manager configure update-project`
Update project configuration.

```bash
ai-dep-manager configure update-project <id> [flags]
```

**Flags:**
- `--name string`: Update project name
- `--auto-update`: Enable/disable automatic updates
- `--priority string`: Update priority level

### `ai-dep-manager configure remove-project`
Remove a project from management.

```bash
ai-dep-manager configure remove-project <id> [flags]
```

**Flags:**
- `--confirm`: Skip confirmation prompt

## Scanning Commands

### `ai-dep-manager scan`
Scan projects for dependency information.

```bash
ai-dep-manager scan [flags]
```

**Flags:**
- `--all`: Scan all projects
- `--project-id int`: Scan specific project by ID
- `--project-name string`: Scan specific project by name
- `--force`: Force rescan (ignore cache)
- `--max-concurrent int`: Maximum concurrent scans (default: 3)
- `--timeout duration`: Scan timeout (default: 5m)
- `--include-dev`: Include development dependencies
- `--exclude string`: Comma-separated list of packages to exclude

**Examples:**
```bash
ai-dep-manager scan --all
ai-dep-manager scan --project-id 1 --force
ai-dep-manager scan --project-name "My App" --include-dev
```

### `ai-dep-manager scan history`
Show scan history for projects.

```bash
ai-dep-manager scan history [flags]
```

**Flags:**
- `--project-id int`: Show history for specific project
- `--limit int`: Limit number of results (default: 10)
- `--format string`: Output format (table, json, yaml) (default: table)

### `ai-dep-manager scan results`
Show latest scan results.

```bash
ai-dep-manager scan results [flags]
```

**Flags:**
- `--project-id int`: Show results for specific project
- `--format string`: Output format (table, json, yaml) (default: table)
- `--detailed`: Show detailed dependency information

## Update Commands

### `ai-dep-manager check`
Check for available updates.

```bash
ai-dep-manager check [flags]
```

**Flags:**
- `--all`: Check all projects
- `--project-id int`: Check specific project
- `--detailed`: Show detailed update information
- `--format string`: Output format (table, json, yaml) (default: table)

### `ai-dep-manager update`
Apply dependency updates.

```bash
ai-dep-manager update [flags]
```

**Flags:**
- `--all`: Update all projects
- `--project-id int`: Update specific project
- `--strategy string`: Update strategy (conservative, balanced, aggressive, security) (default: balanced)
- `--preview`: Show what would be updated without applying
- `--interactive`: Interactive update mode
- `--batch`: Batch update mode
- `--confirm`: Skip confirmation prompts
- `--dry-run`: Simulate updates without applying
- `--packages string`: Comma-separated list of specific packages to update
- `--exclude string`: Comma-separated list of packages to exclude
- `--max-risk string`: Maximum risk level (low, medium, high) (default: medium)
- `--min-priority string`: Minimum priority level (low, medium, high)

**Examples:**
```bash
ai-dep-manager update --all --strategy conservative --preview
ai-dep-manager update --project-id 1 --interactive
ai-dep-manager update --project-id 1 --packages "express,lodash" --confirm
```

### `ai-dep-manager rollback`
Manage and execute rollback operations.

```bash
ai-dep-manager rollback [command] [flags]
```

#### Subcommands

##### `rollback list`
List available rollback plans.

```bash
ai-dep-manager rollback list [flags]
```

**Flags:**
- `--project-id int`: Show rollbacks for specific project
- `--format string`: Output format (table, json, yaml) (default: table)

##### `rollback show`
Show rollback plan details.

```bash
ai-dep-manager rollback show <id> [flags]
```

##### `rollback execute`
Execute a rollback plan.

```bash
ai-dep-manager rollback execute <id> [flags]
```

**Flags:**
- `--confirm`: Skip confirmation prompt
- `--dry-run`: Simulate rollback without applying

## Security Commands

### `ai-dep-manager security`
Security-related operations.

#### Subcommands

##### `security scan`
Scan for security vulnerabilities.

```bash
ai-dep-manager security scan [flags]
```

**Flags:**
- `--all`: Scan all projects
- `--project-id int`: Scan specific project
- `--severity string`: Minimum severity level (low, medium, high, critical)
- `--format string`: Output format (table, json, yaml) (default: table)

##### `security vulnerabilities`
List known vulnerabilities.

```bash
ai-dep-manager security vulnerabilities [flags]
```

**Flags:**
- `--project-id int`: Show vulnerabilities for specific project
- `--severity string`: Filter by severity level
- `--status string`: Filter by status (open, fixed, ignored)

##### `security verify`
Verify package integrity.

```bash
ai-dep-manager security verify [flags]
```

**Flags:**
- `--all`: Verify all projects
- `--project-id int`: Verify specific project
- `--package string`: Verify specific package

##### `security rules`
Manage security rules (whitelist/blacklist).

```bash
ai-dep-manager security rules [command] [flags]
```

**Subcommands:**
- `list`: List all rules
- `add`: Add new rule
- `remove`: Remove rule
- `test`: Test rule against package

##### `security credentials`
Manage registry credentials.

```bash
ai-dep-manager security credentials [command] [flags]
```

**Subcommands:**
- `list`: List stored credentials
- `add`: Add new credentials
- `update`: Update existing credentials
- `remove`: Remove credentials

## Agent Commands

### `ai-dep-manager agent`
Background agent management.

#### Subcommands

##### `agent start`
Start the background agent.

```bash
ai-dep-manager agent start [flags]
```

**Flags:**
- `--foreground`: Run in foreground mode
- `--daemon`: Run as system daemon

##### `agent stop`
Stop the background agent.

```bash
ai-dep-manager agent stop [flags]
```

##### `agent restart`
Restart the background agent.

```bash
ai-dep-manager agent restart [flags]
```

##### `agent status`
Show agent status and statistics.

```bash
ai-dep-manager agent status [flags]
```

**Flags:**
- `--detailed`: Show detailed status information
- `--format string`: Output format (table, json, yaml) (default: table)

##### `agent logs`
Show agent logs.

```bash
ai-dep-manager agent logs [flags]
```

**Flags:**
- `--tail int`: Number of recent log lines to show (default: 100)
- `--follow`: Follow log output
- `--level string`: Filter by log level

## Policy Commands

### `ai-dep-manager policy`
Custom update policy management.

#### Subcommands

##### `policy list`
List all policies.

```bash
ai-dep-manager policy list [flags]
```

##### `policy show`
Show policy details.

```bash
ai-dep-manager policy show <name> [flags]
```

##### `policy create`
Create new policy.

```bash
ai-dep-manager policy create [flags]
```

**Flags:**
- `--name string`: Policy name (required)
- `--template string`: Use policy template
- `--file string`: Load policy from file
- `--interactive`: Interactive policy creation

##### `policy update`
Update existing policy.

```bash
ai-dep-manager policy update <name> [flags]
```

##### `policy delete`
Delete policy.

```bash
ai-dep-manager policy delete <name> [flags]
```

##### `policy test`
Test policy against project.

```bash
ai-dep-manager policy test <name> [flags]
```

**Flags:**
- `--project-id int`: Test against specific project

##### `policy templates`
List available policy templates.

```bash
ai-dep-manager policy templates [flags]
```

## Reporting Commands

### `ai-dep-manager report`
Generate reports and analytics.

#### Subcommands

##### `report generate`
Generate various types of reports.

```bash
ai-dep-manager report generate <type> [flags]
```

**Report Types:**
- `summary`: Overall system summary
- `dependencies`: Detailed dependency report
- `security`: Security vulnerability report
- `updates`: Update history report
- `analytics`: Analytics and trends report

**Flags:**
- `--project-id int`: Generate report for specific project
- `--all`: Include all projects
- `--format string`: Output format (json, csv, html, pdf) (default: json)
- `--output string`: Output file path
- `--email string`: Email report to address
- `--webhook string`: Send report to webhook URL
- `--days int`: Number of days to include in report (default: 30)

##### `report analytics`
Show analytics and insights.

```bash
ai-dep-manager report analytics [command] [flags]
```

**Subcommands:**
- `trends`: Show dependency trends
- `health`: Show project health scores
- `risk`: Show risk assessments

## Notification Commands

### `ai-dep-manager notify`
Notification system management.

#### Subcommands

##### `notify configure`
Configure notification channels.

```bash
ai-dep-manager notify configure <channel> [flags]
```

**Channels:**
- `email`: Email notifications
- `slack`: Slack notifications
- `webhook`: Webhook notifications

##### `notify test`
Test notification channels.

```bash
ai-dep-manager notify test <channel> [flags]
```

**Flags:**
- `--to string`: Test recipient (for email)
- `--message string`: Test message
- `--dry-run`: Show what would be sent without sending

##### `notify send`
Send manual notification.

```bash
ai-dep-manager notify send <message> [flags]
```

**Flags:**
- `--channel string`: Notification channel
- `--priority string`: Message priority (low, medium, high, critical)

##### `notify list`
List configured notification channels.

```bash
ai-dep-manager notify list [flags]
```

## Utility Commands

### `ai-dep-manager status`
Show system status and health.

```bash
ai-dep-manager status [flags]
```

**Flags:**
- `--detailed`: Show detailed status information
- `--check-tools`: Check availability of package manager tools
- `--format string`: Output format (table, json, yaml) (default: table)

### `ai-dep-manager version`
Show version information.

```bash
ai-dep-manager version [flags]
```

**Flags:**
- `--verbose`: Show detailed build information

### `ai-dep-manager lag`
Dependency lag analysis and resolution.

#### Subcommands

##### `lag analyze`
Analyze dependency lag for projects.

```bash
ai-dep-manager lag analyze <project-id> [flags]
```

##### `lag plan`
Create lag resolution plan.

```bash
ai-dep-manager lag plan <project-id> [flags]
```

**Flags:**
- `--strategy string`: Resolution strategy (conservative, balanced, aggressive)

##### `lag execute`
Execute lag resolution plan.

```bash
ai-dep-manager lag execute <project-id> [flags]
```

**Flags:**
- `--plan-id int`: Specific plan to execute
- `--dry-run`: Simulate execution

### `ai-dep-manager logs`
Show application logs.

```bash
ai-dep-manager logs [flags]
```

**Flags:**
- `--tail int`: Number of recent log lines (default: 100)
- `--follow`: Follow log output
- `--level string`: Filter by log level
- `--since string`: Show logs since timestamp

## Exit Codes

The AI Dependency Manager uses the following exit codes:

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Network error
- `4`: Permission error
- `5`: Database error
- `6`: Package manager error
- `7`: Security error
- `8`: Update error

## Environment Variables

The following environment variables can be used to override configuration:

- `AI_DEP_MANAGER_CONFIG_FILE`: Configuration file path
- `AI_DEP_MANAGER_DATA_DIR`: Data directory path
- `AI_DEP_MANAGER_LOG_LEVEL`: Log level
- `AI_DEP_MANAGER_DATABASE_PATH`: Database file path
- `AI_DEP_MANAGER_MASTER_KEY`: Encryption master key
- `AI_DEP_MANAGER_TEST_MODE`: Enable test mode

## Configuration File Format

The configuration file uses YAML format. Here's a complete example:

```yaml
# Database configuration
database:
  path: "~/.ai-dep-manager/data.db"
  max_connections: 10
  connection_timeout: "30s"

# Logging configuration
logging:
  level: "info"
  format: "json"
  file: "~/.ai-dep-manager/logs/app.log"
  max_size: "100MB"
  max_backups: 5

# Agent configuration
agent:
  enabled: true
  schedule: "0 2 * * *"
  max_concurrent_scans: 3
  scan_timeout: "5m"
  auto_update: false
  update_strategy: "conservative"

# Security configuration
security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  whitelist_enabled: false
  master_key: "base64-encoded-key"
  vulnerability_db_update_interval: "24h"

# Package manager configuration
packagemanagers:
  npm:
    enabled: true
    path: "/usr/local/bin/npm"
    registry: "https://registry.npmjs.org/"
  pip:
    enabled: true
    path: "/usr/local/bin/pip"
    index_url: "https://pypi.org/simple/"
  maven:
    enabled: true
    path: "/usr/local/bin/mvn"
  gradle:
    enabled: true
    path: "/usr/local/bin/gradle"

# Notification configuration
notifications:
  email:
    enabled: false
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "user@gmail.com"
    from: "AI Dep Manager <noreply@company.com>"
  slack:
    enabled: false
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#dependencies"
  webhook:
    enabled: false
    url: "https://api.company.com/webhooks/dependencies"
    headers:
      Authorization: "Bearer token"

# Performance configuration
performance:
  max_memory_usage: "1GB"
  cache_size: "256MB"
  gc_percent: 100
```

---

For more information, see the [User Guide](user-guide.md) and [Configuration Guide](configuration.md).
