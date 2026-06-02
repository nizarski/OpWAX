package system

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opwax/opwax/internal/models"
)

// BrowserKind identifies a supported browser product.
type BrowserKind string

const (
	BrowserEdge    BrowserKind = "edge"
	BrowserChrome  BrowserKind = "chrome"
	BrowserFirefox BrowserKind = "firefox"
	BrowserBrave   BrowserKind = "brave"
	BrowserOpera   BrowserKind = "opera"
	BrowserVivaldi BrowserKind = "vivaldi"
)

type browserDef struct {
	kind   BrowserKind
	label  string
	detect func(models.UserProfile) bool
}

var browserCatalog = []browserDef{
	{
		kind:  BrowserEdge,
		label: "Microsoft Edge",
		detect: func(u models.UserProfile) bool {
			return dirExists(filepath.Join(u.AppDataLocal, "Microsoft", "Edge", "User Data"))
		},
	},
	{
		kind:  BrowserChrome,
		label: "Google Chrome",
		detect: func(u models.UserProfile) bool {
			return dirExists(filepath.Join(u.AppDataLocal, "Google", "Chrome", "User Data"))
		},
	},
	{
		kind:  BrowserFirefox,
		label: "Mozilla Firefox",
		detect: func(u models.UserProfile) bool {
			return firefoxInstalled(u.AppDataRoaming)
		},
	},
	{
		kind:  BrowserBrave,
		label: "Brave",
		detect: func(u models.UserProfile) bool {
			return dirExists(filepath.Join(u.AppDataLocal, "BraveSoftware", "Brave-Browser", "User Data"))
		},
	},
	{
		kind:  BrowserOpera,
		label: "Opera",
		detect: func(u models.UserProfile) bool {
			return operaInstalled(u.AppDataLocal, u.AppDataRoaming)
		},
	},
	{
		kind:  BrowserVivaldi,
		label: "Vivaldi",
		detect: func(u models.UserProfile) bool {
			return dirExists(filepath.Join(u.AppDataLocal, "Vivaldi", "User Data"))
		},
	},
}

// DetectedBrowser is an installed browser found on one or more profiles.
type DetectedBrowser struct {
	Kind  BrowserKind
	Label string
	Users []string
}

// DetectInstalledBrowsers scans local user profiles for supported browsers.
func DetectInstalledBrowsers() []DetectedBrowser {
	users, err := EnumerateUsers()
	if err != nil || len(users) == 0 {
		return nil
	}
	profiles, err := ResolveUserProfiles(users)
	if err != nil {
		return nil
	}

	found := map[BrowserKind]*DetectedBrowser{}
	for _, profile := range profiles {
		for _, def := range browserCatalog {
			if !def.detect(profile) {
				continue
			}
			entry, ok := found[def.kind]
			if !ok {
				entry = &DetectedBrowser{Kind: def.kind, Label: def.label}
				found[def.kind] = entry
			}
			entry.Users = appendUniqueBrowserUser(entry.Users, profile.Username)
		}
	}

	out := make([]DetectedBrowser, 0, len(found))
	for _, def := range browserCatalog {
		if entry, ok := found[def.kind]; ok {
			out = append(out, *entry)
		}
	}
	return out
}

// ApplyDetectedBrowsers enables only browsers that were detected on this system.
func ApplyDetectedBrowsers(cfg *models.BrowserConfig, detected []DetectedBrowser) {
	cfg.Edge = false
	cfg.Chrome = false
	cfg.Firefox = false
	cfg.Brave = false
	cfg.Opera = false
	cfg.Vivaldi = false
	for _, b := range detected {
		setBrowserConfig(cfg, b.Kind, true)
	}
}

// SetBrowserConfig updates a single browser flag in config.
func SetBrowserConfig(cfg *models.BrowserConfig, kind BrowserKind, enabled bool) {
	setBrowserConfig(cfg, kind, enabled)
}

func setBrowserConfig(cfg *models.BrowserConfig, kind BrowserKind, enabled bool) {
	switch kind {
	case BrowserEdge:
		cfg.Edge = enabled
	case BrowserChrome:
		cfg.Chrome = enabled
	case BrowserFirefox:
		cfg.Firefox = enabled
	case BrowserBrave:
		cfg.Brave = enabled
	case BrowserOpera:
		cfg.Opera = enabled
	case BrowserVivaldi:
		cfg.Vivaldi = enabled
	}
}

// BrowserKindLabel returns a display label for a browser kind.
func BrowserKindLabel(kind BrowserKind) string {
	for _, def := range browserCatalog {
		if def.kind == kind {
			return def.label
		}
	}
	return string(kind)
}

// BrowserProcessNames returns executable names for enabled browsers.
func BrowserProcessNames(cfg models.BrowserConfig) map[string]string {
	procs := map[string]string{}
	if cfg.Edge {
		procs["msedge.exe"] = "Edge"
	}
	if cfg.Chrome {
		procs["chrome.exe"] = "Chrome"
	}
	if cfg.Firefox {
		procs["firefox.exe"] = "Firefox"
	}
	if cfg.Brave {
		procs["brave.exe"] = "Brave"
	}
	if cfg.Opera {
		procs["opera.exe"] = "Opera"
	}
	if cfg.Vivaldi {
		procs["vivaldi.exe"] = "Vivaldi"
	}
	return procs
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func firefoxInstalled(appDataRoaming string) bool {
	profilesDir := filepath.Join(appDataRoaming, "Mozilla", "Firefox", "Profiles")
	if dirExists(profilesDir) {
		entries, err := os.ReadDir(profilesDir)
		return err == nil && len(entries) > 0
	}
	profilesIni := filepath.Join(appDataRoaming, "Mozilla", "Firefox", "profiles.ini")
	if _, err := os.Stat(profilesIni); err == nil {
		return true
	}
	return false
}

func operaInstalled(localAppData, roamingAppData string) bool {
	candidates := []string{
		filepath.Join(roamingAppData, "Opera Software", "Opera Stable"),
		filepath.Join(localAppData, "Programs", "Opera"),
		filepath.Join(localAppData, "Opera Software", "Opera Stable"),
		filepath.Join(localAppData, "Opera Software", "Opera GX Stable"),
	}
	for _, p := range candidates {
		if dirExists(p) {
			return true
		}
	}
	return false
}

func appendUniqueBrowserUser(users []string, name string) []string {
	for _, u := range users {
		if strings.EqualFold(u, name) {
			return users
		}
	}
	users = append(users, name)
	sort.Strings(users)
	return users
}
