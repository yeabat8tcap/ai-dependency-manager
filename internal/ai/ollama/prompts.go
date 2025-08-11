package ollama

import (
	"fmt"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// generateChangelogAnalysisPrompt creates a prompt for changelog analysis optimized for local models
func (o *OllamaProvider) generateChangelogAnalysisPrompt(request *types.ChangelogAnalysisRequest) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert software dependency analyst. Analyze the following package changelog and provide a detailed JSON response.\n\n")
	
	prompt.WriteString(fmt.Sprintf("**Package Information:**\n"))
	prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.PackageName))
	prompt.WriteString(fmt.Sprintf("- Version Change: %s → %s\n", request.FromVersion, request.ToVersion))
	prompt.WriteString(fmt.Sprintf("- Package Manager: %s\n", request.PackageManager))
	prompt.WriteString(fmt.Sprintf("- Language: %s\n", request.Language))
	
	prompt.WriteString("\n**Changelog Content:**\n")
	prompt.WriteString(request.ChangelogText)
	
	if request.ReleaseNotes != "" {
		prompt.WriteString("\n\n**Release Notes:**\n")
		prompt.WriteString(request.ReleaseNotes)
	}
	
	prompt.WriteString("\n\n**Analysis Instructions:**\n")
	prompt.WriteString("Analyze this changelog and provide a JSON response with the following structure:\n\n")
	
	prompt.WriteString(`{
  "package_name": "` + request.PackageName + `",
  "from_version": "` + request.FromVersion + `",
  "to_version": "` + request.ToVersion + `",
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-1.0,
  "confidence": 0.0-1.0,
  "summary": "Brief summary of changes",
  "recommendations": ["recommendation1", "recommendation2"],
  "breaking_changes": [
    {
      "type": "api|behavior|dependency",
      "description": "Description of breaking change",
      "impact": "Impact description",
      "confidence": 0.0-1.0,
      "mitigation": "How to mitigate this change"
    }
  ],
  "new_features": [
    {
      "name": "Feature name",
      "description": "Feature description",
      "type": "api|ui|performance|security",
      "impact": "Impact description",
      "confidence": 0.0-1.0
    }
  ],
  "bug_fixes": [
    {
      "description": "Bug fix description",
      "impact": "Impact description",
      "severity": "low|medium|high|critical",
      "confidence": 0.0-1.0
    }
  ],
  "security_fixes": [
    {
      "cve": "CVE identifier if available",
      "severity": "low|medium|high|critical",
      "description": "Security fix description",
      "impact": "Security impact description",
      "confidence": 0.0-1.0
    }
  ],
  "deprecations": [
    {
      "api": "Deprecated API or feature",
      "replacement": "Replacement if available",
      "timeline": "Deprecation timeline"
    }
  ]
}`)
	
	prompt.WriteString("\n\n**Important Guidelines:**\n")
	prompt.WriteString("- Focus on identifying breaking changes, security fixes, and significant new features\n")
	prompt.WriteString("- Provide realistic confidence scores based on changelog clarity\n")
	prompt.WriteString("- Risk level should reflect potential impact on existing applications\n")
	prompt.WriteString("- Be concise but comprehensive in descriptions\n")
	prompt.WriteString("- Only include items that are clearly mentioned in the changelog\n")
	prompt.WriteString("- Respond ONLY with valid JSON, no additional text\n")
	
	return prompt.String()
}

// generateVersionDiffAnalysisPrompt creates a prompt for version diff analysis
func (o *OllamaProvider) generateVersionDiffAnalysisPrompt(request *types.VersionDiffAnalysisRequest) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert software dependency analyst. Analyze the following version differences and provide a detailed JSON response.\n\n")
	
	prompt.WriteString(fmt.Sprintf("**Package Information:**\n"))
	prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.PackageName))
	prompt.WriteString(fmt.Sprintf("- Version Change: %s → %s\n", request.FromVersion, request.ToVersion))
	prompt.WriteString(fmt.Sprintf("- Package Manager: %s\n", request.PackageManager))
	prompt.WriteString(fmt.Sprintf("- Language: %s\n", request.Language))
	
	prompt.WriteString("\n**Version Diff Content:**\n")
	prompt.WriteString(request.DiffText)
	
	if len(request.FileChanges) > 0 {
		prompt.WriteString("\n\n**File Changes:**\n")
		for _, fc := range request.FileChanges {
			prompt.WriteString(fmt.Sprintf("- %s (%s): +%d -%d lines\n", fc.Path, fc.Type, fc.LinesAdded, fc.LinesRemoved))
		}
	}
	
	prompt.WriteString("\n\n**Analysis Instructions:**\n")
	prompt.WriteString("Analyze this version diff and provide a JSON response with the following structure:\n\n")
	
	prompt.WriteString(`{
  "package_name": "` + request.PackageName + `",
  "from_version": "` + request.FromVersion + `",
  "to_version": "` + request.ToVersion + `",
  "update_type": "major|minor|patch",
  "semantic_impact": "Description of semantic versioning impact",
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-1.0,
  "confidence": 0.0-1.0,
  "summary": "Brief summary of version differences",
  "recommendations": ["recommendation1", "recommendation2"],
  "api_changes": [
    {
      "type": "added|modified|removed",
      "api": "API name or signature",
      "description": "Description of the change",
      "impact": "Impact on consumers"
    }
  ],
  "behavior_changes": [
    {
      "component": "Component or module name",
      "description": "Description of behavior change",
      "impact": "Impact description"
    }
  ]
}`)
	
	prompt.WriteString("\n\n**Important Guidelines:**\n")
	prompt.WriteString("- Classify update type based on semantic versioning principles\n")
	prompt.WriteString("- Focus on API changes and behavior modifications\n")
	prompt.WriteString("- Assess risk based on potential breaking changes\n")
	prompt.WriteString("- Provide actionable recommendations for consumers\n")
	prompt.WriteString("- Respond ONLY with valid JSON, no additional text\n")
	
	return prompt.String()
}

// generateCompatibilityPredictionPrompt creates a prompt for compatibility prediction
func (o *OllamaProvider) generateCompatibilityPredictionPrompt(request *types.CompatibilityPredictionRequest) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert software dependency analyst. Predict compatibility for the following package update and provide a detailed JSON response.\n\n")
	
	prompt.WriteString(fmt.Sprintf("**Package Information:**\n"))
	prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.PackageName))
	prompt.WriteString(fmt.Sprintf("- Version Change: %s → %s\n", request.FromVersion, request.ToVersion))
	prompt.WriteString(fmt.Sprintf("- Package Manager: %s\n", request.PackageManager))
	
	prompt.WriteString("\n**Project Context:**\n")
	prompt.WriteString(fmt.Sprintf("- Language: %s\n", request.ProjectContext.Language))
	prompt.WriteString(fmt.Sprintf("- Framework: %s\n", request.ProjectContext.Framework))
	
	if len(request.ProjectContext.Dependencies) > 0 {
		prompt.WriteString("- Dependencies:\n")
		for _, dep := range request.ProjectContext.Dependencies {
			prompt.WriteString(fmt.Sprintf("  - %s\n", dep))
		}
	}
	
	prompt.WriteString("\n\n**Compatibility Analysis Instructions:**\n")
	prompt.WriteString("Predict compatibility and provide a JSON response with the following structure:\n\n")
	
	prompt.WriteString(`{
  "package_name": "` + request.PackageName + `",
  "from_version": "` + request.FromVersion + `",
  "to_version": "` + request.ToVersion + `",
  "compatibility_score": 0.0-1.0,
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-1.0,
  "confidence": 0.0-1.0,
  "summary": "Compatibility assessment summary",
  "recommendations": ["recommendation1", "recommendation2"],
  "potential_issues": [
    {
      "type": "breaking_change|dependency_conflict|api_deprecation",
      "description": "Issue description",
      "severity": "low|medium|high|critical",
      "likelihood": 0.0-1.0,
      "mitigation": "How to address this issue"
    }
  ]
}`)
	
	prompt.WriteString("\n\n**Important Guidelines:**\n")
	prompt.WriteString("- Consider project context and existing dependencies\n")
	prompt.WriteString("- Assess likelihood of compatibility issues\n")
	prompt.WriteString("- Provide specific mitigation strategies\n")
	prompt.WriteString("- Higher compatibility scores indicate fewer expected issues\n")
	prompt.WriteString("- Respond ONLY with valid JSON, no additional text\n")
	
	return prompt.String()
}

// generateUpdateClassificationPrompt creates a prompt for update classification
func (o *OllamaProvider) generateUpdateClassificationPrompt(request *types.UpdateClassificationRequest) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert software dependency analyst. Classify the following package update and provide a detailed JSON response.\n\n")
	
	prompt.WriteString(fmt.Sprintf("**Package Information:**\n"))
	prompt.WriteString(fmt.Sprintf("- Package: %s\n", request.PackageName))
	prompt.WriteString(fmt.Sprintf("- Version Change: %s → %s\n", request.FromVersion, request.ToVersion))
	prompt.WriteString(fmt.Sprintf("- Package Manager: %s\n", request.PackageManager))
	
	if request.ChangelogText != "" {
		prompt.WriteString("\n**Changelog:**\n")
		prompt.WriteString(request.ChangelogText)
	}
	
	prompt.WriteString("\n**Project Context:**\n")
	prompt.WriteString(fmt.Sprintf("- Language: %s\n", request.ProjectContext.Language))
	prompt.WriteString(fmt.Sprintf("- Framework: %s\n", request.ProjectContext.Framework))
	
	if len(request.ProjectContext.Dependencies) > 0 {
		prompt.WriteString("- Dependencies:\n")
		for _, dep := range request.ProjectContext.Dependencies {
			prompt.WriteString(fmt.Sprintf("  - %s\n", dep))
		}
	}
	
	prompt.WriteString("\n\n**Classification Instructions:**\n")
	prompt.WriteString("Classify this update and provide a JSON response with the following structure:\n\n")
	
	prompt.WriteString(`{
  "package_name": "` + request.PackageName + `",
  "from_version": "` + request.FromVersion + `",
  "to_version": "` + request.ToVersion + `",
  "update_type": "major|minor|patch",
  "priority": "low|medium|high|critical",
  "urgency": "low|medium|high|immediate",
  "summary": "Classification summary",
  "recommendations": ["recommendation1", "recommendation2"],
  "categories": [
    {
      "name": "security|performance|bug_fix|feature|maintenance",
      "weight": 0.0-1.0,
      "description": "Category description"
    }
  ]
}`)
	
	prompt.WriteString("\n\n**Important Guidelines:**\n")
	prompt.WriteString("- Priority reflects importance for project stability and functionality\n")
	prompt.WriteString("- Urgency reflects time sensitivity (security fixes are typically urgent)\n")
	prompt.WriteString("- Categories should reflect the primary nature of the update\n")
	prompt.WriteString("- Weights should sum to approximately 1.0 across categories\n")
	prompt.WriteString("- Consider project context when determining priority and urgency\n")
	prompt.WriteString("- Respond ONLY with valid JSON, no additional text\n")
	
	return prompt.String()
}
