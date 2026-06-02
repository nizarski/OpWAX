package models

import "fmt"

// ValidateSchedule checks schedule config fields.
func ValidateSchedule(s ScheduleConfig) error {
	if !s.Enabled {
		return nil
	}
	if s.ConfigPath == "" {
		return fmt.Errorf("schedule.config_path required when schedule is enabled")
	}
	switch s.Mode {
	case ScheduleModeAtLogon:
		return nil
	case ScheduleModeDaily, ScheduleModeWeekly, ScheduleModeMonthly, ScheduleModeOnce:
		if s.Time == "" {
			return fmt.Errorf("schedule.time required (HH:MM)")
		}
		return nil
	default:
		return fmt.Errorf("invalid schedule.mode: %q", s.Mode)
	}
}
