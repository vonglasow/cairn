package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead_theWorkspacesAreReadEachUnderItsName(t *testing.T) {
	read, err := readThis(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/cairn/work\"\nhome = \"/elsewhere/home\"\n")
	if err != nil {
		t.Fatal(err)
	}
	folders := map[string]string{}
	for _, one := range read.Workspaces() {
		folders[one.Name] = one.Folder
	}
	if len(folders) != 2 || !strings.HasSuffix(folders["work"], filepath.Join("cairn", "work")) || !filepath.IsAbs(folders["work"]) || folders["home"] != "/elsewhere/home" {
		t.Errorf("Workspaces() = %v, want work under the home and home where it is said, default not among them", read.Workspaces())
	}
	if names := []string{read.Workspaces()[0].Name, read.Workspaces()[1].Name}; names[0] != "home" {
		t.Errorf("Workspaces() named %v, want them by name", names)
	}
	if read.DefaultWorkspace() != "work" {
		t.Errorf("DefaultWorkspace() = %q, want work", read.DefaultWorkspace())
	}
}

func TestRead_noWorkspacesIsNone(t *testing.T) {
	read, err := readThis(t, "notes = \"~/vault\"\n")
	if err != nil || len(read.Workspaces()) != 0 || read.DefaultWorkspace() != "" {
		t.Errorf("Read = %v, %v, %q; want no workspace and no default", err, read.Workspaces(), read.DefaultWorkspace())
	}
}

func TestRead_aWorkspaceNotAbsoluteIsRefusedNamingIt(t *testing.T) {
	_, err := readThis(t, "[workspaces]\nwork = \"cairn/work\"\n")
	if !errors.Is(err, ErrNotAbsolute) || !strings.Contains(err.Error(), "workspaces.work") || !strings.Contains(err.Error(), "cairn/work") {
		t.Errorf("Read = %v, want it refused as not absolute, naming workspaces.work and what it says", err)
	}
}

func TestRead_aKeyNamedLikeTheWorkspacesUnderAnotherTableSaysItIsATable(t *testing.T) {
	read, err := readThis(t, "[pages]\nworkspaces = true\n")
	said := strings.Join(read.Ignored(), "\n")
	if err != nil || !strings.Contains(said, "workspaces is a table of its own, [workspaces]") {
		t.Errorf("Read = %v, ignoring %q; want it saying workspaces is a table", err, said)
	}
}

func TestAbsolute_expandsTheHomeAndRefusesWhatIsRelative(t *testing.T) {
	if at, err := Absolute("~/x", "/home/me"); err != nil || at != "/home/me/x" {
		t.Errorf("Absolute(~/x) = %q, %v", at, err)
	}
	if at, err := Absolute("/x", "/home/me"); err != nil || at != "/x" {
		t.Errorf("Absolute(/x) = %q, %v", at, err)
	}
	if _, err := Absolute("x", "/home/me"); !errors.Is(err, ErrNotAbsolute) || !strings.Contains(err.Error(), `"x"`) {
		t.Errorf("Absolute(x) = %v, want it refused naming what was said", err)
	}
}

func TestWithWorkspace_writesTheTableAloneWhenThereIsNoFile(t *testing.T) {
	got, err := WithWorkspace(nil, "play", "~/cairn/play")
	if err != nil || string(got) != "[workspaces]\nplay = \"~/cairn/play\"\n" {
		t.Errorf("WithWorkspace(nothing) = %q, %v", got, err)
	}
}

func TestWithWorkspace_addsTheTableAfterWhatIsThere(t *testing.T) {
	for _, before := range []string{"notes = \"~/vault\"\n", "notes = \"~/vault\""} {
		got, err := WithWorkspace([]byte(before), "play", "/p")
		if want := "notes = \"~/vault\"\n\n[workspaces]\nplay = \"/p\"\n"; err != nil || string(got) != want {
			t.Errorf("WithWorkspace(%q) = %q, %v; want %q", before, got, err, want)
		}
	}
}

func TestWithWorkspace_joinsTheTableAtItsEnd(t *testing.T) {
	before := "[workspaces]\ndefault = \"work\"\nwork = \"/w\"\n\n# what the window does\n[window]\nkeep_sizes = true\n"
	got, err := WithWorkspace([]byte(before), "play", "/p")
	want := "[workspaces]\ndefault = \"work\"\nwork = \"/w\"\nplay = \"/p\"\n\n# what the window does\n[window]\nkeep_sizes = true\n"
	if err != nil || string(got) != want {
		t.Errorf("WithWorkspace = %q, %v; want %q", got, err, want)
	}
}

func TestWithWorkspace_joinsATableLastInTheFileAfterItsComments(t *testing.T) {
	before := "store = \"/s\"\n\n[workspaces]\n# work = \"~/cairn/work\"\n\n"
	got, err := WithWorkspace([]byte(before), "play", "/p")
	want := "store = \"/s\"\n\n[workspaces]\n# work = \"~/cairn/work\"\nplay = \"/p\"\n\n"
	if err != nil || string(got) != want {
		t.Errorf("WithWorkspace = %q, %v; want %q", got, err, want)
	}
}

func TestWithWorkspace_joinsAnEmptyTableUnderItsHeading(t *testing.T) {
	before := "[workspaces]\n\n# the window\n[window]\n"
	got, err := WithWorkspace([]byte(before), "play", "/p")
	want := "[workspaces]\nplay = \"/p\"\n\n# the window\n[window]\n"
	if err != nil || string(got) != want {
		t.Errorf("WithWorkspace = %q, %v; want %q", got, err, want)
	}
}

func TestWithWorkspace_refusesAFileItCannotAddToWithoutBreakingIt(t *testing.T) {
	for _, before := range []string{"workspaces = { work = \"/w\" }\n", "this is not toml: [[[\n"} {
		if got, err := WithWorkspace([]byte(before), "play", "/p"); !errors.Is(err, ErrNotWritten) {
			t.Errorf("WithWorkspace(%q) = %q, %v; want it refused, the file left to a person", before, got, err)
		}
	}
}

func TestWithDefault_changesTheDefaultWhereItIsSaid(t *testing.T) {
	before := "[workspaces]\nwork = \"/w\"\ndefault = \"work\"\nhome = \"/h\"\n"
	got, err := WithDefault([]byte(before), "home")
	if want := "[workspaces]\nwork = \"/w\"\ndefault = \"home\"\nhome = \"/h\"\n"; err != nil || string(got) != want {
		t.Errorf("WithDefault = %q, %v; want %q", got, err, want)
	}
}

func TestWithDefault_saysTheDefaultUnderTheHeadingWhenNoneIs(t *testing.T) {
	before := "notes = \"/n\"\n[workspaces]\nwork = \"/w\"\n"
	got, err := WithDefault([]byte(before), "work")
	if want := "notes = \"/n\"\n[workspaces]\ndefault = \"work\"\nwork = \"/w\"\n"; err != nil || string(got) != want {
		t.Errorf("WithDefault = %q, %v; want %q", got, err, want)
	}
}

func TestWithDefault_leavesADefaultUnderAnotherTableAlone(t *testing.T) {
	before := "[workspaces]\nwork = \"/w\"\n\n[focus]\ndefault = \"x\"\n"
	got, err := WithDefault([]byte(before), "work")
	if want := "[workspaces]\ndefault = \"work\"\nwork = \"/w\"\n\n[focus]\ndefault = \"x\"\n"; err != nil || string(got) != want {
		t.Errorf("WithDefault = %q, %v; want %q", got, err, want)
	}
}

func TestWithDefault_refusesAFileWithNoTableToSayItIn(t *testing.T) {
	if got, err := WithDefault([]byte("workspaces = { work = \"/w\" }\n"), "work"); !errors.Is(err, ErrNotWritten) {
		t.Errorf("WithDefault = %q, %v; want it refused", got, err)
	}
}

func TestRead_theStoreTheFileNamesIsKeptWhateverAnswers(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Dir(FileIn(home))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(FileIn(home), []byte("store = \"~/s/cairn.db\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := Read(home, func(string) (string, bool) { return "/elsewhere/cairn.db", true })
	if err != nil || read.FileStore() != filepath.Join(home, "s", "cairn.db") {
		t.Errorf("FileStore() = %q, %v; want the file's store under the environment's", read.FileStore(), err)
	}
	if read, _ := readThis(t, "notes = \"~/vault\"\n"); read.FileStore() != "" {
		t.Errorf("FileStore() = %q for a file naming no store", read.FileStore())
	}
}

func TestTheExampleNamesOneWorkspaceTheDefaultOnTheFolderCairnHasAlwaysUsed(t *testing.T) {
	read, err := readThis(t, Example)
	if err != nil {
		t.Fatal(err)
	}
	if got := read.Workspaces(); read.DefaultWorkspace() != "main" || len(got) != 1 || got[0].Name != "main" || filepath.Base(got[0].Folder) != DefaultFolder {
		t.Errorf("the example names %v, the default %q; want main alone on ~/%s", got, read.DefaultWorkspace(), DefaultFolder)
	}
	if read.FileStore() != "" {
		t.Errorf("the example names the store %q, which the default workspace would pass over", read.FileStore())
	}
}
