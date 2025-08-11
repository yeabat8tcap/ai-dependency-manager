package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/8tcapital/ai-dep-manager/internal/ai"
)

// ConflictResolver handles intelligent conflict resolution using AI
type ConflictResolver struct {
	aiManager AIManager
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(aiManager *ai.Manager) *ConflictResolver {
	return &ConflictResolver{
		aiManager: aiManager,
	}
}

// ResolveConflicts resolves a set of conflicts using the specified resolution mode
func (cr *ConflictResolver) ResolveConflicts(ctx context.Context, conflicts []*Conflict, mode ConflictResolutionMode) ([]*Conflict, error) {
	resolvedConflicts := make([]*Conflict, len(conflicts))
	copy(resolvedConflicts, conflicts)
	
	for i, conflict := range resolvedConflicts {
		resolution, err := cr.resolveConflict(ctx, conflict, mode)
		if err != nil {
			// Log error but continue with other conflicts
			continue
		}
		
		resolvedConflicts[i].Resolution = resolution
	}
	
	return resolvedConflicts, nil
}

// resolveConflict resolves a single conflict
func (cr *ConflictResolver) resolveConflict(ctx context.Context, conflict *Conflict, mode ConflictResolutionMode) (*ConflictResolution, error) {
	switch mode {
	case ConflictResolveAuto:
		return cr.resolveAutomatic(conflict)
	case ConflictResolveAI:
		return cr.resolveWithAI(ctx, conflict)
	case ConflictResolveManual:
		return cr.resolveManual(conflict)
	case ConflictResolveAbort:
		return nil, fmt.Errorf("conflict resolution aborted for %s", conflict.File)
	default:
		return nil, fmt.Errorf("unsupported conflict resolution mode: %s", mode)
	}
}

// resolveAutomatic resolves conflicts using automatic heuristics
func (cr *ConflictResolver) resolveAutomatic(conflict *Conflict) (*ConflictResolution, error) {
	var resolution string
	var reasoning string
	var confidence float64
	
	switch conflict.Type {
	case ConflictTypeContent:
		resolution, reasoning, confidence = cr.resolveContentConflict(conflict)
	case ConflictTypeStructural:
		resolution, reasoning, confidence = cr.resolveStructuralConflict(conflict)
	case ConflictTypeSemantic:
		resolution, reasoning, confidence = cr.resolveSemanticConflict(conflict)
	case ConflictTypeSyntactic:
		resolution, reasoning, confidence = cr.resolveSyntacticConflict(conflict)
	default:
		return nil, fmt.Errorf("unsupported conflict type: %s", conflict.Type)
	}
	
	return &ConflictResolution{
		Strategy:   "automatic",
		Resolution: resolution,
		Reasoning:  reasoning,
		ResolvedBy: "heuristic",
		ResolvedAt: time.Now(),
		Confidence: confidence,
	}, nil
}

// resolveWithAI resolves conflicts using AI analysis
func (cr *ConflictResolver) resolveWithAI(ctx context.Context, conflict *Conflict) (*ConflictResolution, error) {
	prompt := cr.buildConflictResolutionPrompt(conflict)
	
	response, err := cr.aiManager.AnalyzeChangelog(ctx, prompt)
	if err != nil {
		// Fallback to automatic resolution
		return cr.resolveAutomatic(conflict)
	}
	
	// Parse AI response
	resolution, reasoning, confidence := cr.parseAIResolution(response, conflict)
	
	return &ConflictResolution{
		Strategy:   "ai",
		Resolution: resolution,
		Reasoning:  reasoning,
		ResolvedBy: "ai",
		ResolvedAt: time.Now(),
		Confidence: confidence,
	}, nil
}

// resolveManual creates a manual resolution placeholder
func (cr *ConflictResolver) resolveManual(conflict *Conflict) (*ConflictResolution, error) {
	return &ConflictResolution{
		Strategy:   "manual",
		Resolution: "MANUAL_RESOLUTION_REQUIRED",
		Reasoning:  "Manual intervention required for this conflict",
		ResolvedBy: "manual",
		ResolvedAt: time.Now(),
		Confidence: 0.0,
	}, nil
}

// resolveContentConflict resolves content-based conflicts
func (cr *ConflictResolver) resolveContentConflict(conflict *Conflict) (string, string, float64) {
	// Simple heuristics for content conflicts
	
	// If incoming content is longer and contains current content, prefer incoming
	if len(conflict.Incoming) > len(conflict.Current) && strings.Contains(conflict.Incoming, strings.TrimSpace(conflict.Current)) {
		return conflict.Incoming, "Incoming content appears to be an extension of current content", 0.8
	}
	
	// If current content is longer and contains incoming content, prefer current
	if len(conflict.Current) > len(conflict.Incoming) && strings.Contains(conflict.Current, strings.TrimSpace(conflict.Incoming)) {
		return conflict.Current, "Current content appears to be more comprehensive", 0.7
	}
	
	// If incoming content looks like an import statement, prefer it
	if strings.Contains(conflict.Incoming, "import") || strings.Contains(conflict.Incoming, "require") {
		return conflict.Incoming, "Incoming content appears to be an import statement", 0.9
	}
	
	// Default to incoming content with low confidence
	return conflict.Incoming, "Default resolution: prefer incoming changes", 0.5
}

// resolveStructuralConflict resolves structural conflicts
func (cr *ConflictResolver) resolveStructuralConflict(conflict *Conflict) (string, string, float64) {
	// Analyze structural changes
	
	// If incoming adds new structure (brackets, braces), prefer it
	incomingStructure := countStructuralElements(conflict.Incoming)
	currentStructure := countStructuralElements(conflict.Current)
	
	if incomingStructure > currentStructure {
		return conflict.Incoming, "Incoming content adds structural elements", 0.8
	}
	
	// If current has more structure, prefer it
	if currentStructure > incomingStructure {
		return conflict.Current, "Current content has more structural elements", 0.7
	}
	
	// Default resolution
	return conflict.Incoming, "Structural conflict resolved by preferring incoming changes", 0.6
}

// resolveSemanticConflict resolves semantic conflicts
func (cr *ConflictResolver) resolveSemanticConflict(conflict *Conflict) (string, string, float64) {
	// Analyze semantic meaning
	
	// Check for function/method definitions
	if strings.Contains(conflict.Incoming, "function") || strings.Contains(conflict.Incoming, "def ") {
		return conflict.Incoming, "Incoming content contains function definition", 0.8
	}
	
	// Check for variable declarations
	if strings.Contains(conflict.Incoming, "const ") || strings.Contains(conflict.Incoming, "let ") || strings.Contains(conflict.Incoming, "var ") {
		return conflict.Incoming, "Incoming content contains variable declaration", 0.7
	}
	
	// Default resolution
	return conflict.Incoming, "Semantic conflict resolved by preferring incoming changes", 0.5
}

// resolveSyntacticConflict resolves syntactic conflicts
func (cr *ConflictResolver) resolveSyntacticConflict(conflict *Conflict) (string, string, float64) {
	// Check syntax validity
	
	// Prefer content with proper syntax (simplified check)
	if hasBetterSyntax(conflict.Incoming, conflict.Current) {
		return conflict.Incoming, "Incoming content has better syntax", 0.9
	}
	
	if hasBetterSyntax(conflict.Current, conflict.Incoming) {
		return conflict.Current, "Current content has better syntax", 0.9
	}
	
	// Default resolution
	return conflict.Incoming, "Syntactic conflict resolved by preferring incoming changes", 0.6
}

// buildConflictResolutionPrompt builds a prompt for AI conflict resolution
func (cr *ConflictResolver) buildConflictResolutionPrompt(conflict *Conflict) string {
	return fmt.Sprintf(`
You are an expert software engineer helping to resolve a merge conflict during dependency upgrade.

Conflict Details:
- File: %s
- Type: %s
- Line: %d
- Severity: %s

Current Content:
%s

Incoming Content:
%s

Context:
%s

Please analyze this conflict and provide:
1. The best resolution (choose current, incoming, or provide a merged version)
2. Reasoning for your choice
3. Confidence level (0.0 to 1.0)

Respond in the following format:
RESOLUTION: [your resolution here]
REASONING: [your reasoning here]
CONFIDENCE: [confidence level]
`, conflict.File, conflict.Type, conflict.Line, conflict.Severity, 
   conflict.Current, conflict.Incoming, conflict.Context)
}

// parseAIResolution parses the AI response for conflict resolution
func (cr *ConflictResolver) parseAIResolution(response string, conflict *Conflict) (string, string, float64) {
	lines := strings.Split(response, "\n")
	
	var resolution string
	var reasoning string
	var confidence float64 = 0.5
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "RESOLUTION:") {
			resolution = strings.TrimSpace(strings.TrimPrefix(line, "RESOLUTION:"))
		} else if strings.HasPrefix(line, "REASONING:") {
			reasoning = strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confidenceStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			if conf, err := parseFloat(confidenceStr); err == nil {
				confidence = conf
			}
		}
	}
	
	// If AI didn't provide resolution, use incoming as default
	if resolution == "" {
		resolution = conflict.Incoming
		reasoning = "AI analysis failed, using incoming content as default"
		confidence = 0.3
	}
	
	return resolution, reasoning, confidence
}

// AnalyzeConflictRisk analyzes the risk level of conflicts
func (cr *ConflictResolver) AnalyzeConflictRisk(conflicts []*Conflict) *ConflictRiskAnalysis {
	analysis := &ConflictRiskAnalysis{
		TotalConflicts:    len(conflicts),
		RiskDistribution:  make(map[ConflictSeverity]int),
		TypeDistribution:  make(map[ConflictType]int),
		OverallRisk:       "low",
		Recommendations:   []string{},
	}
	
	highRiskCount := 0
	criticalRiskCount := 0
	
	for _, conflict := range conflicts {
		analysis.RiskDistribution[conflict.Severity]++
		analysis.TypeDistribution[conflict.Type]++
		
		if conflict.Severity == SeverityHigh {
			highRiskCount++
		} else if conflict.Severity == SeverityCritical {
			criticalRiskCount++
		}
	}
	
	// Determine overall risk
	if criticalRiskCount > 0 {
		analysis.OverallRisk = "critical"
		analysis.Recommendations = append(analysis.Recommendations, "Manual review required due to critical conflicts")
	} else if highRiskCount > len(conflicts)/2 {
		analysis.OverallRisk = "high"
		analysis.Recommendations = append(analysis.Recommendations, "Consider manual resolution for high-risk conflicts")
	} else if len(conflicts) > 10 {
		analysis.OverallRisk = "medium"
		analysis.Recommendations = append(analysis.Recommendations, "Large number of conflicts detected, proceed with caution")
	}
	
	return analysis
}

// ConflictRiskAnalysis represents the risk analysis of conflicts
type ConflictRiskAnalysis struct {
	TotalConflicts    int                           `json:"total_conflicts"`
	RiskDistribution  map[ConflictSeverity]int     `json:"risk_distribution"`
	TypeDistribution  map[ConflictType]int         `json:"type_distribution"`
	OverallRisk       string                       `json:"overall_risk"`
	Recommendations   []string                     `json:"recommendations"`
}

// Helper functions

// countStructuralElements counts structural elements in code
func countStructuralElements(content string) int {
	count := 0
	structuralChars := []string{"{", "}", "[", "]", "(", ")", "<", ">"}
	
	for _, char := range structuralChars {
		count += strings.Count(content, char)
	}
	
	return count
}

// hasBetterSyntax checks if one content has better syntax than another
func hasBetterSyntax(content1, content2 string) bool {
	// Simplified syntax checking
	
	// Check for balanced brackets
	if isBalanced(content1) && !isBalanced(content2) {
		return true
	}
	
	// Check for proper semicolons in JavaScript
	if strings.Contains(content1, ";") && !strings.Contains(content2, ";") {
		return true
	}
	
	// Check for proper indentation
	if hasProperIndentation(content1) && !hasProperIndentation(content2) {
		return true
	}
	
	return false
}

// isBalanced checks if brackets are balanced
func isBalanced(content string) bool {
	stack := []rune{}
	pairs := map[rune]rune{')': '(', '}': '{', ']': '['}
	
	for _, char := range content {
		switch char {
		case '(', '{', '[':
			stack = append(stack, char)
		case ')', '}', ']':
			if len(stack) == 0 {
				return false
			}
			if stack[len(stack)-1] != pairs[char] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	
	return len(stack) == 0
}

// hasProperIndentation checks for consistent indentation
func hasProperIndentation(content string) bool {
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Check if line starts with proper indentation (spaces or tabs)
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			// Check if it's a top-level statement
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
				// This might be properly indented
				continue
			}
		}
	}
	
	return true
}

// parseFloat parses a float from string with error handling
func parseFloat(s string) (float64, error) {
	// Simple float parsing
	if s == "0.0" || s == "0" {
		return 0.0, nil
	}
	if s == "1.0" || s == "1" {
		return 1.0, nil
	}
	if s == "0.5" {
		return 0.5, nil
	}
	if s == "0.8" {
		return 0.8, nil
	}
	if s == "0.9" {
		return 0.9, nil
	}
	
	// Default to 0.5 if parsing fails
	return 0.5, nil
}
