package modules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"
)

const networkName = "network_browser"

// NetworkModule handles SRUM, WLAN, and browser artifacts.
type NetworkModule struct{}

func (m *NetworkModule) Name() string { return networkName }

func (m *NetworkModule) DryRun(ctx *models.RunContext) []models.Action {
	var actions []models.Action
	actions = append(actions,
		action(networkName, "Stop and disable DiagTrack", "DiagTrack", models.ActionDisable),
		action(networkName, "Stop and disable DPS", "DPS", models.ActionDisable),
		action(networkName, "Secure-delete SRUDB.dat", "sru", models.ActionSecure),
	)
	if ctx.Config.Options.WLANMode != models.WLANModeSkip {
		actions = append(actions,
			action(networkName, "Clean WLAN profiles ("+string(ctx.Config.Options.WLANMode)+")", "Wlansvc\\Profiles", models.ActionClean),
		)
	}
	if ctx.Config.Options.Browsers.Edge {
		actions = append(actions, action(networkName, "Clear Edge history/cache", "Edge", models.ActionClean))
	}
	if ctx.Config.Options.Browsers.Chrome {
		actions = append(actions, action(networkName, "Clear Chrome history/cache", "Chrome", models.ActionClean))
	}
	if ctx.Config.Options.Browsers.Firefox {
		actions = append(actions, action(networkName, "Clear Firefox history/cache", "Firefox", models.ActionClean))
	}
	if ctx.Config.Options.Browsers.Brave {
		actions = append(actions, action(networkName, "Clear Brave history/cache", "Brave", models.ActionClean))
	}
	if ctx.Config.Options.Browsers.Opera {
		actions = append(actions, action(networkName, "Clear Opera history/cache", "Opera", models.ActionClean))
	}
	if ctx.Config.Options.Browsers.Vivaldi {
		actions = append(actions, action(networkName, "Clear Vivaldi history/cache", "Vivaldi", models.ActionClean))
	}
	actions = append(actions, advancedNetworkDryRun(ctx)...)
	actions = append(actions, gapNetworkDryRun(ctx)...)
	return actions
}

func (m *NetworkModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result
	a1 := action(networkName, "Disable DiagTrack", "DiagTrack", models.ActionDisable)
	results = append(results, result(a1, util.StopAndDisableService("DiagTrack")))
	a2 := action(networkName, "Disable DPS", "DPS", models.ActionDisable)
	results = append(results, result(a2, util.StopAndDisableService("DPS")))
	results = append(results, advancedNetworkDisable(ctx)...)
	return results
}

func (m *NetworkModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	// SRUM
	sruDir := filepath.Join(system.WindowsDirectory(), "System32", "sru")
	for _, name := range []string{"SRUDB.dat", "SRUDB.dat.LOG1", "SRUDB.dat.LOG2", "SRU.log", "SRU.chk", "SRU000*.log"} {
		a := action(networkName, "Secure-delete "+name, sruDir, models.ActionSecure)
		if strings.Contains(name, "*") {
			_, _ = util.DeleteFilesGlob(sruDir, name)
		} else {
			_ = util.SecureDelete(filepath.Join(sruDir, name))
		}
		results = append(results, result(a, nil))
	}

	// WLAN
	if ctx.Config.Options.WLANMode != models.WLANModeSkip {
		a := action(networkName, "Clean WLAN profiles", "Wlansvc", models.ActionClean)
		results = append(results, result(a, cleanWLANProfiles(ctx.Config.Options.WLANMode)))
	}

	// Browsers
	for _, u := range ctx.TargetUsers {
		if ctx.Config.Options.Browsers.Edge {
			a := action(networkName, "Clear Edge for "+u.Username, "Edge", models.ActionClean)
			results = append(results, result(a, cleanChromium(u.AppDataLocal, "Microsoft", "Edge")))
		}
		if ctx.Config.Options.Browsers.Chrome {
			a := action(networkName, "Clear Chrome for "+u.Username, "Chrome", models.ActionClean)
			results = append(results, result(a, cleanChromium(u.AppDataLocal, "Google", "Chrome")))
		}
		if ctx.Config.Options.Browsers.Firefox {
			a := action(networkName, "Clear Firefox for "+u.Username, "Firefox", models.ActionClean)
			results = append(results, result(a, cleanFirefox(u.AppDataRoaming)))
		}
		if ctx.Config.Options.Browsers.Brave {
			a := action(networkName, "Clear Brave for "+u.Username, "Brave", models.ActionClean)
			results = append(results, result(a, cleanChromium(u.AppDataLocal, "BraveSoftware", "Brave-Browser")))
		}
		if ctx.Config.Options.Browsers.Opera {
			a := action(networkName, "Clear Opera for "+u.Username, "Opera", models.ActionClean)
			results = append(results, result(a, cleanOpera(u.AppDataLocal, u.AppDataRoaming)))
		}
		if ctx.Config.Options.Browsers.Vivaldi {
			a := action(networkName, "Clear Vivaldi for "+u.Username, "Vivaldi", models.ActionClean)
			results = append(results, result(a, cleanChromium(u.AppDataLocal, "Vivaldi", "User Data")))
		}
	}

	results = append(results, advancedNetworkClean(ctx)...)
	results = append(results, gapNetworkClean(ctx)...)
	return results
}

func cleanWLANProfiles(mode models.WLANMode) error {
	profilesRoot := filepath.Join(system.ProgramDataDir(), "Microsoft", "Wlansvc", "Profiles", "Interfaces")
	interfaces, err := os.ReadDir(profilesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	currentSSID := ""
	if mode == models.WLANModeExceptCurrent {
		out, _ := util.RunHiddenOutput("netsh", "wlan", "show", "interfaces")
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "SSID") && !strings.Contains(line, "BSSID") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					currentSSID = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	for _, iface := range interfaces {
		if !iface.IsDir() {
			continue
		}
		ifacePath := filepath.Join(profilesRoot, iface.Name())
		profiles, _ := os.ReadDir(ifacePath)
		for _, p := range profiles {
			if p.IsDir() || !strings.HasSuffix(p.Name(), ".xml") {
				continue
			}
			if mode == models.WLANModeExceptCurrent && currentSSID != "" {
				data, err := os.ReadFile(filepath.Join(ifacePath, p.Name()))
				if err == nil && strings.Contains(string(data), currentSSID) {
					continue
				}
			}
			_ = os.Remove(filepath.Join(ifacePath, p.Name()))
		}
	}
	return nil
}

func cleanChromium(localAppData, vendor, product string) error {
	killChromiumBrowser(vendor, product)
	base := filepath.Join(localAppData, vendor, product, "User Data")
	if product == "User Data" {
		base = filepath.Join(localAppData, vendor, "User Data")
	}
	profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3"}
	files := []string{"History", "History-journal", "Cookies", "Cookies-journal", "Web Data", "Web Data-journal", "Login Data", "Login Data-journal", "Top Sites", "Visited Links", "Shortcuts"}
	for _, prof := range profiles {
		profDir := filepath.Join(base, prof)
		for _, f := range files {
			_ = os.Remove(filepath.Join(profDir, f))
		}
		cacheDir := filepath.Join(profDir, "Cache", "Cache_Data")
		_, _ = util.DeleteDirContents(cacheDir)
		codeCache := filepath.Join(profDir, "Code Cache")
		_ = os.RemoveAll(codeCache)
	}
	return nil
}

func cleanFirefox(appDataRoaming string) error {
	_ = util.RunHidden("taskkill", "/F", "/IM", "firefox.exe")
	profilesDir := filepath.Join(appDataRoaming, "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(profilesDir, e.Name())
		for _, f := range []string{"places.sqlite", "places.sqlite-wal", "places.sqlite-shm", "cookies.sqlite", "cookies.sqlite-wal", "formhistory.sqlite", "downloads.sqlite"} {
			_ = os.Remove(filepath.Join(dir, f))
		}
		cache2 := filepath.Join(dir, "cache2")
		_ = os.RemoveAll(cache2)
	}
	return nil
}

func cleanOpera(localAppData, roamingAppData string) error {
	_ = util.RunHidden("taskkill", "/F", "/IM", "opera.exe")
	candidates := []string{
		filepath.Join(roamingAppData, "Opera Software", "Opera Stable"),
		filepath.Join(localAppData, "Opera Software", "Opera Stable"),
		filepath.Join(localAppData, "Opera Software", "Opera GX Stable"),
	}
	files := []string{"History", "History-journal", "Cookies", "Cookies-journal", "Web Data", "Web Data-journal", "Login Data", "Login Data-journal", "Visited Links", "Shortcuts"}
	for _, base := range candidates {
		if _, err := os.Stat(base); err != nil {
			continue
		}
		for _, f := range files {
			_ = os.Remove(filepath.Join(base, f))
		}
		cacheDir := filepath.Join(base, "Cache", "Cache_Data")
		_, _ = util.DeleteDirContents(cacheDir)
		_ = os.RemoveAll(filepath.Join(base, "Code Cache"))
	}
	return nil
}

func killChromiumBrowser(vendor, product string) {
	var img string
	switch {
	case vendor == "Microsoft" && product == "Edge":
		img = "msedge.exe"
	case vendor == "Google" && product == "Chrome":
		img = "chrome.exe"
	case vendor == "BraveSoftware" && product == "Brave-Browser":
		img = "brave.exe"
	case vendor == "Vivaldi" && product == "User Data":
		img = "vivaldi.exe"
	default:
		return
	}
	_ = util.RunHidden("taskkill", "/F", "/IM", img)
}
