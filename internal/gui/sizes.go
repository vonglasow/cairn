package gui

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// SizesKept keeps how wide and how tall the window was left, past it closing, for the next one to open with.
type SizesKept interface {
	Keep(Sizes) error
	Kept() (Sizes, error)
}

// SizesAt keeps the sizes in a file of their own.
func SizesAt(path string) SizesKept { return fileSizes(path) }

type fileSizes string

func (f fileSizes) Keep(s Sizes) error {
	written, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(string(f)), 0o700); err != nil {
		return err
	}
	return os.WriteFile(string(f), written, 0o600)
}

func (f fileSizes) Kept() (Sizes, error) {
	read, err := os.ReadFile(string(f))
	if errors.Is(err, fs.ErrNotExist) {
		return Sizes{}, nil
	}
	if err != nil {
		return Sizes{}, err
	}
	var s Sizes
	if err := json.Unmarshal(read, &s); err != nil {
		return Sizes{}, err
	}
	return s, nil
}

type noSizes struct{}

func (noSizes) Keep(Sizes) error     { return nil }
func (noSizes) Kept() (Sizes, error) { return Sizes{}, nil }

// KeepingSizes is the window opening with the sizes kept where it is told, and keeping there every size it is given.
func (w Window) KeepingSizes(k SizesKept) Window {
	w.kept = k
	sizes, err := k.Kept()
	if err != nil {
		return w.failing(err)
	}
	w.sizes = sizes
	return w
}

func (w Window) resized(s Sizes) Window {
	w.sizes = s
	if err := w.kept.Keep(s); err != nil {
		return w.failing(err)
	}
	return w
}
