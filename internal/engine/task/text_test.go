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
		got := Text(a.Task(), item.Progress{}, true, Choosing{})
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

func TestText_namesEachObjectiveServedUnderItsParent(t *testing.T) {
	var a AsTold
	a.Numbered(1)
	a.Titled("Pin map")
	a.Serves(2)
	served := Choosing{Serves: []string{"Home automation › Enclosure"}}
	if got := Text(a.Task(), item.Progress{}, true, served); !strings.Contains(got, "#2 Home automation › Enclosure") {
		t.Errorf("the card names the objective as\n%s", got)
	}
	if got := Lines([]Task{a.Task()}, Marks{Serves: map[int][]string{1: served.Serves}}); !strings.Contains(got, "◇ Home automation › Enclosure") {
		t.Errorf("the listing names the objective as %q", got)
	}
}
