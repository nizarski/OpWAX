package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/version"
)

func pad(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.NewPadded(obj)
}

func scrollTab(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewScroll(pad(content))
}

func sectionCard(title, subtitle string, body fyne.CanvasObject) *widget.Card {
	return widget.NewCard(title, subtitle, pad(body))
}

func fieldLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

func hintLabel(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	return l
}

func hintWithIcon(icon fyne.Resource, text string) fyne.CanvasObject {
	ic := widget.NewIcon(icon)
	lbl := hintLabel(text)
	return container.NewBorder(nil, nil, container.NewPadded(ic), nil, lbl)
}

func primaryButton(text string, tapped func()) *widget.Button {
	b := widget.NewButton(text, tapped)
	b.Importance = widget.HighImportance
	return b
}

func dangerButton(text string, tapped func()) *widget.Button {
	b := widget.NewButton(text, tapped)
	b.Importance = widget.DangerImportance
	return b
}

func tabWithIcon(name string, icon fyne.Resource, content fyne.CanvasObject) *container.TabItem {
	return container.NewTabItemWithIcon(name, icon, scrollTab(content))
}

func mainHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(version.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Disable future logging, then remove existing forensic artifacts")
	subtitle.Wrapping = fyne.TextWrapWord
	meta := widget.NewLabel(fmt.Sprintf("v%s · %s · Administrator", version.Version, version.Author))
	meta.Alignment = fyne.TextAlignTrailing

	left := container.NewVBox(title, subtitle)
	row := container.NewBorder(nil, nil, nil, meta, left)
	return container.NewVBox(row, widget.NewSeparator())
}

func statusFooter(status *widget.Label) fyne.CanvasObject {
	status.Wrapping = fyne.TextWrapWord
	icon := widget.NewIcon(theme.InfoIcon())
	body := container.NewBorder(nil, nil, icon, nil, status)
	return container.NewVBox(widget.NewSeparator(), body)
}

func configStatusSummary(cfg models.Config) string {
	var mods []string
	if cfg.Modules.VolatileMemory {
		mods = append(mods, "Volatile")
	}
	if cfg.Modules.RegistryHives {
		mods = append(mods, "Registry")
	}
	if cfg.Modules.NTFSMetadata {
		mods = append(mods, "NTFS")
	}
	if cfg.Modules.ProgramExecution {
		mods = append(mods, "Execution")
	}
	if cfg.Modules.SystemLogs {
		mods = append(mods, "Logs")
	}
	if cfg.Modules.PersistenceStorage {
		mods = append(mods, "Persistence")
	}
	if cfg.Modules.NetworkBrowser {
		mods = append(mods, "Network")
	}
	if len(mods) == 0 {
		mods = append(mods, "none")
	}

	return fmt.Sprintf(
		"Targets: user=%s, drive=%s  ·  Modules: %s  ·  WLAN=%s",
		cfg.Targets.UserMode,
		cfg.Targets.DriveMode,
		strings.Join(mods, ", "),
		cfg.Options.WLANMode,
	)
}

func setAllModules(cfg *models.Config, rows []*moduleRow, enabled bool) {
	cfg.Modules.VolatileMemory = enabled
	cfg.Modules.RegistryHives = enabled
	cfg.Modules.NTFSMetadata = enabled
	cfg.Modules.ProgramExecution = enabled
	cfg.Modules.SystemLogs = enabled
	cfg.Modules.PersistenceStorage = enabled
	cfg.Modules.NetworkBrowser = enabled
	for _, row := range rows {
		row.check.SetChecked(enabled)
	}
}
