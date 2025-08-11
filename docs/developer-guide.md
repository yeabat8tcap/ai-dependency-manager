# AI Dependency Manager - Developer Guide

This guide provides information for developers who want to contribute to or extend the AI Dependency Manager.

## Table of Contents

1. [Development Setup](#development-setup)
2. [Project Structure](#project-structure)
3. [Architecture Overview](#architecture-overview)
4. [Contributing Guidelines](#contributing-guidelines)
5. [Testing](#testing)
6. [Code Style](#code-style)
7. [Adding New Features](#adding-new-features)
8. [Package Manager Integration](#package-manager-integration)
9. [AI Provider Integration](#ai-provider-integration)
10. [Release Process](#release-process)

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Make
- Docker and Docker Compose (for integration tests)
- Node.js and npm (for npm integration testing)
- Python 3.x and pip (for Python integration testing)
- Java and Maven/Gradle (for Java integration testing)

### Getting Started

```bash
# Clone the repository
git clone https://github.com/8tcapital/ai-dep-manager.git
cd ai-dep-manager

# Install dependencies
make deps

# Run tests
make test-all

# Build the application
make build

# Run locally
./bin/ai-dep-manager version
```

### Development Tools

Install recommended development tools:

```bash
# Install golangci-lint for linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt for formatting
go install mvdan.cc/gofumpt@latest

# Install govulncheck for security scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Project Structure

```
ai-dep-manager/
├── cmd/                    # CLI commands and main entry point
│   ├── ai-dep-manager/    # Main application
│   ├── agent.go           # Agent management commands
│   ├── configure.go       # Configuration commands
│   ├── lag.go            # Dependency lag commands
│   ├── notify.go         # Notification commands
│   ├── policy.go         # Policy management commands
│   ├── report.go         # Reporting commands
│   ├── rollback.go       # Rollback commands
│   ├── scan.go           # Scanning commands
│   ├── security.go       # Security commands
│   ├── status.go         # Status commands
│   ├── update.go         # Update commands
│   └── version.go        # Version command
├── internal/              # Private application code
│   ├── ai/               # AI analysis and providers
│   │   ├── heuristic/    # Heuristic-based AI provider
│   │   ├── interface.go  # AI provider interfaces
│   │   └── manager.go    # AI provider manager
│   ├── agent/            # Background agent
│   ├── config/           # Configuration management
│   ├── database/         # Database layer and migrations
│   ├── daemon/           # System daemon management
│   ├── models/           # Data models and schemas
│   ├── notifications/    # Notification system
│   ├── packagemanager/   # Package manager integrations
│   ├── reporting/        # Analytics and reporting
│   ├── security/         # Security features
│   ├── services/         # Core business logic
│   └── testing/          # Testing infrastructure
├── test/                 # Test suites
│   ├── integration/      # Integration tests
│   └── e2e/             # End-to-end tests
├── docs/                # Documentation
├── scripts/             # Build and deployment scripts
├── .github/workflows/   # CI/CD pipeline
├── Dockerfile          # Docker configuration
├── docker-compose.yml  # Docker Compose setup
├── Makefile           # Build automation
└── go.mod             # Go module definition
```

## Architecture Overview

### Core Components

1. **CLI Layer** (`cmd/`): Command-line interface and user interaction
2. **Service Layer** (`internal/services/`): Core business logic
3. **Data Layer** (`internal/models/`, `internal/database/`): Data persistence
4. **Integration Layer** (`internal/packagemanager/`): External system integrations
5. **AI Layer** (`internal/ai/`): AI analysis and decision making

### Key Design Patterns

- **Repository Pattern**: Data access abstraction
- **Strategy Pattern**: Package manager and AI provider implementations
- **Observer Pattern**: Event-driven notifications
- **Factory Pattern**: Service and provider instantiation
- **Command Pattern**: CLI command structure

### Data Flow

```
CLI Command → Service Layer → Package Manager → Registry API
     ↓              ↓              ↓
Database ← AI Analysis ← Changelog Data
     ↓
Notifications
```

## Contributing Guidelines

### Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

### Pull Request Process

1. **Fork the repository** and create a feature branch
2. **Make your changes** following the coding standards
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Run the full test suite** and ensure all tests pass
6. **Submit a pull request** with a clear description

### Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions or changes
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat(scanner): add support for Gradle projects
fix(security): resolve vulnerability in dependency parsing
docs(api): update configuration reference
```

## Testing

### Test Structure

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows

### Running Tests

```bash
# Run all tests
make test-all

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run end-to-end tests
make test-e2e

# Run tests with coverage
make test-coverage

# Run race detection tests
make test-race

# Run benchmarks
make test-bench
```

### Writing Tests

#### Unit Test Example

```go
func TestScannerService_ScanProject(t *testing.T) {
    ctx := testingPkg.SetupTestEnvironment(t)
    defer ctx.Cleanup()

    scanner := services.NewScannerService()

    result, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
    testingPkg.AssertNoError(t, err, "Scan should succeed")
    testingPkg.AssertEqual(t, "completed", result.Status, "Scan should complete")
}
```

#### Integration Test Example

```go
func TestScannerIntegration_FullWorkflow(t *testing.T) {
    testingPkg.SkipIfShort(t, "integration test requires full system")

    ctx := testingPkg.SetupTestEnvironment(t)
    defer ctx.Cleanup()

    // Test full workflow
    scanner := services.NewScannerService()
    updateService := services.NewUpdateService()

    // Scan project
    scanResult, err := scanner.ScanProject(context.Background(), ctx.Projects[0].ID)
    testingPkg.AssertNoError(t, err, "Scan should succeed")

    // Create update plan
    plan, err := updateService.CreateUpdatePlan(context.Background(), ctx.Projects[0].ID, services.UpdatePlanOptions{})
    testingPkg.AssertNoError(t, err, "Update plan creation should succeed")
}
```

### Test Helpers

Use the testing infrastructure in `internal/testing/setup.go`:

```go
// Setup test environment
ctx := testingPkg.SetupTestEnvironment(t)
defer ctx.Cleanup()

// Use helper assertions
testingPkg.AssertNoError(t, err, "Operation should succeed")
testingPkg.AssertEqual(t, expected, actual, "Values should match")
testingPkg.AssertTrue(t, condition, "Condition should be true")
```

## Code Style

### Go Style Guidelines

Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

### Formatting

```bash
# Format code
make fmt

# Run linter
make lint

# Fix linting issues
golangci-lint run --fix
```

### Naming Conventions

- **Packages**: Short, lowercase, single word
- **Functions**: CamelCase, descriptive
- **Variables**: CamelCase, avoid abbreviations
- **Constants**: CamelCase or UPPER_CASE for exported constants
- **Interfaces**: End with "er" when possible (e.g., `Scanner`, `Updater`)

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to scan project %d: %w", projectID, err)
}

// Good: Use custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### Logging

```go
// Use structured logging
logger.Info("Starting project scan",
    "project_id", projectID,
    "project_path", project.Path,
)

// Log errors with context
logger.Error("Failed to parse package file",
    "file", packageFile,
    "error", err,
)
```

## Adding New Features

### Feature Development Process

1. **Create an issue** describing the feature
2. **Design the feature** and get feedback
3. **Implement the feature** with tests
4. **Update documentation**
5. **Submit a pull request**

### Adding a New CLI Command

1. Create command file in `cmd/`:

```go
// cmd/mycommand.go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCommandCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description of my command",
    Long:  "Detailed description of my command",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCommandCmd)
    
    // Add flags
    myCommandCmd.Flags().StringP("option", "o", "", "Option description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    // Implementation
    return nil
}
```

2. Add service logic in `internal/services/`
3. Add tests
4. Update documentation

### Adding a New Service

1. Define interface in `internal/services/`:

```go
type MyService interface {
    DoSomething(ctx context.Context, input Input) (*Output, error)
}
```

2. Implement service:

```go
type myService struct {
    db     database.DB
    logger *slog.Logger
}

func NewMyService() MyService {
    return &myService{
        db:     database.GetDB(),
        logger: slog.Default(),
    }
}

func (s *myService) DoSomething(ctx context.Context, input Input) (*Output, error) {
    // Implementation
    return &Output{}, nil
}
```

3. Add tests and documentation

## Package Manager Integration

### Adding a New Package Manager

1. Implement the `PackageManager` interface:

```go
// internal/packagemanager/newpm.go
type NewPMManager struct {
    config *config.Config
    logger *slog.Logger
}

func NewNewPMManager(cfg *config.Config) PackageManager {
    return &NewPMManager{
        config: cfg,
        logger: slog.Default(),
    }
}

func (n *NewPMManager) DetectProjects(rootPath string) ([]*models.Project, error) {
    // Implementation
}

func (n *NewPMManager) ParseDependencies(project *models.Project) ([]*models.Dependency, error) {
    // Implementation
}

// ... implement other interface methods
```

2. Register the package manager:

```go
// internal/packagemanager/registry.go
func init() {
    RegisterPackageManager("newpm", func(cfg *config.Config) PackageManager {
        return NewNewPMManager(cfg)
    })
}
```

3. Add configuration support:

```yaml
# config.yaml
packagemanagers:
  newpm:
    enabled: true
    path: "/usr/local/bin/newpm"
    timeout: "60s"
```

4. Add tests and documentation

## AI Provider Integration

### Adding a New AI Provider

1. Implement the `AIProvider` interface:

```go
// internal/ai/newprovider/provider.go
type NewProvider struct {
    config *config.Config
    client HTTPClient
}

func NewNewProvider(cfg *config.Config) ai.AIProvider {
    return &NewProvider{
        config: cfg,
        client: &http.Client{},
    }
}

func (p *NewProvider) AnalyzeChangelog(ctx context.Context, req *ai.ChangelogAnalysisRequest) (*ai.ChangelogAnalysis, error) {
    // Implementation
}

// ... implement other interface methods
```

2. Register the provider:

```go
// internal/ai/manager.go
func init() {
    RegisterProvider("newprovider", func(cfg *config.Config) AIProvider {
        return newprovider.NewNewProvider(cfg)
    })
}
```

3. Add configuration:

```yaml
# config.yaml
ai:
  provider: "newprovider"
  newprovider:
    api_key: "${NEW_PROVIDER_API_KEY}"
    endpoint: "https://api.newprovider.com"
```

## Release Process

### Version Management

We use semantic versioning (SemVer):
- `MAJOR.MINOR.PATCH`
- `MAJOR`: Breaking changes
- `MINOR`: New features (backward compatible)
- `PATCH`: Bug fixes (backward compatible)

### Release Steps

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with release notes
3. **Create and push tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. **GitHub Actions** will automatically build and create release
5. **Update documentation** if needed

### Pre-release Checklist

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] CHANGELOG.md is updated
- [ ] Version numbers are bumped
- [ ] Breaking changes are documented
- [ ] Migration guides are provided (if needed)

## Development Workflow

### Git Workflow

1. Create feature branch from `main`
2. Make changes and commit
3. Push branch and create pull request
4. Code review and approval
5. Merge to `main`
6. Delete feature branch

### Branch Naming

- `feature/description`: New features
- `fix/description`: Bug fixes
- `docs/description`: Documentation updates
- `refactor/description`: Code refactoring

### Development Environment

```bash
# Set up development environment
export AI_DEP_MANAGER_LOG_LEVEL=debug
export AI_DEP_MANAGER_DATA_DIR=./dev-data

# Run in development mode
go run . status
```

## Debugging

### Debug Logging

```bash
# Enable debug logging
export AI_DEP_MANAGER_LOG_LEVEL=debug

# Run with verbose output
ai-dep-manager scan --verbose
```

### Using Debugger

```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug application
dlv debug . -- scan --project-id 1
```

### Profiling

```go
// Add profiling to main.go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Then access profiling at `http://localhost:6060/debug/pprof/`

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [GORM ORM](https://gorm.io/)
- [Logrus Logging](https://github.com/sirupsen/logrus)
- [Testify Testing](https://github.com/stretchr/testify)

---

For questions or help, please:
- Open an issue on GitHub
- Join our Discord community
- Email: dev@8tcapital.com
