package models

// UserMode defines which user profiles to target.
type UserMode string

const (
	UserModeCurrent UserMode = "current"
	UserModeAll     UserMode = "all"
	UserModeSelect  UserMode = "select"
)

// DriveMode defines which drives/volumes to target.
type DriveMode string

const (
	DriveModeSystem DriveMode = "system"
	DriveModeAll    DriveMode = "all"
	DriveModeSelect DriveMode = "select"
)

// WLANMode defines wireless profile cleanup behavior.
type WLANMode string

const (
	WLANModeAll           WLANMode = "all"
	WLANModeExceptCurrent WLANMode = "except_current"
	WLANModeSkip          WLANMode = "skip"
)

// Config is the top-level configuration for a cleanup run.
type Config struct {
	Version  int            `json:"version"`
	Targets  TargetConfig   `json:"targets"`
	Options  OptionsConfig  `json:"options"`
	Modules  ModuleConfig   `json:"modules"`
	Schedule ScheduleConfig `json:"schedule"`
}

// TargetConfig defines user and drive targeting.
type TargetConfig struct {
	UserMode        UserMode `json:"user_mode"`
	SelectedUsers   []string `json:"selected_users"`
	DriveMode       DriveMode `json:"drive_mode"`
	SelectedDrives  []string `json:"selected_drives"`
}

// OptionsConfig holds runtime options.
type OptionsConfig struct {
	RebootAfter          bool         `json:"reboot_after"`
	WLANMode             WLANMode     `json:"wlan_mode"`
	Browsers             BrowserConfig `json:"browsers"`
	ManifestDiff         bool         `json:"manifest_diff"`
	PostRunVerification  bool         `json:"post_run_verification"`
	FocusedCleanupMode   bool         `json:"focused_cleanup_mode"`
	SecondPassAfterReboot bool        `json:"second_pass_after_reboot"`
	MFTFreeSpaceScrub    bool         `json:"mft_free_space_scrub"`    // cipher/sdelete on free space + free MFT records
	LogFileResetOnReboot bool         `json:"logfile_reset_on_reboot"` // chkdsk /F schedules $LogFile replay on reboot
	LSASSScrub           bool         `json:"lsass_scrub"`              // cred manager + klist + wdigest
	LSASSRebootAfter     bool         `json:"lsass_reboot_after"`       // reboot after LSASS scrub (optional)
	Advanced             AdvancedOptions `json:"advanced"`
}

// BrowserConfig selects which browsers to clean.
type BrowserConfig struct {
	Edge    bool `json:"edge"`
	Chrome  bool `json:"chrome"`
	Firefox bool `json:"firefox"`
	Brave   bool `json:"brave"`
	Opera   bool `json:"opera"`
	Vivaldi bool `json:"vivaldi"`
}

// ModuleConfig toggles each cleanup category.
type ModuleConfig struct {
	VolatileMemory      bool `json:"volatile_memory"`
	RegistryHives       bool `json:"registry_hives"`
	NTFSMetadata        bool `json:"ntfs_metadata"`
	ProgramExecution    bool `json:"program_execution"`
	SystemLogs          bool `json:"system_logs"`
	PersistenceStorage  bool `json:"persistence_storage"`
	NetworkBrowser      bool `json:"network_browser"`
}

// DefaultConfig returns the safe logs-and-traces default profile.
func DefaultConfig() Config {
	return Config{
		Version: 1,
		Targets: TargetConfig{
			UserMode:  UserModeCurrent,
			DriveMode: DriveModeSystem,
		},
		Options: OptionsConfig{
			RebootAfter:          false,
			WLANMode:             WLANModeExceptCurrent,
			ManifestDiff:         true,
			PostRunVerification:   true,
			FocusedCleanupMode:    false,
			SecondPassAfterReboot: false,
			MFTFreeSpaceScrub:    false,
			LogFileResetOnReboot: false,
			LSASSScrub:           true,
			LSASSRebootAfter:     false,
			Advanced:             DefaultAdvancedOptions(),
			Browsers: BrowserConfig{
				Edge:    true,
				Chrome:  true,
				Firefox: true,
			},
		},
		Modules: ModuleConfig{
			VolatileMemory:     true,
			RegistryHives:      true,
			NTFSMetadata:       true,
			ProgramExecution:   true,
			SystemLogs:         true,
			PersistenceStorage: false,
			NetworkBrowser:     true,
		},
		Schedule: DefaultSchedule(),
	}
}
