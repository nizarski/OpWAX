package artifacts

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
)

// RecallCapturePaths returns CoreAI / Recall capture directories per user.
func RecallCapturePaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		out = append(out, filepath.Join(u.AppDataLocal, "CoreAIPlatform", "Capture"))
	}
	return out
}

// PCALaunchLogPaths returns PCA compatibility trace files.
func PCALaunchLogPaths() []string {
	dir := filepath.Join(system.WindowsDirectory(), "appcompat", "Programs")
	var out []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(strings.ToLower(e.Name()), "pca") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// NotepadTabStatePaths returns Notepad tab-state folders per user.
func NotepadTabStatePaths(users []models.UserProfile) []string {
	var out []string
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
			p := filepath.Join(packages, e.Name(), "LocalState", "TabState")
			if _, err := os.Stat(p); err == nil {
				out = append(out, p)
			}
		}
	}
	return out
}

// OutlookPaths returns Outlook data directories per user.
func OutlookPaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		out = append(out, filepath.Join(u.AppDataLocal, "Microsoft", "Outlook"))
	}
	return out
}

// TeamsPaths returns Teams legacy install paths per user.
func TeamsPaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		out = append(out, filepath.Join(u.AppDataRoaming, "Microsoft", "Teams"))
	}
	return out
}
