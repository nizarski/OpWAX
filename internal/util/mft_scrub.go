package util

import (
	"os"
	"strings"
)

// ScheduleLogFileReset queues chkdsk /F on next reboot for the drive.
func ScheduleLogFileReset(driveLetter string) error {
	letter, err := normalizeDriveLetter(driveLetter)
	if err != nil {
		return err
	}
	return RunHidden("cmd", "/c", "echo Y| chkdsk "+letter+" /F")
}

// ResetLogFileNow attempts immediate log maintenance on a non-system drive.
func ResetLogFileNow(driveRoot string) error {
	letter, err := normalizeDriveLetter(driveRoot)
	if err != nil {
		return err
	}
	_ = RunHidden("fsutil", "volume", "dismount", letter)
	err = RunHidden("chkdsk", letter, "/F", "/X")
	_ = RunHidden("fsutil", "volume", "mount", letter)
	return err
}

// IsSystemDrive reports whether a drive letter is the Windows system volume.
func IsSystemDrive(driveLetter string) bool {
	sys := os.Getenv("SystemRoot")
	if len(sys) < 2 {
		sys = `C:\Windows`
	}
	sysLetter := strings.ToUpper(string(sys[0])) + ":"
	letter, err := normalizeDriveLetter(driveLetter)
	if err != nil {
		return false
	}
	return letter == sysLetter
}

func normalizeDriveLetter(driveLetter string) (string, error) {
	_, letter, err := normalizeDriveRoot(driveLetter)
	return letter, err
}
