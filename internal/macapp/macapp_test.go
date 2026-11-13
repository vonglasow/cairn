package macapp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func installedIn(t *testing.T) (string, []string) {
	t.Helper()
	home := t.TempDir()
	var signed []string
	wasSigning := Signing
	Signing = func(app string) error { signed = append(signed, app); return nil }
	t.Cleanup(func() { Signing = wasSigning })
	return home, signed
}

func TestInstall_writesTheApplicationInTheApplicationsFolder(t *testing.T) {
	home, _ := installedIn(t)
	var signed []string
	Signing = func(app string) error { signed = append(signed, app); return nil }
	app, err := Install(home, "/opt/it's here/cairn", "v1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if app != filepath.Join(home, "Applications", "cairn.app") {
		t.Errorf("installed at %s", app)
	}
	launcher, err := os.ReadFile(filepath.Join(app, "Contents", "MacOS", "cairn"))
	if err != nil || string(launcher) != "#!/bin/sh\nexec '/opt/it'\\''s here/cairn' gui \"$@\"\n" {
		t.Errorf("the launcher reads %q, %v", launcher, err)
	}
	if info, err := os.Stat(filepath.Join(app, "Contents", "MacOS", "cairn")); err != nil || info.Mode().Perm() != 0o755 {
		t.Errorf("the launcher is %v, %v", info.Mode(), err)
	}
	plist, err := os.ReadFile(filepath.Join(app, "Contents", "Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<key>CFBundleIdentifier</key>\n\t<string>" + Identifier + "</string>", "<key>CFBundleShortVersionString</key>\n\t<string>v1.2.3</string>",
		"<key>CFBundleExecutable</key>\n\t<string>cairn</string>", "<key>CFBundleIconFile</key>\n\t<string>cairn.icns</string>", "<key>CFBundlePackageType</key>\n\t<string>APPL</string>"} {
		if !strings.Contains(string(plist), want) {
			t.Errorf("Info.plist holds no %q:\n%s", want, plist)
		}
	}
	if len(signed) != 1 || signed[0] != app {
		t.Errorf("signed %v", signed)
	}
}

func TestInstall_writesAnIconMacOSReads(t *testing.T) {
	home, _ := installedIn(t)
	app, err := Install(home, "/usr/local/bin/cairn", "v1")
	if err != nil {
		t.Fatal(err)
	}
	icon, err := os.ReadFile(filepath.Join(app, "Contents", "Resources", "cairn.icns"))
	if err != nil {
		t.Fatal(err)
	}
	if string(icon[:4]) != "icns" || int(binary.BigEndian.Uint32(icon[4:8])) != len(icon) {
		t.Fatalf("not an icns file: %q, %d bytes", icon[:8], len(icon))
	}
	seen := map[string]bool{}
	for at := 8; at < len(icon); {
		kind, size := string(icon[at:at+4]), int(binary.BigEndian.Uint32(icon[at+4:at+8]))
		if size < 8 || at+size > len(icon) || string(icon[at+9:at+12]) != "PNG" && string(icon[at+8:at+12]) != "ARGB" {
			t.Fatalf("entry %q at %d is %d bytes, neither a PNG nor ARGB", kind, at, size)
		}
		if argb := string(icon[at+8:at+12]) == "ARGB"; argb != (kind == "ic04" || kind == "ic05") {
			t.Errorf("entry %q is ARGB %v: only ic04 and ic05 are, as iconutil writes them", kind, argb)
		}
		seen[kind] = true
		at += size
	}
	for _, kind := range []string{"ic04", "ic05", "ic11", "ic12", "ic07", "ic13", "ic08", "ic14", "ic09", "ic10"} {
		if !seen[kind] {
			t.Errorf("no %s entry: %v", kind, seen)
		}
	}
}

func TestInstall_replacesItsOwnApplicationAndNoOtherOne(t *testing.T) {
	home, _ := installedIn(t)
	if _, err := Install(home, "/old/cairn", "v1"); err != nil {
		t.Fatal(err)
	}
	app, err := Install(home, "/new/cairn", "v2")
	if err != nil {
		t.Fatal(err)
	}
	if launcher, _ := os.ReadFile(filepath.Join(app, "Contents", "MacOS", "cairn")); !strings.Contains(string(launcher), "/new/cairn") {
		t.Errorf("the launcher still reads %q", launcher)
	}
	other := filepath.Join(t.TempDir(), "Applications", "cairn.app")
	if err := os.MkdirAll(filepath.Join(other, "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "Contents", "Info.plist"), []byte("<string>org.someone.else</string>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(filepath.Dir(filepath.Dir(other)), "/new/cairn", "v2"); !errors.Is(err, ErrNotOurs) {
		t.Errorf("err = %v, want ErrNotOurs", err)
	}
}

func TestInstall_saysWhenSigningFails(t *testing.T) {
	home, _ := installedIn(t)
	Signing = func(string) error { return errors.New("no codesign here") }
	if _, err := Install(home, "/usr/local/bin/cairn", "v1"); err == nil || !strings.Contains(err.Error(), "no codesign here") {
		t.Errorf("err = %v", err)
	}
}

func TestSigning_signsAnApplicationForThisMachine(t *testing.T) {
	home, _ := installedIn(t)
	Signing = func(string) error { return nil }
	app, err := Install(home, "/usr/local/bin/cairn", "v1")
	if err != nil {
		t.Fatal(err)
	}
	err = signing(app)
	if runtime.GOOS == "darwin" && err != nil {
		t.Errorf("signing on a Mac: %v", err)
	}
	if runtime.GOOS != "darwin" && err == nil {
		t.Error("signing went through where there is no codesign")
	}
	if err := signing(filepath.Join(t.TempDir(), "nothing.app")); err == nil {
		t.Error("signing nothing went through")
	}
}

func TestInstall_saysWhereItCouldNotMakeTheApplication(t *testing.T) {
	home, _ := installedIn(t)
	applications := filepath.Join(home, "Applications")
	if err := os.Mkdir(applications, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(applications, 0o755) })
	if _, err := Install(home, "/usr/local/bin/cairn", "v1"); err == nil {
		t.Error("an application was made in a folder nobody may write")
	}
}

func TestInstall_saysWhatItCouldNotWrite(t *testing.T) {
	_, _ = installedIn(t)
	home := filepath.Join(t.TempDir(), "a file")
	if err := os.WriteFile(home, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(home, "/usr/local/bin/cairn", "v1"); err == nil {
		t.Error("an application was written under a file")
	}
}

func TestInstall_drawsTheIconToItsCornersForMacOSToRound(t *testing.T) {
	home, _ := installedIn(t)
	app, err := Install(home, "/usr/local/bin/cairn", "v1")
	if err != nil {
		t.Fatal(err)
	}
	icon, err := os.ReadFile(filepath.Join(app, "Contents", "Resources", "cairn.icns"))
	if err != nil {
		t.Fatal(err)
	}
	for at := 8; at < len(icon); {
		size := int(binary.BigEndian.Uint32(icon[at+4 : at+8]))
		if string(icon[at+8:at+12]) == "ARGB" {
			if icon[at+13] != 0xff {
				t.Errorf("%s: its corner is not drawn", icon[at:at+4])
			}
			at += size
			continue
		}
		drawn, err := png.Decode(bytes.NewReader(icon[at+8 : at+size]))
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, alpha := drawn.At(0, 0).RGBA(); alpha != 0xffff {
			t.Errorf("%s: its corner is not drawn", icon[at:at+4])
		}
		at += size
	}
}

func TestPacked_writesARunOfThreeOrMoreAsOneAndTheRestAsItIs(t *testing.T) {
	for _, c := range []struct{ plane, want []byte }{
		{[]byte{5, 5, 5, 5, 1, 2}, []byte{0x81, 5, 0x01, 1, 2}},
		{[]byte{7, 7, 7}, []byte{0x80, 7}},
		{[]byte{1, 2, 2, 3}, []byte{0x03, 1, 2, 2, 3}},
		{bytes.Repeat([]byte{9}, 131), []byte{0xff, 9, 0x00, 9}},
		{bytes.Repeat([]byte{1, 2}, 65), append([]byte{0x7f}, append(bytes.Repeat([]byte{1, 2}, 64), 0x01, 1, 2)...)},
	} {
		if got := packed(c.plane); !bytes.Equal(got, c.want) {
			t.Errorf("packed(%v) = %v, want %v", c.plane, got, c.want)
		}
	}
}

func TestARGB_isTheAlphaThenTheRedTheGreenAndTheBlueOfEveryPixelPacked(t *testing.T) {
	drawn := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	copy(drawn.Pix, []byte{1, 5, 9, 255, 2, 6, 9, 255, 3, 7, 9, 255, 4, 8, 9, 255})
	var said bytes.Buffer
	if err := png.Encode(&said, drawn); err != nil {
		t.Fatal(err)
	}
	got, err := argb(said.Bytes())
	want := []byte{'A', 'R', 'G', 'B', 0x81, 255, 0x03, 1, 2, 3, 4, 0x03, 5, 6, 7, 8, 0x81, 9}
	if err != nil || !bytes.Equal(got, want) {
		t.Errorf("argb = %v, %v; want %v", got, err, want)
	}
	if _, err := argb([]byte("not a picture")); err == nil {
		t.Error("argb read what is not a PNG")
	}
}
