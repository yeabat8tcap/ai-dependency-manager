# AI Dependency Manager - Production Deployment Guide

## 🚀 Production Deployment Overview

This guide provides comprehensive instructions for deploying the AI Dependency Manager to production environments with enterprise-grade reliability, monitoring, and security.

## 📋 Prerequisites

### System Requirements
- **Operating System**: Linux (Ubuntu 20.04+ or CentOS 8+ recommended)
- **Architecture**: x86_64 (AMD64)
- **Memory**: Minimum 2GB RAM, Recommended 4GB+
- **Storage**: Minimum 10GB free space, Recommended 50GB+
- **CPU**: Minimum 2 cores, Recommended 4+ cores
- **Network**: Internet access for dependency registry APIs

### Dependencies
- **Go**: Version 1.21+ (for building from source)
- **SQLite**: Version 3.35+ (embedded database)
- **systemd**: For service management
- **cron**: For scheduled tasks
- **curl**: For health checks and webhooks
- **logrotate**: For log management

## 🔧 Quick Production Deployment

### 1. Automated Deployment Script

The fastest way to deploy to production is using our automated deployment script:

```bash
# Clone the repository
git clone https://github.com/8tcapital/ai-dep-manager.git
cd ai-dep-manager

# Run production deployment (requires root)
sudo ./scripts/deploy-production.sh

# Setup monitoring and logging
sudo ./scripts/setup-monitoring.sh
```

### 2. Manual Deployment Steps

If you prefer manual deployment or need customization:

#### Step 1: Create System User
```bash
sudo groupadd --system ai-dep-manager
sudo useradd --system --gid ai-dep-manager --home-dir /opt/ai-dep-manager \
             --shell /bin/false --comment "AI Dependency Manager" ai-dep-manager
```

#### Step 2: Create Directory Structure
```bash
sudo mkdir -p /opt/ai-dep-manager/{bin,config,logs,data,scripts,monitoring}
sudo chown -R ai-dep-manager:ai-dep-manager /opt/ai-dep-manager
sudo chmod 755 /opt/ai-dep-manager
```

#### Step 3: Build and Deploy Binary
```bash
# Build for production
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=$(git describe --tags --always) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o ai-dep-manager .

# Deploy binary
sudo cp ai-dep-manager /opt/ai-dep-manager/bin/
sudo chown ai-dep-manager:ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager
sudo chmod 755 /opt/ai-dep-manager/bin/ai-dep-manager
```

#### Step 4: Configure Application
```bash
# Create production configuration
sudo tee /opt/ai-dep-manager/config/config.yaml << EOF
# AI Dependency Manager Production Configuration

# Database Configuration
database:
  type: sqlite
  path: /opt/ai-dep-manager/data/ai-dep-manager.db

# Logging Configuration
log_level: info
log_format: json

# Agent Configuration
agent:
  scan_interval: 1h
  auto_update_level: none
  notification_mode: webhook
  max_concurrency: 10

# Security Configuration
security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
  trusted_registries:
    - https://registry.npmjs.org
    - https://pypi.org
    - https://repo1.maven.org/maven2

# Notification Configuration
notifications:
  webhook:
    url: "${WEBHOOK_URL}"
    timeout: 30s
  email:
    smtp_host: "${SMTP_HOST}"
    smtp_port: 587
    from: "${EMAIL_FROM}"
    to: ["${EMAIL_TO}"]

# Performance Configuration
performance:
  cache_ttl: 1h
  request_timeout: 30s
  max_retries: 3
EOF

sudo chown ai-dep-manager:ai-dep-manager /opt/ai-dep-manager/config/config.yaml
sudo chmod 600 /opt/ai-dep-manager/config/config.yaml
```

#### Step 5: Create Systemd Service
```bash
sudo tee /etc/systemd/system/ai-dep-manager.service << EOF
[Unit]
Description=AI Dependency Manager
Documentation=https://github.com/8tcapital/ai-dep-manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=ai-dep-manager
Group=ai-dep-manager
WorkingDirectory=/opt/ai-dep-manager
ExecStart=/opt/ai-dep-manager/bin/ai-dep-manager agent start --config /opt/ai-dep-manager/config/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ai-dep-manager

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/ai-dep-manager/data /opt/ai-dep-manager/logs
CapabilityBoundingSet=
AmbientCapabilities=
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

# Environment
Environment=HOME=/opt/ai-dep-manager
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable ai-dep-manager
sudo systemctl start ai-dep-manager
```

## 📊 Monitoring and Logging

### Health Monitoring
The deployment includes comprehensive health monitoring:

- **Health Checks**: Every 5 minutes via cron
- **Metrics Collection**: Every minute
- **Log Rotation**: Daily with 30-day retention
- **Automated Backups**: Daily at 2 AM

### Log Files
- **Application Logs**: `/opt/ai-dep-manager/logs/application.log`
- **Health Check Logs**: `/opt/ai-dep-manager/logs/health-check.log`
- **Metrics Logs**: `/opt/ai-dep-manager/logs/metrics.log`
- **Alert Logs**: `/opt/ai-dep-manager/logs/alerts.log`

### Monitoring Commands
```bash
# Check service status
sudo systemctl status ai-dep-manager

# View real-time logs
sudo journalctl -u ai-dep-manager -f

# Check health status
sudo -u ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager status

# View recent metrics
tail -f /opt/ai-dep-manager/logs/metrics.log
```

## 🔒 Security Configuration

### SSL/TLS Configuration
For production environments, ensure all external communications use HTTPS:

```yaml
# In config.yaml
security:
  tls:
    enabled: true
    cert_file: /opt/ai-dep-manager/certs/server.crt
    key_file: /opt/ai-dep-manager/certs/server.key
```

### Firewall Configuration
```bash
# Allow only necessary ports
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP (if needed)
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### Credential Management
Store sensitive credentials securely:

```bash
# Create credentials directory
sudo mkdir -p /opt/ai-dep-manager/secrets
sudo chmod 700 /opt/ai-dep-manager/secrets
sudo chown ai-dep-manager:ai-dep-manager /opt/ai-dep-manager/secrets

# Store credentials (example)
echo "your-registry-token" | sudo tee /opt/ai-dep-manager/secrets/npm-token
sudo chmod 600 /opt/ai-dep-manager/secrets/npm-token
```

## 🔄 Backup and Recovery

### Automated Backups
Backups are automatically created daily and stored in `/var/backups/ai-dep-manager/`:

```bash
# Manual backup
sudo /opt/ai-dep-manager/scripts/backup.sh

# Restore from backup
sudo tar -xzf /var/backups/ai-dep-manager/ai-dep-manager_backup_YYYYMMDD_HHMMSS.tar.gz -C /opt/ai-dep-manager/
```

### Database Recovery
```bash
# Stop service
sudo systemctl stop ai-dep-manager

# Restore database
sudo cp /path/to/backup/ai-dep-manager.db /opt/ai-dep-manager/data/

# Fix permissions
sudo chown ai-dep-manager:ai-dep-manager /opt/ai-dep-manager/data/ai-dep-manager.db

# Start service
sudo systemctl start ai-dep-manager
```

## 🚨 Alerting and Notifications

### Webhook Configuration
Configure webhook alerts for critical events:

```bash
# Set webhook URL environment variable
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# Test webhook
sudo -u ai-dep-manager /opt/ai-dep-manager/monitoring/scripts/send-alert.sh \
    "TEST" "Test alert message" "INFO"
```

### Email Configuration
Configure SMTP for email alerts:

```bash
# Install mail utilities
sudo apt-get install mailutils

# Configure postfix or use external SMTP
export SMTP_HOST="smtp.gmail.com"
export EMAIL_FROM="alerts@yourcompany.com"
export EMAIL_TO="admin@yourcompany.com"
```

## 🔧 Maintenance and Updates

### Application Updates
```bash
# Stop service
sudo systemctl stop ai-dep-manager

# Backup current version
sudo cp /opt/ai-dep-manager/bin/ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager.backup

# Deploy new version
sudo cp new-ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager
sudo chown ai-dep-manager:ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager

# Start service
sudo systemctl start ai-dep-manager

# Verify deployment
sudo systemctl status ai-dep-manager
```

### Configuration Updates
```bash
# Edit configuration
sudo nano /opt/ai-dep-manager/config/config.yaml

# Reload service
sudo systemctl reload ai-dep-manager
```

## 📈 Performance Tuning

### Database Optimization
```bash
# Vacuum database periodically
sudo -u ai-dep-manager sqlite3 /opt/ai-dep-manager/data/ai-dep-manager.db "VACUUM;"

# Analyze database
sudo -u ai-dep-manager sqlite3 /opt/ai-dep-manager/data/ai-dep-manager.db "ANALYZE;"
```

### Resource Limits
Configure systemd resource limits:

```ini
# In /etc/systemd/system/ai-dep-manager.service
[Service]
MemoryMax=2G
CPUQuota=200%
TasksMax=1000
```

## 🆘 Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status ai-dep-manager

# Check logs
sudo journalctl -u ai-dep-manager -n 50

# Check configuration
sudo -u ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager --config /opt/ai-dep-manager/config/config.yaml version
```

#### Database Issues
```bash
# Check database permissions
ls -la /opt/ai-dep-manager/data/

# Test database connectivity
sudo -u ai-dep-manager /opt/ai-dep-manager/bin/ai-dep-manager status
```

#### High Resource Usage
```bash
# Check resource usage
top -p $(pgrep ai-dep-manager)

# Check disk usage
du -sh /opt/ai-dep-manager/

# Clean old logs
sudo find /opt/ai-dep-manager/logs -name "*.log" -mtime +7 -delete
```

## 📞 Support and Maintenance

### Regular Maintenance Tasks
- **Daily**: Review health check logs
- **Weekly**: Check disk usage and clean old logs
- **Monthly**: Review and update dependencies
- **Quarterly**: Security audit and configuration review

### Getting Help
- **Documentation**: Check the full documentation in `/docs/`
- **Logs**: Always check application and system logs first
- **Health Checks**: Use built-in health check commands
- **Community**: Submit issues to the GitHub repository

## 🎉 Production Checklist

Before going live, ensure:

- [ ] Service starts automatically on boot
- [ ] Health checks are running every 5 minutes
- [ ] Logs are being rotated properly
- [ ] Backups are created daily
- [ ] Monitoring alerts are configured
- [ ] Security settings are properly configured
- [ ] SSL/TLS is enabled for external communications
- [ ] Firewall rules are in place
- [ ] Resource limits are configured
- [ ] Documentation is accessible to operations team

---

**🚀 Your AI Dependency Manager is now ready for production! 🚀**

For additional support and advanced configuration options, refer to the complete documentation suite in the `/docs/` directory.
