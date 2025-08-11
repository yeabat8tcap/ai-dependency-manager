#!/bin/bash

# AI Dependency Manager - Production Deployment Script
# This script automates the deployment of the AI Dependency Manager to production

set -e

# Configuration
APP_NAME="ai-dep-manager"
VERSION=${VERSION:-$(git describe --tags --always 2>/dev/null || echo "dev")}
BUILD_DIR="./build"
DEPLOY_DIR="/opt/ai-dep-manager"
SERVICE_NAME="ai-dep-manager"
USER="ai-dep-manager"
GROUP="ai-dep-manager"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root for production deployment"
        exit 1
    fi
}

# Create system user and group
create_user() {
    log_info "Creating system user and group..."
    
    if ! getent group "$GROUP" > /dev/null 2>&1; then
        groupadd --system "$GROUP"
        log_success "Created group: $GROUP"
    fi
    
    if ! getent passwd "$USER" > /dev/null 2>&1; then
        useradd --system --gid "$GROUP" --home-dir "$DEPLOY_DIR" \
                --shell /bin/false --comment "AI Dependency Manager" "$USER"
        log_success "Created user: $USER"
    fi
}

# Create directory structure
create_directories() {
    log_info "Creating directory structure..."
    
    mkdir -p "$DEPLOY_DIR"/{bin,config,logs,data,scripts}
    chown -R "$USER:$GROUP" "$DEPLOY_DIR"
    chmod 755 "$DEPLOY_DIR"
    
    log_success "Directory structure created"
}

# Build application
build_application() {
    log_info "Building application..."
    
    # Clean previous builds
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Build for production
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags "-X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o "$BUILD_DIR/$APP_NAME" .
    
    # Make executable
    chmod +x "$BUILD_DIR/$APP_NAME"
    
    log_success "Application built successfully"
}

# Deploy application
deploy_application() {
    log_info "Deploying application..."
    
    # Stop service if running
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_info "Stopping existing service..."
        systemctl stop "$SERVICE_NAME"
    fi
    
    # Copy binary
    cp "$BUILD_DIR/$APP_NAME" "$DEPLOY_DIR/bin/"
    chown "$USER:$GROUP" "$DEPLOY_DIR/bin/$APP_NAME"
    chmod 755 "$DEPLOY_DIR/bin/$APP_NAME"
    
    # Copy configuration files
    if [[ -f "config/production.yaml" ]]; then
        cp config/production.yaml "$DEPLOY_DIR/config/config.yaml"
        chown "$USER:$GROUP" "$DEPLOY_DIR/config/config.yaml"
        chmod 600 "$DEPLOY_DIR/config/config.yaml"
    fi
    
    log_success "Application deployed"
}

# Create systemd service
create_service() {
    log_info "Creating systemd service..."
    
    cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=AI Dependency Manager
Documentation=https://github.com/8tcapital/ai-dep-manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
Group=$GROUP
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/bin/$APP_NAME agent start --config $DEPLOY_DIR/config/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$DEPLOY_DIR/data $DEPLOY_DIR/logs
CapabilityBoundingSet=
AmbientCapabilities=
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

# Environment
Environment=HOME=$DEPLOY_DIR
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    
    log_success "Systemd service created and enabled"
}

# Setup log rotation
setup_logrotate() {
    log_info "Setting up log rotation..."
    
    cat > "/etc/logrotate.d/$SERVICE_NAME" << EOF
$DEPLOY_DIR/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $USER $GROUP
    postrotate
        systemctl reload $SERVICE_NAME > /dev/null 2>&1 || true
    endscript
}
EOF

    log_success "Log rotation configured"
}

# Create backup script
create_backup_script() {
    log_info "Creating backup script..."
    
    cat > "$DEPLOY_DIR/scripts/backup.sh" << 'EOF'
#!/bin/bash

# AI Dependency Manager Backup Script

BACKUP_DIR="/var/backups/ai-dep-manager"
DATA_DIR="/opt/ai-dep-manager/data"
CONFIG_DIR="/opt/ai-dep-manager/config"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="ai-dep-manager_backup_$TIMESTAMP.tar.gz"

mkdir -p "$BACKUP_DIR"

# Create backup
tar -czf "$BACKUP_DIR/$BACKUP_NAME" \
    -C /opt/ai-dep-manager \
    data config

# Keep only last 7 backups
find "$BACKUP_DIR" -name "ai-dep-manager_backup_*.tar.gz" -mtime +7 -delete

echo "Backup created: $BACKUP_DIR/$BACKUP_NAME"
EOF

    chmod +x "$DEPLOY_DIR/scripts/backup.sh"
    chown "$USER:$GROUP" "$DEPLOY_DIR/scripts/backup.sh"
    
    # Add to crontab for daily backups
    echo "0 2 * * * $DEPLOY_DIR/scripts/backup.sh" | crontab -u "$USER" -
    
    log_success "Backup script created and scheduled"
}

# Start service
start_service() {
    log_info "Starting service..."
    
    systemctl start "$SERVICE_NAME"
    sleep 5
    
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_success "Service started successfully"
        systemctl status "$SERVICE_NAME" --no-pager
    else
        log_error "Failed to start service"
        systemctl status "$SERVICE_NAME" --no-pager
        exit 1
    fi
}

# Health check
health_check() {
    log_info "Performing health check..."
    
    # Wait for service to be ready
    sleep 10
    
    # Check if binary responds
    if sudo -u "$USER" "$DEPLOY_DIR/bin/$APP_NAME" version > /dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        exit 1
    fi
}

# Main deployment function
main() {
    log_info "Starting AI Dependency Manager production deployment..."
    log_info "Version: $VERSION"
    
    check_root
    create_user
    create_directories
    build_application
    deploy_application
    create_service
    setup_logrotate
    create_backup_script
    start_service
    health_check
    
    log_success "🎉 AI Dependency Manager deployed successfully!"
    log_info "Service status: $(systemctl is-active $SERVICE_NAME)"
    log_info "Logs: journalctl -u $SERVICE_NAME -f"
    log_info "Config: $DEPLOY_DIR/config/config.yaml"
    log_info "Data: $DEPLOY_DIR/data/"
}

# Run main function
main "$@"
