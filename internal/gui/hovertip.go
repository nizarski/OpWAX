package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	_ fyne.Widget       = (*rebootHintIcon)(nil)
	_ desktop.Hoverable = (*rebootHintIcon)(nil)
	_ fyne.Tappable     = (*rebootHintIcon)(nil)
)

type rebootHintIcon struct {
	widget.BaseWidget
	tooltip string
	canvas  fyne.Canvas
	parent  fyne.Window
	popup   *widget.PopUp
	icon    *widget.Icon
}

func newRebootHintIcon(canvas fyne.Canvas, parent fyne.Window, tooltip string) *rebootHintIcon {
	r := &rebootHintIcon{
		tooltip: tooltip,
		canvas:  canvas,
		parent:  parent,
		icon:    widget.NewIcon(theme.ViewRefreshIcon()),
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *rebootHintIcon) setTooltip(text string) {
	r.tooltip = text
}

func (r *rebootHintIcon) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.icon)
}

func (r *rebootHintIcon) MinSize() fyne.Size {
	return r.icon.MinSize()
}

func (r *rebootHintIcon) MouseIn(*desktop.MouseEvent) {
	if r.tooltip == "" || r.canvas == nil {
		return
	}
	lbl := widget.NewLabel(r.tooltip)
	lbl.Wrapping = fyne.TextWrapWord
	lbl.Resize(fyne.NewSize(280, lbl.MinSize().Height))
	r.popup = widget.NewPopUp(container.NewPadded(lbl), r.canvas)
	r.popup.ShowAtRelativePosition(fyne.NewPos(-140, r.Size().Height+4), r)
}

func (r *rebootHintIcon) MouseMoved(*desktop.MouseEvent) {}

func (r *rebootHintIcon) MouseOut() {
	if r.popup != nil {
		r.popup.Hide()
		r.popup = nil
	}
}

func (r *rebootHintIcon) Tapped(*fyne.PointEvent) {
	if r.tooltip == "" || r.parent == nil {
		return
	}
	dialog.ShowInformation("Reboot may be required", r.tooltip, r.parent)
}
