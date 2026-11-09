package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func readThis(t *testing.T, content string) (Places, error) {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return Read(home, func(string) (string, bool) { return "", false })
}

func TestTheExample_everyKeyItCommentsIsReadOnceItsCommentIsTakenOff(t *testing.T) {
	commented := regexp.MustCompile(`(?m)^# ([a-z_]+) = .*$`)
	found := commented.FindAllStringSubmatchIndex(Example, -1)
	if len(found) == 0 {
		t.Fatal("the example comments no key")
	}
	for _, at := range found {
		line := Example[at[0]:at[1]]
		uncommented := Example[:at[0]] + strings.TrimPrefix(line, "# ") + Example[at[1]:]
		if _, err := readThis(t, uncommented); err != nil {
			t.Errorf("with %q taken out of its comment, the file is refused: %v", line, err)
		}
	}
}

func TestTheExample_writesEveryTableItsHeading(t *testing.T) {
	for _, table := range []string{"[pages]", "[archive]", "[focus]", "[approach]", "[time]", "[window]", "[workspaces]"} {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(table) + `$`).MatchString(Example) {
			t.Errorf("the example does not write %s out of a comment", table)
		}
	}
}

func TestRead_aKeyUnderAnotherTableIsPassedOverSayingTheTableItBelongsUnder(t *testing.T) {
	for _, c := range []struct{ content, names, belongs string }{
		{"[pages]\nkeep_sizes = true\n", "pages.keep_sizes", "[window]"},
		{"[window]\nsession = \"1h\"\n", "window.session", "[focus]"},
		{"[time]\nstore = \"/x/cairn.db\"\n", "time.store", "before any table"},
	} {
		read, err := readThis(t, c.content)
		said := strings.Join(read.Ignored(), "\n")
		if err != nil || !strings.Contains(said, c.names) || !strings.Contains(said, c.belongs) {
			t.Errorf("%q: Read = %v, ignoring %q; want it naming %s and saying %s", c.content, err, said, c.names, c.belongs)
		}
	}
}

func TestRead_aKeyNoTableKnowsSaysNoTable(t *testing.T) {
	read, err := readThis(t, "[pages]\nvault = true\n")
	said := strings.Join(read.Ignored(), "\n")
	if err != nil || !strings.Contains(said, "pages.vault") || strings.Contains(said, "belongs") {
		t.Errorf("Read = %v, ignoring %q; want it passed over, naming no table for a key no table knows", err, said)
	}
}

func TestRead_archiveDaysWithoutAStartIsNoArchiving(t *testing.T) {
	read, err := readThis(t, "[archive]\ndays = 14\n")
	if err != nil {
		t.Fatalf("Read = %v, want the days alone accepted", err)
	}
	if start, days, on := read.Archive(); on || start != "" || days != 14 {
		t.Errorf("Archive() = %q, %d, %v, want nothing archived, in periods of 14 days", start, days, on)
	}
}

func TestRead_archiveDaysOfOneIsTheShortestPeriod(t *testing.T) {
	read, err := readThis(t, "[archive]\ndays = 1\n")
	if err != nil {
		t.Fatalf("Read = %v, want a period of a single day accepted", err)
	}
	if _, days, _ := read.Archive(); days != 1 {
		t.Errorf("Archive() days = %d, want 1", days)
	}
}

func TestRead_archiveDaysAreSaidEvenWhenNothingIsArchived(t *testing.T) {
	read, err := readThis(t, "[archive]\ndays = 7\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, days, _ := read.Archive(); days != 7 {
		t.Errorf("Archive() days = %d, want the 7 the file says", days)
	}
}

func TestRead_noArchiveTableIsPeriodsOfTheDefaultLength(t *testing.T) {
	read, err := readThis(t, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, days, on := read.Archive(); on || days != defaultArchiveDays {
		t.Errorf("Archive() = %d, %v, want nothing archived in periods of %d days", days, on, defaultArchiveDays)
	}
}

func TestRead_aKeepSizesOfFalseIsRead(t *testing.T) {
	read, err := readThis(t, "[window]\nkeep_sizes = false\n")
	if err != nil || read.KeepSizes() {
		t.Errorf("Read = %v, KeepSizes = %v, want false read without a word", err, read.KeepSizes())
	}
}

func TestRead_pagesAreSaidOneByOne(t *testing.T) {
	read, err := readThis(t, "[pages]\nobjectives = false\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := read.Pages(); got != [3]bool{true, false, true} {
		t.Errorf("Pages() = %v, want tasks and time on, objectives off", got)
	}
}

func TestTheExample_writesWeekendsShownUnderTime(t *testing.T) {
	_, under, _ := strings.Cut(Example, "\n[time]\n")
	under, _, _ = strings.Cut(under, "\n[window]\n")
	if !regexp.MustCompile(`(?m)^weekends = "shown"$`).MatchString(under) {
		t.Errorf("the example does not write weekends = \"shown\" under [time]:\n%s", Example)
	}
}
