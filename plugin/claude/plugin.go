// Package claudeplugin is the Claude Code plugin, carried inside the binary.
//
// The plugin was installed from this repository, which a person handed only
// the binary — by brew, or as a file — does not have. So the binary carries
// the files of this folder, and writes them out as a marketplace Claude Code
// installs from: the plugin then always matches the cairn it calls. The files
// stay here, where Claude Code reads them from a clone; this is only what
// makes them travel. See docs/claude-code.md.
package claudeplugin

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// files is the plugin as it lies in this folder. The manifest's folder is
// named outright: a pattern skips a name starting with a dot.
//
//go:embed .claude-plugin/plugin.json .mcp.json hooks skills
var files embed.FS

// Name is what the plugin and its marketplace are called, and the folder the
// plugin is written into under the marketplace.
const Name = "cairn"

// ErrNotItsFolder is returned for a folder that holds something other than a
// marketplace cairn wrote. Writing again clears the plugin written before, and
// a folder named by mistake would lose what it held.
var ErrNotItsFolder = errors.New("the folder holds something cairn did not write; name an empty one, or one cairn wrote the plugin into")

// Write writes the plugin into dir as a marketplace: dir/.claude-plugin holds
// the marketplace, and dir/cairn the plugin, as Claude Code asks for them,
// both saying the version of the cairn writing them — Claude Code reads it to
// tell an update from what it already has. What an older cairn wrote there
// goes first, so nothing it carried and this one does not is left behind.
func Write(dir, version string) error {
	if err := ours(dir); err != nil {
		return err
	}
	market, err := marketplace(version)
	if err != nil {
		return err
	}
	manifest, err := versioned(version)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(dir, Name)); err != nil {
		return fmt.Errorf("clearing the plugin written before: %w", err)
	}
	if err := put(filepath.Join(dir, ".claude-plugin", "marketplace.json"), market, 0o644); err != nil {
		return err
	}
	return fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		held, err := files.ReadFile(path)
		if err != nil {
			return err
		}
		if path == manifestPath {
			held = manifest
		}
		return put(filepath.Join(dir, Name, filepath.FromSlash(path)), held, modeOf(path))
	})
}

// ours refuses a folder that holds anything but a marketplace cairn wrote. A
// folder that is not there yet, or is empty, holds nothing to lose.
func ours(dir string) error {
	held, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && len(held) == 0) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", dir, err)
	}
	said, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "marketplace.json"))
	if err != nil {
		return fmt.Errorf("%s: %w", dir, ErrNotItsFolder)
	}
	var market struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(said, &market) != nil || market.Name != Name {
		return fmt.Errorf("%s: %w", dir, ErrNotItsFolder)
	}
	return nil
}

// manifestPath is where the plugin's own manifest lies in this folder.
const manifestPath = ".claude-plugin/plugin.json"

// versioned is the plugin's manifest with the version of the cairn writing it.
func versioned(version string) ([]byte, error) {
	said, err := files.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var manifest map[string]any
	if err := json.Unmarshal(said, &manifest); err != nil {
		return nil, fmt.Errorf("reading the plugin's manifest: %w", err)
	}
	manifest["version"] = version
	return json.MarshalIndent(manifest, "", "  ")
}

// marketplace is the marketplace offering the plugin, read off the plugin's
// own manifest rather than written a second time beside it.
func marketplace(version string) ([]byte, error) {
	said, err := files.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var plugin struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Author      json.RawMessage `json:"author"`
	}
	if err := json.Unmarshal(said, &plugin); err != nil {
		return nil, fmt.Errorf("reading the plugin's manifest: %w", err)
	}
	return json.MarshalIndent(map[string]any{
		"name":        Name,
		"description": plugin.Description,
		"owner":       plugin.Author,
		"plugins": []map[string]string{{
			"name":        plugin.Name,
			"source":      "./" + Name,
			"description": plugin.Description,
			"version":     version,
		}},
	}, "", "  ")
}

// modeOf is how a file of the plugin is written: a script is run by Claude
// Code, and everything else is only read.
func modeOf(path string) os.FileMode {
	if strings.HasSuffix(path, ".sh") {
		return 0o755
	}
	return 0o644
}

// put writes one file, the folders above it with it.
func put(path string, held []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("making room for %s: %w", path, err)
	}
	if err := os.WriteFile(path, held, mode); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	// WriteFile keeps the mode of a file already there, and the umask may
	// have taken bits off a new one: the mode is said once more, outright.
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
