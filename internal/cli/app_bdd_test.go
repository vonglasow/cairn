package cli

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/internal/linuxapp"
	"cairn/internal/logo"
	"cairn/internal/macapp"
	"cairn/test"
)

func initApp(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerApp(sc)
	}
}

func TestAcceptance_App(t *testing.T) {
	runOneAtATime(t, "app.feature", initApp(t), "")
}

func (w *world) registerApp(sc *godog.ScenarioContext) {
	var signed []string
	var data, tools string
	wasHere, wasLinux, wasSigning, wasRunning := macapp.Here, linuxapp.Here, macapp.Signing, build.Running
	const running = "/usr/local/bin/cairn"
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		signed = nil
		macapp.Here = true
		build.Running = func() (string, error) { return running, nil }
		macapp.Signing = func(app string) error {
			signed = append(signed, app)
			return nil
		}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		macapp.Here, linuxapp.Here, macapp.Signing, build.Running = wasHere, wasLinux, wasSigning, wasRunning
		return ctx, err
	})
	installed := func() string { return filepath.Join(w.Defaults, "Applications", "cairn.app") }
	sc.Step("^I ran `([^`]*)`$", func(line string) error { return w.Seed(line) })
	sc.Step(`^this is neither a Mac nor Linux$`, func() { macapp.Here, linuxapp.Here = false, false })
	sc.Given(`^this is a Linux desktop$`, func() error {
		macapp.Here, linuxapp.Here = false, true
		w.Setenv("XDG_DATA_HOME", "")
		data = filepath.Join(w.Defaults, ".local", "share")
		tools = filepath.Join(w.Defaults, "tools")
		w.SetProcessEnv("PATH", tools)
		if err := os.MkdirAll(tools, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(tools, "xdg-desktop-menu"), []byte("#!/bin/sh\necho \"$*\" >> '"+filepath.Join(tools, "refreshed")+"'\n"), 0o755)
	})
	sc.Given(`^my desktop keeps its data in a folder of its own$`, func() {
		data = filepath.Join(w.Defaults, "data of its own")
		w.Setenv("XDG_DATA_HOME", data)
	})
	sc.Given(`^the desktop has no tool to refresh its menu$`, func() error {
		return os.Remove(filepath.Join(tools, "xdg-desktop-menu"))
	})
	entry := func() string { return filepath.Join(data, "applications", "cairn.desktop") }
	const theirs = "[Desktop Entry]\nName=mine\nExec=mine\n"
	sc.Given(`^a desktop entry of cairn's name written by somebody else$`, func() error {
		if err := os.MkdirAll(filepath.Dir(entry()), 0o755); err != nil {
			return err
		}
		return os.WriteFile(entry(), []byte(theirs), 0o644)
	})
	saysWhere := func() error { return test.WantContains(w.Out.Stdout, entry(), "the answer") }
	sc.Step(`^it says the entry is in my applications folder$`, saysWhere)
	sc.Step(`^it says the entry is in that folder$`, saysWhere)
	sc.Step(`^the entry starts the window of the cairn that wrote it$`, func() error {
		said, err := os.ReadFile(entry())
		if err != nil {
			return err
		}
		if !strings.Contains(string(said), "\nExec="+running+" gui\n") || !strings.Contains(string(said), "\nIcon=cairn\n") {
			return fmt.Errorf("the entry says:\n%s\nwant Exec=%s gui and Icon=cairn", said, running)
		}
		held, err := os.ReadDir(filepath.Dir(entry()))
		if err != nil || len(held) != 1 {
			return fmt.Errorf("the applications folder holds %v, %w: want the one entry", held, err)
		}
		return nil
	})
	sc.Step(`^cairn's logo is in my icon theme at every size$`, func() error {
		for _, side := range []string{"16", "32", "48", "64", "128", "256", "512"} {
			png, err := os.ReadFile(filepath.Join(data, "icons", "hicolor", side+"x"+side, "apps", "cairn.png"))
			if err != nil || !strings.HasPrefix(string(png), "\x89PNG") {
				return fmt.Errorf("no logo at %s pixels: %w", side, err)
			}
		}
		return nil
	})
	refreshed := func(times int) error {
		said, _ := os.ReadFile(filepath.Join(tools, "refreshed"))
		if want := strings.Repeat("forceupdate\n", times); string(said) != want {
			return fmt.Errorf("the menu tool was run with %q, want %q", said, want)
		}
		return nil
	}
	sc.Step(`^the desktop's menu was refreshed$`, func() error { return refreshed(1) })
	sc.Step(`^the desktop's menu was refreshed twice$`, func() error { return refreshed(2) })
	sc.Step(`^it says that entry is left alone$`, func() error {
		return test.WantContains(w.Out.Stderr, "left alone", "the refusal")
	})
	sc.Step(`^the entry is still the one somebody else wrote$`, func() error {
		if said, err := os.ReadFile(entry()); err != nil || string(said) != theirs {
			return fmt.Errorf("the entry now says %q, %w", said, err)
		}
		return nil
	})
	sc.Step(`^it says the application is in my Applications folder$`, func() error {
		return test.WantContains(w.Out.Stdout, installed(), "the answer")
	})
	sc.Step(`^the application runs the window of the cairn that wrote it$`, func() error {
		launcher, err := os.ReadFile(filepath.Join(installed(), "Contents", "MacOS", "cairn"))
		if err != nil {
			return err
		}
		if !strings.Contains(string(launcher), "exec '"+running+"' gui") {
			return fmt.Errorf("the application runs:\n%s\nwant %s gui", launcher, running)
		}
		info, err := os.Stat(filepath.Join(installed(), "Contents", "MacOS", "cairn"))
		if err != nil {
			return err
		}
		if info.Mode()&0o100 == 0 {
			return fmt.Errorf("the launcher cannot be run: %s", info.Mode())
		}
		return nil
	})
	sc.Step(`^the application carries cairn's icon and its version$`, func() error {
		plist, err := os.ReadFile(filepath.Join(installed(), "Contents", "Info.plist"))
		if err != nil {
			return err
		}
		for _, want := range []string{"<string>cairn</string>", "<string>" + build.Read().Version + "</string>", "<string>cairn.icns</string>"} {
			if !strings.Contains(string(plist), want) {
				return fmt.Errorf("Info.plist holds no %s:\n%s", want, plist)
			}
		}
		icon, err := os.ReadFile(filepath.Join(installed(), "Contents", "Resources", "cairn.icns"))
		if err != nil || !strings.HasPrefix(string(icon), "icns") {
			return fmt.Errorf("no icns icon: %w", err)
		}
		return nil
	})
	sc.Step(`^every image of the application's icon is opaque to its corners$`, func() error {
		images, err := iconImages(filepath.Join(installed(), "Contents", "Resources", "cairn.icns"))
		if err != nil {
			return err
		}
		if len(images) == 0 {
			return fmt.Errorf("the icon holds no image")
		}
		for kind, drawn := range images {
			b := drawn.Bounds()
			for _, corner := range []image.Point{b.Min, {b.Max.X - 1, b.Min.Y}, {b.Min.X, b.Max.Y - 1}, b.Max.Sub(image.Pt(1, 1))} {
				if _, _, _, alpha := drawn.At(corner.X, corner.Y).RGBA(); alpha != 0xffff {
					return fmt.Errorf("%s: the corner %v is not opaque", kind, corner)
				}
			}
		}
		return nil
	})
	sc.Step(`^the application's icon holds an image for every size macOS draws an application at, from 16 points to 512, at one and two pixels a point$`, func() error {
		images, err := iconImages(filepath.Join(installed(), "Contents", "Resources", "cairn.icns"))
		if err != nil {
			return err
		}
		for kind := range macOSKinds {
			if _, ok := images[kind]; !ok {
				return fmt.Errorf("the icon has no %s, the %d pixels macOS reads from it", kind, macOSKinds[kind])
			}
		}
		for kind := range images {
			if _, ok := macOSKinds[kind]; !ok {
				return fmt.Errorf("the icon holds %s, which macOS does not read as an image of an application", kind)
			}
		}
		return nil
	})
	sc.Step(`^each image of the application's icon is the logo at the size its kind says$`, func() error {
		images, err := iconImages(filepath.Join(installed(), "Contents", "Resources", "cairn.icns"))
		if err != nil {
			return err
		}
		for kind, drawn := range images {
			side := macOSKinds[kind]
			raw, err := logo.Edge(side)
			if err != nil {
				return fmt.Errorf("%s: %w", kind, err)
			}
			want, err := png.Decode(bytes.NewReader(raw))
			if err != nil {
				return err
			}
			if drawn.Bounds() != want.Bounds() {
				return fmt.Errorf("%s is %v, want the logo's %d pixels", kind, drawn.Bounds(), side)
			}
			for y := range side {
				for x := range side {
					if !sameColour(drawn.At(x, y), want.At(x, y)) {
						return fmt.Errorf("%s at %d,%d is %v, the logo is %v there", kind, x, y, drawn.At(x, y), want.At(x, y))
					}
				}
			}
		}
		return nil
	})
	sc.Step(`^the application was signed for this machine$`, func() error {
		if len(signed) != 1 || signed[0] != installed() {
			return fmt.Errorf("signed %v, want %s once", signed, installed())
		}
		return nil
	})
	sc.Step(`^it says an application is installed on a Mac or a Linux desktop only$`, func() error {
		return test.WantContains(w.Out.Stderr, "only on a Mac or a Linux desktop", "the refusal")
	})
}

var macOSKinds = map[string]int{
	"ic04": 16, "ic05": 32, "ic11": 32, "ic12": 64, "ic07": 128,
	"ic13": 256, "ic08": 256, "ic14": 512, "ic09": 512, "ic10": 1024,
}

func sameColour(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar>>8 == br>>8 && ag>>8 == bg>>8 && ab>>8 == bb>>8 && aa>>8 == ba>>8
}

func iconImages(path string) (map[string]image.Image, error) {
	icon, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	images := map[string]image.Image{}
	for at := 8; at+8 <= len(icon); {
		kind, length := string(icon[at:at+4]), int(binary.BigEndian.Uint32(icon[at+4:at+8]))
		data := icon[at+8 : at+length]
		var drawn image.Image
		if bytes.HasPrefix(data, []byte("ARGB")) {
			drawn, err = unpackedARGB(data[4:], macOSKinds[kind])
		} else {
			drawn, err = png.Decode(bytes.NewReader(data))
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", kind, err)
		}
		images[kind] = drawn
		at += length
	}
	return images, nil
}

func unpackedARGB(packed []byte, side int) (image.Image, error) {
	planes := make([]byte, 0, 4*side*side)
	for i := 0; i < len(packed); {
		c := int(packed[i])
		switch {
		case c >= 0x80 && i+1 < len(packed):
			planes = append(planes, bytes.Repeat(packed[i+1:i+2], c-125)...)
			i += 2
		case c < 0x80 && i+1+c+1 <= len(packed):
			planes = append(planes, packed[i+1:i+2+c]...)
			i += c + 2
		default:
			return nil, fmt.Errorf("the packed planes end inside a run")
		}
	}
	if len(planes) != 4*side*side {
		return nil, fmt.Errorf("the planes unpack to %d bytes, want %d", len(planes), 4*side*side)
	}
	drawn := image.NewNRGBA(image.Rect(0, 0, side, side))
	n := side * side
	for p := range n {
		drawn.Pix[4*p], drawn.Pix[4*p+1], drawn.Pix[4*p+2], drawn.Pix[4*p+3] = planes[n+p], planes[2*n+p], planes[3*n+p], planes[p]
	}
	return drawn, nil
}
