package modules

import (
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"

	"golang.org/x/sys/windows/registry"
)

const volatileName = "volatile_memory"

// VolatileModule handles RAM-exposed caches and credential stores.
type VolatileModule struct{}

func (m *VolatileModule) Name() string { return volatileName }

func (m *VolatileModule) DryRun(ctx *models.RunContext) []models.Action {
	actions := []models.Action{
		action(volatileName, "Flush DNS client cache", "ipconfig /flushdns", models.ActionClean),
		action(volatileName, "Flush ARP cache", "arp -d *", models.ActionClean),
		action(volatileName, "Clear stored credentials (cmdkey)", "cmdkey /list", models.ActionClean),
		action(volatileName, "Disable PowerShell script block logging", `HKCU\PowerShell\ScriptBlockLogging`, models.ActionDisable),
	}
	if ctx.Config.Options.LSASSScrub {
		actions = append(actions,
			action(volatileName, "Disable WDigest plaintext cred caching", "WDigest", models.ActionDisable),
			action(volatileName, "Purge Kerberos tickets (LSA API)", "Kerberos", models.ActionClean),
			action(volatileName, "Delete Credential Manager entries (CredEnumerate)", "CredMgr", models.ActionClean),
			action(volatileName, "Clear Windows Vault stores (native)", "Vault", models.ActionClean),
		)
		if ctx.Config.Options.LSASSRebootAfter {
			actions = append(actions,
				action(volatileName, "Reboot after LSASS scrub (optional)", "reboot", models.ActionClean),
			)
		}
	}
	return actions
}

func (m *VolatileModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	a := action(volatileName, "Disable PS script block logging", "ScriptBlockLogging", models.ActionDisable)
	err := util.SetRegDWORD(registry.CURRENT_USER,
		`SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging`,
		"EnableScriptBlockLogging", 0)
	results = append(results, result(a, err))

	if ctx.Config.Options.LSASSScrub {
		a2 := action(volatileName, "Disable WDigest", "WDigest", models.ActionDisable)
		results = append(results, result(a2, util.SetRegDWORD(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Control\SecurityProviders\WDigest`, "UseLogonCredential", 0)))
	}

	return results
}

func (m *VolatileModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	a1 := action(volatileName, "Flush DNS", "dns", models.ActionClean)
	results = append(results, result(a1, util.RunHidden("ipconfig", "/flushdns")))

	a2 := action(volatileName, "Flush ARP", "arp", models.ActionClean)
	results = append(results, result(a2, util.RunHidden("arp", "-d", "*")))

	a3 := action(volatileName, "Clear cmdkey credentials", "cmdkey", models.ActionClean)
	out, _ := util.RunHiddenOutput("cmdkey", "/list")
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Target:") {
			target := strings.TrimSpace(strings.TrimPrefix(line, "Target:"))
			if target != "" {
				_ = util.RunHidden("cmdkey", "/delete:"+target)
			}
		}
	}
	results = append(results, result(a3, nil))

	if ctx.Config.Options.LSASSScrub {
		a4 := action(volatileName, "Scrub LSASS credential caches", "LSASS", models.ActionSecure)
		err := util.ScrubLSASSCaches()
		results = append(results, result(a4, err))
		note := action(volatileName, util.LSASSScrubNote(ctx.Config.Options.LSASSRebootAfter), "info", models.ActionClean)
		results = append(results, models.Result{Action: note, Success: true})
	}

	return results
}
