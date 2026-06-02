package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/models"
)

type moduleRow struct {
	id         string
	check      *widget.Check
	rebootHint *rebootHintIcon
}

func newModuleRow(win fyne.Window, id, label string, cfg models.Config, onChange func(bool)) *moduleRow {
	row := &moduleRow{
		id:    id,
		check: widget.NewCheck(label, onChange),
	}
	if models.ModuleCanRequireReboot(id) {
		row.rebootHint = newRebootHintIcon(win.Canvas(), win, models.ModuleRebootTooltip(id, cfg))
	}
	return row
}

func (r *moduleRow) updateRebootTooltip(cfg models.Config) {
	if r.rebootHint != nil {
		r.rebootHint.setTooltip(models.ModuleRebootTooltip(r.id, cfg))
	}
}

func (r *moduleRow) canvasObject() fyne.CanvasObject {
	var right fyne.CanvasObject
	if r.rebootHint != nil {
		right = r.rebootHint
	}
	row := container.NewBorder(nil, nil, nil, right, r.check)
	return container.NewPadded(row)
}

func updateModuleRebootTooltips(rows []*moduleRow, cfg models.Config) {
	for _, row := range rows {
		row.updateRebootTooltip(cfg)
	}
}
