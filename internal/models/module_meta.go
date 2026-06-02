package models

// ModuleCanRequireReboot reports modules that may need a reboot when used.
// Shown in the GUI before the user enables them.
func ModuleCanRequireReboot(moduleID string) bool {
	switch moduleID {
	case "volatile_memory", "ntfs_metadata", "persistence_storage":
		return true
	default:
		return false
	}
}

// ModuleNeedsRebootHint reports whether a module may require reboot for full cleanup
// given the current config (e.g. during execution progress).
func ModuleNeedsRebootHint(moduleID string, cfg Config) bool {
	switch moduleID {
	case "volatile_memory":
		return cfg.Modules.VolatileMemory && cfg.Options.LSASSScrub
	case "ntfs_metadata":
		return cfg.Modules.NTFSMetadata && cfg.Options.LogFileResetOnReboot
	case "persistence_storage":
		return cfg.Modules.PersistenceStorage
	default:
		return false
	}
}

// ModuleRebootTooltip returns a short explanation for the reboot indicator.
func ModuleRebootTooltip(moduleID string, cfg Config) string {
	switch moduleID {
	case "volatile_memory":
		return "LSASS credential caches cleared live; optional reboot clears remaining RAM"
	case "ntfs_metadata":
		return "$LogFile reset on system drive runs at next reboot (chkdsk /F)"
	case "persistence_storage":
		return "Pagefile, hiberfil, and locked dumps often need a reboot to clear"
	default:
		return "Reboot may be required"
	}
}
