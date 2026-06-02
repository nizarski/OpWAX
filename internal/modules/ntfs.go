package modules

import (
	"strconv"
	"strings"

	"github.com/opwax/opwax/internal/models"
	"github.com/opwax/opwax/internal/util"
)

const ntfsName = "ntfs_metadata"

// NTFSModule handles USN journal, Zone Identifier ADS, MFT free records, and $LogFile.
type NTFSModule struct{}

func (m *NTFSModule) Name() string { return ntfsName }

func (m *NTFSModule) DryRun(ctx *models.RunContext) []models.Action {
	var actions []models.Action
	actions = append(actions,
		action(ntfsName, "Disable saving Zone Identifier ADS", "Attachments policy", models.ActionDisable),
		action(ntfsName, "Disable NTFS last access timestamps", "fsutil behavior", models.ActionDisable),
	)
	for _, d := range ctx.TargetDrives {
		actions = append(actions,
			action(ntfsName, "Reset USN change journal on "+d.Letter, d.Letter, models.ActionDisable),
		)
		if ctx.Config.Options.LogFileResetOnReboot {
			if util.IsSystemDrive(d.Letter) {
				actions = append(actions,
					action(ntfsName, "Schedule $LogFile reset (chkdsk /F on reboot) "+d.Letter, "chkdsk /F", models.ActionDisable),
				)
			} else {
				actions = append(actions,
					action(ntfsName, "Reset $LogFile now (dismount + chkdsk) "+d.Letter, d.Root, models.ActionSecure),
				)
			}
		}
	}
	if len(ctx.TargetUsers) > 0 {
		roots := util.ZoneIdentifierScanRoots(ctx.TargetUsers)
		actions = append(actions,
			action(ntfsName, "Strip Zone.Identifier ADS in user profile folders", strings.Join(roots, "; "), models.ActionClean),
		)
	}
	actions = append(actions,
		action(ntfsName, "Timestomping detection ($SI vs $FN)", "not performed - OpWAX does not parse live $MFT for timestamp mismatches", models.ActionDisable),
	)
	actions = append(actions, advancedNTFSDryRun(ctx)...)
	return actions
}

func (m *NTFSModule) Disable(ctx *models.RunContext) []models.Result {
	var results []models.Result

	a1 := action(ntfsName, "Disable Zone Identifier saving", "Attachments", models.ActionDisable)
	results = append(results, result(a1, util.DisableZoneIdentifierSaving()))

	a2 := action(ntfsName, "Disable last access time", "fsutil", models.ActionDisable)
	results = append(results, result(a2, util.RunHidden("fsutil", "behavior", "set", "disablelastaccess", "1")))

	for _, d := range ctx.TargetDrives {
		a := action(ntfsName, "Reset USN change journal "+d.Letter, d.Letter, models.ActionDisable)
		results = append(results, result(a, util.ResetUSNJournal(d.Letter)))
	}

	results = append(results, advancedNTFSDisable(ctx)...)
	return results
}

func (m *NTFSModule) Clean(ctx *models.RunContext) []models.Result {
	var results []models.Result

	if len(ctx.TargetUsers) > 0 {
		roots := util.ZoneIdentifierScanRoots(ctx.TargetUsers)
		a := action(ntfsName, "Strip Zone.Identifier in user profile folders", strings.Join(roots, "; "), models.ActionClean)
		count, err := util.StripZoneIdentifiersInPaths(roots)
		if err != nil {
			results = append(results, result(a, err))
		} else {
			a.Description = "Strip Zone.Identifier ADS (" + strconv.Itoa(count) + " stream(s) removed)"
			results = append(results, result(a, nil))
		}
	}

	for _, d := range ctx.TargetDrives {
		if util.USNJournalActive(d.Letter) {
			a1 := action(ntfsName, "Reset USN change journal "+d.Letter, d.Letter, models.ActionDisable)
			results = append(results, result(a1, util.ResetUSNJournal(d.Letter)))
		}

		if ctx.Config.Options.LogFileResetOnReboot {
			if util.IsSystemDrive(d.Letter) {
				a4 := action(ntfsName, "Schedule $LogFile reset "+d.Letter, "chkdsk /F", models.ActionDisable)
				results = append(results, result(a4, util.ScheduleLogFileReset(d.Letter)))
			} else {
				a4 := action(ntfsName, "Reset $LogFile now "+d.Letter, d.Root, models.ActionSecure)
				results = append(results, result(a4, util.ResetLogFileNow(d.Root)))
			}
		}
	}

	results = append(results, advancedNTFSClean(ctx)...)
	return results
}
