package gui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func exportTextToFile(parent fyne.Window, content, defaultName string) {
	if content == "" {
		dialog.ShowInformation("Empty", "Nothing to export.", parent)
		return
	}
	dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		defer uc.Close()
		if _, err := uc.Write([]byte(content)); err != nil {
			dialog.ShowError(err, parent)
		}
	}, parent)
}

func defaultReportFilename() string {
	return "OpWAX-report-" + time.Now().Format("20060102-150405") + ".txt"
}
