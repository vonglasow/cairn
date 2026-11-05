// Package macapp writes cairn as a Mac application: a cairn.app in the person's Applications folder that opens the window of the cairn installed on the machine.
package macapp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"cairn/internal/logo"
)

// Identifier is the bundle identifier the application is known to macOS by, and recognised by when it is written again.
const Identifier = "com.github.vonglasow.cairn"

// Here says whether this is a Mac, where an application can be installed.
var Here = runtime.GOOS == "darwin"

// Signing signs the application for this machine alone, so macOS opens it without a developer's certificate.
var Signing = signing

func signing(app string) error {
	if said, err := exec.Command("codesign", "--force", "--sign", "-", app).CombinedOutput(); err != nil {
		return fmt.Errorf("signing %s: %w: %s", app, err, strings.TrimSpace(string(said)))
	}
	return nil
}

// ErrNotOurs turns installing down over an application somebody else wrote.
var ErrNotOurs = errors.New("an application of that name is already there, and cairn did not write it; it is left alone")

// Install writes cairn.app under home/Applications, its launcher running binary's window, and signs it; one cairn wrote before is replaced.
func Install(home, binary, version string) (string, error) {
	app := filepath.Join(home, "Applications", "cairn.app")
	if said, err := os.ReadFile(filepath.Join(app, "Contents", "Info.plist")); err == nil && !bytes.Contains(said, []byte(Identifier)) {
		return "", fmt.Errorf("%s: %w", app, ErrNotOurs)
	}
	if err := os.RemoveAll(app); err != nil {
		return "", err
	}
	icon, err := icns()
	if err != nil {
		return "", err
	}
	for _, f := range []struct {
		path string
		says []byte
		mode os.FileMode
	}{
		{filepath.Join(app, "Contents", "Info.plist"), []byte(fmt.Sprintf(plist, Identifier, version, version)), 0o644},
		{filepath.Join(app, "Contents", "MacOS", "cairn"), []byte("#!/bin/sh\nexec " + quoted(binary) + " gui \"$@\"\n"), 0o755},
		{filepath.Join(app, "Contents", "Resources", "cairn.icns"), icon, 0o644},
	} {
		if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(f.path, f.says, f.mode); err != nil {
			return "", err
		}
	}
	return app, Signing(app)
}

func quoted(path string) string { return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'" }

const plist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>cairn</string>
	<key>CFBundleDisplayName</key>
	<string>cairn</string>
	<key>CFBundleIdentifier</key>
	<string>%s</string>
	<key>CFBundleShortVersionString</key>
	<string>%s</string>
	<key>CFBundleVersion</key>
	<string>%s</string>
	<key>CFBundleExecutable</key>
	<string>cairn</string>
	<key>CFBundleIconFile</key>
	<string>cairn.icns</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
`

var icnsSizes = []struct {
	kind string
	size int
}{{"icp4", 16}, {"icp5", 32}, {"icp6", 64}, {"ic07", 128}, {"ic08", 256}, {"ic09", 512}, {"ic10", 1024}}

func icns() ([]byte, error) {
	var entries bytes.Buffer
	for _, s := range icnsSizes {
		png, err := logo.At(s.size)
		if err != nil {
			return nil, err
		}
		entries.WriteString(s.kind)
		_ = binary.Write(&entries, binary.BigEndian, uint32(8+len(png)))
		entries.Write(png)
	}
	var out bytes.Buffer
	out.WriteString("icns")
	_ = binary.Write(&out, binary.BigEndian, uint32(8+entries.Len()))
	out.Write(entries.Bytes())
	return out.Bytes(), nil
}
