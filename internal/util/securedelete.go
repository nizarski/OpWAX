package util

import (
	"fmt"
	"os"
)

// SecureDelete overwrites a file once with zeros then removes it.
func SecureDelete(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("secure delete: %s is a directory", path)
	}

	size := info.Size()
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return os.Remove(path)
	}

	buf := make([]byte, 65536)
	if _, err := f.Seek(0, 0); err == nil {
		remaining := size
		for remaining > 0 {
			n := int64(len(buf))
			if remaining < n {
				n = remaining
			}
			clear(buf[:n])
			written, werr := f.Write(buf[:n])
			if werr != nil {
				break
			}
			remaining -= int64(written)
		}
		_ = f.Sync()
	}
	_ = f.Close()

	if err := os.Remove(path); err != nil {
		return scheduleDeleteOnReboot(path)
	}
	return nil
}

// SecureDeleteFast is an alias for SecureDelete (single zero-fill pass).
func SecureDeleteFast(path string) error {
	return SecureDelete(path)
}

// SecureDeleteDirFiles securely deletes all files in a directory tree.
func SecureDeleteDirFiles(root string) (int, error) {
	count := 0
	err := filepathWalk(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if serr := SecureDelete(path); serr == nil {
			count++
		}
		return nil
	})
	return count, err
}

func filepathWalk(root string, fn func(string, os.DirEntry, error) error) error {
	return walkDir(root, fn)
}

func walkDir(root string, fn func(string, os.DirEntry, error) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := root
		if path[len(path)-1] != '\\' && path[len(path)-1] != '/' {
			path += `\`
		}
		path += e.Name()
		if err := fn(path, e, nil); err != nil {
			return err
		}
		if e.IsDir() {
			if err := walkDir(path, fn); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}
