package util

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	advapi32       = windows.NewLazySystemDLL("advapi32.dll")
	credEnumerateW = advapi32.NewProc("CredEnumerateW")
	credFree       = advapi32.NewProc("CredFree")
	credDeleteW    = advapi32.NewProc("CredDeleteW")
)

// credEntry is a minimal view of CREDENTIALW for enumeration.
type credEntry struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     uintptr
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

// ScrubLSASSCaches clears credential stores that LSASS holds in memory (best-effort live).
// Full LSASS RAM zero requires reboot - this removes sources and cached tickets.
func ScrubLSASSCaches() error {
	var errs []string

	if err := disableWDigest(); err != nil {
		errs = append(errs, "wdigest: "+err.Error())
	}
	if err := PurgeKerberosTicketsNative(); err != nil {
		errs = append(errs, "kerberos: "+err.Error())
	}
	if n, err := deleteAllCredManagerEntries(); err != nil {
		errs = append(errs, fmt.Sprintf("credmgr: %v", err))
	} else if n > 0 {
		_ = n
	}
	if err := ClearVaultCredentialsNative(); err != nil {
		errs = append(errs, "vault: "+err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("partial LSASS cache scrub: %s", strings.Join(errs, "; "))
	}
	return nil
}

// disableWDigest stops WDigest from storing plaintext creds in LSASS going forward.
func disableWDigest() error {
	path := `SYSTEM\CurrentControlSet\Control\SecurityProviders\WDigest`
	if err := SetRegDWORD(registry.LOCAL_MACHINE, path, "UseLogonCredential", 0); err != nil {
		return err
	}
	return SetRegDWORD(registry.LOCAL_MACHINE, path, "Negotiate", 0)
}

func deleteAllCredManagerEntries() (int, error) {
	var count uint32
	var creds uintptr
	ret, _, err := credEnumerateW.Call(
		0, // NULL filter
		1, // CRED_ENUMERATE_ALL_CREDENTIALS
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&creds)),
	)
	if ret == 0 {
		return 0, err
	}
	defer credFree.Call(creds)

	deleted := 0
	ptrSize := unsafe.Sizeof(uintptr(0))
	for i := uint32(0); i < count; i++ {
		entryPtr := *(*uintptr)(unsafe.Pointer(creds + uintptr(i)*ptrSize))
		entry := (*credEntry)(unsafe.Pointer(entryPtr))
		if entry.TargetName == nil {
			continue
		}
		target := windows.UTF16PtrToString(entry.TargetName)
		if target == "" {
			continue
		}
		tptr, _ := windows.UTF16PtrFromString(target)
		dret, _, _ := credDeleteW.Call(
			uintptr(unsafe.Pointer(tptr)),
			uintptr(entry.Type),
			0,
		)
		if dret != 0 {
			deleted++
		}
	}
	return deleted, nil
}

// LSASSScrubNote returns human-readable guidance after LSASS cleanup.
func LSASSScrubNote(rebootAfter bool) string {
	msg := "Credential stores and Kerberos tickets cleared; WDigest disabled for future logons."
	if rebootAfter {
		msg += " Reboot will clear remaining LSASS RAM."
	} else {
		msg += " Enable lsass_reboot_after for full RAM cleanup via reboot."
	}
	return msg
}
