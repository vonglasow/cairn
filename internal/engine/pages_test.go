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

func TestPages_takeTheStepsWithTheTasksAndTheLabelsWithTheLastPageTheyLabel(t *testing.T) {
	if PagesOff(PageTasks).Offers(StepNew) {
		t.Error("with the tasks off, a step is still offered")
	}
	if !PagesOff(PageTasks).Offers(LabelList) {
		t.Error("with the objectives still shown, the labels are not offered")
	}
	if PagesOff(PageTasks, PageObjectives).Offers(LabelList) {
		t.Error("with nothing left to label, the labels are still offered")
	}
	if op, found := PagesOff(PageTime).Called("timesheet", "show"); found {
		t.Errorf("with the time off, %s is still called", op.Line())
	}
	if _, found := PagesOff(PageTime).Called("task", "list"); !found {
		t.Error("with the time off, task list is not called")
	}
}

func TestPages_aWeekShowsFiveDaysWhenTheConfigurationHidesTheWeekends(t *testing.T) {
	if got := (Pages{}).WeekDays(); got != 7 {
		t.Errorf("with nothing said, a week shows %d days, want 7", got)
	}
	if got := PagesOff(PageTasks).WithoutWeekends().WeekDays(); got != 5 {
		t.Errorf("without its weekend, a week shows %d days, want 5", got)
	}
	home := t.TempDir()
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[time]\nweekends = \"hidden\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := PagesFrom(Home{Dir: home}).WeekDays(); got != 5 {
		t.Errorf("read from the file, a week shows %d days, want 5", got)
	}
}
