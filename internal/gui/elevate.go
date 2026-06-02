package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/version"
)

// RunElevatePrompt shows a minimal window when not running as admin.
func RunElevatePrompt() {
	a := app.NewWithID("com.opwax.privacy")
	w := a.NewWindow(version.Name + " - Elevation Required")
	w.Resize(fyne.NewSize(480, 220))
	w.CenterOnScreen()

	icon := widget.NewIcon(theme.WarningIcon())
	msg := widget.NewRichTextFromMarkdown(
		"**Administrator privileges are required.**\n\n" +
			"OpWAX must run elevated to modify system artifacts, services, and registry keys.\n\n" +
			"If you declined the UAC prompt, click below to try again.",
	)
	msg.Wrapping = fyne.TextWrapWord

	elevateBtn := primaryButton("Try Again as Administrator", func() {
		if err := system.ElevateSelf(); err != nil {
			dialog.ShowError(err, w)
		}
	})
	closeBtn := widget.NewButton("Close", func() { w.Close() })

	card := widget.NewCard(
		"Elevation required",
		"Approve the UAC prompt when it appears.",
		container.NewVBox(
			container.NewHBox(icon, msg),
			container.NewHBox(elevateBtn, closeBtn),
			hintLabel("Tip: right-click the OpWAX executable → Run as administrator"),
		),
	)

	w.SetContent(pad(card))
	w.ShowAndRun()
}
