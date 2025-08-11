# AI Dependency Manager - Deployment Guide

This guide covers various deployment scenarios for the AI Dependency Manager in production environments.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Installation Methods](#installation-methods)
3. [System Service Deployment](#system-service-deployment)
4. [Docker Deployment](#docker-deployment)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Configuration Management](#configuration-management)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Security Considerations](#security-considerations)
9. [Backup and Recovery](#backup-and-recovery)
10. [Troubleshooting](#troubleshooting)

## System Requirements

### Minimum Requirements
- **OS**: Linux (Ubuntu 18.04+, CentOS 7+, RHEL 7+), macOS 10.15+, Windows 10+
- **CPU**: 2 cores
- **RAM**: 2GB
- **Storage**: 1GB free space
- **Go**: 1.21+ (for building from source)

### Recommended Requirements
- **CPU**: 4+ cores
- **RAM**: 4GB+
- **Storage**: 10GB+ free space (for logs and database)
- **Network**: Stable internet connection for package registry access

### Dependencies
- **Node.js & npm**: For npm project support
- **Python 3.x & pip**: For Python project support
- **Java & Maven/Gradle**: For Java project support
- **Git**: For repository operations

## Installation Methods

### 1. Binary Installation

```bash
# Download latest release
curl -L https://github.com/8tcapital/ai-dep-manager/releases/latest/download/ai-dep-manager-linux-amd64 -o ai-dep-manager
chmod +x ai-dep-manager
sudo mv ai-dep-manager /usr/local/bin/

# Verify installation
ai-dep-manager version
```

### 2. Go Install

```bash
go install github.com/8tcapital/ai-dep-manager@latest
```

### 3. Build from Source

```bash
git clone https://github.com/8tcapital/ai-dep-manager.git
cd ai-dep-manager
make build
sudo make install
```

### 4. Package Managers

```bash
# Homebrew (macOS)
brew install ai-dep-manager

# APT (Ubuntu/Debian)
curl -fsSL https://packages.8tcapital.com/gpg.key | sudo apt-key add -
echo "deb https://packages.8tcapital.com/apt stable main" | sudo tee /etc/apt/sources.list.d/ai-dep-manager.list
sudo apt update
sudo apt install ai-dep-manager

# YUM (CentOS/RHEL)
sudo yum-config-manager --add-repo https://packages.8tcapital.com/yum/ai-dep-manager.repo
sudo yum install ai-dep-manager
```

## System Service Deployment

### Automated Installation

```bash
# Download and run deployment script
curl -fsSL https://raw.githubusercontent.com/8tcapital/ai-dep-manager/main/scripts/deploy.sh | sudo bash -s -- install

# Or download and inspect first
curl -fsSL https://raw.githubusercontent.com/8tcapital/ai-dep-manager/main/scripts/deploy.sh -o deploy.sh
chmod +x deploy.sh
sudo ./deploy.sh install
```

### Manual systemd Service Setup

1. **Create service user:**
```bash
sudo useradd --system --shell /bin/false --home-dir /var/lib/ai-dep-manager ai-dep-manager
sudo mkdir -p /var/lib/ai-dep-manager
sudo chown ai-dep-manager:ai-dep-manager /var/lib/ai-dep-manager
```

2. **Create systemd service file:**
```bash
sudo tee /etc/systemd/system/ai-dep-manager.service > /dev/null <<EOF
[Unit]
Description=AI Dependency Manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=ai-dep-manager
Group=ai-dep-manager
ExecStart=/usr/local/bin/ai-dep-manager agent start --foreground
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
Restart=always
RestartSec=5
TimeoutStopSec=30
Environment=AI_DEP_MANAGER_DATA_DIR=/var/lib/ai-dep-manager
Environment=AI_DEP_MANAGER_CONFIG_FILE=/etc/ai-dep-manager/config.yaml

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/ai-dep-manager
CapabilityBoundingSet=
AmbientCapabilities=
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

[Install]
WantedBy=multi-user.target
EOF
```

3. **Create configuration directory:**
```bash
sudo mkdir -p /etc/ai-dep-manager
sudo cp config.yaml.example /etc/ai-dep-manager/config.yaml
sudo chown -R ai-dep-manager:ai-dep-manager /etc/ai-dep-manager
```

4. **Enable and start service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable ai-dep-manager
sudo systemctl start ai-dep-manager
sudo systemctl status ai-dep-manager
```

### Service Management

```bash
# Start service
sudo systemctl start ai-dep-manager

# Stop service
sudo systemctl stop ai-dep-manager

# Restart service
sudo systemctl restart ai-dep-manager

# Check status
sudo systemctl status ai-dep-manager

# View logs
sudo journalctl -u ai-dep-manager -f

# Enable auto-start
sudo systemctl enable ai-dep-manager
```

## Docker Deployment

### Single Container

```bash
# Pull image
docker pull ai-dep-manager:latest

# Run container
docker run -d \
  --name ai-dep-manager \
  --restart unless-stopped \
  -v /path/to/config:/app/config:ro \
  -v /path/to/data:/app/data \
  -v /path/to/projects:/app/projects:ro \
  -p 8080:8080 \
  ai-dep-manager:latest
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  ai-dep-manager:
    image: ai-dep-manager:latest
    container_name: ai-dep-manager
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config:ro
      - ./data:/app/data
      - ./projects:/app/projects:ro
      - ./logs:/app/logs
    environment:
      - AI_DEP_MANAGER_CONFIG_FILE=/app/config/config.yaml
      - AI_DEP_MANAGER_DATA_DIR=/app/data
      - AI_DEP_MANAGER_LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "ai-dep-manager", "status"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Database (if using PostgreSQL)
  postgres:
    image: postgres:13
    container_name: ai-dep-manager-db
    restart: unless-stopped
    environment:
      POSTGRES_DB: ai_dep_manager
      POSTGRES_USER: ai_dep_manager
      POSTGRES_PASSWORD: secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

Deploy with:
```bash
docker-compose up -d
```

### Multi-Stage Build

Create custom `Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-dep-manager .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates nodejs npm python3 py3-pip openjdk11 maven
WORKDIR /root/

COPY --from=builder /app/ai-dep-manager .
COPY config.yaml.example /app/config/config.yaml

EXPOSE 8080
CMD ["./ai-dep-manager", "agent", "start", "--foreground"]
```

## Kubernetes Deployment

### Namespace and ConfigMap

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ai-dep-manager

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-dep-manager-config
  namespace: ai-dep-manager
data:
  config.yaml: |
    database:
      path: "/data/ai-dep-manager.db"
    logging:
      level: "info"
      format: "json"
    agent:
      enabled: true
      schedule: "0 2 * * *"
    security:
      enable_vulnerability_scanning: true
```

### Deployment and Service

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-dep-manager
  namespace: ai-dep-manager
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ai-dep-manager
  template:
    metadata:
      labels:
        app: ai-dep-manager
    spec:
      containers:
      - name: ai-dep-manager
        image: ai-dep-manager:latest
        ports:
        - containerPort: 8080
        env:
        - name: AI_DEP_MANAGER_CONFIG_FILE
          value: "/config/config.yaml"
        - name: AI_DEP_MANAGER_DATA_DIR
          value: "/data"
        volumeMounts:
        - name: config
          mountPath: /config
        - name: data
          mountPath: /data
        livenessProbe:
          exec:
            command:
            - ai-dep-manager
            - status
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          exec:
            command:
            - ai-dep-manager
            - status
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: ai-dep-manager-config
      - name: data
        persistentVolumeClaim:
          claimName: ai-dep-manager-data

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: ai-dep-manager-service
  namespace: ai-dep-manager
spec:
  selector:
    app: ai-dep-manager
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP

---
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ai-dep-manager-data
  namespace: ai-dep-manager
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

Deploy to Kubernetes:
```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
```

### Ingress (Optional)

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ai-dep-manager-ingress
  namespace: ai-dep-manager
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - ai-dep-manager.yourdomain.com
    secretName: ai-dep-manager-tls
  rules:
  - host: ai-dep-manager.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ai-dep-manager-service
            port:
              number: 80
```

## Configuration Management

### Environment-Specific Configurations

#### Development
```yaml
# config-dev.yaml
database:
  path: "./data/dev.db"
logging:
  level: "debug"
agent:
  enabled: false
security:
  enable_vulnerability_scanning: false
```

#### Staging
```yaml
# config-staging.yaml
database:
  path: "/data/staging.db"
logging:
  level: "info"
agent:
  enabled: true
  schedule: "0 */6 * * *"
security:
  enable_vulnerability_scanning: true
```

#### Production
```yaml
# config-prod.yaml
database:
  path: "/data/production.db"
logging:
  level: "warn"
  format: "json"
agent:
  enabled: true
  schedule: "0 2 * * *"
  max_concurrent_scans: 5
security:
  enable_vulnerability_scanning: true
  enable_integrity_verification: true
notifications:
  email:
    enabled: true
  slack:
    enabled: true
```

### Configuration Validation

```bash
# Validate configuration
ai-dep-manager configure validate

# Test configuration
ai-dep-manager configure test

# Show effective configuration
ai-dep-manager configure show --effective
```

## Monitoring and Logging

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check
curl http://localhost:8080/health/detailed
```

### Metrics Collection

Configure Prometheus metrics:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'ai-dep-manager'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Log Management

#### Centralized Logging with ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/ai-dep-manager/*.log
  fields:
    service: ai-dep-manager
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]

setup.kibana:
  host: "kibana:5601"
```

#### Log Rotation

```bash
# /etc/logrotate.d/ai-dep-manager
/var/log/ai-dep-manager/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 ai-dep-manager ai-dep-manager
    postrotate
        systemctl reload ai-dep-manager
    endscript
}
```

## Security Considerations

### Network Security

1. **Firewall Configuration:**
```bash
# Allow only necessary ports
sudo ufw allow 22/tcp  # SSH
sudo ufw allow 8080/tcp  # AI Dep Manager (if needed)
sudo ufw enable
```

2. **TLS/SSL Configuration:**
```yaml
# config.yaml
server:
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/ai-dep-manager.crt"
    key_file: "/etc/ssl/private/ai-dep-manager.key"
```

### Access Control

1. **Service Account:**
```bash
# Create dedicated user
sudo useradd --system --shell /bin/false ai-dep-manager
```

2. **File Permissions:**
```bash
# Secure configuration files
chmod 600 /etc/ai-dep-manager/config.yaml
chown ai-dep-manager:ai-dep-manager /etc/ai-dep-manager/config.yaml
```

### Credential Management

1. **Environment Variables:**
```bash
# Use environment variables for sensitive data
export AI_DEP_MANAGER_DB_PASSWORD="secure_password"
export AI_DEP_MANAGER_MASTER_KEY="encryption_key"
```

2. **External Secret Management:**
```yaml
# Using HashiCorp Vault
security:
  credentials:
    vault:
      enabled: true
      address: "https://vault.company.com"
      token_file: "/var/lib/ai-dep-manager/vault-token"
```

## Backup and Recovery

### Database Backup

```bash
#!/bin/bash
# backup-db.sh

BACKUP_DIR="/var/backups/ai-dep-manager"
DATE=$(date +%Y%m%d_%H%M%S)
DB_PATH="/var/lib/ai-dep-manager/ai-dep-manager.db"

mkdir -p "$BACKUP_DIR"

# Create backup
sqlite3 "$DB_PATH" ".backup $BACKUP_DIR/backup_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/backup_$DATE.db"

# Remove old backups (keep 30 days)
find "$BACKUP_DIR" -name "backup_*.db.gz" -mtime +30 -delete
```

### Configuration Backup

```bash
#!/bin/bash
# backup-config.sh

BACKUP_DIR="/var/backups/ai-dep-manager"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Backup configuration
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" /etc/ai-dep-manager/

# Remove old backups
find "$BACKUP_DIR" -name "config_*.tar.gz" -mtime +30 -delete
```

### Automated Backup with Cron

```bash
# Add to crontab
0 2 * * * /usr/local/bin/backup-db.sh
0 3 * * * /usr/local/bin/backup-config.sh
```

### Recovery Procedures

1. **Database Recovery:**
```bash
# Stop service
sudo systemctl stop ai-dep-manager

# Restore database
gunzip -c /var/backups/ai-dep-manager/backup_20240101_020000.db.gz > /var/lib/ai-dep-manager/ai-dep-manager.db

# Fix permissions
chown ai-dep-manager:ai-dep-manager /var/lib/ai-dep-manager/ai-dep-manager.db

# Start service
sudo systemctl start ai-dep-manager
```

2. **Configuration Recovery:**
```bash
# Extract configuration backup
tar -xzf /var/backups/ai-dep-manager/config_20240101_030000.tar.gz -C /

# Restart service
sudo systemctl restart ai-dep-manager
```

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

```bash
# Check service status
sudo systemctl status ai-dep-manager

# Check logs
sudo journalctl -u ai-dep-manager -f

# Verify binary
which ai-dep-manager
ai-dep-manager version

# Check permissions
ls -la /usr/local/bin/ai-dep-manager
```

#### 2. Database Connection Issues

```bash
# Check database file
ls -la /var/lib/ai-dep-manager/ai-dep-manager.db

# Test database connection
sqlite3 /var/lib/ai-dep-manager/ai-dep-manager.db ".tables"

# Reset database if corrupted
sudo systemctl stop ai-dep-manager
sudo -u ai-dep-manager ai-dep-manager configure migrate-db
sudo systemctl start ai-dep-manager
```

#### 3. Network Connectivity Issues

```bash
# Test registry connectivity
curl -I https://registry.npmjs.org/
curl -I https://pypi.org/

# Check proxy settings
env | grep -i proxy

# Test DNS resolution
nslookup registry.npmjs.org
```

#### 4. Permission Issues

```bash
# Check service user permissions
sudo -u ai-dep-manager ls -la /var/lib/ai-dep-manager/

# Fix permissions
sudo chown -R ai-dep-manager:ai-dep-manager /var/lib/ai-dep-manager/
sudo chmod 755 /var/lib/ai-dep-manager/
```

### Performance Tuning

1. **Database Optimization:**
```yaml
database:
  max_connections: 20
  connection_timeout: 30s
  query_timeout: 60s
```

2. **Concurrent Processing:**
```yaml
agent:
  max_concurrent_scans: 5
  scan_timeout: 300s
  batch_size: 100
```

3. **Memory Management:**
```yaml
performance:
  max_memory_usage: "1GB"
  gc_percent: 100
  cache_size: "256MB"
```

### Monitoring Commands

```bash
# System resource usage
htop
iostat -x 1
df -h

# Application metrics
ai-dep-manager status --verbose
ai-dep-manager agent stats

# Log analysis
tail -f /var/log/ai-dep-manager/app.log | grep ERROR
journalctl -u ai-dep-manager --since "1 hour ago"
```

---

For additional support, see the [User Guide](user-guide.md) and [Troubleshooting](troubleshooting.md) documentation.
