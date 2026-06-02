package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/opwax/opwax/internal/models"
)

// Load reads config from a JSON file.
func Load(path string) (models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return models.Config{}, err
	}
	return Parse(data)
}

// Parse unmarshals config JSON and validates.
func Parse(data []byte) (models.Config, error) {
	var cfg models.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return models.Config{}, fmt.Errorf("invalid config JSON: %w", err)
	}
	normalizeAdvancedOptions(&cfg.Options.Advanced)
	if err := Validate(cfg); err != nil {
		return models.Config{}, err
	}
	return cfg, nil
}

// Save writes config to a JSON file.
func Save(path string, cfg models.Config) error {
	if err := Validate(cfg); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Validate checks config values.
func Validate(cfg models.Config) error {
	switch cfg.Targets.UserMode {
	case models.UserModeCurrent, models.UserModeAll, models.UserModeSelect:
	default:
		return fmt.Errorf("invalid user_mode: %q", cfg.Targets.UserMode)
	}
	if cfg.Targets.UserMode == models.UserModeSelect && len(cfg.Targets.SelectedUsers) == 0 {
		return fmt.Errorf("selected_users required when user_mode is select")
	}

	switch cfg.Targets.DriveMode {
	case models.DriveModeSystem, models.DriveModeAll, models.DriveModeSelect:
	default:
		return fmt.Errorf("invalid drive_mode: %q", cfg.Targets.DriveMode)
	}
	if cfg.Targets.DriveMode == models.DriveModeSelect && len(cfg.Targets.SelectedDrives) == 0 {
		return fmt.Errorf("selected_drives required when drive_mode is select")
	}

	switch cfg.Options.WLANMode {
	case models.WLANModeAll, models.WLANModeExceptCurrent, models.WLANModeSkip:
	default:
		return fmt.Errorf("invalid wlan_mode: %q", cfg.Options.WLANMode)
	}
	return models.ValidateSchedule(cfg.Schedule)
}

func normalizeAdvancedOptions(adv *models.AdvancedOptions) {
	if adv.OneDrive == "" {
		adv.OneDrive = "off"
	}
	if adv.CloudSync == "" {
		adv.CloudSync = "off"
	}
	if adv.WMI == "" {
		adv.WMI = "off"
	}
	if adv.ScheduledTasks == "" {
		adv.ScheduledTasks = "off"
	}
	if adv.BITS == "" {
		adv.BITS = "off"
	}
	if adv.AlternateRunKeys == "" {
		adv.AlternateRunKeys = "off"
	}
	if adv.RDPCache == "" {
		adv.RDPCache = "off"
	}
	if adv.HyperV == "" {
		adv.HyperV = "off"
	}
	if adv.WSL == "" {
		adv.WSL = "off"
	}
	if adv.Outlook == "" {
		adv.Outlook = "off"
	}
	if adv.Teams == "" {
		adv.Teams = "off"
	}
}

// ToJSON returns formatted JSON for manual editing.
func ToJSON(cfg models.Config) (string, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
