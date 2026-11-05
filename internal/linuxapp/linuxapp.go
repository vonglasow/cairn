// Package linuxapp writes cairn as a Linux desktop application: its entry in the applications menu, and its logo in the icon theme.
package linuxapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"cairn/internal/logo"
)

// Name is what a desktop knows cairn by: its entry, its icon, and the window's class.
const Name = "cairn"

// Here says whether this is Linux, where a desktop entry can be installed.
var Here = runtime.GOOS == "linux"

// ErrNotOurs turns installing down over an entry somebody else wrote.
var ErrNotOurs = errors.New("a desktop entry of that name is already there, and cairn did not write it; it is left alone")

var sizes = []int{16, 32, 48, 64, 128, 256, 512}

const writtenByCairn = "X-Cairn-Written=true"

// DataHome is where a Linux desktop looks for a person's own entries and icons: XDG_DATA_HOME, or the home's .local/share.
func DataHome(home string, lookup func(string) (string, bool)) string {
	if data, ok := lookup("XDG_DATA_HOME"); ok && data != "" {
		return data
	}
	return filepath.Join(home, ".local", "share")
}

// Install writes under data the entry starting binary's window and the logo it names, then refreshes the desktop's menu; an entry somebody else wrote is refused.
func Install(data, binary string) (entry, starts string, err error) {
	entry, starts = entryIn(data), Starter(binary, exec.LookPath)
	if _, other := theirs(entry); other {
		return "", "", fmt.Errorf("%s: %w", entry, ErrNotOurs)
	}
	if err := SettleIn(data, starts); err != nil {
		return "", "", err
	}
	return entry, starts, refreshed()
}

func refreshed() error {
	tool, err := exec.LookPath("xdg-desktop-menu")
	if err != nil {
		return nil
	}
	if said, err := exec.Command(tool, "forceupdate").CombinedOutput(); err != nil {
		return fmt.Errorf("refreshing the menu: %w: %s", err, strings.TrimSpace(string(said)))
	}
	return nil
}

// SettleIn writes under data the logo at every size a desktop asks for, and the desktop entry that names it, starting the window with starts; an entry somebody else wrote is left as it is.
func SettleIn(data, starts string) error {
	for _, size := range sizes {
		held, err := logo.At(size)
		if err != nil {
			return err
		}
		side := strconv.Itoa(size)
		if err := writtenIfChanged(filepath.Join(data, "icons", "hicolor", side+"x"+side, "apps", Name+".png"), held); err != nil {
			return err
		}
	}
	if _, other := theirs(entryIn(data)); other {
		return nil
	}
	return writtenIfChanged(entryIn(data), []byte(DesktopEntry(starts)))
}

func theirs(entry string) ([]byte, bool) {
	was, err := os.ReadFile(entry)
	return was, err == nil && !bytes.Contains(was, []byte(writtenByCairn))
}

func entryIn(data string) string { return filepath.Join(data, "applications", Name+".desktop") }

func keptIn(data string) string { return filepath.Join(data, Name, "entry-kept") }

// Entry is a desktop entry somebody other than cairn wrote, as far as cairn's logo goes: the folder it was found under, the icon it shows, and a sum of what it says, which a refusal to replace it is remembered by.
type Entry struct {
	Data, Icon string
	sum        string
	hides      bool
}

// Hiding reads the entry the desktop finds the window through, and says whether it is somebody else's and shows another icon than cairn's logo.
func Hiding(data string) Entry {
	was, other := theirs(entryIn(data))
	if !other {
		return Entry{Data: data}
	}
	icon := ""
	for line := range strings.Lines(string(was)) {
		if said, ok := strings.CutPrefix(strings.TrimSpace(line), "Icon="); ok && icon == "" {
			icon = said
		}
	}
	sum := sha256.Sum256(was)
	return Entry{Data: data, Icon: icon, sum: hex.EncodeToString(sum[:]), hides: icon != Name}
}

// Asks says whether the person is to be asked about the entry: it hides the logo, and it is not the one they last chose to keep.
func (e Entry) Asks() bool {
	if !e.hides {
		return false
	}
	kept, err := os.ReadFile(keptIn(e.Data))
	return err != nil || strings.TrimSpace(string(kept)) != e.sum
}

// Kept remembers that the entry is to be left as it is, until it changes.
func (e Entry) Kept() error { return writtenIfChanged(keptIn(e.Data), []byte(e.sum+"\n")) }

// Replaced writes cairn's own entry over it, starting the window with starts.
func (e Entry) Replaced(starts string) error {
	return writtenIfChanged(entryIn(e.Data), []byte(DesktopEntry(starts)))
}

// Starter is the program a desktop entry starts the window with: the full path the PATH leads to when it leads to the cairn running, which an upgrade keeps, or else the cairn running; never a bare name, which a desktop whose own PATH does not reach it would ignore the entry for.
func Starter(running string, onPath func(string) (string, error)) string {
	found, err := onPath(Name)
	if err != nil || !sameFile(found, running) {
		return running
	}
	if full, err := filepath.Abs(found); err == nil {
		return full
	}
	return running
}

func sameFile(a, b string) bool {
	a, errA := filepath.EvalSymlinks(a)
	b, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && a == b
}

// DesktopEntry is the entry a Linux desktop finds the window and its logo through: named, and classed, as the window is known.
func DesktopEntry(starts string) string {
	if strings.ContainsAny(starts, " \t\"") {
		starts = strconv.Quote(starts)
	}
	return strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=" + Name,
		"Comment=Objectives and tasks, and the time the work took",
		"Exec=" + starts + " gui",
		"Icon=" + Name,
		"StartupWMClass=" + Name,
		"Terminal=false",
		"Categories=Office;ProjectManagement;",
		writtenByCairn,
	}, "\n") + "\n"
}

func writtenIfChanged(path string, held []byte) error {
	if was, err := os.ReadFile(path); err == nil && bytes.Equal(was, held) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, held, 0o644)
}
