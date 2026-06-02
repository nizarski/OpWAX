package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/version"
)

var moduleLabels = map[string]string{
	"bootstrap":           "Priority bootstrap",
	"volatile_memory":     "Volatile Memory",
	"registry_hives":      "Registry Hives",
	"ntfs_metadata":       "NTFS Metadata",
	"program_execution":   "Program Execution",
	"system_logs":         "System Logs",
	"persistence_storage": "Persistence Storage",
	"network_browser":     "Network & Browser",
}

type moduleProgress struct {
	label   *widget.Label
	reboot  *widget.Icon
	disable *widget.ProgressBar
	clean   *widget.ProgressBar
}

func newModuleProgress(name string, cfg models.Config) *moduleProgress {
	display := moduleLabels[name]
	if display == "" {
		display = name
	}
	mp := &moduleProgress{
		label:   widget.NewLabelWithStyle(display, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		reboot:  widget.NewIcon(theme.ViewRefreshIcon()),
		disable: widget.NewProgressBar(),
		clean:   widget.NewProgressBar(),
	}
	if !models.ModuleNeedsRebootHint(name, cfg) {
		mp.reboot.Hide()
	}
	return mp
}

func (m *moduleProgress) titleRow() fyne.CanvasObject {
	return container.NewBorder(nil, nil, nil, m.reboot, m.label)
}

func (m *moduleProgress) container() fyne.CanvasObject {
	return sectionCard("", "", container.NewVBox(
		m.titleRow(),
		fieldLabel("Disable"),
		m.disable,
		fieldLabel("Clean"),
		m.clean,
	))
}

func (m *moduleProgress) reset() {
	m.disable.SetValue(0)
	m.clean.SetValue(0)
}

type progressTracker struct {
	overall *widget.ProgressBar
	status  *widget.Label
	modules map[string]*moduleProgress
	order   []string
}

func newProgressTracker(moduleNames []string, cfg models.Config) *progressTracker {
	pt := &progressTracker{
		overall: widget.NewProgressBar(),
		status:  widget.NewLabel("Starting…"),
		modules: make(map[string]*moduleProgress),
	}
	for _, n := range moduleNames {
		pt.modules[n] = newModuleProgress(n, cfg)
		pt.order = append(pt.order, n)
	}
	return pt
}

func (pt *progressTracker) content(onCancel func()) fyne.CanvasObject {
	header := widget.NewLabelWithStyle(version.Name+" - Cleanup in progress", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	pt.status.Wrapping = fyne.TextWrapWord

	var rows []fyne.CanvasObject
	rows = append(rows, header, widget.NewSeparator())
	rows = append(rows, fieldLabel("Overall progress"), pt.overall, pt.status, widget.NewSeparator())
	for _, n := range pt.order {
		if mp, ok := pt.modules[n]; ok {
			rows = append(rows, mp.container())
		}
	}
	if onCancel != nil {
		cancelBtn := dangerButton("Cancel Cleanup", onCancel)
		rows = append(rows, widget.NewSeparator(), cancelBtn, hintLabel("Cancel takes effect after the current step finishes."))
	}
	return scrollTab(container.NewVBox(rows...))
}

func (pt *progressTracker) apply(update models.ProgressUpdate) {
	if update.Module != "" {
		if mp, ok := pt.modules[update.Module]; ok {
			switch update.Phase {
			case models.ProgressPhaseDisable, models.ProgressPhaseBootstrap:
				if update.Complete {
					mp.disable.SetValue(1)
				} else {
					mp.disable.SetValue(0.3)
				}
			case models.ProgressPhaseClean:
				if update.Complete {
					mp.clean.SetValue(1)
				} else {
					mp.clean.SetValue(0.3)
				}
			case models.ProgressPhaseSecure:
				if update.Complete {
					mp.clean.SetValue(1)
				} else {
					mp.clean.SetValue(0.5)
				}
			}
		}
	}
	if update.ModuleTotal > 0 && update.ModuleIndex > 0 {
		base := float64(update.ModuleIndex-1) / float64(update.ModuleTotal)
		if update.Complete {
			base = float64(update.ModuleIndex) / float64(update.ModuleTotal)
		}
		pt.overall.SetValue(base)
	}
	phase := string(update.Phase)
	if update.Module != "" {
		label := moduleLabels[update.Module]
		if label == "" {
			label = update.Module
		}
		pt.status.SetText(strings.ToUpper(phase) + ": " + label + " - " + update.Message)
	} else {
		pt.status.SetText(update.Message)
	}
}

func moduleNamesFromConfig(cfg models.ModuleConfig) []string {
	var names []string
	if cfg.VolatileMemory {
		names = append(names, "volatile_memory")
	}
	if cfg.RegistryHives {
		names = append(names, "registry_hives")
	}
	if cfg.NTFSMetadata {
		names = append(names, "ntfs_metadata")
	}
	if cfg.ProgramExecution {
		names = append(names, "program_execution")
	}
	if cfg.SystemLogs {
		names = append(names, "system_logs")
	}
	if cfg.PersistenceStorage {
		names = append(names, "persistence_storage")
	}
	if cfg.NetworkBrowser {
		names = append(names, "network_browser")
	}
	return names
}
