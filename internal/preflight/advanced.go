package preflight

import (
	"fmt"
	"strings"

	"github.com/opwax/opwax/internal/artifacts"
	"github.com/opwax/opwax/internal/models"
)

func addAdvancedChecks(ctx *models.RunContext, add func(models.PreflightSeverity, string, string, bool)) {
	adv := ctx.Config.Options.Advanced

	if ctx.Config.Modules.NTFSMetadata {
		if adv.DeleteVSSShadows {
			add(models.SeverityCritical, "Volume Shadow Copies",
				"All restore points and VSS snapshots will be permanently deleted. System Restore rollback will not be available.", true)
		}
		if adv.FullVolumeUnallocated {
			for _, d := range ctx.TargetDrives {
				add(models.SeverityWarning, "Full-volume unallocated wipe "+d.Letter,
					"cipher /w will 1-pass zero-fill all free space on "+d.Letter+" - can take hours on large drives.", true)
			}
		}
		if adv.WipeFileSlack {
			scope := "known artifact files"
			if adv.WipeFileSlackAllFiles {
				scope = "all files under target user profiles (very slow)"
			}
			add(models.SeverityWarning, "File slack space",
				"Residual bytes at end of file clusters will be zeroed for "+scope+".", true)
		}
		if adv.ScrubBadClusters {
			add(models.SeverityCritical, "Bad cluster scan",
				"chkdsk /B will be scheduled - may require reboot on system drive and can take a long time.", true)
		}
	}

	if ctx.Config.Modules.ProgramExecution {
		if adv.PowerShellHistory {
			add(models.SeverityInfo, "PowerShell history",
				"PSReadLine history saving will be disabled and ConsoleHost_history.txt deleted for target users.", true)
		}
		if adv.UserAssist {
			add(models.SeverityInfo, "UserAssist",
				"All GUI program launch counts and last-run timestamps (ROT13 UserAssist keys) will be cleared.", true)
		}
		if adv.RDPCacheEnabled() {
			add(models.SeverityWarning, "RDP bitmap cache",
				"Terminal Server Client cache*.bin files will be deleted - remote session thumbnails cannot be recovered.", true)
		}
		if adv.WindowsRecall {
			add(models.SeverityInfo, "Windows Recall",
				"Recall / CoreAI capture will be disabled and SQLite capture data under CoreAIPlatform\\Capture deleted.", true)
		}
		if adv.PCAAppCompatLogs {
			add(models.SeverityInfo, "PCA launch logs",
				"PcaAppLaunchDic.txt and related PCA trace files will be deleted.", true)
		}
		if adv.NotepadTabCache {
			add(models.SeverityInfo, "Notepad tab cache",
				"Unsaved Notepad tab/draft state under WindowsNotepad package LocalState will be removed.", true)
		}
	}

	if ctx.Config.Modules.RegistryHives {
		if adv.OfficeTrustRecords {
			add(models.SeverityInfo, "Office TrustRecords",
				"Trusted macro document records under Office Security keys will be cleared for target users.", true)
		}
		if adv.SysinternalsEULA {
			add(models.SeverityInfo, "Sysinternals EULA keys",
				"HKCU\\Software\\Sysinternals EULA acceptance values will be deleted.", true)
		}
	}

	if ctx.Config.Modules.SystemLogs && adv.TimelineActivity {
		add(models.SeverityInfo, "Windows Timeline",
			"Activity History will be disabled and ActivitiesCache.db removed for target users.", true)
	}

	if ctx.Config.Modules.NetworkBrowser {
		if adv.OneDriveEnabled() {
			add(models.SeverityWarning, "OneDrive ("+adv.OneDrive+")",
				"OneDrive sync metadata/logs will be cleared"+oneDriveExtra(adv.OneDrive)+".", true)
		}
		if adv.CloudSyncEnabled() {
			add(models.SeverityWarning, "Cloud sync caches",
				"Dropbox and Google Drive local caches may be cleared; cloud accounts are not removed.", true)
		}
		if adv.OutlookEnabled() {
			sev := models.SeverityWarning
			msg := "Outlook local caches will be cleared."
			if adv.Outlook == "delete_ost_pst" {
				sev = models.SeverityCritical
				msg = "Outlook OST/PST files will be securely deleted - offline mail and archives on this PC will be lost."
			}
			add(sev, "Outlook ("+adv.Outlook+")", msg, true)
		}
		if adv.TeamsEnabled() {
			sev := models.SeverityWarning
			msg := "Teams local cache folders will be cleared."
			if adv.Teams == "full" {
				sev = models.SeverityCritical
				msg = "Full Teams profile data will be removed - local chat history and settings will be lost."
			}
			add(sev, "Teams ("+adv.Teams+")", msg, true)
		}
	}

	if ctx.Config.Modules.PersistenceStorage {
		if adv.WMIEnabled() {
			add(models.SeverityCritical, "WMI repository",
				"The WMI repository (OBJECTS.DATA) will be reset - some management tools may fail until WMI rebuilds.", true)
		}
		if adv.ScheduledTasksEnabled() {
			add(models.SeverityCritical, "Scheduled tasks ("+adv.ScheduledTasks+")",
				"Task XML files under System32\\Tasks will be deleted per selected mode.", true)
		}
		if adv.BITSEnabled() {
			add(models.SeverityWarning, "BITS ("+adv.BITS+")",
				"Background Intelligent Transfer queue database will be cleared"+bitsExtra(adv.BITS)+".", true)
		}
		if adv.AlternateRunKeysEnabled() {
			add(models.SeverityWarning, "Alternate Run keys",
				"RunServices / Explorer\\Run values will be cleared"+runKeysExtra(adv.AlternateRunKeys)+".", true)
		}
		if adv.HyperVEnabled() {
			add(models.SeverityCritical, "Hyper-V snapshots",
				"All Hyper-V snapshot data under ProgramData will be deleted.", true)
		}
		if adv.WSLEnabled() && adv.WSL == "delete_vhdx" {
			paths := artifacts.WSLVHDXPaths(ctx.TargetUsers)
			add(models.SeverityCritical, "WSL virtual disks",
				fmt.Sprintf("%d WSL ext4.vhdx file(s) will be securely deleted - entire Linux distros will be destroyed.", len(paths)), true)
		}
	}
}

func oneDriveExtra(mode string) string {
	if mode == "full" {
		return " and OneDrive sync will be disabled"
	}
	return ""
}

func bitsExtra(mode string) string {
	if mode == "disable_clear" {
		return " and the BITS service will be disabled"
	}
	return ""
}

func runKeysExtra(mode string) string {
	if mode == "clean_disable" {
		return " and future RunServices entries will be blocked by policy"
	}
	return ""
}

// FormatReportMarkdown returns preflight output with severity emphasis for GUI.
func FormatReportMarkdown(r models.PreflightReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("**Pre-flight:** %d check(s), %d warning(s)\n\n", len(r.Checks), r.WarningCount))
	if !r.CanProceed {
		b.WriteString("**CANNOT PROCEED** - fix critical issues first.\n\n")
	}
	for _, c := range r.Checks {
		switch c.Severity {
		case models.SeverityCritical:
			b.WriteString(fmt.Sprintf("🔴 **CRITICAL - %s**\n\n%s\n\n", c.Title, c.Message))
		case models.SeverityWarning:
			b.WriteString(fmt.Sprintf("🟠 **WARNING - %s**\n\n%s\n\n", c.Title, c.Message))
		default:
			b.WriteString(fmt.Sprintf("🔵 **INFO - %s**\n\n%s\n\n", c.Title, c.Message))
		}
	}
	return b.String()
}
