package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

// GenerateManifest scans this system and builds an artifact path map.
func GenerateManifest() (*models.SystemManifest, error) {
	if err := RequireAdmin(); err != nil {
		return nil, err
	}

	host, _ := os.Hostname()
	currentUser, _ := CurrentUsername()
	winDir := WindowsDirectory()

	users, err := EnumerateUsers()
	if err != nil {
		return nil, err
	}
	profiles, err := ResolveUserProfiles(users)
	if err != nil {
		return nil, err
	}

	drives, err := EnumerateDrives()
	if err != nil {
		return nil, err
	}

	m := &models.SystemManifest{
		GeneratedAt: time.Now().UTC(),
		Hostname:    host,
		OSVersion:   runtime.GOOS + " " + osVersionString(),
		WindowsDir:  winDir,
		CurrentUser: currentUser,
		Users:       profiles,
		Drives:      drives,
	}

	m.Artifacts = append(m.Artifacts, discoverSystemArtifacts(winDir, drives)...)
	for _, u := range profiles {
		m.Artifacts = append(m.Artifacts, discoverUserArtifacts(u)...)
	}

	channels, _ := util.ListEventLogChannels()
	m.EventLogs = channels

	return m, nil
}

// ManifestJSON returns indented JSON for a manifest.
func ManifestJSON(m *models.SystemManifest) (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveManifest writes manifest JSON to path.
func SaveManifest(m *models.SystemManifest, path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func osVersionString() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ)
	if err != nil {
		return "unknown"
	}
	defer k.Close()
	product, _, _ := k.GetStringValue("ProductName")
	build, _, _ := k.GetStringValue("CurrentBuildNumber")
	display, _, _ := k.GetStringValue("DisplayVersion")
	if display != "" {
		return fmt.Sprintf("%s %s (build %s)", product, display, build)
	}
	return fmt.Sprintf("%s (build %s)", product, build)
}

func discoverSystemArtifacts(winDir string, drives []models.DriveInfo) []models.ArtifactEntry {
	var entries []models.ArtifactEntry

	add := func(id, cat, name, path string, countItems bool) {
		e := models.ArtifactEntry{ID: id, Category: cat, Name: name, Path: path}
		if info, err := os.Stat(path); err == nil {
			e.Exists = true
			if info.IsDir() {
				if countItems {
					e.ItemCount = countDirItems(path)
				}
			} else {
				e.SizeBytes = info.Size()
			}
		}
		entries = append(entries, e)
	}

	add("prefetch", "execution", "Prefetch files", filepath.Join(winDir, "Prefetch"), true)
	add("amcache", "execution", "Amcache.hve", filepath.Join(winDir, "appcompat", "Programs", "Amcache.hve"), false)
	add("pca_launch", "execution", "PCA launch logs", filepath.Join(winDir, "appcompat", "Programs"), true)
	add("srum", "network", "SRUDB.dat", filepath.Join(winDir, "System32", "sru", "SRUDB.dat"), false)
	add("evtx", "logs", "Event logs directory", filepath.Join(winDir, "System32", "Winevt", "Logs"), true)
	add("ps_evtx", "logs", "PowerShell Operational log",
		filepath.Join(winDir, "System32", "Winevt", "Logs", "Microsoft-Windows-PowerShell%4Operational.evtx"), false)
	add("memory_dmp", "persistence", "MEMORY.DMP", filepath.Join(winDir, "MEMORY.DMP"), false)
	add("minidump", "persistence", "Minidump folder", filepath.Join(winDir, "Minidump"), true)
	add("sam", "registry", "SAM hive", filepath.Join(winDir, "System32", "config", "SAM"), false)
	add("system_hive", "registry", "SYSTEM hive", filepath.Join(winDir, "System32", "config", "SYSTEM"), false)
	add("software_hive", "registry", "SOFTWARE hive", filepath.Join(winDir, "System32", "config", "SOFTWARE"), false)

	wlanRoot := filepath.Join(ProgramDataDir(), "Microsoft", "Wlansvc", "Profiles", "Interfaces")
	add("wlan", "network", "WLAN profiles", wlanRoot, true)

	for _, d := range drives {
		add("usn_"+d.Letter, "ntfs", "USN journal (via fsutil)", d.Letter, false)
		entries[len(entries)-1].Notes = "Use fsutil usn queryjournal " + d.Letter
		add("pagefile_"+d.Letter, "persistence", "pagefile.sys", d.Root+"pagefile.sys", false)
		add("hiberfil", "persistence", "hiberfil.sys", d.Root+"hiberfil.sys", false)
		add("recycle_"+d.Letter, "persistence", "Recycle Bin", filepath.Join(d.Root, "$Recycle.Bin"), true)
	}

	return entries
}

func discoverUserArtifacts(u models.UserProfile) []models.ArtifactEntry {
	var entries []models.ArtifactEntry
	prefix := strings.ToLower(u.Username)

	add := func(id, cat, name, path string, countItems bool) {
		e := models.ArtifactEntry{
			ID:       id + "_" + prefix,
			Category: cat,
			Name:     name + " (" + u.Username + ")",
			Path:     path,
		}
		if info, err := os.Stat(path); err == nil {
			e.Exists = true
			if info.IsDir() {
				if countItems {
					e.ItemCount = countDirItems(path)
				}
			} else {
				e.SizeBytes = info.Size()
			}
		}
		entries = append(entries, e)
	}

	add("ntuser", "registry", "NTUSER.DAT", u.NTUserPath, false)
	add("usrclass", "registry", "UsrClass.dat", u.UsrClassPath, false)
	add("jump_auto", "logs", "AutomaticDestinations",
		filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent", "AutomaticDestinations"), true)
	add("jump_custom", "logs", "CustomDestinations",
		filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent", "CustomDestinations"), true)
	add("recent_lnk", "logs", "Recent shortcuts",
		filepath.Join(u.AppDataRoaming, "Microsoft", "Windows", "Recent"), true)

	recall := filepath.Join(u.AppDataLocal, "CoreAIPlatform", "Capture")
	add("recall", "modern", "Windows Recall capture", recall, true)

	edgeHist := filepath.Join(u.AppDataLocal, "Microsoft", "Edge", "User Data", "Default", "History")
	chromeHist := filepath.Join(u.AppDataLocal, "Google", "Chrome", "User Data", "Default", "History")
	add("edge", "browser", "Edge History", edgeHist, false)
	add("chrome", "browser", "Chrome History", chromeHist, false)

	ffProfiles := filepath.Join(u.AppDataRoaming, "Mozilla", "Firefox", "Profiles")
	if entries2 := discoverFirefoxProfiles(u.Username, ffProfiles); len(entries2) > 0 {
		entries = append(entries, entries2...)
	}

	return entries
}

func discoverFirefoxProfiles(username, profilesDir string) []models.ArtifactEntry {
	var entries []models.ArtifactEntry
	dirs, err := os.ReadDir(profilesDir)
	if err != nil {
		return entries
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		places := filepath.Join(profilesDir, d.Name(), "places.sqlite")
		e := models.ArtifactEntry{
			ID:       "firefox_" + strings.ToLower(username) + "_" + d.Name(),
			Category: "browser",
			Name:     "Firefox places.sqlite (" + username + "/" + d.Name() + ")",
			Path:     places,
		}
		if info, err := os.Stat(places); err == nil {
			e.Exists = true
			e.SizeBytes = info.Size()
		}
		entries = append(entries, e)
	}
	return entries
}

func countDirItems(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(entries)
}
