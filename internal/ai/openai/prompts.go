package openai

import (
	"fmt"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// buildChangelogAnalysisPrompt builds the prompt for changelog analysis
func (o *OpenAIProvider) buildChangelogAnalysisPrompt(request *types.ChangelogAnalysisRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`Analyze the following changelog for package "%s" updating from version %s to %s.

Package Information:
- Package: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Language: %s

Changelog Text:
%s

Release Notes:
%s

Please provide a comprehensive analysis in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "has_breaking_change": boolean,
  "breaking_changes": [
    {
      "type": "api_removal|signature_change|behavior_change|deprecation|removal",
      "description": "detailed description",
      "impact": "impact on existing code",
      "severity": "low|medium|high|critical",
      "confidence": 0.0-1.0,
      "mitigation": "steps to mitigate",
      "affected_apis": ["list of affected APIs"]
    }
  ],
  "new_features": [
    {
      "name": "feature name",
      "description": "detailed description",
      "type": "api|functionality|performance|security",
      "impact": "impact description",
      "confidence": 0.0-1.0,
      "benefits": ["list of benefits"],
      "usage_example": "example usage if applicable"
    }
  ],
  "bug_fixes": [
    {
      "description": "bug fix description",
      "impact": "impact of the fix",
      "severity": "low|medium|high|critical",
      "confidence": 0.0-1.0,
      "affected_components": ["list of affected components"]
    }
  ],
  "security_fixes": [
    {
      "description": "security fix description",
      "severity": "low|medium|high|critical",
      "cve": "CVE identifier if available",
      "cvss": 0.0-10.0,
      "impact": "security impact",
      "confidence": 0.0-1.0,
      "references": ["security advisory URLs"],
      "urgency": "low|medium|high|critical"
    }
  ],
  "deprecations": [
    {
      "api": "deprecated API",
      "replacement": "replacement API",
      "timeline": "deprecation timeline",
      "impact": "impact description",
      "migration_guide": "migration instructions"
    }
  ],
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-10.0,
  "confidence": 0.0-1.0,
  "summary": "concise summary of changes",
  "recommendations": ["actionable recommendations"],
  "migration_steps": ["step-by-step migration guide"],
  "testing_advice": ["testing recommendations"],
  "recommended_timeline": "suggested update timeline",
  "business_impact": "business impact assessment"
}

Focus on:
1. Identifying breaking changes with high accuracy
2. Assessing security implications
3. Providing actionable migration guidance
4. Evaluating business and technical impact
5. Giving realistic confidence scores based on available information`,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ChangelogText, request.ReleaseNotes,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildVersionDiffPrompt builds the prompt for version diff analysis
func (o *OpenAIProvider) buildVersionDiffPrompt(request *types.VersionDiffAnalysisRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`Analyze the version difference for package "%s" from %s to %s.

Package Information:
- Package: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Language: %s

Diff Text:
%s

File Changes:
`, request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language, request.DiffText))

	for _, change := range request.FileChanges {
		prompt.WriteString(fmt.Sprintf("- %s (%s): +%d/-%d lines\n", 
			change.Path, change.Type, change.LinesAdded, change.LinesRemoved))
	}

	prompt.WriteString(fmt.Sprintf(`

Please provide a comprehensive version diff analysis in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "update_type": "major|minor|patch|prerelease",
  "semantic_impact": "detailed semantic versioning impact",
  "api_changes": [
    {
      "type": "addition|modification|removal|deprecation",
      "api": "API name or signature",
      "description": "change description",
      "impact": "impact on existing code",
      "severity": "low|medium|high|critical",
      "examples": ["code examples"],
      "migration": "migration instructions"
    }
  ],
  "behavior_changes": [
    {
      "component": "affected component",
      "description": "behavior change description",
      "impact": "impact assessment",
      "likelihood": 0.0-1.0,
      "testing_advice": "testing recommendations"
    }
  ],
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-10.0,
  "confidence": 0.0-1.0,
  "summary": "concise analysis summary",
  "recommendations": ["actionable recommendations"],
  "migration_effort": "low|medium|high|very_high",
  "backward_compatibility": boolean
}

Focus on:
1. Semantic versioning compliance analysis
2. API surface changes and their implications
3. Behavioral changes that might affect existing code
4. Migration complexity assessment
5. Backward compatibility evaluation`, 
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildCompatibilityPrompt builds the prompt for compatibility prediction
func (o *OpenAIProvider) buildCompatibilityPrompt(request *types.CompatibilityPredictionRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`Predict compatibility issues for updating package "%s" from %s to %s.

Package Information:
- Package: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Language: %s

Project Context:
- Framework: %s
- Language: %s
- Dependencies: %v

Dependency Graph:
`, request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ProjectContext.Framework, request.ProjectContext.Language,
		request.ProjectContext.Dependencies))

	for _, dep := range request.DependencyGraph {
		prompt.WriteString(fmt.Sprintf("- %s@%s (%s)\n", dep.Name, dep.Version, dep.Type))
	}

	prompt.WriteString(fmt.Sprintf(`

Please provide a comprehensive compatibility prediction in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "compatibility_score": 0.0-1.0,
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-10.0,
  "confidence": 0.0-1.0,
  "potential_issues": [
    {
      "type": "breaking_change|dependency_conflict|version_mismatch|api_incompatibility",
      "description": "issue description",
      "severity": "low|medium|high|critical",
      "likelihood": 0.0-1.0,
      "impact": "impact assessment",
      "mitigation": "mitigation strategy",
      "detection": "how to detect this issue"
    }
  ],
  "migration_steps": [
    {
      "step": "step name",
      "description": "detailed description",
      "priority": "low|medium|high|critical",
      "effort": "low|medium|high|very_high",
      "risk": "low|medium|high|critical",
      "validation": "how to validate this step"
    }
  ],
  "testing_recommendations": [
    {
      "type": "unit|integration|e2e|performance|security",
      "description": "testing recommendation",
      "priority": "low|medium|high|critical",
      "test_cases": ["specific test cases"],
      "tools": ["recommended testing tools"]
    }
  ],
  "summary": "compatibility assessment summary",
  "recommendations": ["actionable recommendations"],
  "estimated_effort": "low|medium|high|very_high",
  "rollback_complexity": "low|medium|high|very_high"
}

Focus on:
1. Dependency conflict analysis
2. API compatibility assessment
3. Framework and language version compatibility
4. Migration complexity evaluation
5. Risk mitigation strategies`, 
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildUpdateClassificationPrompt builds the prompt for update classification
func (o *OpenAIProvider) buildUpdateClassificationPrompt(request *types.UpdateClassificationRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`Classify the update for package "%s" from %s to %s.

Package Information:
- Package: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Language: %s

Changelog Text:
%s

Release Notes:
%s

Please provide a comprehensive update classification in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "update_type": "major|minor|patch|security|hotfix|prerelease",
  "priority": "low|medium|high|critical",
  "categories": [
    {
      "name": "security|feature|bugfix|maintenance|performance|documentation",
      "weight": 0.0-1.0,
      "description": "category description",
      "impact": "impact assessment",
      "examples": ["specific examples from changelog"]
    }
  ],
  "urgency": "low|medium|high|critical",
  "recommended_timeline": "immediate|within_week|within_month|next_cycle",
  "business_impact": "impact on business operations",
  "technical_impact": "impact on technical systems",
  "risk_assessment": {
    "level": "low|medium|high|critical",
    "score": 0.0-10.0,
    "factors": ["risk factors"],
    "mitigation": ["mitigation strategies"],
    "monitoring": ["monitoring recommendations"]
  },
  "summary": "classification summary",
  "recommendations": ["actionable recommendations"],
  "dependency_impacts": [
    {
      "dependency": "affected dependency",
      "impact": "impact description",
      "likelihood": 0.0-1.0,
      "mitigation": "mitigation approach"
    }
  ]
}

Focus on:
1. Accurate update type classification
2. Priority assessment based on content and impact
3. Business and technical impact evaluation
4. Risk assessment and mitigation
5. Timeline recommendations based on urgency`,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ChangelogText, request.ReleaseNotes,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}
