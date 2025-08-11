package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/logger"
)

// ValidationService handles patch validation and testing
type ValidationService struct {
	client       *Client
	repositories *RepositoriesService
}

// NewValidationService creates a new validation service
func NewValidationService(client *Client) *ValidationService {
	return &ValidationService{
		client:       client,
		repositories: client.Repositories,
	}
}

// ValidationResult represents the result of patch validation
type ValidationResult struct {
	PatchID         string                `json:"patch_id"`
	Repository      string                `json:"repository"`
	OverallStatus   ValidationStatus      `json:"overall_status"`
	Steps           []*ValidationStepResult `json:"steps"`
	BuildOutput     string                `json:"build_output,omitempty"`
	TestOutput      string                `json:"test_output,omitempty"`
	LintOutput      string                `json:"lint_output,omitempty"`
	Errors          []string              `json:"errors,omitempty"`
	Warnings        []string              `json:"warnings,omitempty"`
	Duration        time.Duration         `json:"duration"`
	ValidatedAt     time.Time             `json:"validated_at"`
	Environment     *ValidationEnvironment `json:"environment"`
}

// ValidationStatus represents the status of validation
type ValidationStatus string

const (
	ValidationStatusPending ValidationStatus = "pending"
	ValidationStatusRunning ValidationStatus = "running"
	ValidationStatusPassed  ValidationStatus = "passed"
	ValidationStatusFailed  ValidationStatus = "failed"
	ValidationStatusSkipped ValidationStatus = "skipped"
)

// ValidationStepResult represents the result of a validation step
type ValidationStepResult struct {
	Step        *ValidationStep  `json:"step"`
	Status      ValidationStatus `json:"status"`
	Output      string           `json:"output,omitempty"`
	Error       string           `json:"error,omitempty"`
	Duration    time.Duration    `json:"duration"`
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at"`
}

// ValidationEnvironment represents the validation environment
type ValidationEnvironment struct {
	OS              string            `json:"os"`
	Architecture    string            `json:"architecture"`
	NodeVersion     string            `json:"node_version,omitempty"`
	PythonVersion   string            `json:"python_version,omitempty"`
	JavaVersion     string            `json:"java_version,omitempty"`
	GoVersion       string            `json:"go_version,omitempty"`
	WorkingDir      string            `json:"working_dir"`
	Environment     map[string]string `json:"environment"`
	PackageManager  string            `json:"package_manager"`
}

// ValidationOptions represents options for validation
type ValidationOptions struct {
	SkipBuild       bool              `json:"skip_build"`
	SkipTests       bool              `json:"skip_tests"`
	SkipLint        bool              `json:"skip_lint"`
	Timeout         time.Duration     `json:"timeout"`
	Environment     map[string]string `json:"environment,omitempty"`
	WorkingDir      string            `json:"working_dir,omitempty"`
	Parallel        bool              `json:"parallel"`
	FailFast        bool              `json:"fail_fast"`
}

// ValidatePatch validates a generated patch
func (v *ValidationService) ValidatePatch(ctx context.Context, patch *GeneratedPatch, options *ValidationOptions) (*ValidationResult, error) {
	logger.Info("Validating patch for repository %s", patch.Repository)
	
	startTime := time.Now()
	
	// Set default options
	if options == nil {
		options = &ValidationOptions{
			Timeout:  30 * time.Minute,
			FailFast: true,
		}
	}
	
	result := &ValidationResult{
		PatchID:       generatePatchID(patch),
		Repository:    patch.Repository,
		OverallStatus: ValidationStatusRunning,
		Steps:         []*ValidationStepResult{},
		Errors:        []string{},
		Warnings:      []string{},
		ValidatedAt:   startTime,
	}
	
	// Setup validation environment
	env, err := v.setupValidationEnvironment(ctx, patch, options)
	if err != nil {
		result.OverallStatus = ValidationStatusFailed
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to setup environment: %v", err))
		return result, err
	}
	result.Environment = env
	
	// Apply patch to temporary environment
	if err := v.applyPatchToEnvironment(ctx, patch, env); err != nil {
		result.OverallStatus = ValidationStatusFailed
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to apply patch: %v", err))
		return result, err
	}
	
	// Run validation steps
	for _, step := range patch.ValidationSteps {
		if ctx.Err() != nil {
			result.OverallStatus = ValidationStatusFailed
			result.Errors = append(result.Errors, "Validation cancelled")
			break
		}
		
		stepResult := v.runValidationStep(ctx, step, env, options)
		result.Steps = append(result.Steps, stepResult)
		
		// Handle step result
		switch stepResult.Status {
		case ValidationStatusFailed:
			if step.Required {
				result.OverallStatus = ValidationStatusFailed
				if options.FailFast {
					logger.Error("Required validation step failed: %s", step.Description)
					break
				}
			} else {
				result.Warnings = append(result.Warnings, 
					fmt.Sprintf("Optional step failed: %s", step.Description))
			}
		case ValidationStatusPassed:
			logger.Debug("Validation step passed: %s", step.Description)
		}
	}
	
	// Set overall status if not already failed
	if result.OverallStatus == ValidationStatusRunning {
		allPassed := true
		for _, stepResult := range result.Steps {
			if stepResult.Step.Required && stepResult.Status != ValidationStatusPassed {
				allPassed = false
				break
			}
		}
		
		if allPassed {
			result.OverallStatus = ValidationStatusPassed
		} else {
			result.OverallStatus = ValidationStatusFailed
		}
	}
	
	result.Duration = time.Since(startTime)
	
	// Cleanup environment
	if err := v.cleanupValidationEnvironment(env); err != nil {
		logger.Warn("Failed to cleanup validation environment: %v", err)
	}
	
	logger.Info("Patch validation completed for %s: %s (duration: %v)", 
		patch.Repository, result.OverallStatus, result.Duration)
	
	return result, nil
}

// setupValidationEnvironment sets up a temporary environment for validation
func (v *ValidationService) setupValidationEnvironment(ctx context.Context, patch *GeneratedPatch, options *ValidationOptions) (*ValidationEnvironment, error) {
	// Parse repository name
	parts := strings.Split(patch.Repository, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", patch.Repository)
	}
	owner, repo := parts[0], parts[1]
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("patch-validation-%s-%s-*", owner, repo))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	logger.Debug("Created validation environment: %s", tempDir)
	
	// Clone repository (in a real implementation, this would clone the actual repo)
	// For now, we'll simulate this by creating the directory structure
	if err := os.MkdirAll(filepath.Join(tempDir, "src"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create src directory: %w", err)
	}
	
	env := &ValidationEnvironment{
		OS:           "linux", // This would be detected
		Architecture: "amd64", // This would be detected
		WorkingDir:   tempDir,
		Environment:  make(map[string]string),
	}
	
	// Detect runtime versions
	if nodeVersion, err := v.getNodeVersion(); err == nil {
		env.NodeVersion = nodeVersion
	}
	if pythonVersion, err := v.getPythonVersion(); err == nil {
		env.PythonVersion = pythonVersion
	}
	if javaVersion, err := v.getJavaVersion(); err == nil {
		env.JavaVersion = javaVersion
	}
	if goVersion, err := v.getGoVersion(); err == nil {
		env.GoVersion = goVersion
	}
	
	// Set environment variables
	if options.Environment != nil {
		for key, value := range options.Environment {
			env.Environment[key] = value
		}
	}
	
	return env, nil
}

// applyPatchToEnvironment applies the patch to the validation environment
func (v *ValidationService) applyPatchToEnvironment(ctx context.Context, patch *GeneratedPatch, env *ValidationEnvironment) error {
	logger.Debug("Applying patch to validation environment")
	
	// Apply configuration changes
	for _, configPatch := range patch.ConfigChanges {
		if err := v.applyConfigPatch(configPatch, env); err != nil {
			return fmt.Errorf("failed to apply config patch %s: %w", configPatch.File, err)
		}
	}
	
	// Apply file changes
	for _, filePatch := range patch.Files {
		if err := v.applyFilePatch(filePatch, env); err != nil {
			return fmt.Errorf("failed to apply file patch %s: %w", filePatch.Path, err)
		}
	}
	
	return nil
}

// applyConfigPatch applies a configuration patch
func (v *ValidationService) applyConfigPatch(patch *ConfigPatch, env *ValidationEnvironment) error {
	filePath := filepath.Join(env.WorkingDir, patch.File)
	
	switch patch.Type {
	case "package.json":
		// Create or update package.json
		content := `{
  "name": "test-project",
  "version": "1.0.0",
  "dependencies": {`
		
		if deps, ok := patch.Changes["dependencies"].(map[string]string); ok {
			var depEntries []string
			for name, version := range deps {
				depEntries = append(depEntries, fmt.Sprintf(`    "%s": "%s"`, name, version))
			}
			content += "\n" + strings.Join(depEntries, ",\n") + "\n"
		}
		
		content += `  }
}`
		
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write package.json: %w", err)
		}
		
	case "requirements.txt":
		if reqs, ok := patch.Changes["requirements"].([]string); ok {
			content := strings.Join(reqs, "\n")
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write requirements.txt: %w", err)
			}
		}
		
	default:
		logger.Warn("Unsupported config patch type: %s", patch.Type)
	}
	
	return nil
}

// applyFilePatch applies a file patch
func (v *ValidationService) applyFilePatch(patch *FilePatch, env *ValidationEnvironment) error {
	filePath := filepath.Join(env.WorkingDir, patch.Path)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	switch patch.Type {
	case "create":
		if err := os.WriteFile(filePath, []byte(patch.NewContent), 0644); err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		
	case "modify":
		// For simplicity, we'll create a new file with the new content
		// In a real implementation, this would apply the specific changes
		if patch.NewContent != "" {
			if err := os.WriteFile(filePath, []byte(patch.NewContent), 0644); err != nil {
				return fmt.Errorf("failed to modify file: %w", err)
			}
		} else {
			// Create a simple file for testing
			content := fmt.Sprintf("// File modified by patch\n// %s\n", patch.Description)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to modify file: %w", err)
			}
		}
		
	case "delete":
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}
	
	return nil
}

// runValidationStep runs a single validation step
func (v *ValidationService) runValidationStep(ctx context.Context, step *ValidationStep, env *ValidationEnvironment, options *ValidationOptions) *ValidationStepResult {
	result := &ValidationStepResult{
		Step:      step,
		Status:    ValidationStatusRunning,
		StartedAt: time.Now(),
	}
	
	logger.Debug("Running validation step: %s", step.Description)
	
	// Check if step should be skipped
	if v.shouldSkipStep(step, options) {
		result.Status = ValidationStatusSkipped
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result
	}
	
	// Run the step
	switch step.Type {
	case "build":
		result = v.runBuildStep(ctx, step, env, result)
	case "test":
		result = v.runTestStep(ctx, step, env, result)
	case "lint":
		result = v.runLintStep(ctx, step, env, result)
	case "manual":
		result = v.runManualStep(ctx, step, env, result)
	default:
		result.Status = ValidationStatusFailed
		result.Error = fmt.Sprintf("Unknown validation step type: %s", step.Type)
	}
	
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)
	
	return result
}

// runBuildStep runs a build validation step
func (v *ValidationService) runBuildStep(ctx context.Context, step *ValidationStep, env *ValidationEnvironment, result *ValidationStepResult) *ValidationStepResult {
	if step.Command == "" {
		result.Status = ValidationStatusSkipped
		result.Output = "No build command specified"
		return result
	}
	
	output, err := v.executeCommand(ctx, step.Command, env)
	result.Output = output
	
	if err != nil {
		result.Status = ValidationStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = ValidationStatusPassed
	}
	
	return result
}

// runTestStep runs a test validation step
func (v *ValidationService) runTestStep(ctx context.Context, step *ValidationStep, env *ValidationEnvironment, result *ValidationStepResult) *ValidationStepResult {
	if step.Command == "" {
		result.Status = ValidationStatusSkipped
		result.Output = "No test command specified"
		return result
	}
	
	output, err := v.executeCommand(ctx, step.Command, env)
	result.Output = output
	
	if err != nil {
		result.Status = ValidationStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = ValidationStatusPassed
	}
	
	return result
}

// runLintStep runs a lint validation step
func (v *ValidationService) runLintStep(ctx context.Context, step *ValidationStep, env *ValidationEnvironment, result *ValidationStepResult) *ValidationStepResult {
	if step.Command == "" {
		result.Status = ValidationStatusSkipped
		result.Output = "No lint command specified"
		return result
	}
	
	output, err := v.executeCommand(ctx, step.Command, env)
	result.Output = output
	
	if err != nil {
		result.Status = ValidationStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = ValidationStatusPassed
	}
	
	return result
}

// runManualStep runs a manual validation step
func (v *ValidationService) runManualStep(ctx context.Context, step *ValidationStep, env *ValidationEnvironment, result *ValidationStepResult) *ValidationStepResult {
	// Manual steps are always marked as passed for automated validation
	// In a real implementation, this might create a checklist or notification
	result.Status = ValidationStatusPassed
	result.Output = "Manual validation step - requires human review"
	return result
}

// executeCommand executes a command in the validation environment
func (v *ValidationService) executeCommand(ctx context.Context, command string, env *ValidationEnvironment) (string, error) {
	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = env.WorkingDir
	
	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range env.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// Execute command
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// shouldSkipStep determines if a validation step should be skipped
func (v *ValidationService) shouldSkipStep(step *ValidationStep, options *ValidationOptions) bool {
	switch step.Type {
	case "build":
		return options.SkipBuild
	case "test":
		return options.SkipTests
	case "lint":
		return options.SkipLint
	default:
		return false
	}
}

// cleanupValidationEnvironment cleans up the validation environment
func (v *ValidationService) cleanupValidationEnvironment(env *ValidationEnvironment) error {
	if env.WorkingDir != "" {
		return os.RemoveAll(env.WorkingDir)
	}
	return nil
}

// Helper functions for version detection
func (v *ValidationService) getNodeVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (v *ValidationService) getPythonVersion() (string, error) {
	cmd := exec.Command("python", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (v *ValidationService) getJavaVersion() (string, error) {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (v *ValidationService) getGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// generatePatchID generates a unique ID for a patch
func generatePatchID(patch *GeneratedPatch) string {
	return fmt.Sprintf("%s-%d", strings.ReplaceAll(patch.Repository, "/", "-"), patch.GeneratedAt.Unix())
}

// ValidatePatchSafety performs safety checks on a patch before application
func (v *ValidationService) ValidatePatchSafety(ctx context.Context, patch *GeneratedPatch) (*ValidationResult, error) {
	logger.Info("Performing safety validation for patch %s", patch.Repository)
	
	result := &ValidationResult{
		PatchID:       generatePatchID(patch),
		Repository:    patch.Repository,
		OverallStatus: ValidationStatusRunning,
		Steps:         []*ValidationStepResult{},
		Errors:        []string{},
		Warnings:      []string{},
		ValidatedAt:   time.Now(),
	}
	
	// Check for dangerous operations
	for _, filePatch := range patch.Files {
		if v.isDangerousFilePatch(filePatch) {
			result.Errors = append(result.Errors, 
				fmt.Sprintf("Dangerous operation detected in %s: %s", filePatch.Path, filePatch.Description))
		}
	}
	
	// Check for suspicious changes
	for _, configPatch := range patch.ConfigChanges {
		if v.isSuspiciousConfigPatch(configPatch) {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("Suspicious config change in %s: %s", configPatch.File, configPatch.Description))
		}
	}
	
	// Check risk assessment
	if patch.RiskAssessment.OverallRisk >= models.RiskHigh {
		result.Warnings = append(result.Warnings, 
			"High risk patch detected - additional review recommended")
	}
	
	// Set overall status
	if len(result.Errors) > 0 {
		result.OverallStatus = ValidationStatusFailed
	} else {
		result.OverallStatus = ValidationStatusPassed
	}
	
	return result, nil
}

// isDangerousFilePatch checks if a file patch contains dangerous operations
func (v *ValidationService) isDangerousFilePatch(patch *FilePatch) bool {
	// Check for dangerous file operations
	if patch.Type == "delete" && strings.Contains(patch.Path, "config") {
		return true
	}
	
	// Check for dangerous content changes
	for _, change := range patch.Changes {
		if strings.Contains(strings.ToLower(change.NewContent), "eval(") ||
		   strings.Contains(strings.ToLower(change.NewContent), "exec(") ||
		   strings.Contains(strings.ToLower(change.NewContent), "system(") {
			return true
		}
	}
	
	return false
}

// isSuspiciousConfigPatch checks if a config patch is suspicious
func (v *ValidationService) isSuspiciousConfigPatch(patch *ConfigPatch) bool {
	// Check for suspicious dependency additions
	if deps, ok := patch.Changes["dependencies"].(map[string]string); ok {
		for name := range deps {
			if strings.Contains(name, "malware") || strings.Contains(name, "backdoor") {
				return true
			}
		}
	}
	
	return false
}
