package modules

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opwax/opwax/internal/artifacts"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/scheduler"
	"github.com/opwax/opwax/internal/util"
)

func advancedNTFSDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.DeleteVSSShadows {
		actions = append(actions,
			action(ntfsName, "Stop Volume Shadow Copy service", "VSS", models.ActionDisable),
			action(ntfsName, "Delete all volume shadow copies", "vssadmin delete shadows /all", models.ActionClean),
		)
	}
	for _, d := range ctx.TargetDrives {
		if adv.FullVolumeUnallocated {
			actions = append(actions,
				action(ntfsName, "1-pass wipe unallocated clusters (cipher /w)", d.Root, models.ActionSecure),
			)
		}
		if adv.ScrubBadClusters {
			actions = append(actions,
				action(ntfsName, "Schedule bad-cluster verification (chkdsk /B)", d.Root+"$BadClust", models.ActionSecure),
			)
		}
	}
	if adv.WipeFileSlack {
		scope := "artifact paths"
		if adv.WipeFileSlackAllFiles {
			scope = "all files in target user profiles"
		}
		actions = append(actions,
			action(ntfsName, "Wipe file slack space ("+scope+")", "target files", models.ActionSecure),
		)
	}
	return actions
}

func advancedNTFSDisable(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.DeleteVSSShadows {
		a := action(ntfsName, "Stop VSS service", "VSS", models.ActionDisable)
		results = append(results, result(a, util.StopService("VSS")))
	}
	return results
}

func advancedNTFSClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.DeleteVSSShadows {
		a := action(ntfsName, "Delete all shadow copies", "vssadmin", models.ActionClean)
		results = append(results, result(a, util.DeleteAllVolumeShadowCopies()))
	}
	for _, d := range ctx.TargetDrives {
		if adv.FullVolumeUnallocated {
			a := action(ntfsName, "Cipher wipe free space "+d.Letter, d.Root, models.ActionSecure)
			results = append(results, result(a, util.CipherWipeFreeSpace(d.Letter)))
		}
		if adv.ScrubBadClusters {
			a := action(ntfsName, "Bad cluster scan "+d.Letter, d.Root, models.ActionSecure)
			results = append(results, result(a, util.ScheduleBadClusterScan(d.Letter)))
		}
	}
	if adv.WipeFileSlack {
		targets := util.CollectSlackTargets(ctx, adv.WipeFileSlackAllFiles)
		a := action(ntfsName, "Wipe slack on files", strings.Join(targets, "; "), models.ActionSecure)
		n, err := util.WipeFileSlackInPaths(targets)
		a.Description = "Wipe file slack (" + strconv.Itoa(n) + " file(s))"
		results = append(results, result(a, err))
	}
	return results
}

func advancedExecutionDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	for _, u := range ctx.TargetUsers {
		if adv.PowerShellHistory {
			p := filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
			actions = append(actions,
				action(executionName, "Disable PSReadLine history saving", p, models.ActionDisable),
				action(executionName, "Delete PSReadLine history", p, models.ActionClean),
			)
		}
		if adv.UserAssist {
			actions = append(actions,
				action(executionName, "Clear UserAssist execution counts", u.Username+`\UserAssist\{GUID}`, models.ActionClean),
			)
		}
		if adv.RDPCacheEnabled() {
			dir := filepath.Join(u.AppDataLocal, "Microsoft", "Terminal Server Client", "Cache")
			for _, p := range artifacts.RDPCachePaths([]models.UserProfile{u}) {
				actions = append(actions, action(executionName, "Delete RDP session cache", p, models.ActionClean))
			}
			if len(artifacts.RDPCachePaths([]models.UserProfile{u})) == 0 {
				actions = append(actions, action(executionName, "Delete RDP session cache (if present)", dir+"\\cache*.bin", models.ActionClean))
			}
		}
	}
	if adv.WindowsRecall {
		for _, p := range artifacts.RecallCapturePaths(ctx.TargetUsers) {
			actions = append(actions,
				action(executionName, "Disable Windows Recall / CoreAI capture", p, models.ActionDisable),
				action(executionName, "Delete Recall capture database", p, models.ActionClean),
			)
		}
	}
	if adv.PCAAppCompatLogs {
		for _, p := range artifacts.PCALaunchLogPaths() {
			actions = append(actions, action(executionName, "Delete PCA launch log", p, models.ActionClean))
		}
	}
	if adv.NotepadTabCache {
		for _, p := range artifacts.NotepadTabStatePaths(ctx.TargetUsers) {
			actions = append(actions, action(executionName, "Delete Notepad tab/draft cache", p, models.ActionClean))
		}
	}
	return actions
}

func advancedExecutionDisable(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.PowerShellHistory {
		a := action(executionName, "Disable PSReadLine history", "PSReadLine", models.ActionDisable)
		results = append(results, result(a, util.DisablePSReadLineHistory()))
	}
	if adv.WindowsRecall {
		a := action(executionName, "Disable Windows Recall", "CoreAIPlatform", models.ActionDisable)
		results = append(results, result(a, util.DisableWindowsRecall()))
	}
	return results
}

func advancedExecutionClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	for _, u := range ctx.TargetUsers {
		if adv.PowerShellHistory {
			p := filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
			a := action(executionName, "Delete PSReadLine history", p, models.ActionClean)
			results = append(results, result(a, util.SecureDelete(p)))
		}
	}
	if adv.UserAssist {
		a := action(executionName, "Clear UserAssist keys", "UserAssist", models.ActionClean)
		results = append(results, result(a, util.ClearUserAssist(ctx)))
	}
	if adv.RDPCacheEnabled() {
		for _, p := range artifacts.RDPCachePaths(ctx.TargetUsers) {
			a := action(executionName, "Delete RDP cache", p, models.ActionClean)
			results = append(results, result(a, util.SecureDelete(p)))
		}
	}
	if adv.WindowsRecall {
		a := action(executionName, "Delete Recall captures", "CoreAIPlatform\\Capture", models.ActionClean)
		results = append(results, result(a, util.DeleteRecallCaptures(ctx.TargetUsers)))
	}
	if adv.PCAAppCompatLogs {
		a := action(executionName, "Delete PCA launch logs", "appcompat\\Programs", models.ActionClean)
		results = append(results, result(a, util.CleanPCALaunchLogs()))
	}
	if adv.NotepadTabCache {
		a := action(executionName, "Delete Notepad tab cache", "Notepad TabState", models.ActionClean)
		results = append(results, result(a, util.CleanNotepadTabState(ctx.TargetUsers)))
	}
	return results
}

func advancedRegistryDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.OfficeTrustRecords {
		for _, u := range ctx.TargetUsers {
			actions = append(actions,
				action(registryName, "Clear Office Trusted Documents", u.Username+`\HKCU\Software\Microsoft\Office\*\Security\Trusted Documents\TrustRecords`, models.ActionClean),
			)
		}
	}
	if adv.SysinternalsEULA {
		actions = append(actions,
			action(registryName, "Clear Sysinternals EULA acceptance keys", `HKCU\Software\Sysinternals`, models.ActionClean),
		)
	}
	return actions
}

func advancedRegistryDisable(_ *models.RunContext) []models.Result {
	return nil
}

func advancedRegistryClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.OfficeTrustRecords {
		a := action(registryName, "Clear Office TrustRecords", "Office", models.ActionClean)
		results = append(results, result(a, util.ClearOfficeTrustRecords(ctx)))
	}
	if adv.SysinternalsEULA {
		a := action(registryName, "Clear Sysinternals EULA keys", "Sysinternals", models.ActionClean)
		results = append(results, result(a, util.ClearSysinternalsEULAKeys()))
	}
	return results
}

func advancedLogsDryRun(ctx *models.RunContext) []models.Action {
	if !ctx.Config.Options.Advanced.TimelineActivity {
		return nil
	}
	var actions []models.Action
	for _, p := range artifacts.TimelineDBPaths(ctx.TargetUsers) {
		actions = append(actions,
			action(logsName, "Disable Windows Timeline / Activity History", p, models.ActionDisable),
			action(logsName, "Delete ActivitiesCache.db", p, models.ActionClean),
		)
	}
	if len(actions) == 0 {
		actions = append(actions,
			action(logsName, "Disable Windows Timeline / Activity History", "ConnectedDevicesPlatform", models.ActionDisable),
		)
	}
	return actions
}

func advancedLogsDisable(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Options.Advanced.TimelineActivity {
		return nil
	}
	a := action(logsName, "Disable Timeline activity feed", "cbdhsvc", models.ActionDisable)
	return []models.Result{result(a, util.DisableTimelineActivity())}
}

func advancedLogsClean(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Options.Advanced.TimelineActivity {
		return nil
	}
	var results []models.Result
	for _, p := range artifacts.TimelineDBPaths(ctx.TargetUsers) {
		a := action(logsName, "Delete Timeline database", p, models.ActionClean)
		results = append(results, result(a, util.SecureDelete(p)))
	}
	return results
}

func advancedPersistenceDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.WMIEnabled() {
		p := filepath.Join(ctx.WindowsDir, "System32", "wbem", "Repository", "OBJECTS.DATA")
		actions = append(actions,
			action(persistenceName, "Stop WMI service", "winmgmt", models.ActionDisable),
			action(persistenceName, "Reset WMI repository", p, models.ActionClean),
		)
	}
	if adv.ScheduledTasksEnabled() {
		root := filepath.Join(ctx.WindowsDir, "System32", "Tasks")
		actions = append(actions,
			action(persistenceName, "Delete scheduled tasks ("+adv.ScheduledTasks+")", root, models.ActionClean),
		)
	}
	if adv.BITSEnabled() {
		p := filepath.Join(ctx.ProgramData, "Microsoft", "Network", "Downloader", "qmgr.db")
		actions = append(actions,
			action(persistenceName, "Stop BITS ("+adv.BITS+")", "bits", models.ActionDisable),
			action(persistenceName, "Clear BITS transfer database", p, models.ActionClean),
		)
	}
	if adv.AlternateRunKeysEnabled() {
		actions = append(actions,
			action(persistenceName, "Disable alternate Run/RunServices keys", "RunServices", models.ActionDisable),
			action(persistenceName, "Clear alternate Run/RunServices values", "HKLM\\...\\RunServices", models.ActionClean),
		)
	}
	if adv.HyperVEnabled() {
		p := filepath.Join(ctx.ProgramData, "Microsoft", "Windows", "Hyper-V", "Snapshots")
		actions = append(actions, action(persistenceName, "Delete Hyper-V snapshots", p, models.ActionClean))
	}
	if adv.WSLEnabled() {
		for _, p := range artifacts.WSLVHDXPaths(ctx.TargetUsers) {
			if adv.WSL == "delete_vhdx" {
				actions = append(actions, action(persistenceName, "Secure-delete WSL distro disk", p, models.ActionSecure))
			} else {
				actions = append(actions, action(persistenceName, "Shutdown WSL ("+adv.WSL+")", p, models.ActionDisable))
			}
		}
	}
	return actions
}

func advancedPersistenceDisable(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.WMIEnabled() {
		a := action(persistenceName, "Stop winmgmt", "winmgmt", models.ActionDisable)
		results = append(results, result(a, util.StopService("winmgmt")))
	}
	if adv.BITSEnabled() {
		a := action(persistenceName, "Stop BITS", "bits", models.ActionDisable)
		results = append(results, result(a, util.StopService("bits")))
	}
	if adv.AlternateRunKeysEnabled() && adv.AlternateRunKeys == "clean_disable" {
		a := action(persistenceName, "Disable alternate Run keys policy", "Policies\\Explorer", models.ActionDisable)
		results = append(results, result(a, util.DisableAlternateRunKeysPolicy()))
	}
	if adv.WSLEnabled() && adv.WSL == "logs" {
		a := action(persistenceName, "Shutdown WSL", "wsl", models.ActionDisable)
		results = append(results, result(a, util.RunHidden("wsl", "--shutdown")))
	}
	return results
}

func advancedPersistenceClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.WMIEnabled() {
		a := action(persistenceName, "Reset WMI repository", "wbem\\Repository", models.ActionClean)
		results = append(results, result(a, util.ResetWMIRepository()))
	}
	if adv.ScheduledTasksEnabled() {
		root := filepath.Join(ctx.WindowsDir, "System32", "Tasks")
		a := action(persistenceName, "Remove scheduled task files", root, models.ActionClean)
		results = append(results, result(a, util.CleanScheduledTasks(adv.ScheduledTasks, scheduler.TaskName())))
	}
	if adv.BITSEnabled() {
		p := filepath.Join(ctx.ProgramData, "Microsoft", "Network", "Downloader", "qmgr.db")
		a := action(persistenceName, "Clear BITS queue DB", p, models.ActionClean)
		results = append(results, result(a, util.CleanBITS(adv.BITS)))
	}
	if adv.AlternateRunKeysEnabled() {
		a := action(persistenceName, "Clear alternate Run keys", "RunServices", models.ActionClean)
		results = append(results, result(a, util.CleanAlternateRunKeys(adv.AlternateRunKeys)))
	}
	if adv.HyperVEnabled() {
		p := filepath.Join(ctx.ProgramData, "Microsoft", "Windows", "Hyper-V", "Snapshots")
		a := action(persistenceName, "Delete Hyper-V snapshots", p, models.ActionClean)
		results = append(results, result(a, util.CleanHyperVSnapshots()))
	}
	if adv.WSLEnabled() {
		a := action(persistenceName, "WSL cleanup ("+adv.WSL+")", "Packages\\*\\ext4.vhdx", models.ActionClean)
		results = append(results, result(a, util.CleanWSL(adv.WSL, ctx.TargetUsers)))
	}
	return results
}

func advancedNetworkDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	for _, u := range ctx.TargetUsers {
		if adv.OneDriveEnabled() {
			settings := filepath.Join(u.AppDataLocal, "Microsoft", "OneDrive", "settings")
			logs := filepath.Join(u.AppDataLocal, "Microsoft", "OneDrive", "logs", "Common")
			actions = append(actions,
				action(networkName, "Disable OneDrive sync ("+adv.OneDrive+")", settings, models.ActionDisable),
				action(networkName, "Clear OneDrive metadata/logs", settings+"; "+logs, models.ActionClean),
			)
		}
	}
	if adv.CloudSyncEnabled() {
		actions = append(actions,
			action(networkName, "Disable cloud sync policies", "Policies", models.ActionDisable),
			action(networkName, "Clear Dropbox/Google Drive caches", "AppData", models.ActionClean),
		)
	}
	if adv.OutlookEnabled() {
		for _, p := range artifacts.OutlookPaths(ctx.TargetUsers) {
			actions = append(actions,
				action(networkName, "Clean Outlook data ("+adv.Outlook+")", p, models.ActionClean),
			)
		}
	}
	if adv.TeamsEnabled() {
		for _, p := range artifacts.TeamsPaths(ctx.TargetUsers) {
			actions = append(actions,
				action(networkName, "Clean Teams data ("+adv.Teams+")", p, models.ActionClean),
			)
		}
	}
	return actions
}

func advancedNetworkDisable(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.OneDriveEnabled() && adv.OneDrive == "full" {
		a := action(networkName, "Disable OneDrive sync policy", "OneDrive", models.ActionDisable)
		results = append(results, result(a, util.DisableOneDriveSync()))
	}
	if adv.CloudSyncEnabled() {
		a := action(networkName, "Disable Google Drive policy", "Drive", models.ActionDisable)
		results = append(results, result(a, util.DisableCloudSyncPolicies()))
	}
	return results
}

func advancedNetworkClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.OneDriveEnabled() {
		a := action(networkName, "Clean OneDrive artifacts", "OneDrive", models.ActionClean)
		results = append(results, result(a, util.CleanOneDrive(adv.OneDrive, ctx.TargetUsers)))
	}
	if adv.CloudSyncEnabled() {
		a := action(networkName, "Clean cloud sync caches", "Dropbox/DriveFS", models.ActionClean)
		results = append(results, result(a, util.CleanCloudSync(ctx.TargetUsers)))
	}
	if adv.OutlookEnabled() {
		a := action(networkName, "Clean Outlook ("+adv.Outlook+")", "Outlook", models.ActionClean)
		results = append(results, result(a, util.CleanOutlookArtifacts(adv.Outlook, ctx.TargetUsers)))
	}
	if adv.TeamsEnabled() {
		a := action(networkName, "Clean Teams ("+adv.Teams+")", "Teams", models.ActionClean)
		results = append(results, result(a, util.CleanTeamsArtifacts(adv.Teams, ctx.TargetUsers)))
	}
	return results
}
