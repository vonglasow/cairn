package task

import (
	"cairn/internal/engine/item"
	"strings"
	"testing"
)

func TestText_saysWhatBlocksTheTaskAndWhy(t *testing.T) {
	made := func(because string) string {
		var a AsTold
		a.Numbered(1)
		a.Titled("Migration Jenkins")
		a.BlockedBy(4)
		if because != "" {
			a.Because(because)
		}
		got := Text(a.Task(), item.Progress{}, true)
		for _, line := range strings.Split(got, "\n") {
			if strings.Contains(line, "blocked by") {
				return strings.TrimSpace(line)
			}
		}
		t.Fatalf("Text has no blocked by line:\n%s", got)
		return ""
	}
	if got := made("the client's answer"); !strings.HasSuffix(got, "#4, the client's answer") {
		t.Errorf("blocked by line = %q, want the blocker then the reason", got)
	}
	if got := made(""); !strings.HasSuffix(got, "#4") {
		t.Errorf("blocked by line = %q, want the blocker alone, no empty reason after it", got)
	}
}
