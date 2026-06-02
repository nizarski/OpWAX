package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"
)

const taskName = "OpWAX-PrivacyCleanup"

// Status describes the installed scheduled task.
type Status struct {
	Installed bool   `json:"installed"`
	Detail    string `json:"detail"`
}

// Install creates or updates the Windows scheduled task.
func Install(exePath, configPath string, sched models.ScheduleConfig) error {
	if err := models.ValidateSchedule(sched); err != nil {
		return err
	}
	if configPath == "" {
		configPath = sched.ConfigPath
	}
	if configPath == "" {
		return fmt.Errorf("config_path required for scheduled runs")
	}
	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}
	absCfg, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absCfg); err != nil {
		return fmt.Errorf("config file not found: %s", absCfg)
	}

	// Use CLI binary if present alongside GUI exe.
	cliPath := strings.TrimSuffix(absExe, filepath.Ext(absExe)) + "-cli.exe"
	if _, err := os.Stat(cliPath); err == nil {
		absExe = cliPath
	}

	tr := fmt.Sprintf(`"%s" -no-gui -config "%s"`, absExe, absCfg)

	args := []string{"/create", "/tn", taskName, "/tr", tr, "/rl", "HIGHEST", "/f"}

	switch sched.Mode {
	case models.ScheduleModeAtLogon:
		args = append(args, "/sc", "ONLOGON")
	case models.ScheduleModeDaily:
		args = append(args, "/sc", "DAILY", "/st", sched.Time)
	case models.ScheduleModeWeekly:
		dow := mapDayOfWeek(sched.DayOfWeek)
		args = append(args, "/sc", "WEEKLY", "/d", dow, "/st", sched.Time)
	case models.ScheduleModeMonthly:
		args = append(args, "/sc", "MONTHLY", "/mo", "FIRST", "/d", fmt.Sprintf("%d", clampDay(sched.DayOfMonth)), "/st", sched.Time)
	case models.ScheduleModeOnce:
		date := time.Now().Format("01/02/2006")
		args = append(args, "/sc", "ONCE", "/sd", date, "/st", sched.Time)
	default:
		return fmt.Errorf("invalid schedule mode: %q", sched.Mode)
	}

	return util.RunHidden("schtasks", args...)
}

// Uninstall removes the scheduled task.
func Uninstall() error {
	return util.RunHidden("schtasks", "/delete", "/tn", taskName, "/f")
}

// GetStatus queries whether the task exists.
func GetStatus() (Status, error) {
	out, err := util.RunHiddenOutput("schtasks", "/query", "/tn", taskName, "/fo", "LIST", "/v")
	if err != nil {
		if strings.Contains(out, "ERROR") || strings.Contains(strings.ToLower(out), "not found") {
			return Status{Installed: false}, nil
		}
		return Status{}, fmt.Errorf("schtasks query: %s", strings.TrimSpace(out))
	}
	return Status{Installed: true, Detail: strings.TrimSpace(out)}, nil
}

func mapDayOfWeek(d int) string {
	days := []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	if d < 0 || d > 6 {
		d = 0
	}
	return days[d]
}

func clampDay(d int) int {
	if d < 1 {
		return 1
	}
	if d > 28 {
		return 28
	}
	return d
}

// ValidateSchedule checks schedule config fields.
func ValidateSchedule(s models.ScheduleConfig) error {
	return models.ValidateSchedule(s)
}

// TaskName returns the Windows task name.
func TaskName() string { return taskName }
