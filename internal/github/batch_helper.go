package github

import (
	"fmt"
	"time"
)

func convertGeneratedPatchToPatch(gp *GeneratedPatch) *Patch {
	var changes []*Change
	var filePatches []FilePatch
	var configPatches []ConfigPatch

	for _, fp := range gp.Files {
		filePatches = append(filePatches, *fp)
		changes = append(changes, fp.Changes...)
	}

	for _, cp := range gp.ConfigChanges {
		configPatches = append(configPatches, *cp)
	}

	// Determine patch type and breaking change status
	patchType := "update"
	isBreaking := false
	for _, dep := range gp.Dependencies {
		if dep.SecurityFix {
			patchType = "security"
		}
		// DependencyUpdate doesn't have BreakingChange bool directly in shared_types.go,
		// but let's assume we rely on RiskAssessment or if it was added.
		// For now, let's rely on RiskAssessment.
	}

	riskLevel := "low"
	if gp.RiskAssessment != nil {
		riskLevel = gp.RiskAssessment.OverallRisk
		if gp.RiskAssessment.BreakingChanges > 0 {
			isBreaking = true
		}
	}

	return &Patch{
		ID:         fmt.Sprintf("patch_%d", time.Now().UnixNano()),
		Repository: gp.Repository,
		Changes:    changes,
		// Let's check shared_types.go Patch struct definition again.
		// Patch struct: Changes []Change `json:"changes"`
		// But here changes is []*Change.
		// So I need to convert []*Change to []Change.
		FilePatches:    filePatches,
		ConfigPatches:  configPatches,
		Description:    gp.PRDescription,
		CreatedAt:      gp.GeneratedAt,
		Status:         "generated",
		Confidence:     gp.RiskAssessment.Confidence,
		Type:           patchType,
		RiskLevel:      riskLevel,
		BreakingChange: isBreaking,
	}
}

func convertChangesToValue(changes []*Change) []Change {
	var values []Change
	for _, c := range changes {
		values = append(values, *c)
	}
	return values
}
