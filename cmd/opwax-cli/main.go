package main

import (
	"flag"

	"github.com/opwax/opwax/internal/runner"
)

func main() {
	configPath := flag.String("config", "", "path to config JSON")
	dryRunOnly := flag.Bool("dry-run", false, "preview actions without executing")
	preflightOnly := flag.Bool("preflight", false, "run pre-flight checks only")
	manifestOut := flag.String("manifest", "", "generate system manifest JSON (use - for stdout)")
	installSchedule := flag.Bool("install-schedule", false, "install Windows scheduled task from config")
	uninstallSchedule := flag.Bool("uninstall-schedule", false, "remove Windows scheduled task")
	scheduleStatus := flag.Bool("schedule-status", false, "print scheduled task status")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		runner.PrintVersion()
	}

	switch {
	case *manifestOut != "":
		runner.RunManifest(*manifestOut)
	case *uninstallSchedule:
		runner.RunUninstallSchedule()
	case *scheduleStatus:
		runner.RunScheduleStatus()
	case *installSchedule:
		runner.RunInstallSchedule(*configPath)
	default:
		runner.RunHeadless(*configPath, *dryRunOnly, *preflightOnly)
	}
}
