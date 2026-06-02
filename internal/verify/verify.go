package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/artifacts"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"
)

type check struct {
	name     string
	category string
	path     string
	kind     checkKind
}

type checkKind int

const (
	checkFileGone checkKind = iota
	checkDirEmpty
	checkUSNInactive
)

// Run scans expected-cleared locations and reports anything still present.
func Run(ctx *models.RunContext) models.VerificationReport {
	report := models.VerificationReport{}
	for _, c := range buildChecks(ctx) {
		report.Checked++
		if gap := runCheck(c); gap != nil {
			report.Gaps = append(report.Gaps, *gap)
		} else {
			report.Clean++
		}
	}
	return report
}

func buildChecks(ctx *models.RunContext) []check {
	adv := ctx.Config.Options.Advanced
	winDir := ctx.WindowsDir
	if winDir == "" {
		winDir = system.WindowsDirectory()
	}
	var checks []check

	addFile := func(cat, name, path string) {
		checks = append(checks, check{name: name, category: cat, path: path, kind: checkFileGone})
	}
	addDir := func(cat, name, path string) {
		checks = append(checks, check{name: name, category: cat, path: path, kind: checkDirEmpty})
	}

	if ctx.Config.Modules.ProgramExecution {
		prefetch := filepath.Join(winDir, "Prefetch")
		addDir("execution", "Prefetch folder", prefetch)
		addFile("execution", "Amcache.hve", filepath.Join(winDir, "appcompat", "Programs", "Amcache.hve"))
		if adv.SyscacheHive {
			addFile("execution", "Syscache.hve", filepath.Join(winDir, "appcompat", "Programs", "Syscache.hve"))
		}
	}

	if ctx.Config.Modules.RegistryHives {
		if adv.WindowsSearchIndex {
			addFile("search", "Windows Search index", artifacts.WindowsSearchIndexPath())
		}
	}

	if adv.EventTranscript && ctx.Config.Modules.NetworkBrowser {
		addFile("diagnosis", "EventTranscript.db", artifacts.EventTranscriptPath())
	}

	if adv.WERReports && ctx.Config.Modules.PersistenceStorage {
		addDir("wer", "WER system reports", artifacts.WERSystemPath())
	}

	if ctx.Config.Modules.SystemLogs {
		addDir("logs", "Winevt\\Logs", filepath.Join(winDir, "System32", "Winevt", "Logs"))
		for _, u := range ctx.TargetUsers {
			recent := filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent")
			addDir("logs", "Jump lists ("+u.Username+")", filepath.Join(recent, "AutomaticDestinations"))
		}
	}

	if ctx.Config.Modules.NTFSMetadata {
		for _, d := range ctx.TargetDrives {
			checks = append(checks, check{
				name:     "USN change journal (" + d.Letter + ")",
				category: "ntfs",
				path:     d.Letter,
				kind:     checkUSNInactive,
			})
		}
	}

	if ctx.Config.Modules.NetworkBrowser {
		srum := filepath.Join(winDir, "System32", "sru", "SRUDB.dat")
		addFile("network", "SRUDB.dat", srum)
	}

	if ctx.Config.Modules.ProgramExecution {
		if adv.WindowsRecall {
			for _, p := range artifacts.RecallCapturePaths(ctx.TargetUsers) {
				addDir("modern", "Windows Recall capture", p)
			}
		}
		if adv.PCAAppCompatLogs {
			for _, p := range artifacts.PCALaunchLogPaths() {
				addFile("modern", "PCA launch log", p)
			}
		}
		if adv.NotepadTabCache {
			for _, p := range artifacts.NotepadTabStatePaths(ctx.TargetUsers) {
				addDir("modern", "Notepad tab cache", p)
			}
		}
		if adv.PowerShellHistory {
			for _, u := range ctx.TargetUsers {
				addFile("execution", "PowerShell history ("+u.Username+")",
					filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"))
			}
		}
	}

	if ctx.Config.Modules.SystemLogs && adv.TimelineActivity {
		for _, p := range artifacts.TimelineDBPaths(ctx.TargetUsers) {
			addFile("logs", "Windows Timeline DB", p)
		}
	}

	if ctx.Config.Modules.NetworkBrowser {
		if adv.OutlookEnabled() && adv.Outlook == "delete_ost_pst" {
			for _, u := range ctx.TargetUsers {
				dir := filepath.Join(u.AppDataLocal, "Microsoft", "Outlook")
				matches, _ := filepath.Glob(filepath.Join(dir, "*.ost"))
				for _, m := range matches {
					addFile("email", "Outlook OST", m)
				}
				matches, _ = filepath.Glob(filepath.Join(dir, "*.pst"))
				for _, m := range matches {
					addFile("email", "Outlook PST", m)
				}
			}
		}
	}

	return checks
}

func runCheck(c check) *models.VerificationGap {
	switch c.kind {
	case checkUSNInactive:
		if util.USNJournalActive(c.path) {
			return &models.VerificationGap{
				Category: c.category,
				Name:     c.name,
				Path:     c.path,
				Detail:   "USN change journal still active",
			}
		}
		return nil
	case checkFileGone:
		info, err := os.Stat(c.path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return &models.VerificationGap{
				Category: c.category,
				Name:     c.name,
				Path:     c.path,
				Detail:   "could not stat: " + err.Error(),
			}
		}
		if info.IsDir() {
			return &models.VerificationGap{
				Category: c.category,
				Name:     c.name,
				Path:     c.path,
				Detail:   "expected file but path is a directory",
			}
		}
		return &models.VerificationGap{
			Category: c.category,
			Name:     c.name,
			Path:     c.path,
			Detail:   fmt.Sprintf("file still exists (%d bytes)", info.Size()),
		}
	case checkDirEmpty:
		entries, err := os.ReadDir(c.path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return &models.VerificationGap{
				Category: c.category,
				Name:     c.name,
				Path:     c.path,
				Detail:   "could not read: " + err.Error(),
			}
		}
		var files []string
		for _, e := range entries {
			name := e.Name()
			if name == "." || name == ".." {
				continue
			}
			files = append(files, name)
		}
		if len(files) == 0 {
			return nil
		}
		detail := fmt.Sprintf("%d item(s) remain", len(files))
		if len(files) <= 3 {
			detail += ": " + strings.Join(files, ", ")
		}
		return &models.VerificationGap{
			Category: c.category,
			Name:     c.name,
			Path:     c.path,
			Detail:   detail,
		}
	default:
		return nil
	}
}

// FormatReport returns human-readable verification output.
func FormatReport(r models.VerificationReport) string {
	if r.Checked == 0 {
		return "Post-run verification: no checks configured.\n"
	}
	if len(r.Gaps) == 0 {
		return fmt.Sprintf("Post-run verification: all %d check(s) passed - no remaining gaps detected.\n", r.Checked)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Post-run verification: %d gap(s) of %d check(s) still present\n\n", len(r.Gaps), r.Checked))
	b.WriteString("REMAINING ARTIFACTS (may need reboot or re-run):\n")
	for _, g := range r.Gaps {
		b.WriteString(fmt.Sprintf("  ! [%s] %s\n     %s\n     → %s\n", g.Category, g.Name, g.Path, g.Detail))
	}
	return b.String()
}
