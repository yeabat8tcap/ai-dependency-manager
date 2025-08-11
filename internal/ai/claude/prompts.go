package claude

import (
	"fmt"
	"strings"

	"github.com/8tcapital/ai-dep-manager/internal/ai/types"
)

// buildChangelogAnalysisPrompt builds the prompt for changelog analysis
func (c *ClaudeProvider) buildChangelogAnalysisPrompt(request *types.ChangelogAnalysisRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`I need you to analyze the following changelog for a software dependency update. Please provide a comprehensive analysis focusing on breaking changes, security implications, and migration guidance.

**Package Information:**
- Package Name: %s
- Current Version: %s
- Target Version: %s
- Package Manager: %s
- Programming Language: %s

**Changelog Content:**
%s

**Release Notes:**
%s

**Analysis Requirements:**

Please analyze this changelog and provide your response in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "has_breaking_change": boolean,
  "breaking_changes": [
    {
      "type": "api_removal|signature_change|behavior_change|deprecation|removal",
      "description": "detailed description of the breaking change",
      "impact": "specific impact on existing code and applications",
      "severity": "low|medium|high|critical",
      "confidence": 0.0-1.0,
      "mitigation": "specific steps to address this breaking change",
      "affected_apis": ["list", "of", "affected", "APIs"]
    }
  ],
  "new_features": [
    {
      "name": "feature name",
      "description": "detailed description of the new feature",
      "type": "api|functionality|performance|security",
      "impact": "impact and benefits of this feature",
      "confidence": 0.0-1.0,
      "benefits": ["list", "of", "benefits"],
      "usage_example": "example of how to use this feature"
    }
  ],
  "bug_fixes": [
    {
      "description": "description of the bug that was fixed",
      "impact": "impact of this fix on existing functionality",
      "severity": "low|medium|high|critical",
      "confidence": 0.0-1.0,
      "affected_components": ["list", "of", "affected", "components"]
    }
  ],
  "security_fixes": [
    {
      "description": "description of the security issue that was fixed",
      "severity": "low|medium|high|critical",
      "cve": "CVE identifier if available",
      "cvss": 0.0-10.0,
      "impact": "security impact and implications",
      "confidence": 0.0-1.0,
      "references": ["security", "advisory", "URLs"],
      "urgency": "low|medium|high|critical"
    }
  ],
  "deprecations": [
    {
      "api": "deprecated API or feature",
      "replacement": "recommended replacement",
      "timeline": "deprecation and removal timeline",
      "impact": "impact of this deprecation",
      "migration_guide": "steps to migrate away from deprecated feature"
    }
  ],
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-10.0,
  "confidence": 0.0-1.0,
  "summary": "concise summary of all changes and their implications",
  "recommendations": ["actionable", "recommendations", "for", "developers"],
  "migration_steps": ["step-by-step", "migration", "guide"],
  "testing_advice": ["specific", "testing", "recommendations"],
  "recommended_timeline": "suggested timeline for implementing this update",
  "business_impact": "assessment of business and operational impact"
}

**Key Analysis Points:**

1. **Breaking Changes**: Identify any changes that could break existing code, including API removals, signature changes, or behavioral modifications.

2. **Security Assessment**: Evaluate any security fixes or implications, including CVE references and urgency levels.

3. **Feature Analysis**: Assess new features and their potential benefits or risks.

4. **Migration Complexity**: Evaluate the effort required to migrate to the new version.

5. **Risk Assessment**: Provide an overall risk level and score based on the changes.

6. **Confidence Scoring**: Base confidence scores on the quality and completeness of the changelog information.

Please ensure your response contains only valid JSON following the exact schema above.`,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ChangelogText, request.ReleaseNotes,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildVersionDiffPrompt builds the prompt for version diff analysis
func (c *ClaudeProvider) buildVersionDiffPrompt(request *types.VersionDiffAnalysisRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`I need you to analyze version differences between two releases of a software package. Focus on semantic versioning compliance, API changes, and behavioral modifications.

**Package Information:**
- Package Name: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Programming Language: %s

**Version Diff Content:**
%s

**File Changes Summary:**
`, request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language, request.DiffText))

	for _, change := range request.FileChanges {
		prompt.WriteString(fmt.Sprintf("- %s (%s): +%d lines added, -%d lines removed\n", 
			change.Path, change.Type, change.LinesAdded, change.LinesRemoved))
	}

	prompt.WriteString(fmt.Sprintf(`

Please analyze these version differences and provide your response in the following JSON format:

{
  "package_name": "%s",
  "from_version": "%s",
  "to_version": "%s",
  "update_type": "major|minor|patch|prerelease",
  "semantic_impact": "detailed analysis of semantic versioning implications",
  "api_changes": [
    {
      "type": "addition|modification|removal|deprecation",
      "api": "specific API name or signature",
      "description": "detailed description of the change",
      "impact": "impact on existing code using this API",
      "severity": "low|medium|high|critical",
      "examples": ["code", "examples", "showing", "changes"],
      "migration": "specific migration instructions"
    }
  ],
  "behavior_changes": [
    {
      "component": "affected component or module",
      "description": "description of behavioral change",
      "impact": "assessment of impact on existing functionality",
      "likelihood": 0.0-1.0,
      "testing_advice": "specific testing recommendations for this change"
    }
  ],
  "risk_level": "low|medium|high|critical",
  "risk_score": 0.0-10.0,
  "confidence": 0.0-1.0,
  "summary": "comprehensive summary of version differences",
  "recommendations": ["actionable", "recommendations", "for", "developers"],
  "migration_effort": "low|medium|high|very_high",
  "backward_compatibility": boolean
}

**Analysis Focus Areas:**

1. **Semantic Versioning**: Analyze whether the version change follows semantic versioning principles and what it implies.

2. **API Surface Changes**: Identify additions, modifications, removals, or deprecations in the public API.

3. **Behavioral Changes**: Detect changes in behavior that might affect existing implementations.

4. **Backward Compatibility**: Assess whether the changes maintain backward compatibility.

5. **Migration Assessment**: Evaluate the complexity and effort required for migration.

Please ensure your response contains only valid JSON following the exact schema above.`,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildCompatibilityPrompt builds the prompt for compatibility prediction
func (c *ClaudeProvider) buildCompatibilityPrompt(request *types.CompatibilityPredictionRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`I need you to predict potential compatibility issues when updating a software dependency. Consider the project context and dependency relationships.

**Package Information:**
- Package Name: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Programming Language: %s

**Project Context:**
- Framework: %s
- Language: %s
- Dependencies: %v

**Current Dependency Graph:**
`, request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ProjectContext.Framework, request.ProjectContext.Language,
		request.ProjectContext.Dependencies))

	for _, dep := range request.DependencyGraph {
		prompt.WriteString(fmt.Sprintf("- %s@%s (type: %s)\n", dep.Name, dep.Version, dep.Type))
	}

	prompt.WriteString(fmt.Sprintf(`

Please analyze potential compatibility issues and provide your response in the following JSON format:

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
      "description": "detailed description of the potential issue",
      "severity": "low|medium|high|critical",
      "likelihood": 0.0-1.0,
      "impact": "assessment of impact if this issue occurs",
      "mitigation": "specific mitigation strategies",
      "detection": "how to detect this issue during testing"
    }
  ],
  "migration_steps": [
    {
      "step": "migration step name",
      "description": "detailed description of this migration step",
      "priority": "low|medium|high|critical",
      "effort": "low|medium|high|very_high",
      "risk": "low|medium|high|critical",
      "validation": "how to validate this step was completed successfully"
    }
  ],
  "testing_recommendations": [
    {
      "type": "unit|integration|e2e|performance|security",
      "description": "specific testing recommendation",
      "priority": "low|medium|high|critical",
      "test_cases": ["specific", "test", "cases", "to", "implement"],
      "tools": ["recommended", "testing", "tools"]
    }
  ],
  "summary": "comprehensive compatibility assessment summary",
  "recommendations": ["actionable", "recommendations", "for", "safe", "migration"],
  "estimated_effort": "low|medium|high|very_high",
  "rollback_complexity": "low|medium|high|very_high"
}

**Compatibility Analysis Focus:**

1. **Dependency Conflicts**: Analyze potential conflicts with other dependencies in the project.

2. **Version Constraints**: Check for version constraint violations or mismatches.

3. **API Compatibility**: Assess compatibility of APIs used by the project.

4. **Framework Integration**: Evaluate compatibility with the project's framework and runtime.

5. **Migration Planning**: Provide structured migration steps with risk assessment.

6. **Testing Strategy**: Recommend comprehensive testing approaches for validation.

Please ensure your response contains only valid JSON following the exact schema above.`,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}

// buildUpdateClassificationPrompt builds the prompt for update classification
func (c *ClaudeProvider) buildUpdateClassificationPrompt(request *types.UpdateClassificationRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf(`I need you to classify a software dependency update based on its content and impact. Focus on categorization, priority assessment, and business implications.

**Package Information:**
- Package Name: %s
- From Version: %s
- To Version: %s
- Package Manager: %s
- Programming Language: %s

**Changelog Content:**
%s

**Release Notes:**
%s

Please classify this update and provide your response in the following JSON format:

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
      "description": "description of this category's relevance",
      "impact": "impact assessment for this category",
      "examples": ["specific", "examples", "from", "changelog"]
    }
  ],
  "urgency": "low|medium|high|critical",
  "recommended_timeline": "immediate|within_week|within_month|next_cycle",
  "business_impact": "assessment of impact on business operations",
  "technical_impact": "assessment of impact on technical systems",
  "risk_assessment": {
    "level": "low|medium|high|critical",
    "score": 0.0-10.0,
    "factors": ["risk", "factors", "identified"],
    "mitigation": ["mitigation", "strategies"],
    "monitoring": ["monitoring", "recommendations"]
  },
  "summary": "comprehensive classification summary",
  "recommendations": ["actionable", "recommendations", "for", "implementation"],
  "dependency_impacts": [
    {
      "dependency": "name of affected dependency",
      "impact": "description of impact on this dependency",
      "likelihood": 0.0-1.0,
      "mitigation": "mitigation approach for this dependency"
    }
  ]
}

**Classification Criteria:**

1. **Update Type**: Classify based on semantic versioning and content (major, minor, patch, security, hotfix, prerelease).

2. **Priority Assessment**: Determine priority based on security fixes, breaking changes, and business impact.

3. **Category Analysis**: Identify primary categories (security, feature, bugfix, maintenance, performance, documentation) with weights.

4. **Urgency Evaluation**: Assess how quickly this update should be implemented.

5. **Impact Assessment**: Evaluate both business and technical implications.

6. **Risk Analysis**: Comprehensive risk assessment with mitigation strategies.

7. **Timeline Recommendation**: Suggest appropriate implementation timeline.

**Priority Guidelines:**
- **Critical**: Security vulnerabilities, critical bug fixes, or breaking changes affecting core functionality
- **High**: Important features, significant bug fixes, or performance improvements
- **Medium**: Minor features, non-critical bug fixes, or maintenance updates
- **Low**: Documentation updates, minor improvements, or optional enhancements

Please ensure your response contains only valid JSON following the exact schema above.`,
		request.PackageName, request.FromVersion, request.ToVersion,
		request.PackageManager, request.Language,
		request.ChangelogText, request.ReleaseNotes,
		request.PackageName, request.FromVersion, request.ToVersion))

	return prompt.String()
}
