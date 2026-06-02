package modules

import (
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

const registryName = "registry_hives"

// RegistryModule cleans registry-based forensic artifacts.
type RegistryModule struct{}

func (m *RegistryModule) Name() string { return registryName }

func (m *RegistryModule) DryRun(ctx *models.RunContext) []models.Action {
	var actions []models.Action
	actions = append(actions,
		action(registryName, "Disable recent docs history", `Explorer\Policies`, models.ActionDisable),
		action(registryName, "Clear USB device enum history", `HKLM\SYSTEM\Enum\USBSTOR`, models.ActionClean),
		action(registryName, "Clear AppCompatCache (Shimcache)", `AppCompatCache`, models.ActionClean),
	)
	for _, u := range ctx.TargetUsers {
		actions = append(actions,
			action(registryName, "Clear ShellBags for "+u.Username, u.NTUserPath, models.ActionClean),
			action(registryName, "Clear UsrClass ShellBags for "+u.Username, u.UsrClassPath, models.ActionClean),
			action(registryName, "Clear Open/Save MRU for "+u.Username, "ComDlg32", models.ActionClean),
		)
	}
	actions = append(actions, advancedRegistryDryRun(ctx)...)
	actions = append(actions, gapRegistryDryRun(ctx)...)
	return actions
}

func (m *RegistryModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	policies := []struct {
		path, name string
		val        uint32
	}{
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "NoRecentDocsHistory", 1},
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "ClearRecentDocsOnExit", 1},
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "NoLowDiskSpaceChecks", 1},
	}
	for _, p := range policies {
		a := action(registryName, "Set policy "+p.name, p.path, models.ActionDisable)
		results = append(results, result(a, util.SetRegDWORD(registry.CURRENT_USER, p.path, p.name, p.val)))
	}
	// Machine-level explorer policy
	a := action(registryName, "Machine NoRecentDocsHistory", "HKLM Policies", models.ActionDisable)
	results = append(results, result(a, util.SetRegDWORD(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`, "NoRecentDocsHistory", 1)))

	results = append(results, advancedRegistryDisable(ctx)...)
	results = append(results, gapRegistryDisable(ctx)...)
	return results
}

func (m *RegistryModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	// USB history - USBSTOR only (safe)
	usbPaths := []string{
		`SYSTEM\CurrentControlSet\Enum\USBSTOR`,
		`SYSTEM\CurrentControlSet\Enum\USB`,
	}
	for _, p := range usbPaths {
		a := action(registryName, "Clear "+p, p, models.ActionClean)
		results = append(results, result(a, util.DeleteRegKey(registry.LOCAL_MACHINE, p)))
	}

	// AppCompatCache
	a := action(registryName, "Clear AppCompatCache", "AppCompatCache", models.ActionClean)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager`, registry.SET_VALUE)
	if err == nil {
		_ = key.DeleteValue("AppCompatCache")
		key.Close()
	}
	results = append(results, result(a, err))

	// Per-user registry keys (loaded hives via API on live system)
	ntuserKeys := []string{
		`Software\Microsoft\Windows\CurrentVersion\Explorer\RecentDocs`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\RunMRU`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\TypedPaths`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\WordWheelQuery`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\StreamMRU`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Bags`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\BagMRU`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\ComDlg32\OpenSavePidlMRU`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\ComDlg32\LastVisitedPidlMRU`,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\ComDlg32\LastVisitedPidlMRULegacy`,
	}

	for _, u := range ctx.TargetUsers {
		for _, k := range ntuserKeys {
			a := action(registryName, "Clear HKCU key for "+u.Username, k, models.ActionClean)
			results = append(results, result(a, clearUserRegKey(ctx, u, k)))
		}
		// UsrClass ShellBags via profile path
		usrClassKeys := []string{
			`Local Settings\Software\Microsoft\Windows\Shell\BagMRU`,
			`Local Settings\Software\Microsoft\Windows\Shell\Bags`,
		}
		for _, k := range usrClassKeys {
			a := action(registryName, "Clear UsrClass for "+u.Username, k, models.ActionClean)
			results = append(results, result(a, clearUsrClassKey(u.ProfilePath, k)))
		}
	}

	// BAM/DAM disable + clean
	for _, svc := range []string{"bam", "dam"} {
		a := action(registryName, "Disable "+svc, svc, models.ActionDisable)
		results = append(results, result(a, util.StopAndDisableService(svc)))
		statePath := `SYSTEM\CurrentControlSet\Services\` + svc + `\State\UserSettings`
		a2 := action(registryName, "Clear "+svc+" UserSettings", statePath, models.ActionClean)
		results = append(results, result(a2, util.DeleteRegKey(registry.LOCAL_MACHINE, statePath)))
	}

	results = append(results, advancedRegistryClean(ctx)...)
	results = append(results, gapRegistryClean(ctx)...)
	return results
}

// clearUserRegKey clears a key for a user by loading their NTUSER hive if needed.
func clearUserRegKey(ctx *models.RunContext, u models.UserProfile, subKey string) error {
	if strings.EqualFold(u.Username, ctx.CurrentUser) {
		return util.DeleteRegKey(registry.CURRENT_USER, subKey)
	}
	tempKey := "OPWAX_NT_" + u.SID
	_ = util.RunHidden("reg", "unload", `HKU\`+tempKey)
	if err := util.RunHidden("reg", "load", `HKU\`+tempKey, u.NTUserPath); err != nil {
		return err
	}
	defer util.RunHidden("reg", "unload", `HKU\`+tempKey)
	return util.RunHidden("reg", "delete", `HKU\`+tempKey+`\`+subKey, "/f")
}

func clearUsrClassKey(profilePath, subKey string) error {
	hivePath := filepath.Join(profilePath, "AppData", "Local", "Microsoft", "Windows", "UsrClass.dat")
	tempKey := "OPWAX_USR_" + filepath.Base(profilePath)
	_ = util.RunHidden("reg", "unload", `HKU\`+tempKey)
	if err := util.RunHidden("reg", "load", `HKU\`+tempKey, hivePath); err != nil {
		return err
	}
	defer util.RunHidden("reg", "unload", `HKU\`+tempKey)
	return util.RunHidden("reg", "delete", `HKU\`+tempKey+`\`+subKey, "/f")
}
