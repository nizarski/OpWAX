package artifacts

import (
	"os"
	"path/filepath"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
)

// Paths returns exact filesystem paths targeted by enabled advanced options.
func Paths(ctx *models.RunContext) []string {
	adv := ctx.Config.Options.Advanced
	var paths []string
	add := func(p ...string) {
		for _, s := range p {
			if s != "" {
				paths = append(paths, s)
			}
		}
	}

	if ctx.Config.Modules.NTFSMetadata {
		if adv.DeleteVSSShadows {
			add(`\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy*`)
		}
		if adv.FullVolumeUnallocated {
			for _, d := range ctx.TargetDrives {
				add(d.Root)
			}
		}
		if adv.ScrubBadClusters {
			for _, d := range ctx.TargetDrives {
				add(filepath.Join(d.Root, "$BadClus"))
			}
		}
	}

	for _, u := range ctx.TargetUsers {
		if ctx.Config.Modules.ProgramExecution {
			if adv.PowerShellHistory {
				add(filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"))
			}
			if adv.UserAssist {
				add(u.Username + `\HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\UserAssist\{*}\Count`)
			}
			if adv.RDPCacheEnabled() {
				add(filepath.Join(u.AppDataLocal, "Microsoft", "Terminal Server Client", "Cache"))
			}
		}
		if ctx.Config.Modules.SystemLogs && adv.TimelineActivity {
			add(filepath.Join(u.AppDataLocal, "ConnectedDevicesPlatform"))
		}
		if ctx.Config.Modules.NetworkBrowser {
			if adv.OneDriveEnabled() {
				add(
					filepath.Join(u.AppDataLocal, "Microsoft", "OneDrive", "settings"),
					filepath.Join(u.AppDataLocal, "Microsoft", "OneDrive", "logs", "Common"),
				)
			}
			if adv.CloudSyncEnabled() {
				add(
					filepath.Join(u.AppDataRoaming, "Dropbox"),
					filepath.Join(u.AppDataLocal, "Google", "DriveFS"),
					filepath.Join(u.AppDataLocal, "Packages"),
				)
			}
		}
	}

	if ctx.Config.Modules.PersistenceStorage {
		if adv.WMIEnabled() {
			add(filepath.Join(system.WindowsDirectory(), "System32", "wbem", "Repository", "OBJECTS.DATA"))
		}
		if adv.ScheduledTasksEnabled() {
			add(filepath.Join(system.WindowsDirectory(), "System32", "Tasks"))
		}
		if adv.BITSEnabled() {
			add(filepath.Join(system.ProgramDataDir(), "Microsoft", "Network", "Downloader", "qmgr.db"))
		}
		if adv.AlternateRunKeysEnabled() {
			add(
				`HKLM\Software\Microsoft\Windows\CurrentVersion\RunServices`,
				`HKLM\Software\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`,
			)
		}
		if adv.HyperVEnabled() {
			add(filepath.Join(system.ProgramDataDir(), "Microsoft", "Windows", "Hyper-V", "Snapshots"))
		}
		if adv.WSLEnabled() {
			for _, u := range ctx.TargetUsers {
				add(filepath.Join(u.AppDataLocal, "Packages", "*", "LocalState", "ext4.vhdx"))
			}
		}
	}

	return dedupe(paths)
}

// TimelineDBPaths finds ActivitiesCache.db files for target users.
func TimelineDBPaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		base := filepath.Join(u.AppDataLocal, "ConnectedDevicesPlatform")
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				p := filepath.Join(base, e.Name(), "ActivitiesCache.db")
				if _, err := os.Stat(p); err == nil {
					out = append(out, p)
				}
			}
		}
	}
	return out
}

// RDPCachePaths returns RDP bitmap cache files for users.
func RDPCachePaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		dir := filepath.Join(u.AppDataLocal, "Microsoft", "Terminal Server Client", "Cache")
		matches, _ := filepath.Glob(filepath.Join(dir, "cache*.bin"))
		out = append(out, matches...)
	}
	return out
}

// WSLVHDXPaths lists ext4.vhdx files for installed WSL distros.
func WSLVHDXPaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		packages := filepath.Join(u.AppDataLocal, "Packages")
		entries, err := os.ReadDir(packages)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			vhdx := filepath.Join(packages, e.Name(), "LocalState", "ext4.vhdx")
			if _, err := os.Stat(vhdx); err == nil {
				out = append(out, vhdx)
			}
		}
	}
	return out
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
