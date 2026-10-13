package build

import (
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func stampFor(t *testing.T, target *string, value string) {
	t.Helper()
	prev := *target
	t.Cleanup(func() { *target = prev })
	*target = value
}

func TestRead_truncatesALongCommit(t *testing.T) {
	stampFor(t, &commit, "4c5a15f700dcabcdef0123456789")
	if got := Read().Commit; got != "4c5a15f700dc" {
		t.Errorf("Commit = %q, want it truncated to %d characters", got, shortCommitLen)
	}
}

func TestRead_keepsAShortCommitWhole(t *testing.T) {
	stampFor(t, &commit, "4c5a15f")
	if got := Read().Commit; got != "4c5a15f" {
		t.Errorf("Commit = %q, want it left alone when already short", got)
	}
}

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

func TestApplyVCS_leavesAStampedCommitAlone(t *testing.T) {
	got := applyVCS(Info{Commit: "from-the-linker"}, []debug.BuildSetting{
		{Key: "vcs.revision", Value: "from-the-working-tree"},
	})
	if got.Commit != "from-the-linker" {
		t.Errorf("Commit = %q, want the linker stamp to win", got.Commit)
	}
}

func TestApplyVCS_leavesAStampedCommitDateAlone(t *testing.T) {
	got := applyVCS(Info{CommitDate: "2026-10-07T21:00:00Z"}, []debug.BuildSetting{
		{Key: "vcs.time", Value: "2026-08-15T13:59:37Z"},
	})
	if got.CommitDate != "2026-10-07T21:00:00Z" {
		t.Errorf("CommitDate = %q, want the linker stamp to win", got.CommitDate)
	}
}

func TestRead_saysTheStampedCommitDate(t *testing.T) {
	stampFor(t, &committed, "2026-10-07T21:00:00Z")
	if got := Read().CommitDate; got != "2026-10-07T21:00:00Z" {
		t.Errorf("CommitDate = %q, want the linker stamp", got)
	}
}

func TestApplyVCS_ignoresUnrelatedSettings(t *testing.T) {
	got := applyVCS(Info{}, []debug.BuildSetting{{Key: "GOARCH", Value: "arm64"}})
	if got.Commit != "" || got.CommitDate != "" {
		t.Errorf("unrelated settings must not fill anything, got %+v", got)
	}
}

func TestRead_readsTheRuntime(t *testing.T) {
	b := Read()
	if b.Go != runtime.Version() {
		t.Errorf("Go = %q, want %q", b.Go, runtime.Version())
	}
	if want := runtime.GOOS + "/" + runtime.GOARCH; b.Platform != want {
		t.Errorf("Platform = %q, want %q", b.Platform, want)
	}
}

func TestString_printsEveryFieldABuildCarries(t *testing.T) {
	told := Info{
		Version:    "v1.2.3",
		Commit:     "4c5a15f700dc",
		Built:      "2026-08-25T09:00:00Z",
		CommitDate: "2026-08-24T20:50:43Z",
		Go:         "go1.26.6",
		Platform:   "darwin/arm64",
	}

	printed := told.String()
	fields := reflect.ValueOf(told)
	for i := range fields.NumField() {
		name := fields.Type().Field(i).Name
		value := fields.Field(i).String()
		if value == "" {
			t.Fatalf("%s was left empty: the walk proves nothing about a field it cannot tell apart", name)
		}
		if !strings.Contains(printed, value) {
			t.Errorf("%s is not in what a build prints:\n%s", name, printed)
		}
	}
}
