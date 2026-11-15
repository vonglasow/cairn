package build

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
)

func TestApplyVCS_fillsAnUnstampedCommit(t *testing.T) {
	got := applyVCS(Info{}, []debug.BuildSetting{
		{Key: "vcs.revision", Value: "4c5a15f700dcabcdef"},
		{Key: "vcs.time", Value: "2026-08-15T13:59:37Z"},
	})
	if got.Commit != "4c5a15f700dcabcdef" {
		t.Errorf("Commit = %q, want the embedded revision", got.Commit)
	}
	if got.CommitDate != "2026-08-15T13:59:37Z" {
		t.Errorf("CommitDate = %q, want the embedded time", got.CommitDate)
	}
}

func TestRunning_isAPathThatIsThere(t *testing.T) {
	running, err := Running()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(running); err != nil || !filepath.IsAbs(running) {
		t.Errorf("Running = %s, %v", running, err)
	}
}

func TestFollowed_keepsAPathThatLeadsNowhere(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	if got := followed(missing); got != missing {
		t.Errorf("followed(%q) = %q, want the path as given", missing, got)
	}
}
