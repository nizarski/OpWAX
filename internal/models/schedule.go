package models

// ScheduleMode defines when a scheduled cleanup runs.
type ScheduleMode string

const (
	ScheduleModeAtLogon  ScheduleMode = "at_logon"
	ScheduleModeDaily    ScheduleMode = "daily"
	ScheduleModeWeekly   ScheduleMode = "weekly"
	ScheduleModeMonthly  ScheduleMode = "monthly"
	ScheduleModeOnce     ScheduleMode = "once"
)

// ScheduleConfig configures Windows Task Scheduler integration.
type ScheduleConfig struct {
	Enabled    bool         `json:"enabled"`
	Mode       ScheduleMode `json:"mode"`
	Time       string       `json:"time"`         // HH:MM (24h) for daily/weekly/monthly/once
	DayOfWeek  int          `json:"day_of_week"`  // 0=Sunday … 6=Saturday (weekly)
	DayOfMonth int          `json:"day_of_month"` // 1–28 (monthly)
	ConfigPath string       `json:"config_path"`  // config JSON used by scheduled run
}

// DefaultSchedule returns disabled schedule defaults.
func DefaultSchedule() ScheduleConfig {
	return ScheduleConfig{
		Enabled:    false,
		Mode:       ScheduleModeDaily,
		Time:       "02:00",
		DayOfWeek:  0,
		DayOfMonth: 1,
	}
}
