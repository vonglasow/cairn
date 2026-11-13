package macapp

import (
	"bytes"
	"encoding/binary"
	"errors"
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
		if size < 8 || at+size > len(icon) || string(icon[at+9:at+12]) != "PNG" {
			t.Fatalf("entry %q at %d is %d bytes, not a PNG", kind, at, size)
		}
		seen[kind] = true
		at += size
	}
	for _, kind := range []string{"icp4", "icp5", "icp6", "ic07", "ic08", "ic09", "ic10"} {
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
