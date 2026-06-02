package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/config"
	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/orchestrator"
	"github.com/opwax/opwax/internal/preflight"
	"github.com/opwax/opwax/internal/scheduler"
	"github.com/opwax/opwax/internal/system"
	"github.com/opwax/opwax/internal/verify"
	"github.com/opwax/opwax/internal/version"
)

// Run starts the OpWAX GUI application.
func Run() {
	if !system.IsAdmin() {
		if !NoAutoElevate {
			_ = system.ElevateSelf()
		}
		RunElevatePrompt()
		return
	}

	application := app.NewWithID("com.opwax.privacy")
	app.SetMetadata(fyne.AppMetadata{
		Name:    version.Name,
		Version: version.Version,
	})
	w := application.NewWindow(version.Name + " - built by " + version.Author)
	w.Resize(fyne.NewSize(960, 740))
	w.CenterOnScreen()
	applyAppIcon(application, w)

	cfg := models.DefaultConfig()
	statusLabel := widget.NewLabel(configStatusSummary(cfg))
	var syncConfig func()
	syncConfig = func() {}
	refreshStatus := func() {
		statusLabel.SetText(configStatusSummary(cfg))
	}
	orch := orchestrator.New()
	exePath, _ := os.Executable()

	// --- Target: Users ---
	userModeSelect := widget.NewSelect([]string{"current", "all", "select"}, func(s string) {
		cfg.Targets.UserMode = models.UserMode(s)
		syncConfig()
	})
	userModeSelect.SetSelected("current")

	userSelectBox := container.NewVBox()
	driveSelectBox := container.NewVBox()

	var userChecks []*widget.Check

	refreshUsers := func() {
		users, err := system.EnumerateUsers()
		if err != nil {
			userChecks = nil
			return
		}
		userChecks = nil
		for _, u := range users {
			name := u
			chk := widget.NewCheck(name, func(v bool) {
				if v {
					cfg.Targets.SelectedUsers = appendUnique(cfg.Targets.SelectedUsers, name)
				} else {
					cfg.Targets.SelectedUsers = removeString(cfg.Targets.SelectedUsers, name)
				}
				syncConfig()
			})
			userChecks = append(userChecks, chk)
		}
	}

	driveModeSelect := widget.NewSelect([]string{"system", "all", "select"}, func(s string) {
		cfg.Targets.DriveMode = models.DriveMode(s)
		syncConfig()
	})
	driveModeSelect.SetSelected("system")

	var driveChecks []*widget.Check

	refreshDrives := func() {
		drives, err := system.EnumerateDrives()
		if err != nil {
			return
		}
		driveChecks = nil
		for _, d := range drives {
			letter := d.Letter
			label := letter
			if d.IsSystem {
				label += " (system)"
			}
			chk := widget.NewCheck(label, func(v bool) {
				if v {
					cfg.Targets.SelectedDrives = appendUnique(cfg.Targets.SelectedDrives, letter)
				} else {
					cfg.Targets.SelectedDrives = removeString(cfg.Targets.SelectedDrives, letter)
				}
				syncConfig()
			})
			driveChecks = append(driveChecks, chk)
		}
	}

	refreshUsers()
	refreshDrives()

	userModeSelect.OnChanged = func(s string) {
		cfg.Targets.UserMode = models.UserMode(s)
		userSelectBox.Objects = nil
		if s == "select" {
			for _, chk := range userChecks {
				userSelectBox.Add(chk)
			}
		}
		userSelectBox.Refresh()
		syncConfig()
	}

	driveModeSelect.OnChanged = func(s string) {
		cfg.Targets.DriveMode = models.DriveMode(s)
		driveSelectBox.Objects = nil
		if s == "select" {
			for _, chk := range driveChecks {
				driveSelectBox.Add(chk)
			}
		}
		driveSelectBox.Refresh()
		syncConfig()
	}

	// --- Module toggles ---
	rowVolatile := newModuleRow(w, "volatile_memory", "Volatile Memory (DNS, ARP, credentials)", cfg, func(v bool) {
		cfg.Modules.VolatileMemory = v
		syncConfig()
	})
	rowVolatile.check.SetChecked(cfg.Modules.VolatileMemory)
	rowRegistry := newModuleRow(w, "registry_hives", "Registry Hives (USB, Shellbags, BAM/DAM)", cfg, func(v bool) {
		cfg.Modules.RegistryHives = v
		syncConfig()
	})
	rowRegistry.check.SetChecked(cfg.Modules.RegistryHives)
	rowNTFS := newModuleRow(w, "ntfs_metadata", "NTFS Metadata (USN journal, Zone.Identifier)", cfg, func(v bool) {
		cfg.Modules.NTFSMetadata = v
		syncConfig()
	})
	rowNTFS.check.SetChecked(cfg.Modules.NTFSMetadata)
	rowExec := newModuleRow(w, "program_execution", "Program Execution (Prefetch, Amcache)", cfg, func(v bool) {
		cfg.Modules.ProgramExecution = v
		syncConfig()
	})
	rowExec.check.SetChecked(cfg.Modules.ProgramExecution)
	rowLogs := newModuleRow(w, "system_logs", "System Logs & Recent Files (.evtx, jump lists)", cfg, func(v bool) {
		cfg.Modules.SystemLogs = v
		syncConfig()
	})
	rowLogs.check.SetChecked(cfg.Modules.SystemLogs)
	rowPersist := newModuleRow(w, "persistence_storage", "Persistence Storage (pagefile, dumps, recycle bin)", cfg, func(v bool) {
		cfg.Modules.PersistenceStorage = v
		syncConfig()
	})
	rowPersist.check.SetChecked(cfg.Modules.PersistenceStorage)
	rowNetwork := newModuleRow(w, "network_browser", "Network & Browser (SRUM, WLAN, browsers)", cfg, func(v bool) {
		cfg.Modules.NetworkBrowser = v
		syncConfig()
	})
	rowNetwork.check.SetChecked(cfg.Modules.NetworkBrowser)
	moduleRows := []*moduleRow{rowVolatile, rowRegistry, rowNTFS, rowExec, rowLogs, rowPersist, rowNetwork}

	refreshModuleHints := func() {
		updateModuleRebootTooltips(moduleRows, cfg)
	}

	rebootCheck := widget.NewCheck("Reboot after cleanup (recommended for locked files)", func(v bool) {
		cfg.Options.RebootAfter = v
		syncConfig()
	})
	rebootCheck.SetChecked(cfg.Options.RebootAfter)
	manifestDiffCheck := widget.NewCheck("Compare before/after manifest (slower, shows cleanup results)", func(v bool) {
		cfg.Options.ManifestDiff = v
		syncConfig()
	})
	manifestDiffCheck.SetChecked(cfg.Options.ManifestDiff)
	verifyCheck := widget.NewCheck("Post-run verification (scan for remaining artifacts)", func(v bool) {
		cfg.Options.PostRunVerification = v
		syncConfig()
	})
	verifyCheck.SetChecked(cfg.Options.PostRunVerification)
	focusedCheck := widget.NewCheck("Focused cleanup (pause Explorer during run)", func(v bool) {
		cfg.Options.FocusedCleanupMode = v
		syncConfig()
	})
	focusedCheck.SetChecked(cfg.Options.FocusedCleanupMode)
	secondPassCheck := widget.NewCheck("Second pass after reboot (if gaps or locked files)", func(v bool) {
		cfg.Options.SecondPassAfterReboot = v
		syncConfig()
	})
	secondPassCheck.SetChecked(cfg.Options.SecondPassAfterReboot)
	mftScrubCheck := widget.NewCheck("Scrub free MFT records + free space (runs last - slow)", func(v bool) {
		cfg.Options.MFTFreeSpaceScrub = v
		syncConfig()
	})
	mftScrubCheck.SetChecked(cfg.Options.MFTFreeSpaceScrub)
	logFileCheck := widget.NewCheck("Reset $LogFile via chkdsk (on reboot for C:)", func(v bool) {
		cfg.Options.LogFileResetOnReboot = v
		refreshModuleHints()
		syncConfig()
	})
	logFileCheck.SetChecked(cfg.Options.LogFileResetOnReboot)
	lsassScrubCheck := widget.NewCheck("LSASS cache scrub (CredMgr, Kerberos, WDigest, Vault)", func(v bool) {
		cfg.Options.LSASSScrub = v
		refreshModuleHints()
		syncConfig()
	})
	lsassScrubCheck.SetChecked(cfg.Options.LSASSScrub)
	lsassRebootCheck := widget.NewCheck("Reboot after LSASS scrub (optional - full RAM clear)", func(v bool) {
		cfg.Options.LSASSRebootAfter = v
		refreshModuleHints()
		syncConfig()
	})
	lsassRebootCheck.SetChecked(cfg.Options.LSASSRebootAfter)

	wlanSelect := widget.NewSelect([]string{"all", "except_current", "skip"}, func(s string) {
		cfg.Options.WLANMode = models.WLANMode(s)
		syncConfig()
	})
	wlanSelect.SetSelected(string(cfg.Options.WLANMode))

	browserBox := container.NewVBox()
	browserStatus := widget.NewLabel("")
	browserChecks := map[system.BrowserKind]*widget.Check{}

	rebuildBrowsers := func(applyAuto bool) {
		detected := system.DetectInstalledBrowsers()
		if applyAuto {
			system.ApplyDetectedBrowsers(&cfg.Options.Browsers, detected)
		}
		browserBox.Objects = nil
		browserChecks = map[system.BrowserKind]*widget.Check{}
		if len(detected) == 0 {
			browserStatus.SetText("No supported browsers detected on this system.")
			browserBox.Add(browserStatus)
			browserBox.Refresh()
			return
		}
		browserStatus.SetText(fmt.Sprintf("Detected %d browser(s)%s.", len(detected), autoBrowserSuffix(applyAuto)))
		browserBox.Add(browserStatus)
		for _, b := range detected {
			kind := b.Kind
			label := b.Label
			if len(b.Users) > 0 {
				label += " - " + strings.Join(b.Users, ", ")
			}
			chk := widget.NewCheck(label, func(v bool) {
				system.SetBrowserConfig(&cfg.Options.Browsers, kind, v)
				syncConfig()
			})
			chk.SetChecked(isBrowserEnabled(cfg.Options.Browsers, kind))
			browserChecks[kind] = chk
			browserBox.Add(chk)
		}
		browserBox.Refresh()
	}
	rebuildBrowsers(true)

	previewText := widget.NewMultiLineEntry()
	previewText.SetPlaceHolder("Click 'Preview (Dry Run)' to see planned actions...")
	previewText.Wrapping = fyne.TextWrapWord

	preflightPreview := widget.NewRichTextFromMarkdown("")
	preflightPreview.Wrapping = fyne.TextWrapWord

	configJSON := widget.NewMultiLineEntry()
	configJSON.Wrapping = fyne.TextWrapOff

	syncConfigToUI := func() {
		userModeSelect.SetSelected(string(cfg.Targets.UserMode))
		driveModeSelect.SetSelected(string(cfg.Targets.DriveMode))
		wlanSelect.SetSelected(string(cfg.Options.WLANMode))
		rebootCheck.SetChecked(cfg.Options.RebootAfter)
		manifestDiffCheck.SetChecked(cfg.Options.ManifestDiff)
		verifyCheck.SetChecked(cfg.Options.PostRunVerification)
		focusedCheck.SetChecked(cfg.Options.FocusedCleanupMode)
		secondPassCheck.SetChecked(cfg.Options.SecondPassAfterReboot)
		mftScrubCheck.SetChecked(cfg.Options.MFTFreeSpaceScrub)
		logFileCheck.SetChecked(cfg.Options.LogFileResetOnReboot)
		lsassScrubCheck.SetChecked(cfg.Options.LSASSScrub)
		lsassRebootCheck.SetChecked(cfg.Options.LSASSRebootAfter)
		rowVolatile.check.SetChecked(cfg.Modules.VolatileMemory)
		rowRegistry.check.SetChecked(cfg.Modules.RegistryHives)
		rowNTFS.check.SetChecked(cfg.Modules.NTFSMetadata)
		rowExec.check.SetChecked(cfg.Modules.ProgramExecution)
		rowLogs.check.SetChecked(cfg.Modules.SystemLogs)
		rowPersist.check.SetChecked(cfg.Modules.PersistenceStorage)
		rowNetwork.check.SetChecked(cfg.Modules.NetworkBrowser)
		for kind, chk := range browserChecks {
			chk.SetChecked(isBrowserEnabled(cfg.Options.Browsers, kind))
		}
		refreshModuleHints()
		syncConfig()
	}

	updateConfigJSON := func() {
		s, err := config.ToJSON(cfg)
		if err == nil {
			configJSON.SetText(s)
		}
		refreshStatus()
	}
	syncConfig = updateConfigJSON
	updateConfigJSON()
	syncConfigToUI()

	// --- Manifest tab ---
	manifestText := widget.NewMultiLineEntry()
	manifestText.SetPlaceHolder("Click 'Generate Manifest' to scan this system for artifact paths…")
	manifestText.Wrapping = fyne.TextWrapWord

	generateManifestBtn := widget.NewButton("Generate Manifest", func() {
		m, err := system.GenerateManifest()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		json, err := system.ManifestJSON(m)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		manifestText.SetText(json)
	})

	saveManifestBtn := widget.NewButton("Save Manifest", func() {
		if manifestText.Text == "" {
			dialog.ShowInformation("Empty", "Generate a manifest first.", w)
			return
		}
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()
			if _, err := uc.Write([]byte(manifestText.Text)); err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
	})

	// --- Schedule tab ---
	scheduleEnabled := widget.NewCheck("Enable scheduled cleanup", func(v bool) {
		cfg.Schedule.Enabled = v
		syncConfig()
	})
	scheduleMode := widget.NewSelect([]string{"at_logon", "daily", "weekly", "monthly", "once"}, func(s string) {
		cfg.Schedule.Mode = models.ScheduleMode(s)
		syncConfig()
	})
	scheduleMode.SetSelected("daily")
	scheduleTime := widget.NewEntry()
	scheduleTime.SetText("02:00")
	scheduleTime.SetPlaceHolder("HH:MM")
	scheduleTime.OnChanged = func(s string) {
		cfg.Schedule.Time = s
		syncConfig()
	}

	dowNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	scheduleDOW := widget.NewSelect(dowNames, func(s string) {
		for i, name := range dowNames {
			if name == s {
				cfg.Schedule.DayOfWeek = i
				break
			}
		}
		syncConfig()
	})
	scheduleDOW.SetSelected("Sunday")

	scheduleDOM := widget.NewEntry()
	scheduleDOM.SetText("1")
	scheduleDOM.OnChanged = func(s string) {
		if d, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			cfg.Schedule.DayOfMonth = d
			syncConfig()
		}
	}

	scheduleConfigPath := widget.NewEntry()
	scheduleConfigPath.SetPlaceHolder(`Path to config JSON for scheduled runs`)
	defaultCfg := filepath.Join(filepath.Dir(exePath), "configs", "default.json")
	if _, err := os.Stat(defaultCfg); err == nil {
		scheduleConfigPath.SetText(defaultCfg)
		cfg.Schedule.ConfigPath = defaultCfg
	}
	scheduleConfigPath.OnChanged = func(s string) {
		cfg.Schedule.ConfigPath = s
		syncConfig()
	}

	scheduleStatus := widget.NewLabel("Schedule status: unknown")

	refreshScheduleStatus := func() {
		st, err := scheduler.GetStatus()
		if err != nil {
			scheduleStatus.SetText("Schedule status: error - " + err.Error())
			return
		}
		if st.Installed {
			scheduleStatus.SetText("Schedule status: INSTALLED (" + scheduler.TaskName() + ")")
		} else {
			scheduleStatus.SetText("Schedule status: not installed")
		}
	}
	refreshScheduleStatus()

	installScheduleBtn := widget.NewButton("Install Schedule", func() {
		cfg.Schedule.Enabled = true
		scheduleEnabled.SetChecked(true)
		cfg.Schedule.Time = scheduleTime.Text
		cfg.Schedule.ConfigPath = scheduleConfigPath.Text
		cfg.Schedule.Mode = models.ScheduleMode(scheduleMode.Selected)
		if err := models.ValidateSchedule(cfg.Schedule); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := config.Save(cfg.Schedule.ConfigPath, cfg); err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := scheduler.Install(exePath, cfg.Schedule.ConfigPath, cfg.Schedule); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Installed", "Scheduled task "+scheduler.TaskName()+" installed.", w)
		refreshScheduleStatus()
		updateConfigJSON()
	})

	removeScheduleBtn := widget.NewButton("Remove Schedule", func() {
		if err := scheduler.Uninstall(); err != nil {
			dialog.ShowError(err, w)
			return
		}
		cfg.Schedule.Enabled = false
		scheduleEnabled.SetChecked(false)
		dialog.ShowInformation("Removed", "Scheduled task removed.", w)
		refreshScheduleStatus()
		updateConfigJSON()
	})

	// --- Actions ---
	dryRunBtn := primaryButton("Preview (Dry Run)", func() {
		if err := config.Validate(cfg); err != nil {
			dialog.ShowError(err, w)
			return
		}
		ctx, err := orch.BuildContext(cfg)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		pf := orch.Preflight(ctx)
		report := orch.DryRun(ctx)
		var b strings.Builder
		b.WriteString(preflight.FormatReport(pf))
		preflightPreview.ParseMarkdown(preflight.FormatReportMarkdown(pf))
		b.WriteString("\n--- PLANNED ACTIONS ---\n\n")
		b.WriteString(orchestrator.FormatDryRun(report))
		previewText.SetText(b.String())
	})

	exportDryRunBtn := widget.NewButton("Export Preview", func() {
		exportTextToFile(w, version.Banner()+"\n\n"+previewText.Text, "OpWAX-preview.txt")
	})

	var lastExecutionSummary string
	exportReportBtn := widget.NewButton("Export Report", func() {
		body := lastExecutionSummary
		if body != "" {
			body = version.Banner() + "\n\n" + body
		}
		exportTextToFile(w, body, defaultReportFilename())
	})
	exportReportBtn.Disable()

	executeBtn := dangerButton("Execute Cleanup", func() {
		if previewText.Text == "" {
			dialog.ShowInformation("Preview Required", "Run Preview (Dry Run) first to review planned actions.", w)
			return
		}
		dialog.ShowConfirm("Confirm Cleanup",
			"This will permanently disable logging and delete forensic artifacts.\nThis cannot be undone.\n\nProceed?",
			func(ok bool) {
				if !ok {
					return
				}
				ctx, err := orch.BuildContext(cfg)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				pf := orch.Preflight(ctx)
				if !pf.CanProceed {
					dialog.ShowError(fmt.Errorf("pre-flight failed - see preview"), w)
					return
				}

				tracker := newProgressTracker(moduleNamesFromConfig(cfg.Modules), cfg)
				progressWin := application.NewWindow("OpWAX - Running")
				progressWin.Resize(fyne.NewSize(560, 520))
				progressWin.CenterOnScreen()

				cancelCh := make(chan struct{})
				var cancelOnce sync.Once
				requestCancel := func() {
					dialog.ShowConfirm("Cancel cleanup?",
						"Stop after the current step finishes?\nPartial changes may already be applied.",
						func(ok bool) {
							if ok {
								cancelOnce.Do(func() { close(cancelCh) })
							}
						}, progressWin)
				}

				progressWin.SetContent(tracker.content(requestCancel))
				progressWin.Show()

				go func() {
					report := orch.ExecuteWithProgress(ctx, func(u models.ProgressUpdate) {
						fyne.Do(func() { tracker.apply(u) })
					}, cancelCh)
					fyne.Do(func() {
						progressWin.Close()
						summary := orchestrator.FormatReport(report)
						lastExecutionSummary = summary
						exportReportBtn.Enable()
						if report.ManifestDiff != nil {
							txt := system.FormatManifestDiff(*report.ManifestDiff)
							if report.Verification != nil {
								txt += "\n" + verify.FormatReport(*report.Verification)
							}
							manifestText.SetText(txt)
						} else if report.Verification != nil {
							manifestText.SetText(verify.FormatReport(*report.Verification))
						}
						title := "Complete"
						if report.Cancelled {
							title = "Cancelled"
						}
						dialog.ShowInformation(title, summary, w)
						previewText.SetText("")
					})
				}()
			}, w)
	})

	saveBtn := widget.NewButton("Save Config", func() {
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()
			if err := config.Save(uc.URI().Path(), cfg); err != nil {
				dialog.ShowError(err, w)
			}
		}, w)
	})

	loadBtn := widget.NewButton("Import Config", func() {
		dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			defer uc.Close()
			loaded, err := config.Load(uc.URI().Path())
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			cfg = loaded
			rebuildBrowsers(false)
			syncConfigToUI()
			scheduleEnabled.SetChecked(cfg.Schedule.Enabled)
			scheduleMode.SetSelected(string(cfg.Schedule.Mode))
			scheduleTime.SetText(cfg.Schedule.Time)
			scheduleConfigPath.SetText(cfg.Schedule.ConfigPath)
			updateConfigJSON()
		}, w)
	})

	applyJSONBtn := widget.NewButton("Apply JSON Edits", func() {
		loaded, err := config.Parse([]byte(configJSON.Text))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		cfg = loaded
		rebuildBrowsers(false)
		syncConfigToUI()
		scheduleEnabled.SetChecked(cfg.Schedule.Enabled)
		scheduleMode.SetSelected(string(cfg.Schedule.Mode))
		scheduleTime.SetText(cfg.Schedule.Time)
		scheduleConfigPath.SetText(cfg.Schedule.ConfigPath)
		dialog.ShowInformation("Applied", "Config updated from JSON.", w)
	})

	targetsTab := tabWithIcon("Targets", theme.ComputerIcon(), container.NewVBox(
		sectionCard("Users", "Which profiles to include in cleanup.", container.NewVBox(
			fieldLabel("User mode"),
			userModeSelect,
			container.NewHBox(
				widget.NewButton("Refresh list", func() {
					refreshUsers()
					userModeSelect.OnChanged(userModeSelect.Selected)
				}),
			),
			userSelectBox,
		)),
		sectionCard("Drives", "Which volumes to process.", container.NewVBox(
			fieldLabel("Drive mode"),
			driveModeSelect,
			container.NewHBox(
				widget.NewButton("Refresh list", func() {
					refreshDrives()
					driveModeSelect.OnChanged(driveModeSelect.Selected)
				}),
			),
			driveSelectBox,
		)),
	))

	modulesToolbar := container.NewHBox(
		widget.NewButton("Enable all", func() {
			setAllModules(&cfg, moduleRows, true)
			updateConfigJSON()
		}),
		widget.NewButton("Disable all", func() {
			setAllModules(&cfg, moduleRows, false)
			updateConfigJSON()
		}),
	)
	modulesTab := tabWithIcon("Modules", theme.ListIcon(), container.NewVBox(
		hintWithIcon(theme.ViewRefreshIcon(), "↻ marks modules that may require a reboot when enabled: Volatile (LSASS), NTFS ($LogFile reset), and Persistence (pagefile/dumps)."),
		modulesToolbar,
		sectionCard("Enabled modules", "", container.NewVBox(
			rowVolatile.canvasObject(),
			rowRegistry.canvasObject(),
			rowNTFS.canvasObject(),
			rowExec.canvasObject(),
			rowLogs.canvasObject(),
			rowPersist.canvasObject(),
			rowNetwork.canvasObject(),
		)),
	))

	optionsTab := tabWithIcon("Options", theme.SettingsIcon(), container.NewVBox(
		sectionCard("Run behavior", "Reboot and verification options.", container.NewVBox(
			rebootCheck,
			manifestDiffCheck,
			verifyCheck,
			focusedCheck,
			secondPassCheck,
		)),
		sectionCard("NTFS & memory", "Advanced filesystem and credential options.", container.NewVBox(
			mftScrubCheck,
			logFileCheck,
			lsassScrubCheck,
			lsassRebootCheck,
		)),
		sectionCard("Network & browsers", "Wireless profiles and detected browsers.", container.NewVBox(
			fieldLabel("WLAN profile cleanup"),
			wlanSelect,
			fieldLabel("Browsers to clean"),
			container.NewHBox(
				widget.NewButton("Refresh detection", func() {
					rebuildBrowsers(true)
					updateConfigJSON()
				}),
			),
			browserBox,
		)),
		advancedOptionsPanel(&cfg, syncConfig),
	))

	previewHint := hintLabel("Step 1: Preview planned actions. Step 2: Execute only after reviewing warnings.")
	previewTab := tabWithIcon("Preview", theme.SearchIcon(), container.NewBorder(
		container.NewVBox(
			previewHint,
			container.NewHBox(dryRunBtn, exportDryRunBtn, exportReportBtn, executeBtn),
			preflightPreview,
		),
		nil, nil, nil,
		container.NewScroll(previewText),
	))

	configTab := tabWithIcon("Config", theme.DocumentSaveIcon(), container.NewBorder(
		container.NewHBox(saveBtn, loadBtn, applyJSONBtn),
		hintLabel("Edit JSON directly or import a saved profile."),
		nil, nil,
		container.NewScroll(configJSON),
	))

	manifestTab := tabWithIcon("Manifest", theme.DocumentIcon(), container.NewBorder(
		container.NewHBox(generateManifestBtn, saveManifestBtn),
		hintLabel("Scan the system for artifact paths before or after cleanup."),
		nil, nil,
		container.NewScroll(manifestText),
	))

	scheduleTab := tabWithIcon("Schedule", theme.HistoryIcon(), container.NewVBox(
		sectionCard("Task Scheduler", "Run OpWAX-cli silently on a schedule (admin).", container.NewVBox(
			scheduleEnabled,
			fieldLabel("Run mode"),
			scheduleMode,
			fieldLabel("Time (HH:MM)"),
			scheduleTime,
			fieldLabel("Day of week (weekly)"),
			scheduleDOW,
			fieldLabel("Day of month 1–28 (monthly)"),
			scheduleDOM,
			fieldLabel("Config file for scheduled runs"),
			scheduleConfigPath,
			scheduleStatus,
			container.NewHBox(installScheduleBtn, removeScheduleBtn),
		)),
	))

	tabs := container.NewAppTabs(
		targetsTab,
		modulesTab,
		optionsTab,
		previewTab,
		manifestTab,
		scheduleTab,
		configTab,
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(container.NewBorder(
		mainHeader(),
		statusFooter(statusLabel),
		nil, nil,
		pad(tabs),
	))

	w.ShowAndRun()
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if strings.EqualFold(v, s) {
			return slice
		}
	}
	return append(slice, s)
}

func removeString(slice []string, s string) []string {
	var out []string
	for _, v := range slice {
		if !strings.EqualFold(v, s) {
			out = append(out, v)
		}
	}
	return out
}

func isBrowserEnabled(b models.BrowserConfig, kind system.BrowserKind) bool {
	switch kind {
	case system.BrowserEdge:
		return b.Edge
	case system.BrowserChrome:
		return b.Chrome
	case system.BrowserFirefox:
		return b.Firefox
	case system.BrowserBrave:
		return b.Brave
	case system.BrowserOpera:
		return b.Opera
	case system.BrowserVivaldi:
		return b.Vivaldi
	default:
		return false
	}
}

func autoBrowserSuffix(applied bool) string {
	if applied {
		return " - enabled automatically"
	}
	return ""
}
