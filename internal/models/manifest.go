package models

import "time"

// ArtifactEntry describes a discovered artifact location on this system.
type ArtifactEntry struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Exists      bool   `json:"exists"`
	SizeBytes   int64  `json:"size_bytes,omitempty"`
	ItemCount   int    `json:"item_count,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// SystemManifest is an auto-discovered map of artifact paths on this machine.
type SystemManifest struct {
	GeneratedAt  time.Time       `json:"generated_at"`
	Hostname     string          `json:"hostname"`
	OSVersion    string          `json:"os_version"`
	WindowsDir   string          `json:"windows_dir"`
	CurrentUser  string          `json:"current_user"`
	Users        []UserProfile   `json:"users"`
	Drives       []DriveInfo     `json:"drives"`
	Artifacts    []ArtifactEntry `json:"artifacts"`
	EventLogs    []string        `json:"event_logs,omitempty"`
}
