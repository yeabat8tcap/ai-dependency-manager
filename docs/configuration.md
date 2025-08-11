# AI Dependency Manager - Configuration Guide

This guide provides detailed information about configuring the AI Dependency Manager for various use cases and environments.

## Table of Contents

1. [Configuration Overview](#configuration-overview)
2. [Configuration File Structure](#configuration-file-structure)
3. [Database Configuration](#database-configuration)
4. [Logging Configuration](#logging-configuration)
5. [Agent Configuration](#agent-configuration)
6. [Security Configuration](#security-configuration)
7. [Package Manager Configuration](#package-manager-configuration)
8. [Notification Configuration](#notification-configuration)
9. [Performance Configuration](#performance-configuration)
10. [Environment-Specific Configurations](#environment-specific-configurations)

## Configuration Overview

The AI Dependency Manager uses a YAML configuration file located at `~/.ai-dep-manager/config.yaml` by default. Configuration can be overridden using environment variables or command-line flags.

### Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

### Configuration File Location

The configuration file is searched in the following order:

1. Path specified by `--config` flag
2. `AI_DEP_MANAGER_CONFIG_FILE` environment variable
3. `~/.ai-dep-manager/config.yaml`
4. `/etc/ai-dep-manager/config.yaml`

## Configuration File Structure

```yaml
# Complete configuration example
database:
  path: "~/.ai-dep-manager/data.db"
  max_connections: 10
  connection_timeout: "30s"
  query_timeout: "60s"

logging:
  level: "info"
  format: "json"
  file: "~/.ai-dep-manager/logs/app.log"
  max_size: "100MB"
  max_backups: 5
  max_age: 30
  compress: true

agent:
  enabled: true
  schedule: "0 2 * * *"
  max_concurrent_scans: 3
  scan_timeout: "5m"
  auto_update: false
  update_strategy: "conservative"
  retry_attempts: 3
  retry_delay: "30s"

security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  whitelist_enabled: false
  master_key: "base64-encoded-key"
  vulnerability_db_update_interval: "24h"
  max_vulnerability_age: "7d"

packagemanagers:
  npm:
    enabled: true
    path: "/usr/local/bin/npm"
    registry: "https://registry.npmjs.org/"
    timeout: "60s"
    max_retries: 3
  pip:
    enabled: true
    path: "/usr/local/bin/pip"
    index_url: "https://pypi.org/simple/"
    timeout: "60s"
    max_retries: 3
  maven:
    enabled: true
    path: "/usr/local/bin/mvn"
    timeout: "120s"
    max_retries: 3
  gradle:
    enabled: true
    path: "/usr/local/bin/gradle"
    timeout: "120s"
    max_retries: 3

notifications:
  email:
    enabled: false
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "user@gmail.com"
    password: "app-password"
    from: "AI Dep Manager <noreply@company.com>"
    tls: true
  slack:
    enabled: false
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#dependencies"
    username: "AI Dep Manager"
  webhook:
    enabled: false
    url: "https://api.company.com/webhooks/dependencies"
    timeout: "30s"
    headers:
      Authorization: "Bearer token"
      Content-Type: "application/json"

performance:
  max_memory_usage: "1GB"
  cache_size: "256MB"
  gc_percent: 100
  max_goroutines: 1000

network:
  timeout: "30s"
  retry_attempts: 3
  retry_delay: "5s"
  proxy:
    http: ""
    https: ""
    no_proxy: "localhost,127.0.0.1"
```

## Database Configuration

### SQLite Configuration (Default)

```yaml
database:
  path: "~/.ai-dep-manager/data.db"
  max_connections: 10
  connection_timeout: "30s"
  query_timeout: "60s"
  pragma:
    journal_mode: "WAL"
    synchronous: "NORMAL"
    cache_size: "-64000"  # 64MB cache
    temp_store: "MEMORY"
```

### PostgreSQL Configuration

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  name: "ai_dep_manager"
  username: "ai_dep_manager"
  password: "secure_password"
  ssl_mode: "require"
  max_connections: 20
  connection_timeout: "30s"
  query_timeout: "60s"
```

### MySQL Configuration

```yaml
database:
  type: "mysql"
  host: "localhost"
  port: 3306
  name: "ai_dep_manager"
  username: "ai_dep_manager"
  password: "secure_password"
  charset: "utf8mb4"
  max_connections: 20
  connection_timeout: "30s"
  query_timeout: "60s"
```

### Database Environment Variables

```bash
export AI_DEP_MANAGER_DATABASE_TYPE="postgres"
export AI_DEP_MANAGER_DATABASE_HOST="localhost"
export AI_DEP_MANAGER_DATABASE_PORT="5432"
export AI_DEP_MANAGER_DATABASE_NAME="ai_dep_manager"
export AI_DEP_MANAGER_DATABASE_USERNAME="ai_dep_manager"
export AI_DEP_MANAGER_DATABASE_PASSWORD="secure_password"
```

## Logging Configuration

### Log Levels

- `debug`: Detailed debugging information
- `info`: General information messages
- `warn`: Warning messages
- `error`: Error messages only

### Log Formats

- `text`: Human-readable text format
- `json`: Structured JSON format (recommended for production)

### File Logging Configuration

```yaml
logging:
  level: "info"
  format: "json"
  file: "~/.ai-dep-manager/logs/app.log"
  max_size: "100MB"      # Maximum file size before rotation
  max_backups: 5         # Number of backup files to keep
  max_age: 30           # Maximum age in days
  compress: true        # Compress rotated files
```

### Console Logging

```yaml
logging:
  level: "info"
  format: "text"
  console: true
  color: true           # Enable colored output
```

### Syslog Configuration

```yaml
logging:
  level: "info"
  format: "json"
  syslog:
    enabled: true
    network: "udp"
    address: "localhost:514"
    facility: "daemon"
    tag: "ai-dep-manager"
```

## Agent Configuration

### Schedule Configuration

The agent uses cron-style scheduling:

```yaml
agent:
  schedule: "0 2 * * *"    # Daily at 2 AM
  # schedule: "0 */6 * * *" # Every 6 hours
  # schedule: "0 0 * * 0"   # Weekly on Sunday
  # schedule: "0 0 1 * *"   # Monthly on 1st
```

### Concurrency Configuration

```yaml
agent:
  max_concurrent_scans: 3      # Maximum parallel scans
  scan_timeout: "5m"           # Timeout per scan
  batch_size: 100              # Dependencies per batch
  worker_pool_size: 10         # Worker goroutines
```

### Auto-Update Configuration

```yaml
agent:
  auto_update: true
  update_strategy: "conservative"  # conservative, balanced, aggressive
  auto_update_filters:
    max_risk: "medium"            # Maximum risk level
    min_priority: "high"          # Minimum priority
    exclude_packages:             # Packages to exclude
      - "react"
      - "vue"
    include_types:                # Update types to include
      - "patch"
      - "minor"
```

## Security Configuration

### Vulnerability Scanning

```yaml
security:
  enable_vulnerability_scanning: true
  vulnerability_sources:
    - "npm_audit"
    - "pypi_safety"
    - "ossindex"
    - "github_advisories"
  vulnerability_db_update_interval: "24h"
  max_vulnerability_age: "7d"
  severity_threshold: "medium"    # Minimum severity to report
```

### Package Integrity

```yaml
security:
  enable_integrity_verification: true
  checksum_algorithms:
    - "sha256"
    - "sha512"
  verify_signatures: true
  trust_store_path: "~/.ai-dep-manager/trust"
```

### Access Control

```yaml
security:
  whitelist_enabled: true
  whitelist_patterns:
    - "express*"
    - "@types/*"
  blacklist_patterns:
    - "malicious-*"
    - "suspicious-package"
  require_approval_for:
    - "major_updates"
    - "new_dependencies"
    - "security_updates"
```

### Credential Encryption

```yaml
security:
  master_key: "base64-encoded-32-byte-key"
  key_derivation:
    algorithm: "pbkdf2"
    iterations: 100000
    salt_length: 32
  encryption:
    algorithm: "aes-256-gcm"
```

## Package Manager Configuration

### npm Configuration

```yaml
packagemanagers:
  npm:
    enabled: true
    path: "/usr/local/bin/npm"
    registry: "https://registry.npmjs.org/"
    timeout: "60s"
    max_retries: 3
    auth:
      token: "${NPM_TOKEN}"
    config:
      cache: "~/.npm"
      audit_level: "moderate"
      fund: false
```

### pip Configuration

```yaml
packagemanagers:
  pip:
    enabled: true
    path: "/usr/local/bin/pip"
    index_url: "https://pypi.org/simple/"
    extra_index_urls:
      - "https://private.pypi.company.com/simple/"
    timeout: "60s"
    max_retries: 3
    config:
      cache_dir: "~/.cache/pip"
      trusted_hosts:
        - "private.pypi.company.com"
```

### Maven Configuration

```yaml
packagemanagers:
  maven:
    enabled: true
    path: "/usr/local/bin/mvn"
    settings_file: "~/.m2/settings.xml"
    local_repository: "~/.m2/repository"
    timeout: "120s"
    max_retries: 3
    profiles:
      - "production"
    properties:
      maven.compiler.source: "11"
      maven.compiler.target: "11"
```

### Gradle Configuration

```yaml
packagemanagers:
  gradle:
    enabled: true
    path: "/usr/local/bin/gradle"
    gradle_home: "/opt/gradle"
    timeout: "120s"
    max_retries: 3
    properties:
      org.gradle.daemon: "true"
      org.gradle.parallel: "true"
      org.gradle.caching: "true"
```

## Notification Configuration

### Email Notifications

```yaml
notifications:
  email:
    enabled: true
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "notifications@company.com"
    password: "${EMAIL_PASSWORD}"
    from: "AI Dep Manager <noreply@company.com>"
    tls: true
    templates:
      vulnerability_found: "templates/vulnerability.html"
      update_available: "templates/update.html"
    recipients:
      critical: ["security@company.com"]
      high: ["dev-team@company.com"]
      medium: ["dev-team@company.com"]
      low: ["dev-team@company.com"]
```

### Slack Notifications

```yaml
notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#dependencies"
    username: "AI Dep Manager"
    icon_emoji: ":robot_face:"
    message_format: "detailed"  # brief, detailed, custom
    thread_replies: true
    channels:
      critical: "#security-alerts"
      high: "#dev-alerts"
      medium: "#dependencies"
      low: "#dependencies"
```

### Webhook Notifications

```yaml
notifications:
  webhook:
    enabled: true
    url: "https://api.company.com/webhooks/dependencies"
    timeout: "30s"
    retry_attempts: 3
    headers:
      Authorization: "Bearer ${WEBHOOK_TOKEN}"
      Content-Type: "application/json"
      X-Source: "ai-dep-manager"
    payload_template: |
      {
        "source": "ai-dep-manager",
        "timestamp": "{{.Timestamp}}",
        "event": "{{.Event}}",
        "project": "{{.Project}}",
        "data": {{.Data}}
      }
```

## Performance Configuration

### Memory Management

```yaml
performance:
  max_memory_usage: "1GB"       # Maximum memory usage
  gc_percent: 100               # Go GC target percentage
  max_goroutines: 1000          # Maximum goroutines
  goroutine_pool_size: 50       # Worker pool size
```

### Caching Configuration

```yaml
performance:
  cache_size: "256MB"           # In-memory cache size
  cache_ttl: "1h"              # Cache time-to-live
  cache_cleanup_interval: "10m" # Cache cleanup frequency
  disk_cache:
    enabled: true
    path: "~/.ai-dep-manager/cache"
    max_size: "1GB"
    ttl: "24h"
```

### Database Performance

```yaml
performance:
  database:
    connection_pool_size: 20
    max_idle_connections: 5
    connection_max_lifetime: "1h"
    query_timeout: "30s"
    batch_size: 1000
```

## Environment-Specific Configurations

### Development Environment

```yaml
# config-dev.yaml
database:
  path: "./data/dev.db"

logging:
  level: "debug"
  format: "text"
  console: true
  color: true

agent:
  enabled: false

security:
  enable_vulnerability_scanning: false
  enable_integrity_verification: false

performance:
  max_memory_usage: "512MB"
  cache_size: "64MB"
```

### Staging Environment

```yaml
# config-staging.yaml
database:
  type: "postgres"
  host: "staging-db.company.com"
  name: "ai_dep_manager_staging"

logging:
  level: "info"
  format: "json"
  file: "/var/log/ai-dep-manager/staging.log"

agent:
  enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours

security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true

notifications:
  slack:
    enabled: true
    channel: "#dev-staging"
```

### Production Environment

```yaml
# config-prod.yaml
database:
  type: "postgres"
  host: "prod-db.company.com"
  name: "ai_dep_manager"
  ssl_mode: "require"
  max_connections: 50

logging:
  level: "warn"
  format: "json"
  file: "/var/log/ai-dep-manager/production.log"
  max_size: "500MB"
  max_backups: 10

agent:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  max_concurrent_scans: 10
  auto_update: false  # Manual approval required

security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  whitelist_enabled: true
  require_approval_for:
    - "major_updates"
    - "new_dependencies"

notifications:
  email:
    enabled: true
  slack:
    enabled: true
    channel: "#production-alerts"

performance:
  max_memory_usage: "4GB"
  cache_size: "1GB"
```

## Configuration Validation

### Built-in Validation

```bash
# Validate configuration file
ai-dep-manager configure validate

# Validate specific configuration
ai-dep-manager configure validate --config /path/to/config.yaml

# Show validation errors in detail
ai-dep-manager configure validate --verbose
```

### Custom Validation Rules

```yaml
# validation.yaml
rules:
  - name: "database_path_exists"
    type: "file_exists"
    path: "database.path"
    required: true
  
  - name: "log_level_valid"
    type: "enum"
    path: "logging.level"
    values: ["debug", "info", "warn", "error"]
  
  - name: "agent_schedule_valid"
    type: "cron"
    path: "agent.schedule"
    required: false
```

## Configuration Templates

### Minimal Configuration

```yaml
# minimal-config.yaml
database:
  path: "~/.ai-dep-manager/data.db"

logging:
  level: "info"

agent:
  enabled: true
  schedule: "0 2 * * *"
```

### Security-Focused Configuration

```yaml
# security-config.yaml
security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  whitelist_enabled: true
  master_key: "${MASTER_KEY}"
  vulnerability_db_update_interval: "6h"

notifications:
  email:
    enabled: true
    recipients:
      critical: ["security@company.com"]

agent:
  auto_update: false  # Require manual approval
```

### High-Performance Configuration

```yaml
# performance-config.yaml
performance:
  max_memory_usage: "8GB"
  cache_size: "2GB"
  max_goroutines: 2000

agent:
  max_concurrent_scans: 20
  batch_size: 500

database:
  max_connections: 100
  connection_timeout: "10s"
```

## Configuration Management Best Practices

### 1. Environment Variables for Secrets

```bash
# Use environment variables for sensitive data
export AI_DEP_MANAGER_DATABASE_PASSWORD="secure_password"
export AI_DEP_MANAGER_MASTER_KEY="base64-encoded-key"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
```

### 2. Configuration Inheritance

```yaml
# base-config.yaml
database: &database
  max_connections: 10
  connection_timeout: "30s"

logging: &logging
  format: "json"
  max_size: "100MB"

# prod-config.yaml
database:
  <<: *database
  host: "prod-db.company.com"

logging:
  <<: *logging
  level: "warn"
```

### 3. Configuration Validation in CI/CD

```bash
# In CI/CD pipeline
ai-dep-manager configure validate --config config/production.yaml
ai-dep-manager configure test --config config/production.yaml
```

### 4. Configuration Backup

```bash
# Backup configuration
cp ~/.ai-dep-manager/config.yaml ~/.ai-dep-manager/config.yaml.backup

# Version control configuration (excluding secrets)
git add config/base-config.yaml
git add config/production.yaml.template
```

---

For more information, see the [User Guide](user-guide.md) and [API Reference](api-reference.md).
