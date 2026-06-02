package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// RunHidden executes a command without creating a visible window.
func RunHidden(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w - %s", name, args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RunHiddenOutput returns combined output from a hidden command.
func RunHiddenOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// DeleteRegKey recursively deletes a registry key if it exists.
func DeleteRegKey(root registry.Key, path string) error {
	key, err := registry.OpenKey(root, path, registry.ALL_ACCESS)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	defer key.Close()

	names, err := key.ReadSubKeyNames(0)
	if err == nil {
		for _, name := range names {
			subPath := path + `\` + name
			if err := DeleteRegKey(root, subPath); err != nil {
				return err
			}
		}
	}

	return registry.DeleteKey(root, path)
}

// SetRegDWORD sets a DWORD registry value, creating keys as needed.
func SetRegDWORD(root registry.Key, path, name string, value uint32) error {
	key, _, err := registry.CreateKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetDWordValue(name, value)
}

// SetRegString sets a string registry value.
func SetRegString(root registry.Key, path, name, value string) error {
	key, _, err := registry.CreateKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(name, value)
}

// DisableService sets service Start=4 (disabled) in registry.
func DisableService(serviceName string) error {
	path := `SYSTEM\CurrentControlSet\Services\` + serviceName
	return SetRegDWORD(registry.LOCAL_MACHINE, path, "Start", 4)
}

// StopService stops a Windows service.
func StopService(serviceName string) error {
	return RunHidden("sc", "stop", serviceName)
}

// StartService starts a Windows service.
func StartService(serviceName string) error {
	return RunHidden("sc", "start", serviceName)
}

// StopAndDisableService stops then disables a service.
func StopAndDisableService(serviceName string) error {
	_ = StopService(serviceName)
	return DisableService(serviceName)
}

// DeleteFilesGlob removes files matching a pattern (simple glob with *).
func DeleteFilesGlob(dir, pattern string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		match := pattern == "*" || strings.HasSuffix(pattern, "*") &&
			strings.HasPrefix(e.Name(), strings.TrimSuffix(pattern, "*"))
		if !match && pattern != "*" {
			if pattern != e.Name() && !strings.Contains(pattern, "*") {
				continue
			}
		}
		if strings.Contains(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			suffix := strings.TrimPrefix(pattern, "*")
			if prefix != "" && !strings.HasPrefix(e.Name(), prefix) {
				continue
			}
			if suffix != "" && !strings.HasSuffix(e.Name(), suffix) {
				continue
			}
		}
		if err := os.Remove(filepathJoin(dir, e.Name())); err == nil {
			count++
		}
	}
	return count, nil
}

func filepathJoin(a, b string) string {
	if strings.HasSuffix(a, `\`) || strings.HasSuffix(a, `/`) {
		return a + b
	}
	return a + `\` + b
}

// DeleteDirContents removes all files in a directory (non-recursive).
func DeleteDirContents(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, e := range entries {
		p := filepathJoin(dir, e.Name())
		var err error
		if e.IsDir() {
			err = os.RemoveAll(p)
		} else {
			err = os.Remove(p)
		}
		if err == nil {
			count++
		}
	}
	return count, nil
}

// ClearEventLogChannel clears a single event log channel.
func ClearEventLogChannel(channel string) error {
	return RunHidden("wevtutil", "cl", channel)
}

// DisableEventLogChannel disables a channel.
func DisableEventLogChannel(channel string) error {
	return RunHidden("wevtutil", "sl", channel, "/e:false")
}

// ListEventLogChannels returns all event log channel names.
func ListEventLogChannels() ([]string, error) {
	out, err := RunHiddenOutput("wevtutil", "el")
	if err != nil {
		return nil, err
	}
	var channels []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			channels = append(channels, line)
		}
	}
	return channels, nil
}
