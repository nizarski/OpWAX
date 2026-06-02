package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"golang.org/x/sys/windows/registry"
)

const zoneIdentifierStream = ":Zone.Identifier"

// ZoneIdentifierScanRoots returns profile folders where Zone.Identifier ADS commonly appears.
func ZoneIdentifierScanRoots(users []models.UserProfile) []string {
	seen := map[string]bool{}
	var roots []string
	add := func(path string) {
		path = filepath.Clean(path)
		if path == "" || seen[path] {
			return
		}
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return
		}
		seen[path] = true
		roots = append(roots, path)
	}

	for _, u := range users {
		add(u.ProfilePath)
		add(filepath.Join(u.ProfilePath, "Downloads"))
		add(filepath.Join(u.ProfilePath, "Desktop"))
		add(filepath.Join(u.ProfilePath, "Documents"))
		add(filepath.Join(u.AppDataLocal, "Temp"))
	}
	return roots
}

// StripZoneIdentifiersInPaths removes Zone.Identifier ADS under the given roots.
// NTFS has no volume-wide ADS index; this is one pass over known user folders (not whole C:\).
func StripZoneIdentifiersInPaths(roots []string) (int, error) {
	count := 0
	for _, root := range roots {
		n, err := stripZoneIdentifiersUnderRoot(root)
		count += n
		if err != nil {
			return count, err
		}
	}
	return count, nil
}

func stripZoneIdentifiersUnderRoot(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := strings.ToLower(d.Name())
			if base == "$recycle.bin" || base == "system volume information" {
				return filepath.SkipDir
			}
			return nil
		}
		adsPath := path + zoneIdentifierStream
		if _, err := os.Stat(adsPath); err == nil {
			if rmErr := os.Remove(adsPath); rmErr == nil {
				count++
			}
		}
		return nil
	})
	return count, err
}

// DisableZoneIdentifierSaving sets registry to stop saving zone info.
func DisableZoneIdentifierSaving() error {
	return SetRegDWORD(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Attachments`,
		"SaveZoneInformation", 1)
}
