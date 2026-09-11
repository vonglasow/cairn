package config_test

// Whether the cell built by Read carries exactly what the file wrote under
// `store` — no expansion, no default. Those are later slices' concern.
//
// TestRead_expandsALeadingTildeAgainstTheHome is the exception: a leading
// "~/" is the one thing this slice teaches Read to change, against the home
// it was itself handed.
//
// TestRead_takesWhatTheEnvironmentSaysOverAnythingElse is the other: the
// environment does not take a turn after the file and before the default —
// it replaces whichever of the two already answered.

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

// silentEnv says nothing, for the tests that are not about the environment
// at all: Read has to be given some lookupEnv, and this is the one that
// leaves the file and the default to answer on their own.
func silentEnv(string) (string, bool) { return "", false }

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

	if got := places.Store(); got != "/tmp/quelquepart/cairn.db" {
		t.Errorf("Store() = %q, want %q", got, "/tmp/quelquepart/cairn.db")
	}
}

// TestRead_fallsBackToTheHomeWhenNothingNamesTheStore covers the one rule
// with its two entries: a file that says nothing about `store`, and no file
// at all. Table-driven so the single test the fiche names carries both cases,
// rather than asserting the rule twice under two names.
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
			if got := places.Store(); got != want {
				t.Errorf("Store() = %q, want %q", got, want)
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
	if got := places.Store(); got != want {
		t.Errorf("Store() = %q, want %q", got, want)
	}
}

// TestRead_takesWhatTheEnvironmentSaysOverAnythingElse is the one rule with
// its four entries: the environment replaces what the file named, replaces
// the default when nothing named anything, undergoes the same "~" expansion
// as the file's own answer, and — set but empty — says nothing at all, so
// the answer stays whatever the file or the default already gave.
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

			if want := c.want(home); places.Store() != want {
				t.Errorf("Store() = %q, want %q", places.Store(), want)
			}
		})
	}
}

// TestRead_refusesANameItDoesNotKnow: a key cairn does not know about is not
// silently ignored — it is refused, and the error names the key that was
// unknown as well as what cairn accepts instead.
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

// TestRead_refusesAFileItCannotMakeSenseOf: a file that is not TOML at all
// is refused, and the error names the file it could not make sense of — not
// just the reason the parser gave up.
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

// TestRead_refusesAPlaceThatIsNotAbsolute is one rule checked under both
// keys: a place the file names — store or notes — has to be absolute once a
// leading "~" is expanded, and the error names the place as the file wrote
// it, before that expansion.
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
				if !strings.Contains(err.Error(), place) {
					t.Errorf("err = %q, want it to name %q", err.Error(), place)
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

// TestRead_resolvesTheNotesTheSameWayAsTheStore is the one rule proven a
// second time under a second name: an emplacement resolves the same way
// whichever one it is. Same three entries as the store's own resolution —
// the file names it, nothing names it, the environment names it and
// replaces both — so a Notes() that answered with the store's value instead
// of its own would fail here exactly as it would for Store().
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

			if want := c.want(home); places.Notes() != want {
				t.Errorf("Notes() = %q, want %q", places.Notes(), want)
			}
		})
	}
}

// TestRead_stillReportsAnUnknownKeyAsSuchRatherThanNotTOML locks the order
// slice 3 relies on by accident: *toml.StrictMissingError unwraps to a
// *toml.DecodeError, so a check for the latter alone would also catch the
// former. Checked in the wrong order, an unknown key would be reported as
// "not TOML" instead of ErrUnknownKey, and SC4 and SC5 would become
// indistinguishable.
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

// readingRecorder is a fake config.ReadingTold: it records whichever single
// fact a Reading told it — its emplacements, or the problem it hit — rather
// than doing anything with it. A test on it is a test on what Reading.Tell
// says, nothing else.
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

// TestReading_reportsWhatItCannotMakeSenseOf is the fiche's named test: the
// tolerant primitive NewReading is a gate over reports, rather than returns,
// the three refusals Read has always carried — naming the file's path
// alongside the problem — and never tells emplacements alongside a problem:
// a Reading that hit one has nothing trustworthy to say about places.
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

// TestReading_saysWhereItReadFromEvenWhenItMadeSenseOfIt is the fiche's named
// test: the ordinary case — no file at all, so the reading falls back to
// defaults and makes sense of it without trouble — still has to say which
// file it read from. "Which file" is half of what a reading answers, whether
// it made sense of the file or not; a Resolved that stayed silent about the
// path would leave anyone reading the emplacements to guess where they came
// from.
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

// told is a fake config.Told: it records, in order, the facts it was told
// rather than doing anything with them. A test on it is a test on what Tell
// says, nothing else.
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

// TestPlaces_tellsWhatNamedEachPlace checks that Places remembers, for each
// of its two emplacements, which of the three provenances named it — the
// file, the environment, or nothing at all — and tells exactly that one
// fact through Tell, carrying the emplacement it resolved to. Six sub-cases,
// one per fact: three provenances times two emplacements.
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
		index int // where the fact under test lands in what Tell says: 0 for the store, 1 for the notes
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
