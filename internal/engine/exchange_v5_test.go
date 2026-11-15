package engine

import (
	"cairn/internal/engine/exchange"
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
	if !slices.Equal(numbers, []int{1, 2, 3, 4}) {
		t.Fatalf("the file holds objectives %v, want every one, served or not", numbers)
	}
	if read.Objectives[1].Parent != 0 || read.Objectives[2].Parent != 2 || read.Objectives[3].Parent != 2 || !slices.Equal(read.Objectives[3].WaitsOn, []int{3}) {
		t.Errorf("the objectives read %+v", read.Objectives)
	}
	if !slices.ContainsFunc(read.Labels, func(l exchange.LabelFiled) bool { return l.Label == "area" && l.Scope == "objectives" }) {
		t.Errorf("labels = %v, want area for objectives", read.Labels)
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
		if !errors.Is(err, exchange.ErrExportPointsOutside) {
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
