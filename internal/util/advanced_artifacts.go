package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"

	"golang.org/x/sys/windows/registry"
)

// --- NTFS / volume ---

func DeleteAllVolumeShadowCopies() error {
	return RunHidden("vssadmin", "delete", "shadows", "/all", "/quiet")
}

func CipherWipeFreeSpace(driveLetter string) error {
	letter := strings.TrimSuffix(strings.TrimSpace(driveLetter), `\`)
	if letter == "" {
		return fmt.Errorf("invalid drive letter")
	}
	if !strings.HasSuffix(letter, ":") {
		letter += ":"
	}
	return RunHidden("cipher", "/w:"+letter)
}

func ScheduleBadClusterScan(driveLetter string) error {
	letter := strings.TrimSuffix(strings.TrimSpace(driveLetter), `\`)
	if !strings.HasSuffix(letter, ":") {
		letter += ":"
	}
	if IsSystemDrive(letter) {
		return RunHidden("chkdsk", letter, "/F", "/R", "/B")
	}
	return RunHidden("chkdsk", letter, "/F", "/B")
}

func WipeFileSlackInPaths(paths []string) (int, error) {
	n := 0
	for _, p := range paths {
		if err := wipeFileSlack(p); err == nil {
			n++
		}
	}
	return n, nil
}

func wipeFileSlack(path string) error {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return err
	}
	// Best-effort: extend file to cluster boundary with zeros then truncate.
	const cluster = 4096
	rem := cluster - int(info.Size()%cluster)
	if rem == cluster {
		return nil
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, rem)
	if _, err := f.WriteAt(buf, info.Size()); err != nil {
		return err
	}
	return f.Truncate(info.Size())
}

func CollectSlackTargets(ctx *models.RunContext, allFiles bool) []string {
	if allFiles {
		var files []string
		for _, u := range ctx.TargetUsers {
			walkCollectFiles(u.ProfilePath, &files, 5000)
		}
		return files
	}
	var paths []string
	for _, u := range ctx.TargetUsers {
		paths = append(paths,
			filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"),
		)
	}
	return paths
}

func walkCollectFiles(root string, out *[]string, limit int) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || len(*out) >= limit {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Size() > 0 {
			*out = append(*out, path)
		}
		return nil
	})
}

// --- Execution ---

func DisablePSReadLineHistory() error {
	_ = SetRegDWORD(registry.CURRENT_USER, `Software\Microsoft\PSReadLine`, "MaximumHistoryCount", 0)
	_ = SetRegDWORD(registry.CURRENT_USER, `Software\Policies\Microsoft\Windows\PowerShell\PSReadLine`, "MaximumHistoryCount", 0)
	return SetRegDWORD(registry.LOCAL_MACHINE, `Software\Policies\Microsoft\Windows\PowerShell\PSReadLine`, "MaximumHistoryCount", 0)
}

func DisableTimelineActivity() error {
	_ = SetRegDWORD(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Privacy`, "PublishUserActivities", 0)
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `Software\Policies\Microsoft\Windows\System`, "PublishUserActivities", 0)
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `Software\Policies\Microsoft\Windows\System`, "EnableActivityFeed", 0)
	return StopAndDisableService("cbdhsvc")
}

func ClearUserAssist(ctx *models.RunContext) error {
	base := `Software\Microsoft\Windows\CurrentVersion\Explorer\UserAssist`
	for _, u := range ctx.TargetUsers {
		if strings.EqualFold(u.Username, ctx.CurrentUser) {
			if err := deleteUserAssistGUIDs(registry.CURRENT_USER, base); err != nil {
				return err
			}
			continue
		}
		tempKey := "OPWAX_UA_" + u.SID
		_ = RunHidden("reg", "unload", `HKU\`+tempKey)
		if err := RunHidden("reg", "load", `HKU\`+tempKey, u.NTUserPath); err != nil {
			continue
		}
		_ = deleteUserAssistGUIDsAt(`HKU\` + tempKey + `\` + base)
		_ = RunHidden("reg", "unload", `HKU\`+tempKey)
	}
	return nil
}

func deleteUserAssistGUIDs(root registry.Key, base string) error {
	key, err := registry.OpenKey(root, base, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	defer key.Close()
	guids, _ := key.ReadSubKeyNames(0)
	for _, g := range guids {
		guidPath := base + `\` + g
		gk, err := registry.OpenKey(root, guidPath, registry.SET_VALUE)
		if err != nil {
			continue
		}
		vals, _ := gk.ReadValueNames(0)
		for _, v := range vals {
			_ = gk.DeleteValue(v)
		}
		gk.Close()
		_ = DeleteRegKey(root, guidPath+`\Count`)
	}
	return nil
}

func deleteUserAssistGUIDsAt(fullKey string) error {
	out, err := RunHiddenOutput("reg", "query", fullKey)
	if err != nil && !strings.Contains(out, "ERROR") {
		return nil
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HKEY_") && strings.Count(line, `\`) > strings.Count(fullKey, `\`) {
			_ = RunHidden("reg", "delete", line, "/f")
		}
	}
	for _, sub := range []string{`{CEBFF5CD-ACE2-4F4F-86AC-176346B0F966}`, `{F4E57C4B-2036-45F0-A9AB-443BCFE0D0F0}`} {
		_ = RunHidden("reg", "delete", fullKey+`\`+sub, "/f")
	}
	guidsOut, _ := RunHiddenOutput("reg", "query", fullKey)
	for _, line := range strings.Split(guidsOut, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "HKEY") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				_ = RunHidden("reg", "delete", parts[len(parts)-1], "/f")
			}
		}
	}
	return nil
}

// --- Persistence ---

func ResetWMIRepository() error {
	_ = StopService("winmgmt")
	path := filepath.Join(os.Getenv("SystemRoot"), "System32", "wbem", "Repository")
	entries, _ := os.ReadDir(path)
	for _, e := range entries {
		if !e.IsDir() {
			_ = os.Remove(filepath.Join(path, e.Name()))
		}
	}
	_ = RunHidden("winmgmt", "/salvagerepository")
	return StartService("winmgmt")
}

func CleanScheduledTasks(mode, excludeTaskName string) error {
	if mode == "" || mode == "off" {
		return nil
	}
	root := filepath.Join(os.Getenv("SystemRoot"), "System32", "Tasks")
	opwax := strings.ToLower(excludeTaskName)
	var remove []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		relLower := strings.ToLower(filepath.ToSlash(rel))
		switch mode {
		case "user":
			if strings.HasPrefix(relLower, "microsoft/windows/") {
				return nil
			}
		case "all_except_opwax":
			if strings.Contains(strings.ToLower(rel), opwax) {
				return nil
			}
		case "all":
		default:
			return nil
		}
		remove = append(remove, path)
		return nil
	})
	for _, p := range remove {
		_ = os.Remove(p)
	}
	return nil
}

func CleanBITS(mode string) error {
	_ = StopService("bits")
	qmgr := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Network", "Downloader", "qmgr.db")
	_ = SecureDelete(qmgr)
	for _, ext := range []string{".dat", ".bak", ".log"} {
		matches, _ := filepath.Glob(qmgr + "*")
		for _, m := range matches {
			if strings.HasSuffix(m, ext) || m == qmgr {
				_ = SecureDelete(m)
			}
		}
	}
	if mode == "disable_clear" {
		return StopAndDisableService("bits")
	}
	return StartService("bits")
}

func CleanAlternateRunKeys(mode string) error {
	paths := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\RunServices`},
		{registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`},
		{registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\RunServicesOnce`},
	}
	for _, p := range paths {
		_ = DeleteRegKey(p.root, p.path)
	}
	if mode == "clean_disable" {
		return DisableAlternateRunKeysPolicy()
	}
	return nil
}

func DisableAlternateRunKeysPolicy() error {
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "DisableLocalMachineRunServices", 1)
	return SetRegDWORD(registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "DisableCurrentUserRunServices", 1)
}

// --- Cloud ---

func CleanOneDrive(mode string, users []models.UserProfile) error {
	for _, u := range users {
		base := filepath.Join(u.AppDataLocal, "Microsoft", "OneDrive")
		if mode == "metadata" || mode == "full" {
			_, _ = DeleteDirContents(filepath.Join(base, "logs", "Common"))
			settings := filepath.Join(base, "settings")
			matches, _ := filepath.Glob(filepath.Join(settings, "**", "*.dat"))
			for _, m := range matches {
				_ = SecureDelete(m)
			}
			entries, _ := os.ReadDir(settings)
			for _, e := range entries {
				if strings.HasSuffix(strings.ToLower(e.Name()), ".dat") {
					_ = SecureDelete(filepath.Join(settings, e.Name()))
				}
			}
		}
		if mode == "full" {
			_ = RunHidden("taskkill", "/F", "/IM", "OneDrive.exe")
			_, _ = DeleteDirContents(filepath.Join(base, "cache"))
		}
	}
	return nil
}

func DisableOneDriveSync() error {
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `Software\Policies\Microsoft\Windows\OneDrive`, "DisableFileSyncNGSC", 1)
	return SetRegDWORD(registry.CURRENT_USER, `Software\Policies\Microsoft\Windows\OneDrive`, "DisableFileSyncNGSC", 1)
}

func CleanCloudSync(users []models.UserProfile) error {
	for _, u := range users {
		paths := []string{
			filepath.Join(u.AppDataRoaming, "Dropbox", ".dropbox.cache"),
			filepath.Join(u.AppDataLocal, "Google", "DriveFS"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				_, _ = DeleteDirContents(p)
			}
		}
	}
	return nil
}

func DisableCloudSyncPolicies() error {
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `Software\Policies\Google\Drive`, "DisableDrive", 1)
	return nil
}

// --- Virtualization ---

func CleanHyperVSnapshots() error {
	root := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Hyper-V", "Snapshots")
	_, err := DeleteDirContents(root)
	return err
}

func CleanWSL(mode string, users []models.UserProfile) error {
	if mode == "logs" {
		_, _ = RunHiddenOutput("wsl", "--shutdown")
		return nil
	}
	if mode == "delete_vhdx" {
		_ = RunHidden("wsl", "--shutdown")
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
					_ = SecureDelete(vhdx)
				}
			}
		}
	}
	return nil
}
