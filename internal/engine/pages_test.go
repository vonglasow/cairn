package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPages_offerOnlyTheActsOfThePagesShown(t *testing.T) {
	off := PagesOff("time")
	if !off.Shows(PageTasks) || !off.Shows(PageObjectives) || off.Shows(PageTime) {
		t.Errorf("with time off, it shows tasks %v, objectives %v, time %v", off.Shows(PageTasks), off.Shows(PageObjectives), off.Shows(PageTime))
	}
	if off.Offers(TimesheetShow) || off.Offers(TimesheetAdd) || !off.Offers(TaskList) || !off.Offers(ConfigShow) {
		t.Error("with time off, the timesheet is offered or something else is not")
	}
	all := PagesOff(PageTasks, PageObjectives, PageTime)
	for _, op := range All() {
		if all.Offers(op) && (op.Thing == "task" || op.Thing == "objective" || op.Thing == "timesheet") {
			t.Errorf("with every page off, %s is offered", op.Line())
		}
	}
	if !all.None() || off.None() || (Pages{}).None() {
		t.Error("None does not say that every page is off, and only then")
	}
	for _, op := range All() {
		if !(Pages{}).Offers(op) {
			t.Errorf("with nothing turned off, %s is not offered", op.Line())
		}
	}
}

func TestPagesFrom_readsTheConfiguration(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[pages]\nobjectives = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	pages := PagesFrom(Home{Dir: home})
	if pages.Shows(PageObjectives) || !pages.Shows(PageTasks) {
		t.Errorf("read from the file, objectives shown %v and tasks %v", pages.Shows(PageObjectives), pages.Shows(PageTasks))
	}
	if got := PagesFrom(Home{}); got.None() || !got.Shows(PageTime) {
		t.Error("with no home, a page is off")
	}
}
