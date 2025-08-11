# AI Dependency Manager - Docker Deployment Guide

This guide provides comprehensive instructions for deploying the AI Dependency Manager using Docker and Docker Compose with the integrated logging system and unified full-stack architecture.

## 🚀 Quick Start

### Prerequisites
- Docker 20.10+ 
- Docker Compose 2.0+
- 4GB+ RAM available
- 10GB+ disk space

### Basic Deployment
```bash
# Clone the repository
git clone <repository-url>
cd AIDepManager

# Create required directories
mkdir -p data logs db config projects scan-targets

# Start the application
docker-compose up -d

# Access the application
open http://localhost:8080
```

## 📋 Deployment Configurations

### Development Environment
```bash
# Uses docker-compose.yml + docker-compose.override.yml automatically
docker-compose up -d

# View logs
docker-compose logs -f ai-dep-manager
```

### Production Environment
```bash
# Use production configuration
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Create production secrets
echo "your-secure-password" | docker secret create grafana_admin_password -

# Enable monitoring stack
docker-compose --profile monitoring up -d
```

### Staging Environment
```bash
# Create staging override
cp docker-compose.override.yml docker-compose.staging.yml

# Modify staging-specific settings
# Then deploy
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
```

## 🏗️ Architecture Overview

### Unified Full-Stack Application
The AI Dependency Manager runs as a single container with:
- **Frontend**: Angular 17+ with Material Design
- **Backend**: Go web server with embedded frontend assets
- **Database**: SQLite with automatic migrations
- **Logging**: Comprehensive logging system with structured output
- **API**: REST API with WebSocket support

### Container Structure
```
ai-dep-manager-app/
├── Frontend (Angular) - Port 8080
├── Backend API - Port 8080/api
├── Database (SQLite) - /data/ai-dep-manager.db
├── Logs - /app/logs/
└── Configuration - /data/config.yaml
```

## 🔧 Configuration

### Environment Variables

#### Application Configuration
```bash
AI_DEP_MANAGER_DATA_DIR=/data              # Data directory
AI_DEP_MANAGER_CONFIG_FILE=/data/config.yaml  # Configuration file
AI_DEP_MANAGER_DB_PATH=/data/ai-dep-manager.db # Database path
```

#### Logging Configuration
```bash
AI_DEP_MANAGER_LOG_LEVEL=info              # Log level (debug, info, warn, error)
AI_DEP_MANAGER_LOG_FORMAT=json             # Log format (json, text)
```

#### Web Server Configuration
```bash
AI_DEP_MANAGER_WEB_PORT=8080               # Web server port
AI_DEP_MANAGER_API_PORT=8081               # API server port (if separate)
GIN_MODE=release                           # Gin mode (debug, release)
```

#### Performance Tuning
```bash
GOMAXPROCS=2                               # Go max processes
GOGC=100                                   # Go garbage collection target
```

### Volume Mounts

#### Required Volumes
- `ai-dep-manager-data:/data` - Application data and database
- `ai-dep-manager-logs:/app/logs` - Application logs
- `ai-dep-manager-db:/app/db` - Database files

#### Optional Volumes
- `./projects:/projects:ro` - Project directories to scan
- `./scan-targets:/scan-targets:ro` - Additional scan targets
- `./config/production.yaml:/data/config.yaml:ro` - Configuration override

## 🔍 Monitoring and Logging

### Built-in Monitoring
- Health check endpoint: `http://localhost:8080/api/health`
- Status endpoint: `http://localhost:8080/api/status`
- Logs endpoint: `http://localhost:8080/api/logs`
- Logging test interface: `http://localhost:8080/logging-test`

### Optional Monitoring Stack
Enable with `--profile monitoring`:

#### Prometheus (Metrics)
- URL: `http://localhost:9090`
- Scrapes application metrics
- 200h data retention

#### Grafana (Dashboards)
- URL: `http://localhost:3000`
- Default credentials: admin/admin
- Pre-configured dashboards

#### Fluent Bit (Log Aggregation)
- Collects and forwards application logs
- Configurable output destinations

### Log Management
```bash
# View application logs
docker-compose logs -f ai-dep-manager

# View all logs
docker-compose logs -f

# Export logs
docker cp ai-dep-manager-app:/app/logs ./exported-logs
```

## 🛠️ Maintenance

### Backup and Restore

#### Backup
```bash
# Stop the application
docker-compose down

# Backup data
tar -czf backup-$(date +%Y%m%d).tar.gz data/ logs/ db/

# Restart
docker-compose up -d
```

#### Restore
```bash
# Stop the application
docker-compose down

# Restore data
tar -xzf backup-YYYYMMDD.tar.gz

# Restart
docker-compose up -d
```

### Updates and Upgrades

#### Application Updates
```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose build --no-cache
docker-compose up -d
```

#### Database Migrations
Database migrations run automatically on startup. Check logs for migration status:
```bash
docker-compose logs ai-dep-manager | grep migration
```

### Troubleshooting

#### Common Issues

**Container won't start:**
```bash
# Check logs
docker-compose logs ai-dep-manager

# Check health
docker-compose ps
```

**Database issues:**
```bash
# Reset database (WARNING: Data loss)
docker-compose down
rm -rf db/
docker-compose up -d
```

**Permission issues:**
```bash
# Fix permissions
sudo chown -R 1001:1001 data/ logs/ db/
```

**Port conflicts:**
```bash
# Change ports in docker-compose.yml
ports:
  - "8090:8080"  # Use different host port
```

## 🔒 Security

### Production Security Checklist
- [ ] Change default passwords
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Set up log rotation
- [ ] Enable security scanning
- [ ] Configure backup encryption
- [ ] Set resource limits
- [ ] Enable audit logging

### Network Security
```bash
# Restrict network access
networks:
  ai-dep-manager-network:
    driver: bridge
    internal: true  # Disable external access
```

### Secret Management
```bash
# Create secrets
echo "secure-password" | docker secret create db_password -
echo "api-key" | docker secret create api_key -
```

## 📊 Performance Tuning

### Resource Allocation
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
```

### Database Optimization
- Regular VACUUM operations
- Index optimization
- Connection pooling
- Query optimization

### Logging Optimization
- Appropriate log levels
- Log rotation
- Structured logging
- Centralized log aggregation

## 🌐 Production Deployment

### Load Balancing
```yaml
# Use with Traefik or nginx
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.ai-dep-manager.rule=Host(`yourdomain.com`)"
  - "traefik.http.routers.ai-dep-manager.tls=true"
```

### High Availability
- Multiple replicas
- Health checks
- Graceful shutdowns
- Database clustering
- Load balancing

### Scaling
```bash
# Scale horizontally
docker-compose up -d --scale ai-dep-manager=3
```

## 📞 Support

### Health Checks
- Application: `curl http://localhost:8080/api/health`
- Database: Check logs for connection status
- Logging: `curl http://localhost:8080/api/logs/test`

### Monitoring URLs
- Application: `http://localhost:8080`
- API Documentation: `http://localhost:8080/api`
- Logging Test: `http://localhost:8080/logging-test`
- Metrics: `http://localhost:9090` (if monitoring enabled)
- Dashboards: `http://localhost:3000` (if monitoring enabled)

For additional support, check the application logs and monitoring dashboards for detailed information about system status and performance.
