package util

import (
	"os"
	"path/filepath"
)

// ClearVaultCredentialsNative removes Windows Vault credential stores (replaces vaultcmd).
func ClearVaultCredentialsNative() error {
	_ = StopService("VaultSvc")

	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		_ = os.RemoveAll(filepath.Join(local, "Microsoft", "Vault"))
		_ = os.RemoveAll(filepath.Join(local, "Microsoft", "Credentials"))
	}
	if roaming := os.Getenv("APPDATA"); roaming != "" {
		_ = os.RemoveAll(filepath.Join(roaming, "Microsoft", "Credentials"))
		_ = os.RemoveAll(filepath.Join(roaming, "Microsoft", "SystemCertificates"))
	}
	return nil
}
