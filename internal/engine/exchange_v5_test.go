package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestTaskExport_writesTheTreeWhatWaitsAndTheScopes(t *testing.T) {
	read := exportedFrom(t,
		asked(ObjectiveNew, "title", "Garden"),
		asked(ObjectiveNew, "title", "Home automation"),
		asked(ObjectiveNew, "title", "Firmware", "parent", "2"),
		asked(ObjectiveNew, "title", "Features", "parent", "2"),
		waitsOn("4", "3"),
		asked(TaskNew, "title", "Dimmer scenes", "objective", "4"),
		made("label", "area", "scope", "objectives"),
	)
	numbers := []int{}
	for _, o := range read.Objectives {
		numbers = append(numbers, o.Number)
	}
	if !slices.Equal(numbers, []int{2, 3, 4}) {
		t.Fatalf("the file holds objectives %v, want what the task serves, its parent and what it waits on", numbers)
	}
	if read.Objectives[0].Parent != 0 || read.Objectives[1].Parent != 2 || read.Objectives[2].Parent != 2 || !slices.Equal(read.Objectives[2].WaitsOn, []int{3}) {
		t.Errorf("the objectives read %+v", read.Objectives)
	}
	if !slices.ContainsFunc(read.Labels, func(l labelFiled) bool { return l.Label == "area" && l.Scope == "objectives" }) {
		t.Errorf("labels = %v, want area for objectives", read.Labels)
	}
}

func TestTaskImport_bringsTheTreeWhatWaitsAndTheScopes(t *testing.T) {
	e := imported(t, `{"cairn": "tasks", "version": 5, "objectives": [
		{"number": 7, "title": "Features", "state": "active", "priority": "normal", "parent": 5, "waits_on": [6]},
		{"number": 5, "title": "Home automation", "state": "active", "priority": "normal"},
		{"number": 6, "title": "Firmware", "state": "active", "priority": "normal", "parent": 5}],
		"labels": [{"label": "area", "scope": "objectives"}],
		"tasks": [{"number": 1, "title": "Dimmer scenes", "state": "todo", "priority": "normal", "objectives": [7]}]}`)
	features, err := e.showObjective(1)
	if err != nil {
		t.Fatal(err)
	}
	if c := factsOfObjective(features); c.Title != "Features" || c.Parent != 2 || !slices.Equal(c.WaitsOn, []int{3}) {
		t.Errorf("Features came back as %+v", c)
	}
	if firmware, _ := e.showObjective(3); factsOfObjective(firmware).Parent != 2 {
		t.Errorf("Firmware came back under %d", factsOfObjective(firmware).Parent)
	}
	scopes, _ := e.scopes()
	if scopes.of("area") != "objectives" {
		t.Errorf("area came back for %s", scopes.of("area"))
	}
}

func TestTaskImport_refusesATreeOrAWaitOutsideTheFile(t *testing.T) {
	for _, file := range []string{
		`{"cairn": "tasks", "version": 5, "objectives": [{"number": 1, "title": "x", "state": "active", "priority": "normal", "parent": 9}], "tasks": []}`,
		`{"cairn": "tasks", "version": 5, "objectives": [{"number": 1, "title": "x", "state": "active", "priority": "normal", "waits_on": [9]}], "tasks": []}`,
	} {
		from := filepath.Join(t.TempDir(), "held.json")
		if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
		read, err := readExport(from)
		if err == nil {
			_, err = openEngine(t).importing(read)
		}
		if !errors.Is(err, ErrExportPointsOutside) {
			t.Errorf("%s reads %v", file, err)
		}
	}
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(`{"cairn": "tasks", "version": 5, "objectives": [], "labels": [{"label": "x", "scope": "nowhere"}], "tasks": []}`), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err == nil {
		_, err = openEngine(t).importing(read)
	}
	if err == nil {
		t.Error("a label of no such scope was read")
	}
}
