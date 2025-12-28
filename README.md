# Superintelligence Dependency Manager (AutoUpdateAgent)

[![CI](https://github.com/yeabat8tcap/ai-dependency-manager/workflows/CI/badge.svg)](https://github.com/yeabat8tcap/ai-dependency-manager/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yeabat8tcap/ai-dependency-manager)](https://goreportcard.com/report/github.com/yeabat8tcap/ai-dependency-manager)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://codecov.io/gh/yeabat8tcap/ai-dependency-manager/branch/main/graph/badge.svg)](https://codecov.io/gh/yeabat8tcap/ai-dependency-manager)
[![GitHub Release](https://img.shields.io/github/release/yeabat8tcap/ai-dependency-manager.svg)](https://github.com/yeabat8tcap/ai-dependency-manager/releases)
[![Docker](https://img.shields.io/badge/docker-available-blue.svg)](https://github.com/yeabat8tcap/ai-dependency-manager)

An autonomous Superintelligence-powered CLI agent that intelligently manages software dependencies across multiple package managers with advanced security, risk assessment, and automated update capabilities.

## 🚀 Features

### Core Capabilities
- 🔍 **Multi-Package Manager Support**: Native integration with npm, pip, Maven, and Gradle
- 🤖 **AI-Powered Analysis**: Intelligent changelog analysis and breaking change detection
- 🛡️ **Advanced Security**: Package integrity verification, vulnerability scanning, and malicious package detection
- 🔄 **Background Agent**: Autonomous monitoring with configurable scheduling and daemon support
- ⚡ **Smart Updates**: Risk-based update grouping with rollback capabilities
- 📊 **Comprehensive Analytics**: Detailed reporting, dependency lag analysis, and audit trails

### Advanced Features
- 🎯 **Custom Policies**: Flexible update policies with condition-based rules
- 📧 **Multi-Channel Notifications**: Email, Slack, and webhook integrations
- 🐳 **Container Ready**: Docker support with multi-stage builds
- 🔐 **Secure Credential Management**: Encrypted storage for registry authentication
- 📈 **Performance Optimized**: Concurrent processing with configurable limits
- 🧪 **Comprehensive Testing**: Unit, integration, and end-to-end test suites

## 📦 Installation

### Prerequisites
- Go 1.21 or later
- Node.js and npm (for npm projects)
- Python 3.x and pip (for Python projects)
- Java and Maven/Gradle (for Java projects)

### Quick Install

```bash
# Using Go install
go install github.com/8tcapital/superint-dep-manager@latest

# Or build from source
git clone https://github.com/8tcapital/superint-dep-manager.git
cd superint-dep-manager
make build
sudo make install
```

### Docker Installation

```bash
# Pull the image
docker pull superint-dep-manager:latest

# Run with Docker Compose
docker-compose up -d
```

## 🎯 Quick Start

### 1. Initial Configuration

```bash
# Initialize the system
superint-dep-manager configure

# Add a project
superint-dep-manager configure add-project /path/to/your/project

# Set up security preferences
superint-dep-manager security configure
```

### 2. Basic Operations

```bash
# Check system status
superint-dep-manager status

# Scan for updates
superint-dep-manager scan --all

# Preview available updates
superint-dep-manager update --preview

# Apply safe updates
superint-dep-manager update --strategy conservative

# Check security vulnerabilities
superint-dep-manager security scan --all
```

### 3. Background Agent

```bash
# Start the background agent
superint-dep-manager agent start

# Check agent status
superint-dep-manager agent status

# Configure scheduling
superint-dep-manager configure set agent.schedule "0 2 * * *"  # Daily at 2 AM
```

## 📚 Documentation

- [User Guide](docs/user-guide.md) - Comprehensive usage documentation
- [API Reference](docs/api-reference.md) - Complete API documentation
- [Configuration Guide](docs/configuration.md) - Detailed configuration options
- [Security Guide](docs/security.md) - Security features and best practices
- [Developer Guide](docs/developer-guide.md) - Contributing and development setup
- [Deployment Guide](docs/deployment.md) - Production deployment instructions

## 🔧 Configuration

Copy the example configuration file and customize it:

```bash
mkdir -p ~/.superint-dep-manager
cp config.yaml.example ~/.superint-dep-manager/config.yaml
```

Edit `~/.superint-dep-manager/config.yaml` to configure:
- Log levels and formats
- Database settings
- Background agent behavior
- Security preferences
- Project-specific settings

## 🏗️ Architecture

The Superintelligence Dependency Manager is built with a modular, production-ready architecture:

```
├── cmd/                    # CLI commands and main entry point
├── internal/
│   ├── ai/                # AI analysis and heuristic providers
│   ├── agent/             # Background agent and scheduling
│   ├── config/            # Configuration management
│   ├── database/          # Database layer and migrations
│   ├── daemon/            # System daemon and service management
│   ├── models/            # Data models and database schemas
│   ├── notifications/     # Multi-channel notification system
│   ├── packagemanager/    # Package manager integrations
│   ├── reporting/         # Analytics and reporting engine
│   ├── security/          # Security scanning and verification
│   ├── services/          # Core business logic services
│   └── testing/           # Testing infrastructure and helpers
├── test/
│   ├── integration/       # Integration test suites
│   └── e2e/              # End-to-end testing scenarios
├── scripts/               # Build, deployment, and utility scripts
├── docs/                  # Comprehensive documentation
└── .github/workflows/     # CI/CD pipeline configuration
```

## 🛠️ Development

### Prerequisites
- Go 1.21+
- Docker and Docker Compose
- Make
- golangci-lint (for linting)

### Setup
```bash
# Clone and setup
git clone https://github.com/8tcapital/superint-dep-manager.git
cd superint-dep-manager

# Install dependencies
make deps

# Run tests
make test-all

# Build
make build
```

### Testing
```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Coverage report
make test-coverage

# All tests with benchmarks
make test-all
```

## 🚀 Production Deployment

### System Service
```bash
# Install as system service
sudo ./scripts/deploy.sh install

# Start the service
sudo systemctl start superint-dep-manager
sudo systemctl enable superint-dep-manager
```

### Docker Deployment
```bash
# Using Docker Compose
docker-compose up -d

# Or standalone Docker
docker run -d \
  -v /path/to/config:/app/config \
  -v /path/to/data:/app/data \
  superint-dep-manager:latest
```

## 📊 Monitoring and Observability

- **Health Checks**: Built-in health endpoints for monitoring
- **Metrics**: Prometheus-compatible metrics export
- **Logging**: Structured logging with configurable levels
- **Audit Trail**: Complete audit log of all operations
- **Reporting**: Comprehensive analytics and dependency insights

## 🔒 Security Features

- **Package Integrity**: SHA-256 checksum verification
- **Vulnerability Scanning**: Integration with security databases
- **Malicious Package Detection**: Pattern-based threat detection
- **Secure Credentials**: AES-GCM encrypted credential storage
- **Audit Logging**: Complete security event tracking
- **Access Control**: Role-based permissions (enterprise)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go and modern DevOps practices
- Inspired by the need for intelligent dependency management
- Thanks to all contributors and the open-source community

## 📞 Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/8tcapital/superint-dep-manager/issues)
- 💬 [Discussions](https://github.com/8tcapital/superint-dep-manager/discussions)
- 📧 Email: support@8tcapital.com

---

**Made with ❤️ by the 8tcapital team**

### Prerequisites

- Go 1.21 or later
- SQLite3

### Building from Source

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Run locally
go run main.go status
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Roadmap

- [x] **Phase 1**: Foundation & Architecture
- [ ] **Phase 2**: Package Manager Integration
- [ ] **Phase 3**: Dependency Scanning & Discovery
- [ ] **Phase 4**: Basic CLI Interface
- [ ] **Phase 5**: AI Integration Framework
- [ ] **Phase 6**: Update Management System
- [ ] **Phase 7**: Background Agent
- [ ] **Phase 8**: Security & Safety Features
- [ ] **Phase 9**: Advanced Features
- [ ] **Phase 10**: Testing & Quality Assurance
- [ ] **Phase 11**: Documentation & Deployment
- [ ] **Phase 12**: Future Enhancements

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions, issues, or contributions, please visit our [GitHub repository](https://github.com/8tcapital/superint-dep-manager).
