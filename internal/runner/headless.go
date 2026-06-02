package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opwax/opwax/internal/config"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/orchestrator"
	"github.com/opwax/opwax/internal/preflight"
	"github.com/opwax/opwax/internal/scheduler"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/version"
)

// RunManifest generates and saves or prints a system manifest.
func RunManifest(outPath string) {
	m, err := system.GenerateManifest()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if outPath == "-" {
		json, err := system.ManifestJSON(m)
		if err != nil {
			os.Exit(1)
		}
		fmt.Print(json)
		return
	}
	if err := system.SaveManifest(m, outPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// RunInstallSchedule registers the Windows scheduled task.
func RunInstallSchedule(configPath string) {
	if configPath == "" {
		fmt.Fprintln(os.Stderr, "-config required with -install-schedule")
		os.Exit(1)
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg.Schedule.Enabled = true
	if cfg.Schedule.ConfigPath == "" {
		cfg.Schedule.ConfigPath = configPath
	}
	exe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	if err := scheduler.Install(exe, cfg.Schedule.ConfigPath, cfg.Schedule); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Scheduled task installed:", scheduler.TaskName())
}

// RunUninstallSchedule removes the scheduled task.
func RunUninstallSchedule() {
	if err := scheduler.Uninstall(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Scheduled task removed")
}

// RunScheduleStatus prints scheduled task status.
func RunScheduleStatus() {
	st, err := scheduler.GetStatus()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if st.Installed {
		fmt.Println("INSTALLED")
		fmt.Println(st.Detail)
	} else {
		fmt.Println("NOT INSTALLED")
	}
}

// RunHeadless executes cleanup or dry-run from the CLI.
func RunHeadless(configPath string, dryRunOnly, preflightOnly bool) {
	cfg := models.DefaultConfig()
	if configPath != "" {
		loaded, err := config.Load(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cfg = loaded
	} else {
		exe, _ := os.Executable()
		defaultCfg := filepath.Join(filepath.Dir(exe), "configs", "default.json")
		if loaded, err := config.Load(defaultCfg); err == nil {
			cfg = loaded
		}
	}

	orch := orchestrator.New()
	ctx, err := orch.BuildContext(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(version.Banner())
	pf := orch.Preflight(ctx)
	fmt.Print(preflight.FormatReport(pf))

	if preflightOnly {
		if !pf.CanProceed {
			os.Exit(2)
		}
		os.Exit(0)
	}

	report := orch.DryRun(ctx)
	fmt.Print("\n--- PLANNED ACTIONS ---\n\n")
	fmt.Print(orchestrator.FormatDryRun(report))

	if dryRunOnly {
		os.Exit(0)
	}

	if !pf.CanProceed {
		fmt.Fprintln(os.Stderr, "pre-flight failed")
		os.Exit(2)
	}

	execReport := orch.Execute(ctx)
	fmt.Print(orchestrator.FormatReport(execReport))
}
