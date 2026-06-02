package modules

import (
	"path/filepath"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

const bootstrapName = "bootstrap"

// BootstrapDryRun returns priority disable actions that always run first.
func BootstrapDryRun(ctx *models.RunContext) []models.Action {
	var actions []models.Action
	if ctx.Config.Modules.SystemLogs {
		actions = append(actions,
			action(bootstrapName, "Priority: clear audit policy", "auditpol", models.ActionDisable),
			action(bootstrapName, "Priority: stop/disable EventLog collector stack", "EventLog; Wecsvc; WdiSystemHost", models.ActionDisable),
		)
	}
	if ctx.Config.Modules.ProgramExecution {
		actions = append(actions,
			action(bootstrapName, "Priority: disable Prefetcher + SysMain + PcaSvc", "PrefetchParameters; SysMain; PcaSvc", models.ActionDisable),
		)
	}
	if ctx.Config.Modules.NTFSMetadata {
		for _, d := range ctx.TargetDrives {
			actions = append(actions,
				action(bootstrapName, "Priority: reset USN change journal "+d.Letter, d.Letter, models.ActionDisable),
			)
		}
	}
	if gapCollectorsNeeded(ctx) {
		actions = append(actions,
			action(bootstrapName, "Priority: stop WSearch + DiagTrack + DPS + SysMain", "WSearch; DiagTrack; DPS; SysMain", models.ActionDisable),
		)
	}
	if ctx.Config.Options.FocusedCleanupMode {
		actions = append(actions,
			action(bootstrapName, "Priority: pause Explorer shell during cleanup", "explorer.exe", models.ActionDisable),
		)
	}
	return actions
}

// BootstrapDisable stops EventLog, Prefetch, and USN collectors before any other work.
func BootstrapDisable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	if ctx.Config.Modules.SystemLogs {
		a1 := action(bootstrapName, "Priority: clear audit policy", "auditpol", models.ActionDisable)
		results = append(results, result(a1, util.RunHidden("auditpol", "/clear", "/y")))

		a2 := action(bootstrapName, "Priority: disable EventLog collector stack", "EventLog; Wecsvc; WdiSystemHost", models.ActionDisable)
		results = append(results, result(a2, util.DisableEventLogStack()))
	}

	if ctx.Config.Modules.ProgramExecution {
		prefetchPath := `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management\PrefetchParameters`
		a := action(bootstrapName, "Priority: disable Prefetcher", prefetchPath, models.ActionDisable)
		err := util.SetRegDWORD(registry.LOCAL_MACHINE, prefetchPath, "EnablePrefetcher", 0)
		_ = util.SetRegDWORD(registry.LOCAL_MACHINE, prefetchPath, "EnableSuperfetch", 0)
		results = append(results, result(a, err))

		a2 := action(bootstrapName, "Priority: disable SysMain", "SysMain", models.ActionDisable)
		results = append(results, result(a2, util.StopAndDisableService("SysMain")))

		a3 := action(bootstrapName, "Priority: disable PcaSvc", "PcaSvc", models.ActionDisable)
		results = append(results, result(a3, util.StopAndDisableService("PcaSvc")))
	}

	if ctx.Config.Modules.NTFSMetadata {
		for _, d := range ctx.TargetDrives {
			a := action(bootstrapName, "Priority: reset USN change journal "+d.Letter, d.Letter, models.ActionDisable)
			results = append(results, result(a, util.ResetUSNJournal(d.Letter)))
		}
	}

	if gapCollectorsNeeded(ctx) {
		a := action(bootstrapName, "Priority: collector kill (WSearch/DiagTrack/DPS/SysMain)", "WSearch; DiagTrack; DPS; SysMain", models.ActionDisable)
		results = append(results, result(a, util.CollectorKillDisable()))
	}

	return results
}

// MFTScrubDryRun returns final-phase MFT scrub actions (after all deletes).
func MFTScrubDryRun(ctx *models.RunContext) []models.Action {
	if !ctx.Config.Modules.NTFSMetadata || !ctx.Config.Options.MFTFreeSpaceScrub {
		return nil
	}
	var actions []models.Action
	for _, d := range ctx.TargetDrives {
		actions = append(actions,
			action(ntfsName, "Final: scrub free MFT records + free space on "+d.Letter, d.Letter, models.ActionSecure),
		)
	}
	return actions
}

// MFTScrub runs free MFT record / free-space scrub after all artifact deletes.
func MFTScrub(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Modules.NTFSMetadata || !ctx.Config.Options.MFTFreeSpaceScrub {
		return nil
	}
	var results []models.Result
	for _, d := range ctx.TargetDrives {
		a := action(ntfsName, "Final: scrub free MFT records "+d.Letter, d.Letter, models.ActionSecure)
		results = append(results, result(a, util.ScrubMFTFreeSpace(d.Letter)))
	}
	return results
}

// BootstrapCleanPaths returns paths that must not be touched until bootstrap disable completes.
func BootstrapCleanPaths(ctx *models.RunContext) []string {
	var paths []string
	if ctx.Config.Modules.ProgramExecution {
		paths = append(paths, filepath.Join(system.WindowsDirectory(), "Prefetch"))
	}
	if ctx.Config.Modules.SystemLogs {
		paths = append(paths, filepath.Join(system.WindowsDirectory(), "System32", "Winevt", "Logs"))
	}
	return paths
}
