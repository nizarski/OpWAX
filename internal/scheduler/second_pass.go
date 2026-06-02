package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opwax/opwax/internal/util"
)

const secondPassTaskName = "OpWAX-SecondPass"

// ScheduleSecondPass creates a one-time logon task (/Z auto-delete) for post-reboot cleanup.
func ScheduleSecondPass(exePath, configPath string) error {
	if configPath == "" {
		return fmt.Errorf("config_path required for second pass")
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
	cliPath := strings.TrimSuffix(absExe, filepath.Ext(absExe)) + "-cli.exe"
	if _, err := os.Stat(cliPath); err == nil {
		absExe = cliPath
	}
	tr := fmt.Sprintf(`"%s" -no-gui -config "%s"`, absExe, absCfg)
	st := time.Now().Add(2 * time.Minute).Format("15:04")
	return util.RunHidden("schtasks", "/create", "/tn", secondPassTaskName,
		"/tr", tr, "/sc", "ONLOGON", "/rl", "HIGHEST", "/f", "/Z", "/st", st)
}

// UninstallSecondPass removes the one-time second-pass task if present.
func UninstallSecondPass() error {
	return util.RunHidden("schtasks", "/delete", "/tn", secondPassTaskName, "/f")
}

// SecondPassScheduled reports whether a second-pass task exists.
func SecondPassScheduled() bool {
	_, err := util.RunHiddenOutput("schtasks", "/query", "/tn", secondPassTaskName)
	return err == nil
}
