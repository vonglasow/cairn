package gui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTheEntryStartsCairnByAFullPath(t *testing.T) {
	dir := t.TempDir()
	running := filepath.Join(dir, "Cellar", Known)
	linked := filepath.Join(dir, "bin", Known)
	other := filepath.Join(dir, "other", Known)
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
	if err := os.WriteFile(Known, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	full, err := filepath.Abs(Known)
	if err != nil {
		t.Fatal(err)
	}
	got := Starter(full, func(string) (string, error) { return Known, nil })
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
				path := filepath.Join(data, "applications", Known+".desktop")
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
	path := filepath.Join(data, "applications", Known+".desktop")
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
