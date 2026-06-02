package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"

	"golang.org/x/sys/windows/registry"
)

// TargetedExecutionEventChannels are high-value execution-evidence EVTX channels.
var TargetedExecutionEventChannels = []string{
	"Microsoft-Windows-TaskScheduler/Operational",
	"Microsoft-Windows-Application-Experience/Program-Telemetry/Operational",
	"Microsoft-Windows-Application-Experience/Program-Inventory/Operational",
	"Microsoft-Windows-Application-Experience/Compatibility-Assistant/Operational",
}

// DisableWindowsSearch stops indexing and blocks future search index writes.
func DisableWindowsSearch() error {
	_ = StopAndDisableService("WSearch")
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\WSearch`, "Start", 4)
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\Windows Search`, "AllowIndexingEncryptedStoresOrItems", 0)
	return SetRegDWORD(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\Windows Search`, "PreventIndexingAutomatic", 1)
}

// DeleteWindowsSearchIndex removes the main Windows Search ESE database.
func DeleteWindowsSearchIndex() error {
	base := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Search", "Data", "Applications", "Windows")
	for _, name := range []string{"Windows.edb", "Windows.jfm", "Windows.edb.jfm"} {
		_ = SecureDelete(filepath.Join(base, name))
	}
	_, _ = DeleteDirContents(base)
	return nil
}

// DeleteSyscacheHive removes Syscache and Win7-era RecentFileCache.
func DeleteSyscacheHive() error {
	dir := filepath.Join(os.Getenv("SystemRoot"), "appcompat", "Programs")
	for _, name := range []string{
		"Syscache.hve", "Syscache.hve.LOG1", "Syscache.hve.LOG2",
		"RecentFileCache.bcf",
	} {
		_ = SecureDelete(filepath.Join(dir, name))
	}
	return nil
}

// DeleteEventTranscriptDB removes Diagnosis EventTranscript SQLite DB.
func DeleteEventTranscriptDB() error {
	p := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Diagnosis", "EventTranscript", "EventTranscript.db")
	_ = SecureDelete(p)
	_, _ = DeleteDirContents(filepath.Dir(p))
	return nil
}

// DeleteWERReports clears system and per-user WER report folders.
func DeleteWERReports(users []models.UserProfile) error {
	sys := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "WER")
	_, _ = DeleteDirContents(sys)
	for _, u := range users {
		userWER := filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "WER")
		_, _ = DeleteDirContents(userWER)
	}
	return nil
}

// DeleteShellIconCaches removes Explorer icon/thumb cache DBs for target users.
func DeleteShellIconCaches(users []models.UserProfile) error {
	for _, u := range users {
		explorer := filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "Explorer")
		matches, _ := filepath.Glob(filepath.Join(explorer, "iconcache_*.db"))
		matches2, _ := filepath.Glob(filepath.Join(explorer, "thumbcache_*.db"))
		for _, m := range append(matches, matches2...) {
			_ = SecureDelete(m)
		}
	}
	return nil
}

// ClearExecutionRegistryKeys clears MUICache, RecentApps, AppCompatFlags per user.
func ClearExecutionRegistryKeys(ctx *models.RunContext) error {
	ntuserKeys := []string{
		`Software\Microsoft\Windows\CurrentVersion\Search\RecentApps`,
		`Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Layers`,
		`Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Compatibility Assistant\Persisted`,
	}
	usrClassKeys := []string{
		`Local Settings\Software\Microsoft\Windows\Shell\MuiCache`,
	}
	for _, u := range ctx.TargetUsers {
		for _, k := range ntuserKeys {
			_ = clearUserRegKeyForGap(ctx, u, k)
		}
		for _, k := range usrClassKeys {
			_ = clearUsrClassKeyForGap(u.ProfilePath, k)
		}
	}
	return nil
}

func clearUserRegKeyForGap(ctx *models.RunContext, u models.UserProfile, subKey string) error {
	if strings.EqualFold(u.Username, ctx.CurrentUser) {
		return DeleteRegKey(registry.CURRENT_USER, subKey)
	}
	tempKey := "OPWAX_GAP_" + u.SID
	_ = RunHidden("reg", "unload", `HKU\`+tempKey)
	if err := RunHidden("reg", "load", `HKU\`+tempKey, u.NTUserPath); err != nil {
		return err
	}
	defer RunHidden("reg", "unload", `HKU\`+tempKey)
	return RunHidden("reg", "delete", `HKU\`+tempKey+`\`+subKey, "/f")
}

func clearUsrClassKeyForGap(profilePath, subKey string) error {
	hivePath := filepath.Join(profilePath, "AppData", "Local", "Microsoft", "Windows", "UsrClass.dat")
	tempKey := "OPWAX_GAPU_" + filepath.Base(profilePath)
	_ = RunHidden("reg", "unload", `HKU\`+tempKey)
	if err := RunHidden("reg", "load", `HKU\`+tempKey, hivePath); err != nil {
		return err
	}
	defer RunHidden("reg", "unload", `HKU\`+tempKey)
	return RunHidden("reg", "delete", `HKU\`+tempKey+`\`+subKey, "/f")
}

// DeleteServicingLogs clears Panther, CBS, and Windows Update trace logs.
func DeleteServicingLogs() error {
	win := os.Getenv("SystemRoot")
	paths := []string{
		filepath.Join(win, "Panther"),
		filepath.Join(win, "Logs", "CBS"),
		filepath.Join(win, "Logs", "WindowsUpdate"),
		filepath.Join(win, "SoftwareDistribution", "DataStore", "Logs"),
	}
	for _, p := range paths {
		_, _ = DeleteDirContents(p)
	}
	return nil
}

// CleanPrintSpooler stops spooler and clears job files.
func CleanPrintSpooler() error {
	_ = StopService("Spooler")
	defer StartService("Spooler")
	dir := filepath.Join(os.Getenv("SystemRoot"), "System32", "spool", "PRINTERS")
	_, _ = DeleteDirContents(dir)
	return nil
}

// DeleteDeliveryOptimizationCache clears DO peer cache.
func DeleteDeliveryOptimizationCache() error {
	dir := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "DeliveryOptimization", "Cache")
	_, _ = DeleteDirContents(dir)
	return nil
}

// DeleteSmartScreenCache clears local SmartScreen edge cache.
func DeleteSmartScreenCache(users []models.UserProfile) error {
	for _, u := range users {
		dir := filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "Safety", "edge", "remote")
		_, _ = DeleteDirContents(dir)
	}
	return nil
}

// DeleteDeveloperTraces removes common dev tool history caches.
func DeleteDeveloperTraces(users []models.UserProfile) error {
	for _, u := range users {
		home := u.ProfilePath
		paths := []string{
			filepath.Join(home, ".gitconfig"),
			filepath.Join(home, ".lesshst"),
			filepath.Join(home, ".python_history"),
			filepath.Join(home, ".node_repl_history"),
			filepath.Join(home, ".npm", "_logs"),
			filepath.Join(home, "AppData", "Roaming", "Code", "User", "History"),
			filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "History"),
			filepath.Join(home, ".cursor"),
		}
		for _, p := range paths {
			if info, err := os.Stat(p); err == nil {
				if info.IsDir() {
					_, _ = DeleteDirContents(p)
				} else {
					_ = SecureDelete(p)
				}
			}
		}
		gitHist := filepath.Join(home, ".bash_history")
		_ = SecureDelete(gitHist)
	}
	return nil
}

// DisableTargetedEventChannels disables high-value execution EVTX channels.
func DisableTargetedEventChannels() error {
	for _, ch := range TargetedExecutionEventChannels {
		_ = DisableEventLogChannel(ch)
	}
	return nil
}

// ClearTargetedEventChannels clears high-value execution EVTX channels.
func ClearTargetedEventChannels() error {
	for _, ch := range TargetedExecutionEventChannels {
		_ = ClearEventLogChannel(ch)
	}
	return nil
}

// BeginFocusedCleanup stops Explorer to reduce shell artifacts during cleanup.
func BeginFocusedCleanup() func() {
	_ = RunHidden("taskkill", "/F", "/IM", "explorer.exe")
	return func() {
		_ = RunHidden("cmd", "/c", "start", "explorer.exe")
	}
}

// CollectorKillDisable stops WSearch, DiagTrack, DPS, and SysMain (idempotent).
func CollectorKillDisable() error {
	for _, svc := range []string{"WSearch", "DiagTrack", "DPS", "SysMain"} {
		_ = StopAndDisableService(svc)
	}
	return nil
}
