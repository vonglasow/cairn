package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPagesFrom_readsTheConfiguration(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[pages]\nobjectives = false\nagents = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	pages := PagesFrom(Home{Dir: home})
	if pages.Shows(PageObjectives) || !pages.Shows(PageTasks) {
		t.Errorf("read from the file, objectives shown %v and tasks %v", pages.Shows(PageObjectives), pages.Shows(PageTasks))
	}
	if pages.Shows(PageAgents) || pages.Offers(AgentList) {
		t.Error("read from the file, the agents page is shown or its list offered")
	}
	if !PagesOff(PageTasks, PageObjectives, PageTime).Offers(AgentList) || PagesOff(PageTasks, PageObjectives, PageTime).None() {
		t.Error("with only the agents on, its list is not offered or no page is on")
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
