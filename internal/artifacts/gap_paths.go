package artifacts

import (
	"os"
	"path/filepath"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
)

// WindowsSearchIndexPath returns the main Search index database path.
func WindowsSearchIndexPath() string {
	return filepath.Join(system.ProgramDataDir(), "Microsoft", "Search", "Data", "Applications", "Windows", "Windows.edb")
}

// SyscachePaths returns Syscache hive and Win7 RecentFileCache paths.
func SyscachePaths() []string {
	dir := filepath.Join(system.WindowsDirectory(), "appcompat", "Programs")
	return []string{
		filepath.Join(dir, "Syscache.hve"),
		filepath.Join(dir, "RecentFileCache.bcf"),
	}
}

// EventTranscriptPath returns EventTranscript.db path.
func EventTranscriptPath() string {
	return filepath.Join(system.ProgramDataDir(), "Microsoft", "Diagnosis", "EventTranscript", "EventTranscript.db")
}

// WERSystemPath returns system WER folder.
func WERSystemPath() string {
	return filepath.Join(system.ProgramDataDir(), "Microsoft", "Windows", "WER")
}

// WERUserPaths returns per-user WER folders.
func WERUserPaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		out = append(out, filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "WER"))
	}
	return out
}

// ShellCacheGlobPatterns returns glob patterns for icon/thumb caches per user.
func ShellCacheGlobPatterns(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		base := filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "Explorer")
		out = append(out, filepath.Join(base, "iconcache_*.db"))
		out = append(out, filepath.Join(base, "thumbcache_*.db"))
	}
	return out
}

// ServicingLogDirs returns Panther/CBS/WindowsUpdate log directories.
func ServicingLogDirs() []string {
	win := system.WindowsDirectory()
	return []string{
		filepath.Join(win, "Panther"),
		filepath.Join(win, "Logs", "CBS"),
		filepath.Join(win, "Logs", "WindowsUpdate"),
	}
}

// PrintSpoolerDir returns the print spool folder.
func PrintSpoolerDir() string {
	return filepath.Join(system.WindowsDirectory(), "System32", "spool", "PRINTERS")
}

// DeliveryOptimizationCacheDir returns DO cache path.
func DeliveryOptimizationCacheDir() string {
	return filepath.Join(system.ProgramDataDir(), "Microsoft", "Windows", "DeliveryOptimization", "Cache")
}

// SmartScreenCachePaths returns SmartScreen local cache dirs per user.
func SmartScreenCachePaths(users []models.UserProfile) []string {
	var out []string
	for _, u := range users {
		out = append(out, filepath.Join(u.AppDataLocal, "Microsoft", "Windows", "Safety", "edge", "remote"))
	}
	return out
}

// DeveloperTracePaths returns common developer history paths for a user.
func DeveloperTracePaths(u models.UserProfile) []string {
	home := u.ProfilePath
	return []string{
		filepath.Join(home, ".gitconfig"),
		filepath.Join(home, ".lesshst"),
		filepath.Join(home, ".python_history"),
		filepath.Join(home, ".node_repl_history"),
		filepath.Join(home, ".bash_history"),
		filepath.Join(home, "AppData", "Roaming", "Code", "User", "History"),
		filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "History"),
	}
}

// ExistingPath filters paths that exist on disk.
func ExistingPath(paths []string) []string {
	var out []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}
