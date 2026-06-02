package util

import (
	"strings"
)

// DisableUSNJournal disables the NTFS USN change journal on a volume.
// Uses /N so the call returns only after the journal is fully disabled.
func DisableUSNJournal(driveLetter string) error {
	return resetUSNJournal(driveLetter)
}

// ResetUSNJournal deletes all USN journal data and disables the journal (timeline defeat).
func ResetUSNJournal(driveLetter string) error {
	return resetUSNJournal(driveLetter)
}

func resetUSNJournal(driveLetter string) error {
	letter, err := normalizeDriveLetter(driveLetter)
	if err != nil {
		return err
	}
	return RunHidden("fsutil", "usn", "deletejournal", "/N", letter)
}

// USNJournalActive reports whether an NTFS change journal is active on the volume.
func USNJournalActive(driveLetter string) bool {
	letter, err := normalizeDriveLetter(driveLetter)
	if err != nil {
		return false
	}
	out, err := RunHiddenOutput("fsutil", "usn", "queryjournal", letter)
	if err != nil {
		return false
	}
	lower := strings.ToLower(out)
	if strings.Contains(lower, "not active") || strings.Contains(lower, "does not exist") {
		return false
	}
	return strings.Contains(lower, "journal id") || strings.Contains(lower, "usn journal")
}
