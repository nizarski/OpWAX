package main

import (
	"flag"
	"os"
	"os/exec"

	"github.com/opwax/opwax/internal/gui"
	"github.com/opwax/opwax/internal/runner"
)

func main() {
	configPath := flag.String("config", "", "path to config JSON (delegates to CLI)")
	dryRunOnly := flag.Bool("dry-run", false, "preview actions without executing")
	preflightOnly := flag.Bool("preflight", false, "run pre-flight checks only")
	noGUI := flag.Bool("no-gui", false, "force headless mode")
	noAutoElevate := flag.Bool("no-auto-elevate", false, "do not automatically request UAC elevation on GUI launch")
	manifestOut := flag.String("manifest", "", "generate system manifest JSON")
	installSchedule := flag.Bool("install-schedule", false, "install Windows scheduled task")
	uninstallSchedule := flag.Bool("uninstall-schedule", false, "remove Windows scheduled task")
	scheduleStatus := flag.Bool("schedule-status", false, "print scheduled task status")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		runner.PrintVersion()
	}

	if *manifestOut != "" {
		runner.RunManifest(*manifestOut)
		return
	}
	if *uninstallSchedule {
		runner.RunUninstallSchedule()
		return
	}
	if *scheduleStatus {
		runner.RunScheduleStatus()
		return
	}
	if *installSchedule {
		runner.RunInstallSchedule(*configPath)
		return
	}
	if *configPath != "" || *noGUI || *dryRunOnly || *preflightOnly {
		runHeadlessMode(*configPath, *dryRunOnly, *preflightOnly)
		return
	}

	gui.NoAutoElevate = *noAutoElevate
	gui.Run()
}

func runHeadlessMode(configPath string, dryRun, preflight bool) {
	exe, err := os.Executable()
	if err == nil {
		cli := exe[:len(exe)-len(".exe")] + "-cli.exe"
		if _, err := os.Stat(cli); err == nil {
			cmd := exec.Command(cli, os.Args[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					os.Exit(exit.ExitCode())
				}
				os.Exit(1)
			}
			return
		}
	}
	runner.RunHeadless(configPath, dryRun, preflight)
}
