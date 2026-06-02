package preflight

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"
)

// Run executes pre-flight checks for a resolved run context.
func Run(ctx *models.RunContext) models.PreflightReport {
	var checks []models.PreflightCheck
	canProceed := true
	warnings := 0

	add := func(sev models.PreflightSeverity, title, msg string, proceed bool) {
		checks = append(checks, models.PreflightCheck{
			Severity:   sev,
			Title:      title,
			Message:    msg,
			CanProceed: proceed,
		})
		if !proceed {
			canProceed = false
		}
		if sev == models.SeverityWarning || sev == models.SeverityCritical {
			warnings++
		}
	}

	if len(ctx.TargetUsers) == 0 {
		add(models.SeverityCritical, "No users", "No target user profiles resolved.", false)
	}
	if len(ctx.TargetDrives) == 0 {
		add(models.SeverityCritical, "No drives", "No target drives resolved.", false)
	}

	if ctx.Config.Modules.NetworkBrowser {
		running := detectRunningBrowsers(ctx.Config.Options.Browsers)
		if len(running) > 0 {
			add(models.SeverityWarning, "Browsers running",
				fmt.Sprintf("%s - will be force-closed during cleanup.", strings.Join(running, ", ")), true)
		}
	}

	if ctx.Config.Modules.SystemLogs {
		add(models.SeverityWarning, "Event Log service",
			"EventLog, Wecsvc, and related collectors will be stopped/disabled; the entire Winevt\\Logs folder will be purged.", true)
		add(models.SeverityWarning, "Security log clearing (Event ID 1102)",
			"Clearing Security logs is itself auditable - examiners may see evidence that logs were wiped (including Event ID 1102 on some systems).", true)
		evtxDir := filepath.Join(ctx.WindowsDir, "System32", "Winevt", "Logs")
		if n := countFiles(evtxDir, ".evtx"); n > 0 {
			add(models.SeverityInfo, "Event logs",
				fmt.Sprintf("%d .evtx files will be cleared.", n), true)
		}
	}

	if ctx.Config.Options.MFTFreeSpaceScrub {
		for _, d := range ctx.TargetDrives {
			add(models.SeverityWarning, "MFT free-space scrub "+d.Letter,
				"Native Go scrubber will 1-pass zero-fill free space on "+d.Letter+" - can still take a long time on large drives.", true)
		}
	}
	if ctx.Config.Options.LogFileResetOnReboot {
		add(models.SeverityInfo, "$LogFile reset",
			"chkdsk /F will be scheduled on system drive - runs on next reboot.", true)
	}
	if ctx.Config.Options.LSASSScrub {
		add(models.SeverityWarning, "LSASS scrub",
			"Credential caches cleared live; full LSASS RAM requires reboot (no injection).", true)
		if ctx.Config.Options.LSASSRebootAfter {
			add(models.SeverityInfo, "LSASS reboot",
				"System will reboot after cleanup to clear remaining LSASS RAM.", true)
		}
	}

	if ctx.Config.Modules.NTFSMetadata {
		for _, d := range ctx.TargetDrives {
			add(models.SeverityWarning, "USN journal "+d.Letter,
				"NTFS change journal will be disabled (fsutil usn deletejournal /N). Windows Search indexing may rescan the volume.", true)
			add(models.SeverityInfo, "Zone.Identifier scan",
				fmt.Sprintf("User profile folders on %s will be scanned for Zone.Identifier ADS (Downloads, Desktop, Documents, Temp).", d.Letter), true)
		}
	}

	if ctx.Config.Modules.ProgramExecution {
		add(models.SeverityInfo, "SysMain",
			"SysMain (SuperFetch) and PcaSvc will be disabled. Prefetch files deleted.", true)
	}

	if ctx.Config.Modules.PersistenceStorage {
		for _, d := range ctx.TargetDrives {
			if _, err := os.Stat(d.Root + "pagefile.sys"); err == nil {
				add(models.SeverityWarning, "Pagefile",
					fmt.Sprintf("pagefile.sys on %s may remain locked until reboot.", d.Letter), true)
			}
			if d.IsSystem {
				if _, err := os.Stat(d.Root + "hiberfil.sys"); err == nil {
					add(models.SeverityInfo, "Hibernation",
						"Hibernation will be disabled; hiberfil.sys removed if not locked.", true)
				}
			}
		}
		add(models.SeverityInfo, "Recycle Bin",
			"Recycle bin file content ($R*) will be overwritten once (fast zero-fill), then metadata removed.", true)
	}

	if ctx.Config.Targets.UserMode == models.UserModeAll || len(ctx.TargetUsers) > 1 {
		add(models.SeverityInfo, "Multi-user",
			fmt.Sprintf("Cleanup targets %d user profile(s). Logged-in users' hives may be partially locked.", len(ctx.TargetUsers)), true)
	}

	if ctx.Config.Modules.NetworkBrowser {
		sru := filepath.Join(ctx.WindowsDir, "System32", "sru", "SRUDB.dat")
		if _, err := os.Stat(sru); err == nil {
			add(models.SeverityInfo, "SRUM",
				"DiagTrack and DPS will be disabled; SRUDB.dat securely deleted.", true)
		}
	}

	if len(checks) == 0 {
		add(models.SeverityInfo, "Ready", "All pre-flight checks passed.", true)
	}

	addAdvancedChecks(ctx, add)
	addGapArtifactChecks(ctx, add)
	addCentralCopyWarnings(ctx, add)

	return models.PreflightReport{
		Checks:       checks,
		CanProceed:   canProceed,
		WarningCount: warnings,
	}
}

func detectRunningBrowsers(b models.BrowserConfig) []string {
	var running []string
	for exe, name := range system.BrowserProcessNames(b) {
		out, err := util.RunHiddenOutput("tasklist", "/FI", "IMAGENAME eq "+exe, "/NH")
		if err == nil && strings.Contains(strings.ToLower(out), strings.ToLower(exe)) {
			running = append(running, name)
		}
	}
	return running
}

func countFiles(dir, ext string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ext) {
			n++
		}
	}
	return n
}

// FormatReport returns human-readable preflight text.
func FormatReport(r models.PreflightReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Pre-flight: %d check(s), %d warning(s)\n", len(r.Checks), r.WarningCount))
	if !r.CanProceed {
		b.WriteString("CANNOT PROCEED - fix critical issues first.\n")
	}
	b.WriteString("\n")
	for _, c := range r.Checks {
		tag := "INFO"
		switch c.Severity {
		case models.SeverityWarning:
			tag = "WARN"
		case models.SeverityCritical:
			tag = "CRIT"
		}
		b.WriteString(fmt.Sprintf("[%s] %s\n    %s\n\n", tag, c.Title, c.Message))
	}
	return b.String()
}
