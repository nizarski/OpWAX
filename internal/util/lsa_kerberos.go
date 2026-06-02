package util

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	kerbPurgeTktCacheMessage = 6
	statusSuccess            = 0
)

type lsaUnicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type luid struct {
	LowPart  uint32
	HighPart int32
}

type kerbPurgeTktCacheRequest struct {
	MessageType uint32
	LogonId     luid
	ServerName  lsaUnicodeString
	RealmName   lsaUnicodeString
}

var (
	secur32                        = windows.NewLazySystemDLL("secur32.dll")
	procLsaConnectUntrusted        = secur32.NewProc("LsaConnectUntrusted")
	procLsaLookupAuthenticationPkg = secur32.NewProc("LsaLookupAuthenticationPackage")
	procLsaCallAuthenticationPkg   = secur32.NewProc("LsaCallAuthenticationPackage")
	procLsaDeregisterLogonProcess  = secur32.NewProc("LsaDeregisterLogonProcess")
	procLsaFreeReturnBuffer        = secur32.NewProc("LsaFreeReturnBuffer")
)

// PurgeKerberosTicketsNative purges Kerberos ticket cache via LSA (replaces klist purge).
func PurgeKerberosTicketsNative() error {
	var lsaHandle uintptr
	ret, _, err := procLsaConnectUntrusted.Call(uintptr(unsafe.Pointer(&lsaHandle)))
	if status := lsaStatus(ret); status != statusSuccess {
		return fmt.Errorf("LsaConnectUntrusted: %s", err)
	}
	defer procLsaDeregisterLogonProcess.Call(lsaHandle)

	kerbName, _ := windows.UTF16PtrFromString("Kerberos")
	var authPackage uint32
	ret, _, err = procLsaLookupAuthenticationPkg.Call(
		lsaHandle,
		uintptr(unsafe.Pointer(kerbName)),
		uintptr(unsafe.Pointer(&authPackage)),
	)
	if status := lsaStatus(ret); status != statusSuccess {
		return fmt.Errorf("LsaLookupAuthenticationPackage: %s", err)
	}

	req := kerbPurgeTktCacheRequest{
		MessageType: kerbPurgeTktCacheMessage,
		LogonId:     luid{0, 0},
	}
	var respPtr uintptr
	var respLen uint32
	var protoStatus uint32

	ret, _, err = procLsaCallAuthenticationPkg.Call(
		lsaHandle,
		uintptr(authPackage),
		uintptr(unsafe.Pointer(&req)),
		uintptr(unsafe.Sizeof(req)),
		uintptr(unsafe.Pointer(&respPtr)),
		uintptr(unsafe.Pointer(&respLen)),
		uintptr(unsafe.Pointer(&protoStatus)),
	)
	if status := lsaStatus(ret); status != statusSuccess {
		return fmt.Errorf("LsaCallAuthenticationPackage: %s", err)
	}
	if respPtr != 0 {
		_, _, _ = procLsaFreeReturnBuffer.Call(respPtr)
	}
	if protoStatus != statusSuccess {
		return fmt.Errorf("Kerberos purge protocol status 0x%x", protoStatus)
	}
	return nil
}

func lsaStatus(ret uintptr) uint32 {
	// NTSTATUS - LSA returns NTSTATUS in ret; success is 0
	return uint32(ret)
}
