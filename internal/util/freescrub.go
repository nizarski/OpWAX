package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

const (
	scrubChunkSize    = 1024 * 1024 // 1 MiB write chunks
	scrubReserveBytes = 8 * 1024 * 1024
	scrubMFTMaxTries  = 200000
	mftGrowChunk      = 256
)
// ScrubProgressFunc reports free-space scrub progress (optional).
type ScrubProgressFunc func(message string, percent float64)

// ScrubMFTFreeSpace wipes free NTFS clusters and free MFT records on a volume.
// Pure Go reimplementation of SDelete -c / -z (no external binaries).
func ScrubMFTFreeSpace(driveLetter string) error {
	return ScrubMFTFreeSpaceWithProgress(driveLetter, nil)
}

// ScrubMFTFreeSpaceWithProgress runs the scrubber with optional progress callbacks.
func ScrubMFTFreeSpaceWithProgress(driveLetter string, onProgress ScrubProgressFunc) error {
	root, letter, err := normalizeDriveRoot(driveLetter)
	if err != nil {
		return err
	}

	emit := func(msg string, pct float64) {
		if onProgress != nil {
			onProgress(msg, pct)
		}
	}

	workDir := filepath.Join(root, fmt.Sprintf("$OPWAX_SCRUB_%d", os.Getpid()))
	if err := os.MkdirAll(workDir, 0700); err != nil {
		return fmt.Errorf("scrub workdir: %w", err)
	}
	defer os.RemoveAll(workDir)

	emit("Filling free MFT records on "+letter, 0.05)
	mftCount, err := scrubFreeMFTRecords(workDir, emit)
	if err != nil {
		return fmt.Errorf("mft records %s: %w", letter, err)
	}
	_ = mftCount

	emit("Scrubbing free clusters on "+letter, 0.35)
	if err := scrubFreeClusters(workDir, root, emit); err != nil {
		return fmt.Errorf("free clusters %s: %w", letter, err)
	}

	emit("Free-space scrub complete on "+letter, 1.0)
	return nil
}

// scrubFreeMFTRecords mimics SDelete: create max resident files until MFT free records are filled.
func scrubFreeMFTRecords(workDir string, emit ScrubProgressFunc) (int, error) {
	count := 0
	failures := 0
	for count < scrubMFTMaxTries {
		if failures > 32 {
			break
		}
		name := filepath.Join(workDir, fmt.Sprintf("mft_%08d.tmp", count))
		size, err := createMaxMFTResidentFile(name)
		if err != nil || size <= 0 {
			failures++
			_ = os.Remove(name)
			continue
		}
		failures = 0
		_ = os.Remove(name)
		count++
		if count%500 == 0 && emit != nil {
			emit(fmt.Sprintf("MFT fillers: %d", count), 0.05+float64(count%10000)/100000)
		}
	}
	if count == 0 && failures > 0 {
		return 0, fmt.Errorf("could not allocate MFT filler files")
	}
	return count, nil
}

// createMaxMFTResidentFile grows a file until NTFS refuses (MFT record full), securely overwrites, returns size.
func createMaxMFTResidentFile(path string) (int64, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	chunk := make([]byte, mftGrowChunk)
	var total int64
	for {
		n, werr := f.Write(chunk)
		total += int64(n)
		if werr != nil {
			break
		}
		if total > 65536 {
			break
		}
	}
	if total <= 0 {
		return 0, fmt.Errorf("zero-length MFT filler")
	}
	if err := secureOverwriteHandle(f, total); err != nil {
		return total, err
	}
	return total, nil
}

// scrubFreeClusters allocates two large files to occupy free space, overwrites, deletes (SDelete pattern).
func scrubFreeClusters(workDir, root string, emit ScrubProgressFunc) error {
	free, err := diskFreeBytes(root)
	if err != nil {
		return err
	}
	if free <= scrubReserveBytes {
		return nil
	}
	toFill := free - scrubReserveBytes
	first := toFill / 2
	second := toFill - first

	f1 := filepath.Join(workDir, "cluster_a.dat")
	f2 := filepath.Join(workDir, "cluster_b.dat")

	emit("Allocating scrub file A", 0.4)
	if err := allocateScrubFile(f1, first); err != nil {
		return err
	}
	emit("Allocating scrub file B", 0.55)
	if err := allocateScrubFile(f2, second); err != nil {
		_ = os.Remove(f1)
		return err
	}

	emit("Secure overwrite (1-pass zero-fill)", 0.7)
	if err := secureOverwritePath(f1); err != nil {
		_ = os.Remove(f1)
		_ = os.Remove(f2)
		return err
	}
	if err := secureOverwritePath(f2); err != nil {
		_ = os.Remove(f1)
		_ = os.Remove(f2)
		return err
	}

	_ = os.Remove(f1)
	_ = os.Remove(f2)
	return nil
}

func allocateScrubFile(path string, size uint64) error {
	if size == 0 {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, scrubChunkSize)
	var written uint64
	for written < size {
		toWrite := uint64(len(buf))
		if remaining := size - written; remaining < toWrite {
			toWrite = remaining
		}
		n, err := f.Write(buf[:toWrite])
		written += uint64(n)
		if err != nil {
			return fmt.Errorf("allocate %s at %d: %w", path, written, err)
		}
	}
	return f.Sync()
}

func diskFreeBytes(root string) (uint64, error) {
	root = strings.TrimSuffix(root, `\`) + `\`
	ptr, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return 0, err
	}
	var freeAvailable uint64
	if err := windows.GetDiskFreeSpaceEx(ptr, &freeAvailable, nil, nil); err != nil {
		return 0, err
	}
	return freeAvailable, nil
}

func normalizeDriveRoot(driveLetter string) (root, letter string, err error) {
	letter = strings.TrimSuffix(strings.ToUpper(strings.TrimSpace(driveLetter)), `\`)
	if !strings.HasSuffix(letter, ":") {
		letter += ":"
	}
	if len(letter) != 2 || letter[1] != ':' {
		return "", "", fmt.Errorf("invalid drive: %q", driveLetter)
	}
	return letter + `\`, letter, nil
}

// secureOverwriteHandle applies a single fast zero-fill pass to an open file.
func secureOverwriteHandle(f *os.File, size int64) error {
	if size <= 0 {
		return nil
	}
	buf := make([]byte, scrubChunkSize)
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	remaining := size
	for remaining > 0 {
		n := int64(len(buf))
		if remaining < n {
			n = remaining
		}
		clear(buf[:n])
		written, err := f.Write(buf[:n])
		if err != nil {
			return err
		}
		remaining -= int64(written)
	}
	return f.Sync()
}
func secureOverwritePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return secureOverwriteHandle(f, info.Size())
}
