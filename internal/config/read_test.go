package config_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/config"
)

func silentEnv(string) (string, bool) { return "", false }

type placesHeard struct{ store, notes string }

func (p *placesHeard) StoreNamedByTheFile(at string)        { p.store = at }
func (p *placesHeard) StoreNamedByTheEnvironment(at string) { p.store = at }
func (p *placesHeard) StoreNamedByNothing(at string)        { p.store = at }
func (p *placesHeard) NotesNamedByTheFile(at string)        { p.notes = at }
func (p *placesHeard) NotesNamedByTheEnvironment(at string) { p.notes = at }
func (p *placesHeard) NotesNamedByNothing(at string)        { p.notes = at }

func storeOf(p config.Places) string {
	heard := &placesHeard{}
	p.Tell(heard)
	return heard.store
}

func notesOf(p config.Places) string {
	heard := &placesHeard{}
	p.Tell(heard)
	return heard.notes
}

func TestRead_refusesAPlaceTheEnvironmentNamesThatIsNotAbsolute(t *testing.T) {
	for _, variable := range []string{config.StoreVariable, config.NotesVariable} {
		env := func(name string) (string, bool) {
			if name == variable {
				return "relative/place", true
			}
			return "", false
		}
		_, err := config.Read(t.TempDir(), env)
		if !errors.Is(err, config.ErrNotAbsolute) || !strings.Contains(err.Error(), variable+`="relative/place"`) {
			t.Errorf("%s: Read = %v, want ErrNotAbsolute naming the variable and what it was set to", variable, err)
		}
	}
}

func TestRead_takesTheStoreTheFileNames(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	content := "store = \"/tmp/quelquepart/cairn.db\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}

	places, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatalf("Read returned an error: %v", err)
	}

	if got := storeOf(places); got != "/tmp/quelquepart/cairn.db" {
		t.Errorf("the store = %q, want %q", got, "/tmp/quelquepart/cairn.db")
	}
}

func TestRead_fallsBackToTheHomeWhenNothingNamesTheStore(t *testing.T) {
	cases := []struct {
		name  string
		setup func(home string)
	}{
		{
			name: "file present but silent about store",
			setup: func(home string) {
				dir := filepath.Join(home, ".config", "cairn")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
				content := "notes = \"/tmp/unrelated\"\n"
				if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
			},
		},
		{
			name:  "no file at all",
			setup: func(home string) {},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			c.setup(home)

			places, err := config.Read(home, silentEnv)
			if err != nil {
				t.Fatalf("Read returned an error: %v", err)
			}

			want := filepath.Join(home, ".cairn", "cairn.db")
			if got := storeOf(places); got != want {
				t.Errorf("the store = %q, want %q", got, want)
			}
		})
	}
}

func TestRead_expandsALeadingTildeAgainstTheHome(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	content := "store = \"~/work/cairn.db\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}

	places, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatalf("Read returned an error: %v", err)
	}

	want := filepath.Join(home, "work", "cairn.db")
	if got := storeOf(places); got != want {
		t.Errorf("the store = %q, want %q", got, want)
	}
}

func TestRead_takesWhatTheEnvironmentSaysOverAnythingElse(t *testing.T) {
	withFile := func(home string) {
		dir := filepath.Join(home, ".config", "cairn")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
		content := "store = \"/tmp/quelquepart/cairn.db\"\n"
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
	}

	cases := []struct {
		name  string
		setup func(home string)
		env   func(string) (string, bool)
		want  func(home string) string
	}{
		{
			name:  "replaces what the file named",
			setup: withFile,
			env:   func(string) (string, bool) { return "/somewhere/of/my/own.db", true },
			want:  func(string) string { return "/somewhere/of/my/own.db" },
		},
		{
			name:  "replaces the default under the home",
			setup: func(string) {},
			env:   func(string) (string, bool) { return "/somewhere/of/my/own.db", true },
			want:  func(string) string { return "/somewhere/of/my/own.db" },
		},
		{
			name:  "undergoes the same tilde expansion as the file's",
			setup: func(string) {},
			env:   func(string) (string, bool) { return "~/work/cairn.db", true },
			want:  func(home string) string { return filepath.Join(home, "work", "cairn.db") },
		},
		{
			name:  "set but empty says nothing: the answer stays what it was",
			setup: withFile,
			env:   func(string) (string, bool) { return "", true },
			want:  func(string) string { return "/tmp/quelquepart/cairn.db" },
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			c.setup(home)

			places, err := config.Read(home, c.env)
			if err != nil {
				t.Fatalf("Read returned an error: %v", err)
			}

			if want := c.want(home); storeOf(places) != want {
				t.Errorf("the store = %q, want %q", storeOf(places), want)
			}
		})
	}
}

func TestRead_refusesANameItDoesNotKnow(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	content := "vault = \"/tmp/quelquepart\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}

	_, err := config.Read(home, silentEnv)
	if err == nil {
		t.Fatalf("Read returned no error, want one naming %q", "vault")
	}
	if !errors.Is(err, config.ErrUnknownKey) {
		t.Errorf("err = %v, want it to be config.ErrUnknownKey", err)
	}
	if !strings.Contains(err.Error(), "vault") {
		t.Errorf("err = %q, want it to name the unknown key %q", err.Error(), "vault")
	}
	if !strings.Contains(err.Error(), "store") {
		t.Errorf("err = %q, want it to name what cairn accepts instead", err.Error())
	}
}

func TestRead_refusesAFileItCannotMakeSenseOf(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	path := filepath.Join(dir, "config.toml")
	content := "not { toml : at all\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}

	_, err := config.Read(home, silentEnv)
	if err == nil {
		t.Fatalf("Read returned no error, want one naming %q", path)
	}
	if !errors.Is(err, config.ErrNotTOML) {
		t.Errorf("err = %v, want it to be config.ErrNotTOML", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("err = %q, want it to name the file %q", err.Error(), path)
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

				_, err := config.Read(home, silentEnv)
				if err == nil {
					t.Fatalf("Read returned no error, want one refusing %q", place)
				}
				if !errors.Is(err, config.ErrNotAbsolute) {
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

				_, err := config.Read(home, silentEnv)
				if err != nil {
					t.Fatalf("Read returned an error for %q: %v", place, err)
				}
			})
		}
	}
}

func TestRead_resolvesTheNotesTheSameWayAsTheStore(t *testing.T) {
	cases := []struct {
		name  string
		setup func(home string)
		env   func(string) (string, bool)
		want  func(home string) string
	}{
		{
			name: "the file names it",
			setup: func(home string) {
				dir := filepath.Join(home, ".config", "cairn")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
				content := "notes = \"~/mynotes\"\n"
				if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
			},
			env:  silentEnv,
			want: func(home string) string { return filepath.Join(home, "mynotes") },
		},
		{
			name:  "nothing names it",
			setup: func(string) {},
			env:   silentEnv,
			want:  func(home string) string { return filepath.Join(home, ".cairn", "notes") },
		},
		{
			name: "the environment names it, replacing both",
			setup: func(home string) {
				dir := filepath.Join(home, ".config", "cairn")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
				content := "notes = \"/tmp/quelquepart/notes\"\n"
				if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
					t.Fatalf("could not set up the fixture: %v", err)
				}
			},
			env: func(key string) (string, bool) {
				if key == config.NotesVariable {
					return "/somewhere/of/my/own/notes", true
				}
				return "", false
			},
			want: func(string) string { return "/somewhere/of/my/own/notes" },
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			c.setup(home)

			places, err := config.Read(home, c.env)
			if err != nil {
				t.Fatalf("Read returned an error: %v", err)
			}

			if want := c.want(home); notesOf(places) != want {
				t.Errorf("the notes = %q, want %q", notesOf(places), want)
			}
		})
	}
}

func TestRead_stillReportsAnUnknownKeyAsSuchRatherThanNotTOML(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	content := "vault = \"/tmp/quelquepart\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}

	_, err := config.Read(home, silentEnv)
	if err == nil {
		t.Fatalf("Read returned no error, want one naming %q", "vault")
	}
	if !errors.Is(err, config.ErrUnknownKey) {
		t.Errorf("err = %v, want it to be config.ErrUnknownKey, not config.ErrNotTOML", err)
	}
	if errors.Is(err, config.ErrNotTOML) {
		t.Errorf("err = %v, want it not to be config.ErrNotTOML", err)
	}
}

type readingRecorder struct {
	resolvedCalled bool
	places         config.Places
	path           string
	readFrom       string
	err            error
}

func (r *readingRecorder) ReadFrom(path string) {
	r.readFrom = path
}

func (r *readingRecorder) Resolved(p config.Places) {
	r.resolvedCalled = true
	r.places = p
}

func (r *readingRecorder) NotUnderstood(path string, err error) {
	r.path = path
	r.err = err
}

func TestReading_reportsWhatItCannotMakeSenseOf(t *testing.T) {
	writeConfig := func(t *testing.T, home, content string) string {
		t.Helper()
		dir := filepath.Join(home, ".config", "cairn")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
		path := filepath.Join(dir, "config.toml")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
		return path
	}

	cases := []struct {
		name    string
		content string
		wantErr error
		wantSub string
	}{
		{
			name:    "not TOML at all",
			content: "not { toml : at all\n",
			wantErr: config.ErrNotTOML,
		},
		{
			name:    "a name it does not know",
			content: "vault = \"/tmp/quelquepart\"\n",
			wantErr: config.ErrUnknownKey,
			wantSub: "vault",
		},
		{
			name:    "a place that is not absolute",
			content: "store = \"vault\"\n",
			wantErr: config.ErrNotAbsolute,
			wantSub: "vault",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			path := writeConfig(t, home, c.content)

			reading, err := config.NewReading(home, silentEnv)
			if err != nil {
				t.Fatalf("NewReading returned an error: %v", err)
			}

			got := &readingRecorder{}
			reading.Tell(got)

			if got.resolvedCalled {
				t.Errorf("Tell said Resolved(%v), want it not to tell emplacements alongside a problem", got.places)
			}
			if got.err == nil {
				t.Fatalf("Tell reported no problem, want one wrapping %v", c.wantErr)
			}
			if !errors.Is(got.err, c.wantErr) {
				t.Errorf("err = %v, want it to be %v", got.err, c.wantErr)
			}
			if got.path != path {
				t.Errorf("path = %q, want %q", got.path, path)
			}
			if c.wantSub != "" && !strings.Contains(got.err.Error(), c.wantSub) {
				t.Errorf("err = %q, want it to mention %q", got.err.Error(), c.wantSub)
			}
		})
	}
}

func TestReading_saysWhereItReadFromEvenWhenItMadeSenseOfIt(t *testing.T) {
	home := t.TempDir()

	reading, err := config.NewReading(home, silentEnv)
	if err != nil {
		t.Fatalf("NewReading returned an error: %v", err)
	}

	got := &readingRecorder{}
	reading.Tell(got)

	if !got.resolvedCalled {
		t.Fatal("Tell did not call Resolved, want it to since no file at all still resolves to defaults")
	}
	want := filepath.Join(home, ".config", "cairn", "config.toml")
	if got.readFrom != want {
		t.Errorf("readFrom = %q, want %q", got.readFrom, want)
	}
}

type told struct {
	said []string
}

func (r *told) StoreNamedByTheFile(s string) {
	r.said = append(r.said, fmt.Sprintf("StoreNamedByTheFile(%q)", s))
}
func (r *told) StoreNamedByTheEnvironment(s string) {
	r.said = append(r.said, fmt.Sprintf("StoreNamedByTheEnvironment(%q)", s))
}
func (r *told) StoreNamedByNothing(s string) {
	r.said = append(r.said, fmt.Sprintf("StoreNamedByNothing(%q)", s))
}
func (r *told) NotesNamedByTheFile(s string) {
	r.said = append(r.said, fmt.Sprintf("NotesNamedByTheFile(%q)", s))
}
func (r *told) NotesNamedByTheEnvironment(s string) {
	r.said = append(r.said, fmt.Sprintf("NotesNamedByTheEnvironment(%q)", s))
}
func (r *told) NotesNamedByNothing(s string) {
	r.said = append(r.said, fmt.Sprintf("NotesNamedByNothing(%q)", s))
}

func TestPlaces_tellsWhatNamedEachPlace(t *testing.T) {
	writeConfig := func(t *testing.T, home, content string) {
		t.Helper()
		dir := filepath.Join(home, ".config", "cairn")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
			t.Fatalf("could not set up the fixture: %v", err)
		}
	}

	cases := []struct {
		name  string
		setup func(home string)
		env   func(string) (string, bool)
		index int
		want  func(home string) string
	}{
		{
			name:  "store named by the file",
			setup: func(home string) { writeConfig(t, home, "store = \"/tmp/somewhere/cairn.db\"\n") },
			env:   silentEnv,
			index: 0,
			want:  func(string) string { return `StoreNamedByTheFile("/tmp/somewhere/cairn.db")` },
		},
		{
			name:  "store named by the environment",
			setup: func(string) {},
			env: func(key string) (string, bool) {
				if key == config.StoreVariable {
					return "/somewhere/of/my/own.db", true
				}
				return "", false
			},
			index: 0,
			want:  func(string) string { return `StoreNamedByTheEnvironment("/somewhere/of/my/own.db")` },
		},
		{
			name:  "store named by nothing",
			setup: func(string) {},
			env:   silentEnv,
			index: 0,
			want: func(home string) string {
				return fmt.Sprintf("StoreNamedByNothing(%q)", filepath.Join(home, ".cairn", "cairn.db"))
			},
		},
		{
			name:  "notes named by the file",
			setup: func(home string) { writeConfig(t, home, "notes = \"/tmp/somewhere/notes\"\n") },
			env:   silentEnv,
			index: 1,
			want:  func(string) string { return `NotesNamedByTheFile("/tmp/somewhere/notes")` },
		},
		{
			name:  "notes named by the environment",
			setup: func(string) {},
			env: func(key string) (string, bool) {
				if key == config.NotesVariable {
					return "/somewhere/of/my/own/notes", true
				}
				return "", false
			},
			index: 1,
			want:  func(string) string { return `NotesNamedByTheEnvironment("/somewhere/of/my/own/notes")` },
		},
		{
			name:  "notes named by nothing",
			setup: func(string) {},
			env:   silentEnv,
			index: 1,
			want: func(home string) string {
				return fmt.Sprintf("NotesNamedByNothing(%q)", filepath.Join(home, ".cairn", "notes"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home := t.TempDir()
			c.setup(home)

			places, err := config.Read(home, c.env)
			if err != nil {
				t.Fatalf("Read returned an error: %v", err)
			}

			got := &told{}
			places.Tell(got)

			if len(got.said) != 2 {
				t.Fatalf("Tell said %v, want exactly two facts", got.said)
			}
			if want := c.want(home); got.said[c.index] != want {
				t.Errorf("Tell said %v, want %q at index %d", got.said, want, c.index)
			}
		})
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

func TestRead_saysWhichPagesAreTurnedOff(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[pages]\ntime = false\ntasks = true\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := read.PagesOff(); len(got) != 1 || got[0] != "time" {
		t.Errorf("the pages off are %v, want [time]", got)
	}
	none, err := config.Read(t.TempDir(), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := none.PagesOff(); len(got) != 0 {
		t.Errorf("with no file the pages off are %v", got)
	}
	all, err := config.Read(writtenConfig(t, "[pages]\nobjectives = false\ntime = false\ntasks = false\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(all.PagesOff(), ","); got != "tasks,objectives,time" {
		t.Errorf("every page off reads %q, want them in the order they are drawn", got)
	}
	if _, err := config.Read(writtenConfig(t, "[pages]\ncalendar = false\n"), silentEnv); !errors.Is(err, config.ErrUnknownKey) {
		t.Errorf("a page cairn does not have: %v, want ErrUnknownKey", err)
	}
}

func TestRead_takesThePluginFolderTheFileNames(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("plugin = \"~/claude/cairn\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got.Plugin() != filepath.Join(home, "claude", "cairn") {
		t.Errorf("Plugin = %q", got.Plugin())
	}
	if unnamed, _ := config.Read(t.TempDir(), silentEnv); unnamed.Plugin() != "" {
		t.Errorf("with nothing said, Plugin = %q, want nothing", unnamed.Plugin())
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("plugin = \"claude/cairn\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Read(home, silentEnv); !errors.Is(err, config.ErrNotAbsolute) {
		t.Errorf("a plugin folder that is not absolute = %v, want ErrNotAbsolute", err)
	}
}

func TestRead_saysFromWhenAndHowOftenDoneTasksAreArchived(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[archive]\nstart = 2026-10-21\ndays = 10\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	start, days, ok := read.Archive()
	if !ok || start != "2026-10-21" || days != 10 {
		t.Errorf("archiving counts from %q in periods of %d days (%v)", start, days, ok)
	}
	unsaid, err := config.Read(writtenConfig(t, "[archive]\nstart = 2026-10-21\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if _, days, _ := unsaid.Archive(); days != 14 {
		t.Errorf("a period whose length is not said lasts %d days, want 14", days)
	}
	none, _ := config.Read(t.TempDir(), silentEnv)
	if _, _, ok := none.Archive(); ok {
		t.Error("with no file, tasks are archived on their own")
	}
	for _, refused := range []struct{ written, names string }{
		{"[archive]\nstart = 2026-10-21\ndays = 0\n", "archive.days"},
		{"[archive]\nstart = \"soon\"\n", "archive.start"},
		{"[archive]\ndays = 14\n", "archive.start"},
	} {
		_, err := config.Read(writtenConfig(t, refused.written), silentEnv)
		if !errors.Is(err, config.ErrArchive) || !strings.Contains(err.Error(), refused.names) {
			t.Errorf("%q gives %v, want ErrArchive naming %s", refused.written, err, refused.names)
		}
	}
}

func TestRead_saysHowLongAFocusSessionLasts(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[focus]\nsession = \"1h\"\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := read.FocusSession(); got != "1h" {
		t.Errorf("a focus session lasts %q, want 1h", got)
	}
	none, _ := config.Read(t.TempDir(), silentEnv)
	if got := none.FocusSession(); got != "" {
		t.Errorf("with no file, a focus session lasts %q", got)
	}
}

func TestRead_saysWhatADeletedTaskDoesWithItsTime(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[time]\ndeleted = \"taken\"\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if got := read.TimeDeleted(); got != config.TimeTaken {
		t.Errorf("a deleted task's time is %q, want taken", got)
	}
	none, _ := config.Read(t.TempDir(), silentEnv)
	if got := none.TimeDeleted(); got != config.TimeKept {
		t.Errorf("with no file, a deleted task's time is %q, want kept", got)
	}
	if _, err := config.Read(writtenConfig(t, "[time]\ndeleted = \"gone\"\n"), silentEnv); !errors.Is(err, config.ErrTime) {
		t.Errorf("err = %v, want ErrTime", err)
	}
}

func TestRead_saysWhetherTheWindowKeepsItsSizes(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[window]\nkeep_sizes = true\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	if !read.KeepSizes() {
		t.Error("keep_sizes = true does not keep the window's sizes")
	}
	none, _ := config.Read(t.TempDir(), silentEnv)
	if none.KeepSizes() {
		t.Error("with no file, the window keeps its sizes")
	}
	_, err = config.Read(writtenConfig(t, "[window]\nkeep_sizes = \"always\"\n"), silentEnv)
	if !errors.Is(err, config.ErrKeepSizes) || !strings.Contains(err.Error(), "window.keep_sizes") {
		t.Errorf("err = %v, want ErrKeepSizes naming window.keep_sizes", err)
	}
}
