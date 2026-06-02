package preflight

import "github.com/opwax/opwax/internal/models"

func addCentralCopyWarnings(ctx *models.RunContext, add func(models.PreflightSeverity, string, string, bool)) {
	if !anyCleanupModule(ctx.Config.Modules) {
		return
	}
	add(models.SeverityWarning, "Central & cloud copies",
		"Local cleanup does NOT remove copies forwarded off-box: Windows Event Forwarding (WEF) to SIEM, "+
			"EDR/Defender cloud telemetry, Entra ID / Microsoft 365 sign-in logs, Intune compliance logs, "+
			"OneDrive version history, or corporate backup/DR snapshots. Assume central copies may still exist.", true)
	add(models.SeverityInfo, "BitLocker & offline forensics",
		"Full-disk encryption (BitLocker) reduces offline disk imaging risk but does not affect live logged-in artifacts OpWAX targets.", true)
}

func anyCleanupModule(m models.ModuleConfig) bool {
	return m.VolatileMemory || m.RegistryHives || m.NTFSMetadata || m.ProgramExecution ||
		m.SystemLogs || m.PersistenceStorage || m.NetworkBrowser
}

func addGapArtifactChecks(ctx *models.RunContext, add func(models.PreflightSeverity, string, string, bool)) {
	adv := ctx.Config.Options.Advanced
	if ctx.Config.Modules.RegistryHives && adv.WindowsSearchIndex {
		add(models.SeverityInfo, "Windows Search index",
			"WSearch will be stopped; Windows.edb and related index files deleted.", true)
	}
	if ctx.Config.Modules.RegistryHives && adv.ExecutionRegistries {
		add(models.SeverityInfo, "Execution registry keys",
			"MUICache, RecentApps, and AppCompatFlags will be cleared for target users.", true)
	}
	if ctx.Config.Modules.RegistryHives && adv.ShellIconCaches {
		add(models.SeverityInfo, "Shell caches",
			"Explorer iconcache_*.db and thumbcache_*.db will be deleted (icons may rebuild on next login).", true)
	}
	if ctx.Config.Modules.ProgramExecution && adv.SyscacheHive {
		add(models.SeverityInfo, "Syscache.hve",
			"Syscache.hve and RecentFileCache.bcf will be securely deleted.", true)
	}
	if ctx.Config.Modules.SystemLogs && adv.TargetedEventChannels {
		add(models.SeverityInfo, "Targeted execution EVTX",
			"Task Scheduler and Application Experience operational channels will be disabled and cleared.", true)
	}
	if ctx.Config.Modules.NetworkBrowser && adv.EventTranscript {
		add(models.SeverityInfo, "EventTranscript.db",
			"Diagnosis EventTranscript SQLite database will be deleted.", true)
	}
	if ctx.Config.Modules.PersistenceStorage && adv.WERReports {
		add(models.SeverityInfo, "WER reports",
			"Windows Error Reporting folders (system + user) will be cleared.", true)
	}
	if adv.ServicingLogs {
		add(models.SeverityWarning, "Servicing logs",
			"Panther, CBS, and Windows Update log folders will be cleared - may affect troubleshooting history.", true)
	}
	if adv.PrintSpooler {
		add(models.SeverityInfo, "Print spooler",
			"Print queue files will be deleted; Spooler service briefly stopped.", true)
	}
	if adv.DeveloperTraces {
		add(models.SeverityInfo, "Developer traces",
			"Git/npm/Python/VS Code/Cursor local histories will be cleared for target users.", true)
	}
	if ctx.Config.Options.FocusedCleanupMode {
		add(models.SeverityWarning, "Focused cleanup mode",
			"Explorer will be stopped during cleanup to reduce new Prefetch/USN/shell artifacts; desktop restarts when finished.", true)
	}
	if ctx.Config.Options.SecondPassAfterReboot {
		add(models.SeverityInfo, "Second pass after reboot",
			"A one-time logon scheduled task will re-run OpWAX after reboot to clear locked files (auto-removes itself).", true)
	}
}
