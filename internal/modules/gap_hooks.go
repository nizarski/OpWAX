package modules

import (
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/artifacts"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"
)

func gapRegistryDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.WindowsSearchIndex {
		actions = append(actions,
			action(registryName, "Disable Windows Search (WSearch)", "WSearch", models.ActionDisable),
			action(registryName, "Delete Windows Search index (Windows.edb)", artifacts.WindowsSearchIndexPath(), models.ActionClean),
		)
	}
	if adv.ExecutionRegistries {
		for _, u := range ctx.TargetUsers {
			actions = append(actions,
				action(registryName, "Clear MUICache / RecentApps / AppCompatFlags for "+u.Username, u.Username+`\NTUSER+UsrClass`, models.ActionClean),
			)
		}
	}
	if adv.ShellIconCaches {
		for _, p := range artifacts.ShellCacheGlobPatterns(ctx.TargetUsers) {
			actions = append(actions, action(registryName, "Delete shell icon/thumb cache", p, models.ActionClean))
		}
	}
	if adv.SmartScreenCache {
		for _, p := range artifacts.SmartScreenCachePaths(ctx.TargetUsers) {
			actions = append(actions, action(registryName, "Clear SmartScreen cache", p, models.ActionClean))
		}
	}
	return actions
}

func gapRegistryDisable(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.WindowsSearchIndex {
		a := action(registryName, "Disable Windows Search", "WSearch", models.ActionDisable)
		results = append(results, result(a, util.DisableWindowsSearch()))
	}
	return results
}

func gapRegistryClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.WindowsSearchIndex {
		a := action(registryName, "Delete Windows Search index", artifacts.WindowsSearchIndexPath(), models.ActionClean)
		results = append(results, result(a, util.DeleteWindowsSearchIndex()))
	}
	if adv.ExecutionRegistries {
		a := action(registryName, "Clear MUICache / RecentApps / AppCompatFlags", "execution registry keys", models.ActionClean)
		results = append(results, result(a, util.ClearExecutionRegistryKeys(ctx)))
	}
	if adv.ShellIconCaches {
		a := action(registryName, "Delete shell icon/thumb caches", "Explorer\\*.db", models.ActionClean)
		results = append(results, result(a, util.DeleteShellIconCaches(ctx.TargetUsers)))
	}
	if adv.SmartScreenCache {
		a := action(registryName, "Clear SmartScreen cache", "Safety\\edge\\remote", models.ActionClean)
		results = append(results, result(a, util.DeleteSmartScreenCache(ctx.TargetUsers)))
	}
	return results
}

func gapExecutionDryRun(ctx *models.RunContext) []models.Action {
	if !ctx.Config.Options.Advanced.SyscacheHive {
		return nil
	}
	var actions []models.Action
	for _, p := range artifacts.SyscachePaths() {
		actions = append(actions, action(executionName, "Delete Syscache / RecentFileCache", p, models.ActionClean))
	}
	return actions
}

func gapExecutionClean(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Options.Advanced.SyscacheHive {
		return nil
	}
	a := action(executionName, "Delete Syscache hive", "Syscache.hve", models.ActionClean)
	return []models.Result{result(a, util.DeleteSyscacheHive())}
}

func gapLogsDryRun(ctx *models.RunContext) []models.Action {
	if !ctx.Config.Options.Advanced.TargetedEventChannels {
		return nil
	}
	var actions []models.Action
	for _, ch := range util.TargetedExecutionEventChannels {
		actions = append(actions,
			action(logsName, "Disable execution evidence channel", ch, models.ActionDisable),
			action(logsName, "Clear execution evidence channel", ch, models.ActionClean),
		)
	}
	return actions
}

func gapLogsDisable(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Options.Advanced.TargetedEventChannels {
		return nil
	}
	a := action(logsName, "Disable targeted execution EVTX channels", strings.Join(util.TargetedExecutionEventChannels, "; "), models.ActionDisable)
	return []models.Result{result(a, util.DisableTargetedEventChannels())}
}

func gapLogsClean(ctx *models.RunContext) []models.Result {
	if !ctx.Config.Options.Advanced.TargetedEventChannels {
		return nil
	}
	a := action(logsName, "Clear targeted execution EVTX channels", "TaskScheduler; AppExperience", models.ActionClean)
	return []models.Result{result(a, util.ClearTargetedEventChannels())}
}

func gapNetworkDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.EventTranscript {
		actions = append(actions,
			action(networkName, "Delete EventTranscript.db", artifacts.EventTranscriptPath(), models.ActionClean),
		)
	}
	if adv.DeliveryOptimization {
		actions = append(actions,
			action(networkName, "Clear Delivery Optimization cache", artifacts.DeliveryOptimizationCacheDir(), models.ActionClean),
		)
	}
	if adv.DeveloperTraces {
		for _, u := range ctx.TargetUsers {
			actions = append(actions,
				action(networkName, "Clear developer tool histories for "+u.Username, u.ProfilePath, models.ActionClean),
			)
		}
	}
	return actions
}

func gapNetworkClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.EventTranscript {
		a := action(networkName, "Delete EventTranscript.db", artifacts.EventTranscriptPath(), models.ActionClean)
		results = append(results, result(a, util.DeleteEventTranscriptDB()))
	}
	if adv.DeliveryOptimization {
		a := action(networkName, "Clear Delivery Optimization cache", artifacts.DeliveryOptimizationCacheDir(), models.ActionClean)
		results = append(results, result(a, util.DeleteDeliveryOptimizationCache()))
	}
	if adv.DeveloperTraces {
		a := action(networkName, "Clear developer traces", "git; vscode; npm", models.ActionClean)
		results = append(results, result(a, util.DeleteDeveloperTraces(ctx.TargetUsers)))
	}
	return results
}

func gapPersistenceDryRun(ctx *models.RunContext) []models.Action {
	adv := ctx.Config.Options.Advanced
	var actions []models.Action
	if adv.WERReports {
		actions = append(actions,
			action(persistenceName, "Clear WER report folders", artifacts.WERSystemPath(), models.ActionClean),
		)
	}
	if adv.ServicingLogs {
		for _, p := range artifacts.ServicingLogDirs() {
			actions = append(actions, action(persistenceName, "Clear servicing logs", p, models.ActionClean))
		}
	}
	if adv.PrintSpooler {
		actions = append(actions,
			action(persistenceName, "Clear print spooler jobs", artifacts.PrintSpoolerDir(), models.ActionClean),
		)
	}
	return actions
}

func gapPersistenceClean(ctx *models.RunContext) []models.Result {
	adv := ctx.Config.Options.Advanced
	var results []models.Result
	if adv.WERReports {
		a := action(persistenceName, "Clear WER reports", artifacts.WERSystemPath(), models.ActionClean)
		results = append(results, result(a, util.DeleteWERReports(ctx.TargetUsers)))
	}
	if adv.ServicingLogs {
		a := action(persistenceName, "Clear servicing logs", filepath.Join(ctx.WindowsDir, "Panther"), models.ActionClean)
		results = append(results, result(a, util.DeleteServicingLogs()))
	}
	if adv.PrintSpooler {
		a := action(persistenceName, "Clear print spooler", artifacts.PrintSpoolerDir(), models.ActionClean)
		results = append(results, result(a, util.CleanPrintSpooler()))
	}
	return results
}

// gapCollectorsNeeded reports if extended bootstrap collector kill should run.
func gapCollectorsNeeded(ctx *models.RunContext) bool {
	adv := ctx.Config.Options.Advanced
	if adv.WindowsSearchIndex && ctx.Config.Modules.RegistryHives {
		return true
	}
	if ctx.Config.Modules.NetworkBrowser || adv.EventTranscript {
		return true
	}
	if ctx.Config.Modules.ProgramExecution {
		return true
	}
	return false
}
