package modules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

const persistenceName = "persistence_storage"

// PersistenceModule handles pagefile, dumps, hiberfil, recycle bin.
type PersistenceModule struct{}

func (m *PersistenceModule) Name() string { return persistenceName }

func (m *PersistenceModule) DryRun(ctx *models.RunContext) []models.Action {
	var actions []models.Action
	actions = append(actions,
		action(persistenceName, "Disable crash dumps", "CrashControl", models.ActionDisable),
		action(persistenceName, "Clear pagefile at shutdown + disable pagefile", "Memory Management", models.ActionDisable),
		action(persistenceName, "Disable hibernation", "powercfg", models.ActionDisable),
		action(persistenceName, "Secure-delete MEMORY.DMP and minidumps", system.WindowsDirectory(), models.ActionSecure),
	)
	for _, d := range ctx.TargetDrives {
		actions = append(actions,
			action(persistenceName, "Secure-delete recycle bin on "+d.Letter, d.Root+"$Recycle.Bin", models.ActionSecure),
		)
	}
	actions = append(actions, advancedPersistenceDryRun(ctx)...)
	actions = append(actions, gapPersistenceDryRun(ctx)...)
	return actions
}

func (m *PersistenceModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	crashPath := `SYSTEM\CurrentControlSet\Control\CrashControl`
	a1 := action(persistenceName, "Disable crash dumps", crashPath, models.ActionDisable)
	_ = util.SetRegDWORD(registry.LOCAL_MACHINE, crashPath, "CrashDumpEnabled", 0)
	_ = util.SetRegDWORD(registry.LOCAL_MACHINE, crashPath, "LogEvent", 0)
	results = append(results, result(a1, nil))

	memPath := `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`
	a2 := action(persistenceName, "Clear pagefile at shutdown", memPath, models.ActionDisable)
	_ = util.SetRegDWORD(registry.LOCAL_MACHINE, memPath, "ClearPageFileAtShutdown", 1)
	// Disable pagefile
	pagePath := `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`
	_ = util.SetRegString(registry.LOCAL_MACHINE, pagePath, "PagingFiles", "")
	results = append(results, result(a2, util.RunHidden("wmic", "computersystem", "set", "AutomaticManagedPagefile=False")))

	a3 := action(persistenceName, "Disable hibernation", "powercfg", models.ActionDisable)
	results = append(results, result(a3, util.RunHidden("powercfg", "/hibernate", "off")))

	results = append(results, advancedPersistenceDisable(ctx)...)
	return results
}

func (m *PersistenceModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	winDir := system.WindowsDirectory()
	dumpFiles := []string{
		filepath.Join(winDir, "MEMORY.DMP"),
	}
	for _, f := range dumpFiles {
		a := action(persistenceName, "Secure-delete "+f, f, models.ActionSecure)
		results = append(results, result(a, util.SecureDelete(f)))
	}

	minidump := filepath.Join(winDir, "Minidump")
	a := action(persistenceName, "Secure-delete minidumps", minidump, models.ActionSecure)
	_, err := util.SecureDeleteDirFiles(minidump)
	results = append(results, result(a, err))

	// hiberfil.sys on system drive
	for _, d := range ctx.TargetDrives {
		if d.IsSystem {
			hiber := d.Root + "hiberfil.sys"
			a2 := action(persistenceName, "Remove hiberfil.sys", hiber, models.ActionClean)
			results = append(results, result(a2, util.SecureDelete(hiber)))
		}
		pagefile := d.Root + "pagefile.sys"
		a3 := action(persistenceName, "Remove pagefile.sys", pagefile, models.ActionSecure)
		results = append(results, result(a3, util.SecureDelete(pagefile)))
	}

	// Recycle bin secure overwrite
	for _, d := range ctx.TargetDrives {
		rb := filepath.Join(d.Root, "$Recycle.Bin")
		a4 := action(persistenceName, "Secure-delete recycle bin "+d.Letter, rb, models.ActionSecure)
		results = append(results, result(a4, secureDeleteRecycleBin(rb)))
	}

	results = append(results, advancedPersistenceClean(ctx)...)
	results = append(results, gapPersistenceClean(ctx)...)
	return results
}

func secureDeleteRecycleBin(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, sidDir := range entries {
		if !sidDir.IsDir() {
			continue
		}
		sidPath := filepath.Join(root, sidDir.Name())
		files, _ := os.ReadDir(sidPath)
		for _, f := range files {
			name := f.Name()
			full := filepath.Join(sidPath, name)
			if strings.HasPrefix(name, "$R") {
				_ = util.SecureDelete(full)
			} else if strings.HasPrefix(name, "$I") {
				_ = os.Remove(full)
			}
		}
	}
	return nil
}
