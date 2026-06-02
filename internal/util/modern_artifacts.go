package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"

	"golang.org/x/sys/windows/registry"
)

// DisableWindowsRecall blocks Windows Recall / CoreAI snapshot capture.
func DisableWindowsRecall() error {
	_ = SetRegDWORD(registry.CURRENT_USER, `Software\Microsoft\Windows\Windows AI`, "DisableAIDataAnalysis", 1)
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\Windows AI`, "DisableAIDataAnalysis", 1)
	_ = SetRegDWORD(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\Windows AI`, "TurnOffSavingSnapshots", 1)
	return SetRegDWORD(registry.LOCAL_MACHINE, `SOFTWARE\Policies\Microsoft\Windows\WindowsAI`, "DisableAIDataAnalysis", 1)
}

// DeleteRecallCaptures removes CoreAI / Recall capture databases.
func DeleteRecallCaptures(users []models.UserProfile) error {
	for _, u := range users {
		base := filepath.Join(u.AppDataLocal, "CoreAIPlatform", "Capture")
		if err := secureRemoveTree(base); err != nil {
			return err
		}
	}
	return nil
}

// CleanPCALaunchLogs removes Program Compatibility Assistant trace files.
func CleanPCALaunchLogs() error {
	dir := filepath.Join(os.Getenv("SystemRoot"), "appcompat", "Programs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		name := strings.ToLower(e.Name())
		if strings.HasPrefix(name, "pca") {
			_ = SecureDelete(filepath.Join(dir, e.Name()))
		}
	}
	return nil
}

// CleanNotepadTabState removes unsaved Notepad tab/draft cache.
func CleanNotepadTabState(users []models.UserProfile) error {
	for _, u := range users {
		packages := filepath.Join(u.AppDataLocal, "Packages")
		entries, err := os.ReadDir(packages)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || !strings.HasPrefix(strings.ToLower(e.Name()), "microsoft.windowsnotepad") {
				continue
			}
			tab := filepath.Join(packages, e.Name(), "LocalState", "TabState")
			_ = secureRemoveTree(tab)
		}
	}
	return nil
}

// ClearOfficeTrustRecords removes Office macro trusted-document records.
func ClearOfficeTrustRecords(ctx *models.RunContext) error {
	const officeRoot = `Software\Microsoft\Office`
	for _, u := range ctx.TargetUsers {
		if strings.EqualFold(u.Username, ctx.CurrentUser) {
			if err := clearOfficeTrustUnder(registry.CURRENT_USER, officeRoot); err != nil {
				return err
			}
			continue
		}
		tempKey := "OPWAX_OFF_" + u.SID
		_ = RunHidden("reg", "unload", `HKU\`+tempKey)
		if err := RunHidden("reg", "load", `HKU\`+tempKey, u.NTUserPath); err != nil {
			continue
		}
		_ = clearOfficeTrustRegHive(`HKU\` + tempKey + `\` + officeRoot)
		_ = RunHidden("reg", "unload", `HKU\`+tempKey)
	}
	return nil
}

func clearOfficeTrustRegHive(fullOfficeRoot string) error {
	out, err := RunHiddenOutput("reg", "query", fullOfficeRoot, "/s")
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "TrustRecords") && strings.HasPrefix(line, "HKEY") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				_ = RunHidden("reg", "delete", parts[0], "/f")
			}
		}
	}
	return nil
}

func clearOfficeTrustUnder(root registry.Key, officeRoot string) error {
	office, err := registry.OpenKey(root, officeRoot, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	defer office.Close()
	versions, _ := office.ReadSubKeyNames(0)
	for _, ver := range versions {
		apps, err := registry.OpenKey(root, officeRoot+`\`+ver, registry.READ)
		if err != nil {
			continue
		}
		appNames, _ := apps.ReadSubKeyNames(0)
		apps.Close()
		for _, app := range appNames {
			trust := officeRoot + `\` + ver + `\` + app + `\Security\Trusted Documents\TrustRecords`
			_ = DeleteRegKey(root, trust)
		}
	}
	return nil
}

// ClearSysinternalsEULAKeys removes Sysinternals tool EULA acceptance records.
func ClearSysinternalsEULAKeys() error {
	return DeleteRegKey(registry.CURRENT_USER, `Software\Sysinternals`)
}

// CleanOutlookArtifacts clears Outlook local cache or OST/PST files.
func CleanOutlookArtifacts(mode string, users []models.UserProfile) error {
	if mode == "" || mode == "off" {
		return nil
	}
	for _, u := range users {
		dir := filepath.Join(u.AppDataLocal, "Microsoft", "Outlook")
		switch mode {
		case "cache":
			for _, sub := range []string{"RoamCache", "NewOutlook", "Offline Address Books"} {
				_, _ = DeleteDirContents(filepath.Join(dir, sub))
			}
			matches, _ := filepath.Glob(filepath.Join(dir, "*.nst"))
			for _, m := range matches {
				_ = SecureDelete(m)
			}
		case "delete_ost_pst":
			matches, _ := filepath.Glob(filepath.Join(dir, "*.ost"))
			matches2, _ := filepath.Glob(filepath.Join(dir, "*.pst"))
			for _, m := range append(matches, matches2...) {
				_ = SecureDelete(m)
			}
		}
	}
	return nil
}

// CleanTeamsArtifacts clears Teams local cache or full profile data.
func CleanTeamsArtifacts(mode string, users []models.UserProfile) error {
	if mode == "" || mode == "off" {
		return nil
	}
	for _, u := range users {
		legacy := filepath.Join(u.AppDataRoaming, "Microsoft", "Teams")
		packages := filepath.Join(u.AppDataLocal, "Packages")
		switch mode {
		case "cache":
			for _, sub := range []string{"Cache", "blob_storage", "databases", "GPUCache", "tmp"} {
				_, _ = DeleteDirContents(filepath.Join(legacy, sub))
			}
			entries, _ := os.ReadDir(packages)
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), "msteams") {
					for _, sub := range []string{"LocalCache", "AC", "TempState"} {
						_, _ = DeleteDirContents(filepath.Join(packages, e.Name(), sub))
					}
				}
			}
		case "full":
			_ = RunHidden("taskkill", "/F", "/IM", "ms-teams.exe")
			_ = RunHidden("taskkill", "/F", "/IM", "Teams.exe")
			_ = secureRemoveTree(legacy)
			entries, _ := os.ReadDir(packages)
			for _, e := range entries {
				if strings.Contains(strings.ToLower(e.Name()), "msteams") {
					_ = secureRemoveTree(filepath.Join(packages, e.Name()))
				}
			}
		}
	}
	return nil
}

// DisableEventLogStack stops event log collector services.
func DisableEventLogStack() error {
	for _, svc := range []string{"EventLog", "Wecsvc", "WdiSystemHost"} {
		_ = StopAndDisableService(svc)
	}
	return nil
}

// PurgeWinevtLogsDirectory removes all files under Winevt\Logs.
func PurgeWinevtLogsDirectory(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	n := 0
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if e.IsDir() {
			sub, err := PurgeWinevtLogsDirectory(p)
			n += sub
			if err != nil {
				return n, err
			}
			_ = os.Remove(p)
			continue
		}
		if err := os.Remove(p); err == nil {
			n++
		} else {
			_ = SecureDelete(p)
			n++
		}
	}
	return n, nil
}

func secureRemoveTree(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return SecureDelete(root)
	}
	_ = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		_ = SecureDelete(path)
		return nil
	})
	return os.RemoveAll(root)
}
