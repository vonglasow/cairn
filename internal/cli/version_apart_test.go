package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestARepositoryAScenarioMakesWritesNothingIntoTheOneCommitting(t *testing.T) {
	committing := filepath.Join(t.TempDir(), "index.lock")
	t.Setenv("GIT_INDEX_FILE", committing)
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		if err := command(repo, "git", args...); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(committing); err == nil {
		t.Fatal("git, run for a scenario, wrote into the index of the commit the hooks run for")
	}
}
