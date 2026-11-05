package linuxapp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheEntryStartsCairnByAFullPath(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "Cellar", Name)
	linked := filepath.Join(dir, "bin", Name)
	other := filepath.Join(dir, "other", Name)
	for _, made := range []string{running, other} {
		if err := os.MkdirAll(filepath.Dir(made), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(made, nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(linked), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(running, linked); err != nil {
		t.Fatal(err)
	}
	leadsTo := func(to string) func(string) (string, error) {
		return func(string) (string, error) { return to, nil }
	}
	nowhere := func(string) (string, error) { return "", errors.New("not on the PATH") }
	for _, c := range []struct {
		name   string
		onPath func(string) (string, error)
		want   string
	}{
		{"the PATH leads to it", leadsTo(linked), linked},
		{"the PATH leads to another", leadsTo(other), running},
		{"the PATH leads nowhere", nowhere, running},
		{"the PATH leads to a file that is not there", leadsTo(filepath.Join(dir, "gone")), running},
	} {
		if got := Starter(running, c.onPath); got != c.want {
			t.Errorf("%s: the entry starts %q, want %q", c.name, got, c.want)
		}
	}
}

func TestAStarterFoundByARelativePathIsMadeFull(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(Name, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	full, err := filepath.Abs(Name)
	if err != nil {
		t.Fatal(err)
	}
	got := Starter(full, func(string) (string, error) { return Name, nil })
	if got != full {
		t.Errorf("the entry starts %q, want %q", got, full)
	}
}

func TestAnEntryIsAskedAboutOnlyWhenItHidesTheLogo(t *testing.T) {
	for _, c := range []struct {
		what, written string
		asks          bool
	}{
		{"no entry", "", false},
		{"cairn's own", DesktopEntry("cairn"), false},
		{"somebody else's showing the logo", "[Desktop Entry]\nIcon=cairn\n", false},
		{"somebody else's showing another icon", "[Desktop Entry]\nIcon=view-list-details\n", true},
		{"somebody else's showing no icon", "[Desktop Entry]\nName=mine\n", true},
	} {
		t.Run(c.what, func(t *testing.T) {
			data := t.TempDir()
			if c.written != "" {
				path := filepath.Join(data, "applications", Name+".desktop")
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(c.written), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if got := Hiding(data); got.Asks() != c.asks {
				t.Errorf("asks %v, want %v: %+v", got.Asks(), c.asks, got)
			}
		})
	}
}

func TestAnEntryKeptIsNotAskedAboutUntilItChanges(t *testing.T) {
	data := t.TempDir()
	path := filepath.Join(data, "applications", Name+".desktop")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[Desktop Entry]\nIcon=view-list-details\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	hiding := Hiding(data)
	if hiding.Icon != "view-list-details" {
		t.Errorf("the entry shows %q", hiding.Icon)
	}
	if err := hiding.Kept(); err != nil {
		t.Fatal(err)
	}
	if Hiding(data).Asks() {
		t.Error("an entry kept is asked about again, unchanged")
	}
	if err := os.WriteFile(path, []byte("[Desktop Entry]\nIcon=other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !Hiding(data).Asks() {
		t.Error("an entry kept and then changed is not asked about")
	}
	if err := Hiding(data).Replaced("cairn"); err != nil {
		t.Fatal(err)
	}
	if Hiding(data).Asks() {
		t.Error("an entry replaced is still asked about")
	}
}

func TestDataHome_isTheDesktopsOwnFolderOrElseTheHomes(t *testing.T) {
	set := func(v string) func(string) (string, bool) {
		return func(key string) (string, bool) { return v, key == "XDG_DATA_HOME" && v != "" }
	}
	if got := DataHome("/home/me", set("/elsewhere")); got != "/elsewhere" {
		t.Errorf("with XDG_DATA_HOME set, %s", got)
	}
	if got, want := DataHome("/home/me", set("")), filepath.Join("/home/me", ".local", "share"); got != want {
		t.Errorf("with nothing set, %s, want %s", got, want)
	}
}

func menuTool(t *testing.T) string {
	t.Helper()
	tools := t.TempDir()
	t.Setenv("PATH", tools)
	said := filepath.Join(tools, "said")
	if err := os.WriteFile(filepath.Join(tools, "xdg-desktop-menu"), []byte("#!/bin/sh\necho \"$*\" >> '"+said+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return said
}

func TestInstall_writesTheEntryAndTheLogoAndRefreshesTheMenu(t *testing.T) {
	said := menuTool(t)
	data := t.TempDir()
	entry, starts, err := Install(data, "/opt/it's here/cairn")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(data, "applications", Name+".desktop"); entry != want {
		t.Errorf("the entry is %s, want %s", entry, want)
	}
	if starts != "/opt/it's here/cairn" {
		t.Errorf("it starts %s", starts)
	}
	if written, err := os.ReadFile(entry); err != nil || string(written) != DesktopEntry(starts) {
		t.Errorf("the entry says %q, %v", written, err)
	}
	for _, side := range []string{"16", "32", "48", "64", "128", "256", "512"} {
		if _, err := os.Stat(filepath.Join(data, "icons", "hicolor", side+"x"+side, "apps", Name+".png")); err != nil {
			t.Errorf("no logo at %s: %v", side, err)
		}
	}
	if ran, _ := os.ReadFile(said); string(ran) != "forceupdate\n" {
		t.Errorf("the menu tool was run with %q", ran)
	}
}

func TestInstall_leavesAnEntryOfSomebodyElsesAlone(t *testing.T) {
	said := menuTool(t)
	data := t.TempDir()
	path := filepath.Join(data, "applications", Name+".desktop")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[Desktop Entry]\nIcon=cairn\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Install(data, "/usr/bin/cairn"); !errors.Is(err, ErrNotOurs) {
		t.Errorf("over somebody else's entry: %v", err)
	}
	if was, _ := os.ReadFile(path); string(was) != "[Desktop Entry]\nIcon=cairn\n" {
		t.Errorf("their entry now says %q", was)
	}
	if ran, err := os.ReadFile(said); err == nil {
		t.Errorf("the menu was refreshed for nothing: %q", ran)
	}
}

func TestInstall_needsNoMenuToolButSaysWhenItFails(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, _, err := Install(t.TempDir(), "/usr/bin/cairn"); err != nil {
		t.Errorf("with no menu tool: %v", err)
	}
	tools := t.TempDir()
	t.Setenv("PATH", tools)
	if err := os.WriteFile(filepath.Join(tools, "xdg-desktop-menu"), []byte("#!/bin/sh\necho broken >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Install(t.TempDir(), "/usr/bin/cairn"); err == nil || !strings.Contains(err.Error(), "broken") {
		t.Errorf("a menu tool failing: %v", err)
	}
}

func TestInstall_saysWhatItCouldNotWrite(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	data := t.TempDir()
	if err := os.WriteFile(filepath.Join(data, "icons"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Install(data, "/usr/bin/cairn"); err == nil {
		t.Error("it wrote under a file")
	}
}
