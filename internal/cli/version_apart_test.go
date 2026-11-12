package cli

import (
	"os"
	"os/exec"
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

func TestAProgramAScenarioRunsReadsNoGitConfigurationOfTheMachine(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = the machine's own\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	said, _ := apart(exec.Command("git", "config", "--get", "user.name")).Output()
	if len(said) != 0 {
		t.Errorf("git, run for a scenario, read the machine's configuration: %s", said)
	}
}
