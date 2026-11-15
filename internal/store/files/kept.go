package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Kept is a file cairn keeps one thing in, whole: written over, read back, taken away.
type Kept string

// Keep writes the file whole, making its folder first.
func (k Kept) Keep(said []byte) error {
	if err := os.MkdirAll(filepath.Dir(string(k)), 0o700); err != nil {
		return err
	}
	return os.WriteFile(string(k), said, 0o600)
}

// Read is what the file holds; nothing when it is not there.
func (k Kept) Read() ([]byte, error) {
	read, err := os.ReadFile(string(k))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return read, err
}

// Drop takes the file away; a file not there is gone already.
func (k Kept) Drop() error {
	if err := os.Remove(string(k)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
