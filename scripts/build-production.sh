#!/bin/bash

# AI Dependency Manager - Production Build Script
# This script creates a comprehensive production build with all GitHub Integration Bot features

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
BUILD_DIR="build"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

echo -e "${BLUE}🚀 AI Dependency Manager - Production Build${NC}"
echo -e "${BLUE}================================================${NC}"
echo -e "Version: ${GREEN}$VERSION${NC}"
echo -e "Build Time: ${GREEN}$BUILD_TIME${NC}"
echo -e "Git Commit: ${GREEN}$GIT_COMMIT${NC}"
echo ""

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
if [ -d "$BUILD_DIR" ]; then
    rm -rf $BUILD_DIR || {
        echo -e "${RED}❌ Failed to clean build directory. You may need to run: sudo rm -rf $BUILD_DIR${NC}"
        exit 1
    }
fi
mkdir -p $BUILD_DIR/{bin,web,docs,config,data,logs,db,projects,scan-targets}

# Build frontend with Geist Mono font
echo -e "${YELLOW}🎨 Building frontend with Geist Mono font...${NC}"
cd web
npm ci --silent
npm run build:production
cd ..

# Copy frontend build to production directory
echo -e "${YELLOW}📦 Packaging frontend assets...${NC}"
cp -r web/dist/ai-dep-manager-frontend/browser/* $BUILD_DIR/web/

# Build Go backend with all GitHub Integration Bot features
echo -e "${YELLOW}🔨 Building Go backend with GitHub Integration Bot...${NC}"

# Set build flags
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT -w -s"

# Build for multiple platforms
echo -e "${BLUE}Building for multiple platforms...${NC}"

# Linux AMD64
echo -e "  ${GREEN}→${NC} Linux AMD64"
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/bin/ai-dep-manager-linux-amd64 .

# Linux ARM64
echo -e "  ${GREEN}→${NC} Linux ARM64"
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/bin/ai-dep-manager-linux-arm64 .

# macOS AMD64
echo -e "  ${GREEN}→${NC} macOS AMD64"
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/bin/ai-dep-manager-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
echo -e "  ${GREEN}→${NC} macOS ARM64"
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/bin/ai-dep-manager-darwin-arm64 .

# Windows AMD64
echo -e "  ${GREEN}→${NC} Windows AMD64"
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o $BUILD_DIR/bin/ai-dep-manager-windows-amd64.exe .

# Create Docker production image
echo -e "${YELLOW}🐳 Building Docker production image...${NC}"
docker build -t ai-dep-manager:$VERSION -t ai-dep-manager:latest .

# Copy documentation
echo -e "${YELLOW}📚 Packaging documentation...${NC}"
cp -r docs/* $BUILD_DIR/docs/
cp README.md $BUILD_DIR/
cp LICENSE $BUILD_DIR/
cp CONTRIBUTING.md $BUILD_DIR/

# Copy configuration templates
echo -e "${YELLOW}⚙️  Packaging configuration templates...${NC}"
cp config.yaml.example $BUILD_DIR/config/config.example.yaml
cp docker-compose.yml $BUILD_DIR/
cp Dockerfile $BUILD_DIR/

# Create deployment scripts
echo -e "${YELLOW}🚀 Creating deployment scripts...${NC}"

# Production deployment script
cat > $BUILD_DIR/deploy-production.sh << 'EOF'
#!/bin/bash

# AI Dependency Manager - Production Deployment Script
# Deploys the AI Dependency Manager with GitHub Integration Bot features

set -e

echo "🚀 Deploying AI Dependency Manager to Production"
echo "=============================================="

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo "❌ This script should not be run as root for security reasons"
   exit 1
fi

# Create application user
sudo useradd -r -s /bin/false ai-dep-manager 2>/dev/null || true

# Create directories
sudo mkdir -p /opt/ai-dep-manager/{bin,config,logs,data}
sudo mkdir -p /var/log/ai-dep-manager
sudo mkdir -p /etc/ai-dep-manager

# Copy binary
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

if [[ "$OS" == "linux" && "$ARCH" == "x86_64" ]]; then
    BINARY="ai-dep-manager-linux-amd64"
elif [[ "$OS" == "linux" && "$ARCH" == "aarch64" ]]; then
    BINARY="ai-dep-manager-linux-arm64"
elif [[ "$OS" == "darwin" && "$ARCH" == "x86_64" ]]; then
    BINARY="ai-dep-manager-darwin-amd64"
elif [[ "$OS" == "darwin" && "$ARCH" == "arm64" ]]; then
    BINARY="ai-dep-manager-darwin-arm64"
else
    echo "❌ Unsupported platform: $OS $ARCH"
    exit 1
fi

echo "📦 Installing binary: $BINARY"
sudo cp bin/$BINARY /opt/ai-dep-manager/bin/ai-dep-manager
sudo chmod +x /opt/ai-dep-manager/bin/ai-dep-manager

# Copy configuration
sudo cp config/config.example.yaml /etc/ai-dep-manager/config.yaml

# Set ownership
sudo chown -R ai-dep-manager:ai-dep-manager /opt/ai-dep-manager
sudo chown -R ai-dep-manager:ai-dep-manager /var/log/ai-dep-manager
sudo chown -R ai-dep-manager:ai-dep-manager /etc/ai-dep-manager

# Create systemd service
sudo tee /etc/systemd/system/ai-dep-manager.service > /dev/null << 'SYSTEMD_EOF'
[Unit]
Description=AI Dependency Manager with GitHub Integration Bot
Documentation=https://github.com/8tcapital/ai-dep-manager
After=network.target
Wants=network.target

[Service]
Type=simple
User=ai-dep-manager
Group=ai-dep-manager
ExecStart=/opt/ai-dep-manager/bin/ai-dep-manager serve --config /etc/ai-dep-manager/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ai-dep-manager

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/ai-dep-manager /var/log/ai-dep-manager /etc/ai-dep-manager
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_BIND_SERVICE

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
SYSTEMD_EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable ai-dep-manager
sudo systemctl start ai-dep-manager

echo "✅ AI Dependency Manager deployed successfully!"
echo "📊 Service status:"
sudo systemctl status ai-dep-manager --no-pager -l

echo ""
echo "🔧 Configuration file: /etc/ai-dep-manager/config.yaml"
echo "📝 Logs: journalctl -u ai-dep-manager -f"
echo "🌐 Web interface: http://localhost:8081"
echo ""
echo "🚀 GitHub Integration Bot Features Available:"
echo "   • Automated dependency patching"
echo "   • Enterprise approval workflows"
echo "   • Batch processing capabilities"
echo "   • Comprehensive analytics and reporting"
echo "   • Project management integration (Jira, Linear, Asana)"
echo "   • Custom patch rules and policies"
EOF

chmod +x $BUILD_DIR/deploy-production.sh

# Docker deployment script
cat > $BUILD_DIR/deploy-docker.sh << 'EOF'
#!/bin/bash

# AI Dependency Manager - Docker Deployment Script

set -e

echo "🐳 Deploying AI Dependency Manager with Docker"
echo "============================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install docker-compose and try again."
    exit 1
fi

# Deploy with docker-compose
echo "🚀 Starting AI Dependency Manager with docker-compose..."
docker-compose up -d

echo "✅ AI Dependency Manager deployed successfully with Docker!"
echo ""
echo "📊 Container status:"
docker-compose ps

echo ""
echo "🌐 Web interface: http://localhost:8080"
echo "📝 Logs: docker-compose logs -f"
echo "🛑 Stop: docker-compose down"
echo ""
echo "🚀 GitHub Integration Bot Features Available:"
echo "   • Automated dependency patching"
echo "   • Enterprise approval workflows"
echo "   • Batch processing capabilities"
echo "   • Comprehensive analytics and reporting"
echo "   • Project management integration"
echo "   • Custom patch rules and policies"
EOF

chmod +x $BUILD_DIR/deploy-docker.sh

# Create comprehensive README for production build
cat > $BUILD_DIR/PRODUCTION-README.md << 'EOF'
# AI Dependency Manager - Production Build

This is a production-ready build of the AI Dependency Manager with comprehensive GitHub Integration Bot features.

## 🚀 Features Included

### Core Features
- ✅ Multi-package manager support (npm, pip, Maven, Gradle)
- ✅ AI-powered dependency analysis (OpenAI, Claude, Ollama)
- ✅ Real-time dependency scanning and monitoring
- ✅ Interactive update management with risk assessment
- ✅ Background agent for continuous monitoring
- ✅ Comprehensive notification system

### GitHub Integration Bot (Phases 1-5)
- ✅ **Phase 1**: GitHub API integration with authentication and webhooks
- ✅ **Phase 2**: AI-powered patch generation and code analysis
- ✅ **Phase 3**: Smart patch application with conflict resolution
- ✅ **Phase 4**: Comprehensive PR management with automated testing
- ✅ **Phase 5**: Enterprise features with governance and analytics

### Enterprise Features
- 🏢 **Approval Workflows**: Configurable approval processes with escalation
- 📊 **Batch Processing**: Intelligent processing of multiple dependency updates
- 📈 **Analytics & Reporting**: Comprehensive patch success tracking and insights
- 🔗 **Project Management Integration**: Jira, Linear, and Asana connectivity
- 📋 **Custom Policies**: Organization-wide governance and compliance rules
- 🎨 **Modern UI**: Geist Mono font for improved developer experience

## 📦 Deployment Options

### Option 1: Native Deployment
```bash
./deploy-production.sh
```

### Option 2: Docker Deployment
```bash
./deploy-docker.sh
```

## 🔧 Configuration

1. Copy the example configuration:
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

2. Edit the configuration file with your settings:
   - Database configuration
   - GitHub integration settings
   - AI provider API keys
   - Project management tool credentials

## 🌐 Web Interface

After deployment, access the web interface at:
- Native: http://localhost:8081
- Docker: http://localhost:8080

## 📚 Documentation

- `docs/user-guide.md` - Complete user guide
- `docs/deployment.md` - Detailed deployment instructions
- `docs/api-reference.md` - CLI command reference
- `docs/configuration.md` - Configuration options
- `docs/security.md` - Security guidelines

## 🔒 Security Features

- Non-root execution
- Systemd security hardening
- Resource limits and capability restrictions
- Secure credential management
- Audit logging and compliance tracking

## 📊 Monitoring & Logging

- Structured JSON logging
- Health check endpoints
- Performance metrics collection
- Real-time status monitoring
- Comprehensive error tracking

## 🚀 GitHub Integration Bot Commands

```bash
# Setup GitHub integration
ai-dep-manager github setup --token YOUR_TOKEN --repositories owner/repo

# Create batch update job
ai-dep-manager github batch create

# View analytics report
ai-dep-manager github analytics report owner/repo

# Manage organization policies
ai-dep-manager github policy list

# Check approval workflows
ai-dep-manager github approval status workflow-id
```

## 🆘 Support

For support and documentation, visit:
- GitHub: https://github.com/8tcapital/ai-dep-manager
- Documentation: ./docs/
- Issues: https://github.com/8tcapital/ai-dep-manager/issues

## 📄 License

MIT License - see LICENSE file for details.
EOF

# Create build manifest
cat > $BUILD_DIR/BUILD-MANIFEST.json << EOF
{
  "version": "$VERSION",
  "buildTime": "$BUILD_TIME",
  "gitCommit": "$GIT_COMMIT",
  "features": {
    "coreFeatures": [
      "Multi-package manager support",
      "AI-powered dependency analysis",
      "Real-time scanning and monitoring",
      "Interactive update management",
      "Background agent",
      "Notification system"
    ],
    "githubIntegrationBot": {
      "phase1": "GitHub API integration with authentication and webhooks",
      "phase2": "AI-powered patch generation and code analysis",
      "phase3": "Smart patch application with conflict resolution",
      "phase4": "Comprehensive PR management with automated testing",
      "phase5": "Enterprise features with governance and analytics"
    },
    "enterpriseFeatures": [
      "Approval workflows with escalation",
      "Batch processing capabilities",
      "Analytics and reporting platform",
      "Project management integration",
      "Custom policies and governance",
      "Modern UI with Geist Mono font"
    ]
  },
  "platforms": [
    "linux-amd64",
    "linux-arm64",
    "darwin-amd64",
    "darwin-arm64",
    "windows-amd64"
  ],
  "deployment": {
    "docker": "docker-compose up -d",
    "native": "./deploy-production.sh"
  }
}
EOF

# Calculate build sizes
echo -e "${YELLOW}📊 Build Statistics:${NC}"
echo -e "Frontend assets: ${GREEN}$(du -sh $BUILD_DIR/web | cut -f1)${NC}"
echo -e "Documentation: ${GREEN}$(du -sh $BUILD_DIR/docs | cut -f1)${NC}"
echo -e "Binaries:"
for binary in $BUILD_DIR/bin/*; do
    if [[ -f "$binary" ]]; then
        size=$(du -sh "$binary" | cut -f1)
        name=$(basename "$binary")
        echo -e "  ${GREEN}→${NC} $name: ${GREEN}$size${NC}"
    fi
done

# Create checksums
echo -e "${YELLOW}🔐 Generating checksums...${NC}"
cd $BUILD_DIR
find bin -type f -exec sha256sum {} \; > CHECKSUMS.txt
cd ..

echo ""
echo -e "${GREEN}✅ Production build completed successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo -e "Build directory: ${BLUE}$BUILD_DIR${NC}"
echo -e "Version: ${BLUE}$VERSION${NC}"
echo -e "Docker image: ${BLUE}ai-dep-manager:$VERSION${NC}"
echo ""
echo -e "${YELLOW}🚀 Ready for deployment:${NC}"
echo -e "  Native: ${BLUE}cd $BUILD_DIR && ./deploy-production.sh${NC}"
echo -e "  Docker: ${BLUE}cd $BUILD_DIR && ./deploy-docker.sh${NC}"
echo ""
echo -e "${GREEN}🎉 GitHub Integration Bot with all Phase 1-5 features is ready for production!${NC}"
