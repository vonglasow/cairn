package macapp

import (
	"bytes"
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
