package modules

import (
	"path/filepath"
	"strconv"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"
)

const logsName = "system_logs"

// LogsModule handles event logs, jump lists, and recent shortcuts.
type LogsModule struct{}

func (m *LogsModule) Name() string { return logsName }

func (m *LogsModule) DryRun(ctx *models.RunContext) []models.Action {
	logsDir := filepath.Join(system.WindowsDirectory(), "System32", "Winevt", "Logs")
	var actions []models.Action
	actions = append(actions,
		action(logsName, "Clear audit policies", "auditpol", models.ActionDisable),
		action(logsName, "Stop and disable EventLog + collector services", "EventLog; Wecsvc; WdiSystemHost", models.ActionDisable),
		action(logsName, "Disable all event log channels", "wevtutil", models.ActionDisable),
		action(logsName, "Clear all event log channels", "wevtutil", models.ActionClean),
		action(logsName, "Purge entire Winevt\\Logs directory", logsDir, models.ActionClean),
	)
	for _, u := range ctx.TargetUsers {
		recent := filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent")
		actions = append(actions,
			action(logsName, "Delete jump lists for "+u.Username, recent, models.ActionClean),
			action(logsName, "Delete .lnk shortcuts for "+u.Username, recent, models.ActionClean),
		)
	}
	actions = append(actions, advancedLogsDryRun(ctx)...)
	actions = append(actions, gapLogsDryRun(ctx)...)
	return actions}

func (m *LogsModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	a1 := action(logsName, "Clear audit policy", "auditpol", models.ActionDisable)
	results = append(results, result(a1, util.RunHidden("auditpol", "/clear", "/y")))

	a2 := action(logsName, "Disable event log stack", "EventLog; Wecsvc", models.ActionDisable)
	results = append(results, result(a2, util.DisableEventLogStack()))

	channels, _ := util.ListEventLogChannels()
	for _, ch := range channels {
		a := action(logsName, "Disable channel "+ch, ch, models.ActionDisable)
		results = append(results, result(a, util.DisableEventLogChannel(ch)))
	}

	results = append(results, advancedLogsDisable(ctx)...)
	results = append(results, gapLogsDisable(ctx)...)
	return results}

func (m *LogsModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	channels, _ := util.ListEventLogChannels()
	for _, ch := range channels {
		a := action(logsName, "Clear channel "+ch, ch, models.ActionClean)
		results = append(results, result(a, util.ClearEventLogChannel(ch)))
	}

	logsDir := filepath.Join(system.WindowsDirectory(), "System32", "Winevt", "Logs")
	a := action(logsName, "Purge Winevt\\Logs", logsDir, models.ActionClean)
	count, err := util.PurgeWinevtLogsDirectory(logsDir)
	if err == nil {
		a.Description = "Purge Winevt\\Logs (" + strconv.Itoa(count) + " file(s))"
	}
	results = append(results, result(a, err))

	for _, u := range ctx.TargetUsers {
		recent := filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent")
		for _, sub := range []string{"AutomaticDestinations", "CustomDestinations"} {
			dir := filepath.Join(recent, sub)
			a := action(logsName, "Clear "+sub+" for "+u.Username, dir, models.ActionClean)
			_, err := util.DeleteDirContents(dir)
			results = append(results, result(a, err))
		}
		a2 := action(logsName, "Delete .lnk for "+u.Username, recent, models.ActionClean)
		_, _ = util.DeleteFilesGlob(recent, "*.lnk")
		results = append(results, result(a2, nil))
	}

	results = append(results, advancedLogsClean(ctx)...)
	results = append(results, gapLogsClean(ctx)...)
	return results}
