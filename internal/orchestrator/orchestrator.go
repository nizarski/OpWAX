package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/opwax/opwax/internal/config"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/modules"
	"github.com/opwax/opwax/internal/preflight"
	"github.com/opwax/opwax/internal/scheduler"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"
	"github.com/opwax/opwax/internal/verify"
)

// Orchestrator coordinates dry-run and execution phases.
type Orchestrator struct{}

// New creates an Orchestrator.
func New() *Orchestrator {
	return &Orchestrator{}
}

// BuildContext resolves targets from config and system discovery.
func (o *Orchestrator) BuildContext(cfg models.Config) (*models.RunContext, error) {
	if err := system.RequireAdmin(); err != nil {
		return nil, err
	}

	currentUser, err := system.CurrentUsername()
	if err != nil {
		return nil, err
	}

	var targetUsernames []string
	switch cfg.Targets.UserMode {
	case models.UserModeCurrent:
		targetUsernames = []string{currentUser}
	case models.UserModeAll:
		targetUsernames, err = system.EnumerateUsers()
		if err != nil {
			return nil, err
		}
	case models.UserModeSelect:
		targetUsernames = cfg.Targets.SelectedUsers
	}

	profiles, err := system.ResolveUserProfiles(targetUsernames)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, fmt.Errorf("no matching user profiles found")
	}

	allDrives, err := system.EnumerateDrives()
	if err != nil {
		return nil, err
	}

	var targetDrives []models.DriveInfo
	switch cfg.Targets.DriveMode {
	case models.DriveModeSystem:
		for _, d := range allDrives {
			if d.IsSystem {
				targetDrives = append(targetDrives, d)
			}
		}
	case models.DriveModeAll:
		targetDrives = allDrives
	case models.DriveModeSelect:
		want := map[string]bool{}
		for _, letter := range cfg.Targets.SelectedDrives {
			want[strings.ToUpper(strings.TrimSuffix(letter, `\`))] = true
		}
		for _, d := range allDrives {
			if want[strings.ToUpper(d.Letter)] {
				targetDrives = append(targetDrives, d)
			}
		}
	}
	if len(targetDrives) == 0 {
		return nil, fmt.Errorf("no matching drives found")
	}

	return &models.RunContext{
		Config:       cfg,
		CurrentUser:  currentUser,
		TargetUsers:  profiles,
		TargetDrives: targetDrives,
		WindowsDir:   system.WindowsDirectory(),
		ProgramData:  system.ProgramDataDir(),
	}, nil
}

// DryRun collects all planned actions from enabled modules.
func (o *Orchestrator) DryRun(ctx *models.RunContext) models.DryRunReport {
	var actions []models.Action
	actions = append(actions, modules.BootstrapDryRun(ctx)...)
	for _, mod := range modules.EnabledModules(ctx.Config.Modules) {
		actions = append(actions, mod.DryRun(ctx)...)
	}
	actions = append(actions, modules.MFTScrubDryRun(ctx)...)
	return models.DryRunReport{Actions: actions}
}

// Preflight runs pre-execution validation checks.
func (o *Orchestrator) Preflight(ctx *models.RunContext) models.PreflightReport {
	return preflight.Run(ctx)
}

// Execute runs disable → clean → secure phases across all modules.
func (o *Orchestrator) Execute(ctx *models.RunContext) models.ExecutionReport {
	return o.ExecuteWithProgress(ctx, nil, nil)
}

// ExecuteWithProgress runs cleanup and reports per-module progress.
// Cancel is optional; when closed, execution stops before the next module step.
func (o *Orchestrator) ExecuteWithProgress(ctx *models.RunContext, onProgress models.ProgressFunc, cancel <-chan struct{}) models.ExecutionReport {
	report := models.ExecutionReport{}
	mods := modules.EnabledModules(ctx.Config.Modules)
	total := len(mods)

	var manifestBefore *models.SystemManifest
	if ctx.Config.Options.ManifestDiff {
		manifestBefore, _ = system.GenerateManifest()
	}

	emit := func(phase models.ProgressPhase, idx int, modName, msg string, complete bool) {
		if onProgress == nil {
			return
		}
		onProgress(models.ProgressUpdate{
			Phase:       phase,
			Module:      modName,
			ModuleIndex: idx + 1,
			ModuleTotal: total,
			Message:     msg,
			Complete:    complete,
		})
	}

	var restoreExplorer func()
	if ctx.Config.Options.FocusedCleanupMode {
		emit(models.ProgressPhaseBootstrap, 0, "bootstrap", "Pausing Explorer shell…", false)
		restoreExplorer = util.BeginFocusedCleanup()
		defer func() {
			if restoreExplorer != nil {
				restoreExplorer()
			}
		}()
	}

	stopCancelled := func(phase models.ProgressPhase, idx int, modName string) bool {
		if !cancelled(cancel) {
			return false
		}
		report.Cancelled = true
		if modName != "" {
			report.CancelledAt = string(phase) + ":" + modName
		} else {
			report.CancelledAt = string(phase)
		}
		emit(phase, idx, modName, "Cancelled", true)
		return true
	}

	// Phase 0: Priority bootstrap - stop EventLog, Prefetch, USN before any other work.
	if stopCancelled(models.ProgressPhaseBootstrap, 0, "bootstrap") {
		return report
	}
	emit(models.ProgressPhaseBootstrap, 0, "bootstrap", "Stopping collectors (EventLog, Prefetch, USN)…", false)
	report.Results = append(report.Results, modules.BootstrapDisable(ctx)...)
	emit(models.ProgressPhaseBootstrap, 0, "bootstrap", "Collectors stopped", true)

	// Phase 1: Disable all (stop future recording)
	for i, mod := range mods {
		name := mod.Name()
		if stopCancelled(models.ProgressPhaseDisable, i, name) {
			return report
		}
		emit(models.ProgressPhaseDisable, i, name, "Disabling future logging…", false)
		report.Results = append(report.Results, mod.Disable(ctx)...)
		emit(models.ProgressPhaseDisable, i, name, "Disable complete", true)
	}

	// Phase 2: Clean all (remove existing artifacts)
	for i, mod := range mods {
		name := mod.Name()
		if stopCancelled(models.ProgressPhaseClean, i, name) {
			return report
		}
		emit(models.ProgressPhaseClean, i, name, "Removing artifacts…", false)
		report.Results = append(report.Results, mod.Clean(ctx)...)
		emit(models.ProgressPhaseClean, i, name, "Clean complete", true)
	}

	// Phase 3: Secure - MFT free-record scrub after all deletes (optional).
	if len(modules.MFTScrubDryRun(ctx)) > 0 {
		if stopCancelled(models.ProgressPhaseSecure, total-1, "ntfs_metadata") {
			return report
		}
		emit(models.ProgressPhaseSecure, total-1, "ntfs_metadata", "Scrubbing free MFT records + free space…", false)
		report.Results = append(report.Results, modules.MFTScrub(ctx)...)
		emit(models.ProgressPhaseSecure, total-1, "ntfs_metadata", "MFT scrub complete", true)
	}

	if stopCancelled(models.ProgressPhaseFinalize, total-1, "") {
		return report
	}
	emit(models.ProgressPhaseFinalize, total-1, "", "Finalizing…", false)

	if ctx.Config.Options.ManifestDiff && manifestBefore != nil {
		if stopCancelled(models.ProgressPhaseFinalize, total-1, "") {
			return report
		}
		emit(models.ProgressPhaseFinalize, total-1, "", "Generating after-manifest…", false)
		if manifestAfter, err := system.GenerateManifest(); err == nil {
			diff := system.CompareManifests(manifestBefore, manifestAfter)
			report.ManifestDiff = &diff
		}
	}

	if ctx.Config.Options.PostRunVerification && !report.Cancelled {
		emit(models.ProgressPhaseFinalize, total-1, "", "Verifying artifact removal…", false)
		v := verify.Run(ctx)
		report.Verification = &v
	}

	// Check if reboot recommended for locked files
	for _, r := range report.Results {
		if !r.Success && strings.Contains(strings.ToLower(r.Error), "being used") {
			report.NeedsReboot = true
		}
	}

	if ctx.Config.Options.RebootAfter || (ctx.Config.Options.LSASSScrub && ctx.Config.Options.LSASSRebootAfter) {
		if !report.Cancelled {
			if err := rebootNow(); err == nil {
				report.RebootQueued = true
			}
		}
	} else if report.NeedsReboot {
		report.NeedsReboot = true
	}

	if ctx.Config.Options.LSASSScrub && ctx.Config.Options.LSASSRebootAfter && !report.RebootQueued {
		report.NeedsReboot = true
	}

	if ctx.Config.Options.SecondPassAfterReboot && !report.Cancelled {
		needsSecondPass := report.NeedsReboot || report.RebootQueued || ctx.Config.Options.RebootAfter
		if report.Verification != nil && len(report.Verification.Gaps) > 0 {
			needsSecondPass = true
		}
		if needsSecondPass {
			if cfgPath, err := saveSecondPassConfig(ctx.Config); err == nil {
				if exe, err := os.Executable(); err == nil {
					if err := scheduler.ScheduleSecondPass(exe, cfgPath); err == nil {
						report.SecondPassScheduled = true
					}
				}
			}
		}
	}

	emit(models.ProgressPhaseFinalize, total-1, "", "Done", true)
	return report
}

func saveSecondPassConfig(cfg models.Config) (string, error) {
	dir := filepath.Join(os.Getenv("ProgramData"), "OpWAX")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "second-pass-config.json")
	// Second pass should not re-schedule itself.
	cfg2 := cfg
	cfg2.Options.SecondPassAfterReboot = false
	cfg2.Options.FocusedCleanupMode = false
	if err := config.Save(path, cfg2); err != nil {
		return "", err
	}
	return path, nil
}

func cancelled(cancel <-chan struct{}) bool {
	if cancel == nil {
		return false
	}
	select {
	case <-cancel:
		return true
	default:
		return false
	}
}

func rebootNow() error {
	cmd := exec.Command("shutdown", "/r", "/t", "5", "/c", "OpWAX cleanup complete")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// FormatDryRun returns human-readable dry-run text for GUI preview.
func FormatDryRun(report models.DryRunReport) string {
	var b strings.Builder
	b.WriteString("Execution order: bootstrap (collectors + optional Explorer pause) → disable all → clean all → MFT scrub (if enabled) → verify → optional second pass\n\n")
	b.WriteString(fmt.Sprintf("Planned actions: %d\n\n", len(report.Actions)))
	curModule := ""
	for i, a := range report.Actions {
		if a.Module != curModule {
			curModule = a.Module
			b.WriteString(fmt.Sprintf("\n[%s]\n", strings.ToUpper(curModule)))
		}
		b.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, a.Kind, a.Description))
		if a.Target != "" {
			b.WriteString(fmt.Sprintf("     → %s\n", a.Target))
		}
	}
	return b.String()
}

// FormatReport returns execution summary (in-memory display only).
func FormatReport(report models.ExecutionReport) string {
	if report.Cancelled {
		msg := "Cleanup cancelled - partial changes may have been applied.\n"
		if report.CancelledAt != "" {
			msg += "Stopped before: " + report.CancelledAt + "\n"
		}
		ok, fail := 0, 0
		for _, r := range report.Results {
			if r.Success {
				ok++
			} else {
				fail++
			}
		}
		msg += fmt.Sprintf("Completed steps: %d succeeded, %d failed\n", ok, fail)
		if fail > 0 {
			msg += "\nFailed actions:\n"
			for _, r := range report.Results {
				if !r.Success {
					msg += fmt.Sprintf("  • %s: %s\n", r.Action.Description, r.Error)
				}
			}
		}
		if report.ManifestDiff != nil {
			msg += "\n" + system.FormatManifestDiff(*report.ManifestDiff)
		}
		if report.Verification != nil {
			msg += "\n" + verify.FormatReport(*report.Verification)
		}
		return msg
	}

	ok, fail := 0, 0
	for _, r := range report.Results {
		if r.Success {
			ok++
		} else {
			fail++
		}
	}
	msg := fmt.Sprintf("Complete: %d succeeded, %d failed\n", ok, fail)
	if report.RebootQueued {
		msg += "System reboot scheduled in 5 seconds.\n"
	} else if report.NeedsReboot {
		msg += "Some files were locked - reboot recommended to finish cleanup.\n"
	}
	if report.SecondPassScheduled {
		msg += "One-time second-pass cleanup scheduled at next logon (task auto-deletes after run).\n"
	}
	if fail > 0 {
		msg += "\nFailed actions:\n"
		for _, r := range report.Results {
			if !r.Success {
				msg += fmt.Sprintf("  • %s: %s\n", r.Action.Description, r.Error)
			}
		}
	}
	if report.ManifestDiff != nil {
		msg += "\n" + system.FormatManifestDiff(*report.ManifestDiff)
	}
	if report.Verification != nil {
		msg += "\n" + verify.FormatReport(*report.Verification)
	}
	return msg
}
