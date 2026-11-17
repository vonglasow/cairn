package linuxapp

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
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

func TestALogoRewrittenMakesTheIconCacheStale(t *testing.T) {
	data := t.TempDir()
	theme := filepath.Join(data, "icons", "hicolor")
	for _, size := range sizes {
		side := strconv.Itoa(size)
		if err := os.MkdirAll(filepath.Join(theme, side+"x"+side, "apps"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(theme, "48x48", "apps", Name+".png"), []byte("an old logo"), 0o644); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(theme, "icon-theme.cache")
	long := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.WriteFile(cache, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{cache, theme} {
		if err := os.Chtimes(path, long, long); err != nil {
			t.Fatal(err)
		}
	}
	if err := SettleIn(data, "/usr/bin/cairn"); err != nil {
		t.Fatal(err)
	}
	if dir, _ := os.Stat(theme); !dir.ModTime().After(long) {
		t.Errorf("the theme was left at %v, as old as its cache: the old logo stays shown", dir.ModTime())
	}
	if err := os.Chtimes(theme, long, long); err != nil {
		t.Fatal(err)
	}
	if err := SettleIn(data, "/usr/bin/cairn"); err != nil {
		t.Fatal(err)
	}
	if dir, _ := os.Stat(theme); dir.ModTime().After(long) {
		t.Error("the theme was touched though no logo changed")
	}
}
