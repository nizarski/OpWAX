package models

// AdvancedOptions controls niche / deep forensic artifact handling.
// Folded into existing modules; each option requires its parent module enabled.
type AdvancedOptions struct {
	DeleteVSSShadows      bool `json:"delete_vss_shadows"`
	FullVolumeUnallocated bool `json:"full_volume_unallocated"`
	WipeFileSlack         bool `json:"wipe_file_slack"`
	WipeFileSlackAllFiles bool `json:"wipe_file_slack_all_files"`
	ScrubBadClusters      bool `json:"scrub_bad_clusters"`

	OneDrive         string `json:"onedrive"`           // off, metadata, full
	CloudSync        string `json:"cloud_sync"`         // off, all
	WMI              string `json:"wmi"`                // off, reset
	ScheduledTasks   string `json:"scheduled_tasks"`    // off, user, all_except_opwax, all
	BITS             string `json:"bits"`               // off, clear, disable_clear
	AlternateRunKeys string `json:"alternate_run_keys"` // off, clean, clean_disable

	PowerShellHistory bool   `json:"powershell_history"`
	TimelineActivity  bool   `json:"timeline_activity"`
	RDPCache          string `json:"rdp_cache"` // off, target_users
	UserAssist        bool   `json:"userassist"`

	HyperV string `json:"hyperv"` // off, snapshots
	WSL    string `json:"wsl"`    // off, logs, delete_vhdx

	WindowsRecall      bool   `json:"windows_recall"`
	PCAAppCompatLogs   bool   `json:"pca_appcompat_logs"`
	NotepadTabCache    bool   `json:"notepad_tab_cache"`
	OfficeTrustRecords bool   `json:"office_trust_records"`
	SysinternalsEULA   bool   `json:"sysinternals_eula"`
	Outlook            string `json:"outlook"` // off, cache, delete_ost_pst
	Teams              string `json:"teams"`   // off, cache, full

	// Tier A - high-value execution gaps (safe defaults on)
	WindowsSearchIndex    bool `json:"windows_search_index"`
	ShellIconCaches       bool `json:"shell_icon_caches"`
	ExecutionRegistries   bool `json:"execution_registries"`
	SyscacheHive          bool `json:"syscache_hive"`
	EventTranscript       bool `json:"event_transcript"`
	WERReports            bool `json:"wer_reports"`
	TargetedEventChannels bool `json:"targeted_event_channels"`

	// Tier B - optional deep traces (default off)
	ServicingLogs        bool `json:"servicing_logs"`
	PrintSpooler         bool `json:"print_spooler"`
	DeveloperTraces      bool `json:"developer_traces"`
	DeliveryOptimization bool `json:"delivery_optimization"`
	SmartScreenCache     bool `json:"smartscreen_cache"`
}

// DefaultAdvancedOptions returns safe-profile defaults (low-risk execution artifacts on).
func DefaultAdvancedOptions() AdvancedOptions {
	return AdvancedOptions{
		OneDrive:              "off",
		CloudSync:             "off",
		WMI:                   "off",
		ScheduledTasks:        "off",
		BITS:                  "off",
		AlternateRunKeys:      "off",
		RDPCache:              "off",
		HyperV:                "off",
		WSL:                   "off",
		PowerShellHistory:     true,
		TimelineActivity:      true,
		UserAssist:            true,
		WindowsRecall:         true,
		PCAAppCompatLogs:      true,
		NotepadTabCache:       true,
		OfficeTrustRecords:    true,
		SysinternalsEULA:      true,
		Outlook:               "off",
		Teams:                 "off",
		WindowsSearchIndex:    true,
		ShellIconCaches:       true,
		ExecutionRegistries:   true,
		SyscacheHive:          true,
		EventTranscript:       true,
		WERReports:            true,
		TargetedEventChannels: true,
	}
}

func (a AdvancedOptions) OneDriveEnabled() bool  { return a.OneDrive != "" && a.OneDrive != "off" }
func (a AdvancedOptions) CloudSyncEnabled() bool { return a.CloudSync != "" && a.CloudSync != "off" }
func (a AdvancedOptions) WMIEnabled() bool       { return a.WMI != "" && a.WMI != "off" }
func (a AdvancedOptions) ScheduledTasksEnabled() bool {
	return a.ScheduledTasks != "" && a.ScheduledTasks != "off"
}
func (a AdvancedOptions) BITSEnabled() bool { return a.BITS != "" && a.BITS != "off" }
func (a AdvancedOptions) AlternateRunKeysEnabled() bool {
	return a.AlternateRunKeys != "" && a.AlternateRunKeys != "off"
}
func (a AdvancedOptions) RDPCacheEnabled() bool { return a.RDPCache != "" && a.RDPCache != "off" }
func (a AdvancedOptions) HyperVEnabled() bool   { return a.HyperV != "" && a.HyperV != "off" }
func (a AdvancedOptions) WSLEnabled() bool      { return a.WSL != "" && a.WSL != "off" }
func (a AdvancedOptions) OutlookEnabled() bool  { return a.Outlook != "" && a.Outlook != "off" }
func (a AdvancedOptions) TeamsEnabled() bool    { return a.Teams != "" && a.Teams != "off" }
