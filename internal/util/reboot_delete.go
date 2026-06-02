package util

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const moveDelayUntilReboot = 0x4

// scheduleDeleteOnReboot registers a file for deletion at next boot via MoveFileExW.
func scheduleDeleteOnReboot(path string) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	moveFileEx := kernel32.NewProc("MoveFileExW")
	ret, _, err := moveFileEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		moveDelayUntilReboot,
	)
	if ret == 0 {
		return fmt.Errorf("MoveFileExW: %w", err)
	}
	return nil
}
