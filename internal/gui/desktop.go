package gui

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed logo/cairn-16.png logo/cairn-32.png logo/cairn-48.png logo/cairn-64.png logo/cairn-128.png logo/cairn-256.png logo/cairn-512.png
var logos embed.FS

var (
	desktopSizes = []int{16, 32, 48, 64, 128, 256, 512}
	x11Sizes     = []int{16, 32, 48, 64, 128}
)

const writtenByCairn = "X-Cairn-Written=true"

func logoAt(size int) ([]byte, error) {
	return logos.ReadFile("logo/cairn-" + strconv.Itoa(size) + ".png")
}

// DataHome is where a Linux desktop looks for a person's own entries and icons: XDG_DATA_HOME, or ~/.local/share.
func DataHome() string {
	if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		return data
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

// SettleIn writes under data the logo at every size a desktop asks for, and the desktop entry that names it, starting the window with starts; an entry somebody else wrote is left as it is.
func SettleIn(data, starts string) error {
	for _, size := range desktopSizes {
		held, err := logoAt(size)
		if err != nil {
			return err
		}
		side := strconv.Itoa(size)
		if err := writtenIfChanged(filepath.Join(data, "icons", "hicolor", side+"x"+side, "apps", Known+".png"), held); err != nil {
			return err
		}
	}
	entry := entryIn(data)
	if was, err := os.ReadFile(entry); err == nil && !strings.Contains(string(was), writtenByCairn) {
		return nil
	}
	return writtenIfChanged(entry, []byte(DesktopEntry(starts)))
}

func entryIn(data string) string { return filepath.Join(data, "applications", Known+".desktop") }

func keptIn(data string) string { return filepath.Join(data, Known, "entry-kept") }

// Entry is a desktop entry somebody other than cairn wrote, as far as cairn's logo goes: the folder it was found under, the icon it shows, and a sum of what it says, which a refusal to replace it is remembered by.
type Entry struct {
	Data, Icon string
	sum        string
	hides      bool
}

// Hiding reads the entry the desktop finds the window through, and says whether it is somebody else's and shows another icon than cairn's logo.
func Hiding(data string) Entry {
	was, err := os.ReadFile(entryIn(data))
	if err != nil || strings.Contains(string(was), writtenByCairn) {
		return Entry{Data: data}
	}
	icon := ""
	for line := range strings.Lines(string(was)) {
		if said, ok := strings.CutPrefix(strings.TrimSpace(line), "Icon="); ok && icon == "" {
			icon = said
		}
	}
	sum := sha256.Sum256(was)
	return Entry{Data: data, Icon: icon, sum: hex.EncodeToString(sum[:]), hides: icon != Known}
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
	found, err := onPath(Known)
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
		"Name=" + Known,
		"Comment=Objectives and tasks, and the time the work took",
		"Exec=" + starts + " gui",
		"Icon=" + Known,
		"StartupWMClass=" + Known,
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

// X11Icon is the logo as X11's _NET_WM_ICON holds it: at each size a bar asks for, its width, its height, then every pixel as ARGB, row by row.
func X11Icon() []uint32 {
	var held []uint32
	for _, size := range x11Sizes {
		raw, err := logoAt(size)
		if err != nil {
			panic(fmt.Sprintf("the logo at %d pixels is not embedded: %v", size, err))
		}
		drawn, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			panic(fmt.Sprintf("the logo at %d pixels does not decode: %v", size, err))
		}
		b := drawn.Bounds()
		held = append(held, uint32(b.Dx()), uint32(b.Dy()))
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(drawn.At(x, y)).(color.NRGBA)
				held = append(held, uint32(c.A)<<24|uint32(c.R)<<16|uint32(c.G)<<8|uint32(c.B))
			}
		}
	}
	return held
}

// Settling settles the window in on a Linux desktop under data, started with starts, and asks about an entry somebody else wrote that hides the logo, unless something else is being asked already.
func (w Window) Settling(data, starts string) Window {
	_ = SettleIn(data, starts)
	if hiding := Hiding(data); hiding.Asks() && w.screen != OnAsking {
		w.hiding, w.starts = hiding, starts
		w.screen, w.question = OnAsking, aboutHiding
	}
	return w
}

func (w Window) answeringHiding(replace bool) Window {
	w.screen, w.question = OnTheBoard, aboutDeleting
	err := w.hiding.Kept()
	if replace {
		err = w.hiding.Replaced(w.starts)
	}
	if err != nil {
		return w.failing(err)
	}
	return w
}
