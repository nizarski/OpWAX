package system

import (
	"fmt"
	"strings"

	"github.com/opwax/opwax/internal/models"
)

// CompareManifests diffs before and after cleanup manifests.
func CompareManifests(before, after *models.SystemManifest) models.ManifestDiff {
	diff := models.ManifestDiff{}
	if before == nil || after == nil {
		return diff
	}

	afterMap := map[string]models.ArtifactEntry{}
	for _, a := range after.Artifacts {
		afterMap[a.ID] = a
	}

	for _, b := range before.Artifacts {
		if !b.Exists {
			continue
		}
		a, ok := afterMap[b.ID]
		descBefore := artifactSummary(b)

		if !ok || !a.Exists {
			diff.Removed = append(diff.Removed, models.ManifestDiffEntry{
				ID: b.ID, Name: b.Name, Path: b.Path,
				Before: descBefore, After: "gone",
				Status: "removed",
			})
			continue
		}

		descAfter := artifactSummary(a)
		if b.ItemCount > 0 && a.ItemCount >= 0 && a.ItemCount < b.ItemCount {
			diff.Reduced = append(diff.Reduced, models.ManifestDiffEntry{
				ID: b.ID, Name: b.Name, Path: b.Path,
				Before: descBefore, After: descAfter,
				Status: "reduced",
			})
		} else if a.Exists && (b.SizeBytes > 0 && a.SizeBytes > 0) {
			diff.StillPresent = append(diff.StillPresent, models.ManifestDiffEntry{
				ID: b.ID, Name: b.Name, Path: b.Path,
				Before: descBefore, After: descAfter,
				Status: "still_present",
			})
		} else if !a.Exists {
			diff.Removed = append(diff.Removed, models.ManifestDiffEntry{
				ID: b.ID, Name: b.Name, Path: b.Path,
				Before: descBefore, After: "gone",
				Status: "removed",
			})
		} else {
			diff.Unchanged++
		}
	}
	return diff
}

func artifactSummary(a models.ArtifactEntry) string {
	if a.ItemCount > 0 {
		return fmt.Sprintf("exists (%d items)", a.ItemCount)
	}
	if a.SizeBytes > 0 {
		return fmt.Sprintf("exists (%d bytes)", a.SizeBytes)
	}
	if a.Exists {
		return "exists"
	}
	return "missing"
}

// FormatManifestDiff returns human-readable diff text.
func FormatManifestDiff(d models.ManifestDiff) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Manifest diff: %d removed, %d reduced, %d still present, %d unchanged\n\n",
		len(d.Removed), len(d.Reduced), len(d.StillPresent), d.Unchanged))

	if len(d.Removed) > 0 {
		b.WriteString("REMOVED:\n")
		for _, e := range d.Removed {
			b.WriteString(fmt.Sprintf("  + %s (%s)\n", e.Name, e.Path))
		}
		b.WriteString("\n")
	}
	if len(d.Reduced) > 0 {
		b.WriteString("REDUCED:\n")
		for _, e := range d.Reduced {
			b.WriteString(fmt.Sprintf("  ~ %s: %s -> %s\n", e.Name, e.Before, e.After))
		}
		b.WriteString("\n")
	}
	if len(d.StillPresent) > 0 {
		b.WriteString("STILL PRESENT (may need reboot):\n")
		for _, e := range d.StillPresent {
			b.WriteString(fmt.Sprintf("  ! %s: %s\n", e.Name, e.After))
		}
	}
	return b.String()
}
