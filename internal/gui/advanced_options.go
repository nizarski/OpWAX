package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opwax/opwax/internal/models"
)

func advancedOptionsPanel(cfg *models.Config, syncConfig func()) fyne.CanvasObject {
	adv := &cfg.Options.Advanced

	vssCheck := widget.NewCheck("Delete all VSS / shadow copies", func(v bool) {
		adv.DeleteVSSShadows = v
		syncConfig()
	})
	vssCheck.SetChecked(adv.DeleteVSSShadows)

	fullVolCheck := widget.NewCheck("Full-volume unallocated wipe (cipher /w)", func(v bool) {
		adv.FullVolumeUnallocated = v
		syncConfig()
	})
	fullVolCheck.SetChecked(adv.FullVolumeUnallocated)

	slackCheck := widget.NewCheck("Wipe file slack space", func(v bool) {
		adv.WipeFileSlack = v
		syncConfig()
	})
	slackCheck.SetChecked(adv.WipeFileSlack)

	slackAllCheck := widget.NewCheck("Slack wipe: all profile files (slow)", func(v bool) {
		adv.WipeFileSlackAllFiles = v
		syncConfig()
	})
	slackAllCheck.SetChecked(adv.WipeFileSlackAllFiles)

	badClustCheck := widget.NewCheck("Bad cluster scan (chkdsk /B)", func(v bool) {
		adv.ScrubBadClusters = v
		syncConfig()
	})
	badClustCheck.SetChecked(adv.ScrubBadClusters)

	psCheck := widget.NewCheck("PowerShell history (disable + delete)", func(v bool) {
		adv.PowerShellHistory = v
		syncConfig()
	})
	psCheck.SetChecked(adv.PowerShellHistory)

	timelineCheck := widget.NewCheck("Windows Timeline (disable + delete)", func(v bool) {
		adv.TimelineActivity = v
		syncConfig()
	})
	timelineCheck.SetChecked(adv.TimelineActivity)

	userAssistCheck := widget.NewCheck("UserAssist (all GUI execution counts)", func(v bool) {
		adv.UserAssist = v
		syncConfig()
	})
	userAssistCheck.SetChecked(adv.UserAssist)

	onedriveSelect := widget.NewSelect([]string{"off", "metadata", "full"}, func(s string) {
		adv.OneDrive = s
		syncConfig()
	})
	if adv.OneDrive == "" {
		adv.OneDrive = "off"
	}
	onedriveSelect.SetSelected(adv.OneDrive)

	cloudSelect := widget.NewSelect([]string{"off", "all"}, func(s string) {
		adv.CloudSync = s
		syncConfig()
	})
	if adv.CloudSync == "" {
		adv.CloudSync = "off"
	}
	cloudSelect.SetSelected(adv.CloudSync)

	rdpSelect := widget.NewSelect([]string{"off", "target_users"}, func(s string) {
		adv.RDPCache = s
		syncConfig()
	})
	if adv.RDPCache == "" {
		adv.RDPCache = "off"
	}
	rdpSelect.SetSelected(adv.RDPCache)

	wmiSelect := widget.NewSelect([]string{"off", "reset"}, func(s string) {
		adv.WMI = s
		syncConfig()
	})
	if adv.WMI == "" {
		adv.WMI = "off"
	}
	wmiSelect.SetSelected(adv.WMI)

	tasksSelect := widget.NewSelect([]string{"off", "user", "all_except_opwax", "all"}, func(s string) {
		adv.ScheduledTasks = s
		syncConfig()
	})
	if adv.ScheduledTasks == "" {
		adv.ScheduledTasks = "off"
	}
	tasksSelect.SetSelected(adv.ScheduledTasks)

	bitsSelect := widget.NewSelect([]string{"off", "clear", "disable_clear"}, func(s string) {
		adv.BITS = s
		syncConfig()
	})
	if adv.BITS == "" {
		adv.BITS = "off"
	}
	bitsSelect.SetSelected(adv.BITS)

	runKeysSelect := widget.NewSelect([]string{"off", "clean", "clean_disable"}, func(s string) {
		adv.AlternateRunKeys = s
		syncConfig()
	})
	if adv.AlternateRunKeys == "" {
		adv.AlternateRunKeys = "off"
	}
	runKeysSelect.SetSelected(adv.AlternateRunKeys)

	hypervSelect := widget.NewSelect([]string{"off", "snapshots"}, func(s string) {
		adv.HyperV = s
		syncConfig()
	})
	if adv.HyperV == "" {
		adv.HyperV = "off"
	}
	hypervSelect.SetSelected(adv.HyperV)

	wslSelect := widget.NewSelect([]string{"off", "logs", "delete_vhdx"}, func(s string) {
		adv.WSL = s
		syncConfig()
	})
	if adv.WSL == "" {
		adv.WSL = "off"
	}
	wslSelect.SetSelected(adv.WSL)

	recallCheck := widget.NewCheck("Windows Recall (disable + delete)", func(v bool) {
		adv.WindowsRecall = v
		syncConfig()
	})
	recallCheck.SetChecked(adv.WindowsRecall)

	pcaCheck := widget.NewCheck("PCA launch logs (PcaAppLaunchDic.txt)", func(v bool) {
		adv.PCAAppCompatLogs = v
		syncConfig()
	})
	pcaCheck.SetChecked(adv.PCAAppCompatLogs)

	notepadCheck := widget.NewCheck("Notepad tab/draft cache", func(v bool) {
		adv.NotepadTabCache = v
		syncConfig()
	})
	notepadCheck.SetChecked(adv.NotepadTabCache)

	officeCheck := widget.NewCheck("Office Trusted Documents (TrustRecords)", func(v bool) {
		adv.OfficeTrustRecords = v
		syncConfig()
	})
	officeCheck.SetChecked(adv.OfficeTrustRecords)

	sysintCheck := widget.NewCheck("Sysinternals EULA acceptance keys", func(v bool) {
		adv.SysinternalsEULA = v
		syncConfig()
	})
	sysintCheck.SetChecked(adv.SysinternalsEULA)

	outlookSelect := widget.NewSelect([]string{"off", "cache", "delete_ost_pst"}, func(s string) {
		adv.Outlook = s
		syncConfig()
	})
	if adv.Outlook == "" {
		adv.Outlook = "off"
	}
	outlookSelect.SetSelected(adv.Outlook)

	teamsSelect := widget.NewSelect([]string{"off", "cache", "full"}, func(s string) {
		adv.Teams = s
		syncConfig()
	})
	if adv.Teams == "" {
		adv.Teams = "off"
	}
	teamsSelect.SetSelected(adv.Teams)

	searchCheck := widget.NewCheck("Windows Search index (WSearch + Windows.edb)", func(v bool) {
		adv.WindowsSearchIndex = v
		syncConfig()
	})
	searchCheck.SetChecked(adv.WindowsSearchIndex)

	shellCacheCheck := widget.NewCheck("Shell icon/thumb caches", func(v bool) {
		adv.ShellIconCaches = v
		syncConfig()
	})
	shellCacheCheck.SetChecked(adv.ShellIconCaches)

	execRegCheck := widget.NewCheck("MUICache / RecentApps / AppCompatFlags", func(v bool) {
		adv.ExecutionRegistries = v
		syncConfig()
	})
	execRegCheck.SetChecked(adv.ExecutionRegistries)

	syscacheCheck := widget.NewCheck("Syscache.hve + RecentFileCache.bcf", func(v bool) {
		adv.SyscacheHive = v
		syncConfig()
	})
	syscacheCheck.SetChecked(adv.SyscacheHive)

	evtxTargetCheck := widget.NewCheck("Task Scheduler + App Experience EVTX", func(v bool) {
		adv.TargetedEventChannels = v
		syncConfig()
	})
	evtxTargetCheck.SetChecked(adv.TargetedEventChannels)

	eventTranscriptCheck := widget.NewCheck("EventTranscript.db (Diagnosis)", func(v bool) {
		adv.EventTranscript = v
		syncConfig()
	})
	eventTranscriptCheck.SetChecked(adv.EventTranscript)

	werCheck := widget.NewCheck("Windows Error Reporting (WER)", func(v bool) {
		adv.WERReports = v
		syncConfig()
	})
	werCheck.SetChecked(adv.WERReports)

	servicingCheck := widget.NewCheck("Servicing logs (Panther/CBS/WU)", func(v bool) {
		adv.ServicingLogs = v
		syncConfig()
	})
	servicingCheck.SetChecked(adv.ServicingLogs)

	spoolCheck := widget.NewCheck("Print spooler queue files", func(v bool) {
		adv.PrintSpooler = v
		syncConfig()
	})
	spoolCheck.SetChecked(adv.PrintSpooler)

	devCheck := widget.NewCheck("Developer traces (git/npm/VS Code/Cursor)", func(v bool) {
		adv.DeveloperTraces = v
		syncConfig()
	})
	devCheck.SetChecked(adv.DeveloperTraces)

	doCheck := widget.NewCheck("Delivery Optimization cache", func(v bool) {
		adv.DeliveryOptimization = v
		syncConfig()
	})
	doCheck.SetChecked(adv.DeliveryOptimization)

	smartScreenCheck := widget.NewCheck("SmartScreen local cache", func(v bool) {
		adv.SmartScreenCache = v
		syncConfig()
	})
	smartScreenCheck.SetChecked(adv.SmartScreenCache)

	return container.NewVBox(
		hintLabel("Advanced artifacts require the matching module enabled. Destructive options default off."),
		sectionCard("NTFS module", "Volume shadows, free space, slack, bad clusters.", container.NewVBox(
			vssCheck, fullVolCheck, slackCheck, slackAllCheck, badClustCheck,
		)),
		sectionCard("Execution module", "PowerShell, UserAssist, RDP cache.", container.NewVBox(
			psCheck, userAssistCheck,
			fieldLabel("RDP bitmap cache"), rdpSelect,
		)),
		sectionCard("Logs module", "Windows Timeline / Activity History.", container.NewVBox(
			timelineCheck,
		)),
		sectionCard("Network module", "OneDrive and other cloud sync.", container.NewVBox(
			fieldLabel("OneDrive"), onedriveSelect,
			fieldLabel("Other cloud sync"), cloudSelect,
		)),
		sectionCard("Persistence module", "WMI, tasks, BITS, Run keys, Hyper-V, WSL.", container.NewVBox(
			fieldLabel("WMI repository"), wmiSelect,
			fieldLabel("Scheduled tasks"), tasksSelect,
			fieldLabel("BITS queue"), bitsSelect,
			fieldLabel("Alternate Run keys"), runKeysSelect,
			fieldLabel("Hyper-V"), hypervSelect,
			fieldLabel("WSL"), wslSelect,
		)),
		sectionCard("Modern Windows 11", "Recall, PCA, Notepad (Execution module).", container.NewVBox(
			recallCheck, pcaCheck, notepadCheck,
		)),
		sectionCard("Registry module", "Office macros, Sysinternals.", container.NewVBox(
			officeCheck, sysintCheck,
		)),
		sectionCard("Execution gaps (Tier A)", "Search index, shell caches, targeted logs; safe defaults on.", container.NewVBox(
			searchCheck, shellCacheCheck, execRegCheck, syscacheCheck,
			evtxTargetCheck, eventTranscriptCheck, werCheck,
		)),
		sectionCard("Deep traces (Tier B)", "Servicing, spooler, dev tools; default off.", container.NewVBox(
			servicingCheck, spoolCheck, devCheck, doCheck, smartScreenCheck,
		)),
		sectionCard("Email & chat", "Outlook and Teams (Network module).", container.NewVBox(
			fieldLabel("Outlook"), outlookSelect,
			fieldLabel("Microsoft Teams"), teamsSelect,
		)),
	)
}
