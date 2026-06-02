package gui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// applyAppIcon sets the Fyne app/window icon from assets/opwax.png when present.
func applyAppIcon(a fyne.App, w fyne.Window) {
	paths := iconSearchPaths()
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		res := fyne.NewStaticResource("opwax.png", data)
		a.SetIcon(res)
		w.SetIcon(res)
		return
	}
}

func iconSearchPaths() []string {
	var paths []string
	if exe, err := os.Executable(); err == nil {
		root := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(root, "assets", "opwax.png"),
			filepath.Join(root, "opwax.png"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, "assets", "opwax.png"))
	}
	return paths
}
