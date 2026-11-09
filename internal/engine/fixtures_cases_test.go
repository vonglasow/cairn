package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestFixtures_holdEveryCaseABoardDraws(t *testing.T) {
	var read exchanged
	if err := json.Unmarshal(fixtures, &read); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	saw := func(what string, yes bool) {
		if yes {
			seen[what] = true
		}
	}
	perColumn := map[string]int{}
	focused := 0
	today := time.Now()
	for _, f := range read.Tasks {
		perColumn[f.State]++
		saw("priority "+f.Priority, true)
		saw("a title too long for a card", len([]rune(f.Title)) > 80)
		saw("accents", strings.ContainsAny(f.Title, "éèàçô"))
		saw("an emoji", strings.ContainsFunc(f.Title, func(r rune) bool { return r > 0x1F000 }))
		saw("a long description", len(f.Description) > 400)
		saw("no description", f.Description == "")
		done := 0
		for _, s := range f.Steps {
			if s.Done {
				done++
			}
		}
		saw("no step", len(f.Steps) == 0)
		saw("every step done", len(f.Steps) > 0 && done == len(f.Steps))
		saw("half the steps done", len(f.Steps) > 1 && done*2 == len(f.Steps))
		saw("twenty steps", len(f.Steps) >= 20)
		saw("no label", len(f.Labels) == 0)
		saw("one label", len(f.Labels) == 1)
		saw("too many labels", len(f.Labels) >= 6)
		if day, err := domain.FromToday(f.Deadline, today); err == nil && day != "" {
			due, _ := domain.ParseDeadline(day)
			days := int(due.Sub(startOfDay(today)).Hours() / 24)
			saw("a deadline past", days < 0)
			saw("a deadline today", days == 0)
			saw("a deadline tomorrow", days == 1)
			saw("a deadline in a week", days >= 5 && days <= 9)
			saw("a deadline far off", days > 60)
		}
		saw("an estimate", f.Estimate != "")
		saw("time over several weeks", len(f.Spent) >= 5)
		saw("blocked by a task", len(f.BlockedBy) > 0)
		saw("blocked by a reason", f.Because != "")
		saw("linked", len(f.Linked) > 0)
		saw("archived", f.Archived)
		if f.Focused {
			focused++
		}
	}
	for _, column := range domain.States {
		if perColumn[string(column)] == 0 {
			t.Errorf("no task in %s", column)
		}
	}
	if slices.Max(slicesOf(perColumn)) < 20 {
		t.Error("no column holds enough to scroll")
	}
	if focused != 2 {
		t.Errorf("%d tasks in focus, want 2", focused)
	}
	for _, p := range domain.Priorities {
		if !seen["priority "+string(p)] {
			t.Errorf("no task of priority %s", p)
		}
	}
	for _, want := range []string{"a title too long for a card", "accents", "an emoji", "a long description", "no description",
		"no step", "every step done", "half the steps done", "twenty steps", "no label", "one label", "too many labels",
		"a deadline past", "a deadline today", "a deadline tomorrow", "a deadline in a week", "a deadline far off",
		"an estimate", "time over several weeks", "blocked by a task", "blocked by a reason", "linked", "archived"} {
		if !seen[want] {
			t.Errorf("no task with %s", want)
		}
	}
	standings := map[string]bool{}
	under, waiting := 0, false
	for _, o := range read.Objectives {
		standings[o.State] = true
		if o.Parent != 0 {
			under++
		}
		waiting = waiting || len(o.WaitsOn) > 0
	}
	if len(standings) < 4 {
		t.Errorf("objectives stand only %v", standings)
	}
	if under < 2 || !waiting {
		t.Errorf("%d sub-objectives and one waiting on another %v; want a parent holding several, one of them waiting on another", under, waiting)
	}
	coloured, idle := false, false
	carried := map[string]bool{}
	for _, f := range read.Tasks {
		for _, l := range f.Labels {
			carried[l] = true
		}
	}
	scoped := false
	for _, l := range read.Labels {
		coloured = coloured || l.Colour != ""
		idle = idle || !carried[l.Label]
		scoped = scoped || l.Scope == "objectives"
	}
	if !coloured || !idle || !scoped {
		t.Errorf("labels: a colour chosen %v, one nothing carries %v, one for objectives only %v", coloured, idle, scoped)
	}
}

func slicesOf(m map[string]int) []int {
	out := []int{}
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
