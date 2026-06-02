package system

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/opwax/opwax/internal/models"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// IsAdmin returns true if the process has an elevated administrator token.
func IsAdmin() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()

	if elevated, ok := tokenIsElevated(token); ok && elevated {
		return true
	}

	// When UAC is disabled, elevated tokens may not set TokenIsElevated.
	if !isUACEnabled() && isAdministratorsMember() {
		return true
	}
	return false
}

func tokenIsElevated(token windows.Token) (elevated bool, ok bool) {
	var elevation struct {
		TokenIsElevated uint32
	}
	var outLen uint32
	err := windows.GetTokenInformation(token, windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &outLen)
	if err != nil {
		return false, false
	}
	return elevation.TokenIsElevated != 0, true
}

func isUACEnabled() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer key.Close()
	enableLUA, _, err := key.GetIntegerValue("EnableLUA")
	if err != nil {
		return true
	}
	return enableLUA != 0
}

func isAdministratorsMember() bool {
	var sid *windows.SID
	if err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	); err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	var member int32
	advapi32 := windows.NewLazySystemDLL("advapi32.dll")
	checkTokenMembership := advapi32.NewProc("CheckTokenMembership")
	ret, _, _ := checkTokenMembership.Call(0, uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(&member)))
	return ret != 0 && member != 0
}

// RequireAdmin exits with error message if not elevated.
func RequireAdmin() error {
	if !IsAdmin() {
		return fmt.Errorf("administrator privileges required - run as Administrator")
	}
	return nil
}

// WindowsDirectory returns %SystemRoot%.
func WindowsDirectory() string {
	if sysRoot := os.Getenv("SystemRoot"); sysRoot != "" {
		return sysRoot
	}
	return `C:\Windows`
}

// ProgramDataDir returns %ProgramData%.
func ProgramDataDir() string {
	if pd := os.Getenv("ProgramData"); pd != "" {
		return pd
	}
	return `C:\ProgramData`
}

// CurrentUsername returns the current user's username.
func CurrentUsername() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	name := u.Username
	if idx := strings.LastIndex(name, `\`); idx >= 0 {
		name = name[idx+1:]
	}
	return name, nil
}

// EnumerateUsers returns local user profile folder names.
func EnumerateUsers() ([]string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProfileList`, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	sids, err := key.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}

	var users []string
	seen := map[string]bool{}

	for _, sid := range sids {
		if !strings.HasPrefix(sid, "S-1-5-21-") {
			continue
		}
		sub, err := registry.OpenKey(key, sid, registry.READ)
		if err != nil {
			continue
		}
		profilePath, _, err := sub.GetStringValue("ProfileImagePath")
		sub.Close()
		if err != nil {
			continue
		}
		base := filepath.Base(profilePath)
		if base == "" || seen[base] {
			continue
		}
		lower := strings.ToLower(base)
		if lower == "systemprofile" || lower == "localservice" ||
			lower == "networkservice" || lower == "defaultuser0" {
			continue
		}
		seen[base] = true
		users = append(users, base)
	}
	return users, nil
}

// ResolveUserProfiles builds profile paths for target usernames.
func ResolveUserProfiles(usernames []string) ([]models.UserProfile, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProfileList`, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	want := map[string]bool{}
	for _, u := range usernames {
		want[strings.ToLower(u)] = true
	}

	var profiles []models.UserProfile
	sids, _ := key.ReadSubKeyNames(0)

	for _, sid := range sids {
		sub, err := registry.OpenKey(key, sid, registry.READ)
		if err != nil {
			continue
		}
		profilePath, _, err := sub.GetStringValue("ProfileImagePath")
		sub.Close()
		if err != nil {
			continue
		}
		username := filepath.Base(profilePath)
		if !want[strings.ToLower(username)] {
			continue
		}
		profiles = append(profiles, models.UserProfile{
			Username:       username,
			SID:            sid,
			ProfilePath:    profilePath,
			NTUserPath:     filepath.Join(profilePath, "NTUSER.DAT"),
			UsrClassPath:   filepath.Join(profilePath, "AppData", "Local", "Microsoft", "Windows", "UsrClass.dat"),
			AppDataRoaming: filepath.Join(profilePath, "AppData", "Roaming"),
			AppDataLocal:   filepath.Join(profilePath, "AppData", "Local"),
		})
	}
	return profiles, nil
}

// EnumerateDrives returns fixed drives on the system.
func EnumerateDrives() ([]models.DriveInfo, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")
	getDriveType := kernel32.NewProc("GetDriveTypeW")

	ret, _, _ := getLogicalDrives.Call()
	mask := uint32(ret)

	systemRoot := strings.ToUpper(string([]rune(WindowsDirectory())[0])) + ":"

	var drives []models.DriveInfo
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		letter := string(rune('A' + i))
		root := letter + `:\`
		rootPtr, _ := windows.UTF16PtrFromString(root)
		dt, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(rootPtr)))
		if dt != 3 { // DRIVE_FIXED
			continue
		}
		drives = append(drives, models.DriveInfo{
			Letter:   letter + ":",
			Root:     root,
			IsSystem: strings.EqualFold(letter+":", systemRoot),
		})
	}
	return drives, nil
}
