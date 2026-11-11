package engine

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cairn/internal/config"
)

//go:embed fixtures/fixtures.json
var fixtures []byte

const fixturesMark = ".made-by-cairn-fixtures"

// ErrNotMadeByFixtures is returned when the folder named for the fixtures holds something cairn fixtures did not put there: it is never emptied.
var ErrNotMadeByFixtures = errors.New("the folder was not made by cairn fixtures, and is left alone")

// Fixtures says where the fixtures are, and whether they were made just now rather than opened as they were left.
type Fixtures struct {
	Store string
	Notes string
	Made  bool
}

// FixturesFile is the fixtures the binary carries, as a file task import reads.
func FixturesFile() []byte { return fixtures }

// FixturesFolder is where the fixtures are made when no folder is named: one of their own in the temporary folder, never the store in use.
func FixturesFolder() string { return filepath.Join(os.TempDir(), "cairn-fixtures") }

// MakeFixtures fills a store in the folder with the fixtures the binary carries, or with the tasks of the file named; a folder it made before is opened as it was left, unless it is to be made again.
func MakeFixtures(folder, from string, fresh bool) (Fixtures, error) {
	at := Fixtures{Store: filepath.Join(folder, "cairn.db"), Notes: filepath.Join(folder, "notes")}
	entries, err := os.ReadDir(folder)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Fixtures{}, err
	}
	_, marked := os.Stat(filepath.Join(folder, fixturesMark))
	if len(entries) > 0 && marked != nil {
		return Fixtures{}, fmt.Errorf("%s: %w", folder, ErrNotMadeByFixtures)
	}
	if len(entries) > 0 && !fresh {
		return at, nil
	}
	if err := os.RemoveAll(folder); err != nil {
		return Fixtures{}, err
	}
	if err := os.MkdirAll(folder, 0o700); err != nil {
		return Fixtures{}, err
	}
	if err := os.WriteFile(filepath.Join(folder, fixturesMark), nil, 0o600); err != nil {
		return Fixtures{}, err
	}
	if err := os.WriteFile(config.SettingsBeside(at.Store), []byte("[recurrence]\non = true\n\n[workspace]\ncolour = \"peach\"\nname = \"fixtures\"\n"), 0o600); err != nil {
		return Fixtures{}, err
	}
	if from == "" {
		from = filepath.Join(folder, "fixtures.json")
		if err := os.WriteFile(from, fixtures, 0o600); err != nil {
			return Fixtures{}, err
		}
	}
	if _, err := DoAt(TaskImport, Places{Store: at.Store, Notes: at.Notes}, Input{Given: map[string]string{"from": from}}, Acting{Through: "fixtures"}); err != nil {
		return Fixtures{}, err
	}
	at.Made = true
	return at, nil
}
