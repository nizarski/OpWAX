package modules

import (
	"os"
	"path/filepath"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

const executionName = "program_execution"

// ExecutionModule handles Prefetch, SuperFetch, Amcache.
type ExecutionModule struct{}

func (m *ExecutionModule) Name() string { return executionName }

func (m *ExecutionModule) DryRun(ctx *models.RunContext) []models.Action {
	prefetch := filepath.Join(system.WindowsDirectory(), "Prefetch")
	actions := []models.Action{
		action(executionName, "Disable Prefetcher", "PrefetchParameters", models.ActionDisable),
		action(executionName, "Stop and disable SysMain service", "SysMain", models.ActionDisable),
		action(executionName, "Stop and disable PcaSvc (Amcache)", "PcaSvc", models.ActionDisable),
		action(executionName, "Delete Prefetch *.pf files", prefetch, models.ActionClean),
		action(executionName, "Delete SuperFetch Ag*.db", prefetch, models.ActionClean),
		action(executionName, "Delete Amcache.hve", "appcompat\\Programs", models.ActionClean),
	}
	actions = append(actions, advancedExecutionDryRun(ctx)...)
	actions = append(actions, gapExecutionDryRun(ctx)...)
	return actions}

func (m *ExecutionModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	prefetchPath := `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management\PrefetchParameters`
	a1 := action(executionName, "Disable Prefetcher", prefetchPath, models.ActionDisable)
	err := util.SetRegDWORD(registry.LOCAL_MACHINE, prefetchPath, "EnablePrefetcher", 0)
	_ = util.SetRegDWORD(registry.LOCAL_MACHINE, prefetchPath, "EnableSuperfetch", 0)
	results = append(results, result(a1, err))

	a2 := action(executionName, "Disable SysMain", "SysMain", models.ActionDisable)
	results = append(results, result(a2, util.StopAndDisableService("SysMain")))

	a3 := action(executionName, "Disable PcaSvc", "PcaSvc", models.ActionDisable)
	results = append(results, result(a3, util.StopAndDisableService("PcaSvc")))

	results = append(results, advancedExecutionDisable(ctx)...)
	return results
}

func (m *ExecutionModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result
	prefetch := filepath.Join(system.WindowsDirectory(), "Prefetch")

	a1 := action(executionName, "Delete *.pf", prefetch, models.ActionClean)
	count, err := util.DeleteFilesGlob(prefetch, "*.pf")
	results = append(results, result(a1, err))
	_ = count

	a2 := action(executionName, "Delete Ag*.db", prefetch, models.ActionClean)
	_, _ = util.DeleteFilesGlob(prefetch, "Ag*.db")
	entries, _ := os.ReadDir(prefetch)
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".db" {
			_ = os.Remove(filepath.Join(prefetch, e.Name()))
		}
	}
	results = append(results, result(a2, nil))

	amcacheDir := filepath.Join(system.WindowsDirectory(), "appcompat", "Programs")
	a3 := action(executionName, "Delete Amcache", amcacheDir, models.ActionClean)
	for _, name := range []string{"Amcache.hve", "Amcache.hve.LOG1", "Amcache.hve.LOG2", "Amcache.hve.bak"} {
		_ = os.Remove(filepath.Join(amcacheDir, name))
	}
	results = append(results, result(a3, nil))

	results = append(results, advancedExecutionClean(ctx)...)
	results = append(results, gapExecutionClean(ctx)...)
	return results}
