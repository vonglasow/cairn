package files_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func silentEnv(string) (string, bool) { return "", false }

func TestRead_refusesAPlaceTheEnvironmentNamesThatIsNotAbsolute(t *testing.T) {
	for _, variable := range []string{files.StoreVariable, files.NotesVariable} {
		env := func(name string) (string, bool) {
			if name == variable {
				return "relative/place", true
			}
			return "", false
		}
		_, err := files.Read(t.TempDir(), env)
		if !errors.Is(err, files.ErrNotAbsolute) || !strings.Contains(err.Error(), variable+`="relative/place"`) {
			t.Errorf("%s: Read = %v, want ErrNotAbsolute naming the variable and what it was set to", variable, err)
		}
	}
}

func TestRead_refusesAPlaceThatIsNotAbsolute(t *testing.T) {
	refused := []string{"vault", "./vault", "../vault", ""}
	accepted := []string{"~/vault", "/tmp/x"}

	writeConfig := func(t *testing.T, home, key, value string) {
		t.Helper()
		dir := filepath.Join(home, ".config", "cairn")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
		content := key + " = " + strconv.Quote(value) + "\n"
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
	}

	for _, key := range []string{"store", "notes"} {
		for _, place := range refused {
			t.Run(key+" refuses "+strconv.Quote(place), func(t *testing.T) {
				home := t.TempDir()
				writeConfig(t, home, key, place)

				_, err := files.Read(home, silentEnv)
				if err == nil {
					t.Fatalf("Read returned no error, want one refusing %q", place)
				}
				if !errors.Is(err, files.ErrNotAbsolute) {
					t.Errorf("err = %v, want it to be config.ErrNotAbsolute", err)
				}
				if !strings.HasPrefix(err.Error(), strconv.Quote(place)) {
					t.Errorf("err = %q, want it to begin with %q", err.Error(), strconv.Quote(place))
				}
			})
		}
		for _, place := range accepted {
			t.Run(key+" accepts "+strconv.Quote(place), func(t *testing.T) {
				home := t.TempDir()
				writeConfig(t, home, key, place)

				_, err := files.Read(home, silentEnv)
				if err != nil {
					t.Fatalf("Read returned an error for %q: %v", place, err)
				}
			})
		}
	}
}

func writtenConfig(t *testing.T, content string) string {
	t.Helper()
	home := t.TempDir()
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

func TestRead_saysFromWhenAndHowOftenDoneTasksAreArchived(t *testing.T) {
	read, err := files.Read(writtenConfig(t, "[archive]\nstart = 2026-10-21\ndays = 10\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	start, days, ok := read.Archive()
	if !ok || start != "2026-10-21" || days != 10 {
		t.Errorf("archiving counts from %q in periods of %d days (%v)", start, days, ok)
	}
	unsaid, err := files.Read(writtenConfig(t, "[archive]\nstart = 2026-10-21\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if _, days, _ := unsaid.Archive(); days != 14 {
		t.Errorf("a period whose length is not said lasts %d days, want 14", days)
	}
	none, _ := files.Read(t.TempDir(), silentEnv)
	if _, _, ok := none.Archive(); ok {
		t.Error("with no file, tasks are archived on their own")
	}
	for _, refused := range []struct{ written, names string }{
		{"[archive]\nstart = 2026-10-21\ndays = 0\n", "archive.days"},
		{"[archive]\nstart = \"soon\"\n", "archive.start"},
		{"[archive]\ndays = 0\n", "archive.days"},
	} {
		_, err := files.Read(writtenConfig(t, refused.written), silentEnv)
		if !errors.Is(err, files.ErrArchive) || !strings.Contains(err.Error(), refused.names) {
			t.Errorf("%q gives %v, want ErrArchive naming %s", refused.written, err, refused.names)
		}
	}
}

func TestRead_saysWhichIntroTheBoardArrivesBy(t *testing.T) {
	read, err := files.Read(writtenConfig(t, "[window]\nintro = \"deck\"\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := read.Intro(); got != files.IntroDeck {
		t.Errorf("the intro is %q, want deck", got)
	}
	none, _ := files.Read(t.TempDir(), silentEnv)
	if got := none.Intro(); got != files.IntroRipple {
		t.Errorf("with no file, the intro is %q, want ripple", got)
	}
	if _, err := files.Read(writtenConfig(t, "[window]\nintro = \"confetti\"\n"), silentEnv); !errors.Is(err, files.ErrIntro) || !strings.Contains(err.Error(), "window.intro") {
		t.Errorf("err = %v, want ErrIntro naming window.intro", err)
	}
	if len(files.Intros()) != 11 {
		t.Errorf("the intros are %v, want eleven", files.Intros())
	}
}

func TestRead_saysWhetherTasksComeBack(t *testing.T) {
	read, err := files.Read(writtenConfig(t, "[series]\non = true\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if !read.Recurrence() {
		t.Error("on = true does not turn recurrence on")
	}
	none, _ := files.Read(t.TempDir(), silentEnv)
	if none.Recurrence() {
		t.Error("with no file, recurrence is on")
	}
	_, err = files.Read(writtenConfig(t, "[series]\non = 3\n"), silentEnv)
	if !errors.Is(err, files.ErrRecurrence) || !strings.Contains(err.Error(), "series.on") {
		t.Errorf("err = %v, want ErrRecurrence naming series.on", err)
	}
}

func TestRead_stillReadsTheOldNameOfTheSeriesKey(t *testing.T) {
	read, err := files.Read(writtenConfig(t, "[recurrence]\non = true\n"), silentEnv)
	if err != nil || !read.Recurrence() || !slices.Contains(read.Written(), "series.on") {
		t.Errorf("[recurrence] on = true reads %v (%v), written %v", read.Recurrence(), err, read.Written())
	}
}
