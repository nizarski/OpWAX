package modules

import (
	"github.com/opwax/opwax/internal/models"
)

// Module defines the disable-then-clean interface for each artifact category.
type Module interface {
	Name() string
	DryRun(ctx *models.RunContext) []models.Action
	Disable(ctx *models.RunContext) []models.Result
	Clean(ctx *models.RunContext) []models.Result
}

// EnabledModules returns modules based on config toggles.
func EnabledModules(cfg models.ModuleConfig) []Module {
	var mods []Module
	if cfg.VolatileMemory {
		mods = append(mods, &VolatileModule{})
	}
	if cfg.RegistryHives {
		mods = append(mods, &RegistryModule{})
	}
	if cfg.NTFSMetadata {
		mods = append(mods, &NTFSModule{})
	}
	if cfg.ProgramExecution {
		mods = append(mods, &ExecutionModule{})
	}
	if cfg.SystemLogs {
		mods = append(mods, &LogsModule{})
	}
	if cfg.PersistenceStorage {
		mods = append(mods, &PersistenceModule{})
	}
	if cfg.NetworkBrowser {
		mods = append(mods, &NetworkModule{})
	}
	return mods
}

func action(module, desc, target string, kind models.ActionKind) models.Action {
	return models.Action{
		Module:      module,
		Kind:        kind,
		Description: desc,
		Target:      target,
	}
}

func result(a models.Action, err error) models.Result {
	r := models.Result{Action: a, Success: err == nil}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}
