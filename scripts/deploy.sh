#!/bin/bash

# AI Dependency Manager Deployment Script
# This script helps deploy the AI Dependency Manager in various environments

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="ai-dep-manager"
SERVICE_NAME="ai-dep-manager"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warn "Running as root. Consider using a non-root user for better security."
    fi
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."
    
    # Check Go version
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go version: $GO_VERSION"
    else
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Check package managers
    local missing_tools=()
    
    if ! command -v npm &> /dev/null; then
        missing_tools+=("npm")
    fi
    
    if ! command -v pip &> /dev/null && ! command -v pip3 &> /dev/null; then
        missing_tools+=("pip")
    fi
    
    if ! command -v mvn &> /dev/null; then
        missing_tools+=("maven")
    fi
    
    if ! command -v gradle &> /dev/null; then
        missing_tools+=("gradle")
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_warn "Missing package managers: ${missing_tools[*]}"
        log_warn "Some features may not work without these tools."
    fi
    
    log_success "System requirements check completed"
}

# Build the application
build_app() {
    log_info "Building AI Dependency Manager..."
    
    cd "$PROJECT_DIR"
    
    # Clean previous builds
    if [ -f "$BINARY_NAME" ]; then
        rm "$BINARY_NAME"
    fi
    
    # Build with version information
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    
    log_info "Building version: $VERSION (commit: $COMMIT)"
    
    go build -ldflags "-X main.Version=$VERSION -X main.GitCommit=$COMMIT -X main.BuildTime=$BUILD_TIME" -o "$BINARY_NAME" .
    
    if [ $? -eq 0 ]; then
        log_success "Build completed successfully"
    else
        log_error "Build failed"
        exit 1
    fi
}

# Install the application
install_app() {
    log_info "Installing AI Dependency Manager..."
    
    local install_dir="/usr/local/bin"
    local config_dir="/etc/ai-dep-manager"
    local data_dir="/var/lib/ai-dep-manager"
    
    # Create directories
    sudo mkdir -p "$config_dir" "$data_dir"
    
    # Copy binary
    sudo cp "$PROJECT_DIR/$BINARY_NAME" "$install_dir/"
    sudo chmod +x "$install_dir/$BINARY_NAME"
    
    # Copy example configuration
    if [ -f "$PROJECT_DIR/config.yaml.example" ]; then
        sudo cp "$PROJECT_DIR/config.yaml.example" "$config_dir/config.yaml"
        log_info "Example configuration copied to $config_dir/config.yaml"
        log_warn "Please edit the configuration file before starting the service"
    fi
    
    # Set permissions
    sudo chown -R $USER:$USER "$data_dir"
    
    log_success "Installation completed"
    log_info "Binary installed to: $install_dir/$BINARY_NAME"
    log_info "Configuration: $config_dir/config.yaml"
    log_info "Data directory: $data_dir"
}

# Install systemd service
install_service() {
    log_info "Installing systemd service..."
    
    local service_file="/etc/systemd/system/$SERVICE_NAME.service"
    local binary_path="/usr/local/bin/$BINARY_NAME"
    local config_path="/etc/ai-dep-manager/config.yaml"
    local data_dir="/var/lib/ai-dep-manager"
    
    # Create service file
    sudo tee "$service_file" > /dev/null <<EOF
[Unit]
Description=AI Dependency Manager
Documentation=https://github.com/8tcapital/ai-dep-manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$data_dir
ExecStart=$binary_path agent start --daemon
ExecStop=$binary_path agent stop
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment=AI_DEP_MANAGER_CONFIG_FILE=$config_path
Environment=AI_DEP_MANAGER_DATA_DIR=$data_dir

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$data_dir

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd and enable service
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    
    log_success "Systemd service installed and enabled"
    log_info "Service file: $service_file"
    log_info "Start service: sudo systemctl start $SERVICE_NAME"
    log_info "Check status: sudo systemctl status $SERVICE_NAME"
}

# Deploy with Docker
deploy_docker() {
    log_info "Deploying with Docker..."
    
    cd "$PROJECT_DIR"
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Build Docker image
    log_info "Building Docker image..."
    docker build -t ai-dep-manager:latest .
    
    # Check if docker-compose is available
    if command -v docker-compose &> /dev/null; then
        log_info "Starting with docker-compose..."
        docker-compose up -d
        
        log_success "Docker deployment completed"
        log_info "Check status: docker-compose ps"
        log_info "View logs: docker-compose logs -f ai-dep-manager"
        log_info "Stop: docker-compose down"
    else
        log_info "Starting with docker run..."
        
        # Create data volume
        docker volume create ai-dep-manager-data
        
        # Run container
        docker run -d \
            --name ai-dep-manager \
            --restart unless-stopped \
            -v ai-dep-manager-data:/data \
            -v "$(pwd)/projects:/projects:ro" \
            ai-dep-manager:latest
        
        log_success "Docker deployment completed"
        log_info "Check status: docker ps"
        log_info "View logs: docker logs -f ai-dep-manager"
        log_info "Stop: docker stop ai-dep-manager"
    fi
}

# Uninstall the application
uninstall_app() {
    log_info "Uninstalling AI Dependency Manager..."
    
    # Stop and disable service if it exists
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        sudo systemctl stop "$SERVICE_NAME"
        log_info "Service stopped"
    fi
    
    if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
        sudo systemctl disable "$SERVICE_NAME"
        log_info "Service disabled"
    fi
    
    # Remove service file
    if [ -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
        sudo rm "/etc/systemd/system/$SERVICE_NAME.service"
        sudo systemctl daemon-reload
        log_info "Service file removed"
    fi
    
    # Remove binary
    if [ -f "/usr/local/bin/$BINARY_NAME" ]; then
        sudo rm "/usr/local/bin/$BINARY_NAME"
        log_info "Binary removed"
    fi
    
    # Ask about configuration and data
    read -p "Remove configuration files? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo rm -rf "/etc/ai-dep-manager"
        log_info "Configuration files removed"
    fi
    
    read -p "Remove data directory? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo rm -rf "/var/lib/ai-dep-manager"
        log_info "Data directory removed"
    fi
    
    log_success "Uninstallation completed"
}

# Show usage
usage() {
    echo "AI Dependency Manager Deployment Script"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  build       Build the application"
    echo "  install     Install the application (requires sudo)"
    echo "  service     Install systemd service (requires sudo)"
    echo "  docker      Deploy with Docker"
    echo "  uninstall   Uninstall the application (requires sudo)"
    echo "  check       Check system requirements"
    echo "  help        Show this help message"
    echo
    echo "Examples:"
    echo "  $0 build                # Build the application"
    echo "  $0 install              # Build and install"
    echo "  $0 service              # Install with systemd service"
    echo "  $0 docker               # Deploy with Docker"
}

# Main script logic
main() {
    case "${1:-}" in
        build)
            check_requirements
            build_app
            ;;
        install)
            check_root
            check_requirements
            build_app
            install_app
            ;;
        service)
            check_root
            check_requirements
            build_app
            install_app
            install_service
            ;;
        docker)
            deploy_docker
            ;;
        uninstall)
            check_root
            uninstall_app
            ;;
        check)
            check_requirements
            ;;
        help|--help|-h)
            usage
            ;;
        "")
            usage
            exit 1
            ;;
        *)
            log_error "Unknown command: $1"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
