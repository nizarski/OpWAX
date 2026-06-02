package models

// PreflightSeverity classifies a pre-run warning.
type PreflightSeverity string

const (
	SeverityInfo    PreflightSeverity = "info"
	SeverityWarning PreflightSeverity = "warning"
	SeverityCritical PreflightSeverity = "critical"
)

// PreflightCheck is a single pre-run validation result.
type PreflightCheck struct {
	Severity    PreflightSeverity `json:"severity"`
	Title       string            `json:"title"`
	Message     string            `json:"message"`
	CanProceed  bool              `json:"can_proceed"`
}

// PreflightReport aggregates pre-run checks.
type PreflightReport struct {
	Checks      []PreflightCheck `json:"checks"`
	CanProceed  bool             `json:"can_proceed"`
	WarningCount int             `json:"warning_count"`
}

// ManifestDiffEntry describes a change between before/after manifests.
type ManifestDiffEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Before   string `json:"before"` // e.g. "exists (142 items)"
	After    string `json:"after"`
	Status   string `json:"status"` // removed, reduced, unchanged, still_present
}

// ManifestDiff compares two system manifests.
type ManifestDiff struct {
	Removed      []ManifestDiffEntry `json:"removed"`
	Reduced      []ManifestDiffEntry `json:"reduced"`
	StillPresent []ManifestDiffEntry `json:"still_present"`
	Unchanged    int                 `json:"unchanged_count"`
}
