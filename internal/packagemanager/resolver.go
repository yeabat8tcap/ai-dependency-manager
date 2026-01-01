package packagemanager

import (
	"strings"
)

// ConflictResolver analyzes proposed updates for potential conflicts
type ConflictResolver struct {
	dependencies []DependencyEntry
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(dependencies []DependencyEntry) *ConflictResolver {
	return &ConflictResolver{
		dependencies: dependencies,
	}
}

// Conflict represents a detected dependency conflict
type Conflict struct {
	DependencyName  string
	ProposedVersion string
	ParentName      string
	Constraint      string
	Type            string // diamond, direct, transitive
}

// AnalyzeUpdate checks if updating a dependency causes any conflicts
func (r *ConflictResolver) AnalyzeUpdate(depName string, newVersion string) ([]Conflict, error) {
	var conflicts []Conflict

	for _, d := range r.dependencies {
		if len(d.Requirements) == 0 {
			continue
		}

		requirements := d.Requirements

		// Check if this dependency depends on the one we're updating
		if constraint, exists := requirements[depName]; exists {
			if !isCompatible(newVersion, constraint) {
				conflicts = append(conflicts, Conflict{
					DependencyName:  depName,
					ProposedVersion: newVersion,
					ParentName:      d.Name,
					Constraint:      constraint,
					Type:            "transitive",
				})
			}
		}
	}

	return conflicts, nil
}

// isCompatible checks if a version satisfies a constraint
// This is a simplified version for the demo
func isCompatible(version, constraint string) bool {
	// Remove prefixes like ^, ~, ==, >=, <=
	cleanConstraint := strings.TrimLeft(constraint, "^~=><!")

	// If it's a direct equality constraint
	if strings.HasPrefix(constraint, "==") || (!strings.ContainsAny(constraint, "><^~") && constraint != "") {
		return version == cleanConstraint
	}

	// For the demo, we'll implement simple logic for ^ and >=
	if strings.HasPrefix(constraint, "^") {
		// Simplified: same major version
		vParts := strings.Split(version, ".")
		cParts := strings.Split(cleanConstraint, ".")
		if len(vParts) > 0 && len(cParts) > 0 {
			return vParts[0] == cParts[0]
		}
	}

	if strings.HasPrefix(constraint, ">=") {
		// Simplified: just check string comparison if lengths match
		return version >= cleanConstraint
	}

	// Default to true for unknown constraints in this simplified demo
	return true
}
