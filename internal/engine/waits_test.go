package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"cairn/internal/objective"
)

func waitsOn(number string, on ...string) func() (Operation, Input) {
	return func() (Operation, Input) {
		return ObjectiveEdit, Input{Given: map[string]string{"number": number}, Listed: map[string][]string{"waits-on": on}}
	}
}

func shownObjective(t *testing.T, path, number string) (Answer, Objective) {
	t.Helper()
	shown, err := Do(ObjectiveShow, path, given("number", number))
	if err != nil {
		t.Fatal(err)
	}
	said, _ := Of(ObjectiveShow, shown).(Objective)
	return shown, said
}

func TestObjectiveWaitsOn_isChangedTheWayASetIs(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Enclosure"),
		asked(ObjectiveNew, "title", "Firmware"),
		asked(ObjectiveNew, "title", "Features"),
		waitsOn("3", "2", "+1"),
	)
	shown, said := shownObjective(t, path, "3")
	if !slices.Equal(said.WaitsOn, []int{1, 2}) || said.WaitedOnBy == nil || len(said.WaitedOnBy) != 0 {
		t.Errorf("#3 waits on %v and is waited on by %v", said.WaitsOn, said.WaitedOnBy)
	}
	if text := RenderText(ObjectiveShow, shown); !strings.Contains(text, "#1 Enclosure, #2 Firmware") {
		t.Errorf("a person reads:\n%s", text)
	}
	shown, said = shownObjective(t, path, "1")
	if !slices.Equal(said.WaitedOnBy, []int{3}) || said.WaitsOn == nil || len(said.WaitsOn) != 0 {
		t.Errorf("#1 waits on %v and is waited on by %v", said.WaitsOn, said.WaitedOnBy)
	}
	if text := RenderText(ObjectiveShow, shown); !strings.Contains(text, "#3 Features waits on it") {
		t.Errorf("a person reads:\n%s", text)
	}
	ran(t, path, waitsOn("3", "-1"))
	if _, said = shownObjective(t, path, "3"); !slices.Equal(said.WaitsOn, []int{2}) {
		t.Errorf("#3 waits on %v once #1 is taken off", said.WaitsOn)
	}
	listed, err := Do(ObjectiveList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if all := Of(ObjectiveList, listed).(Objectives).Objectives; !slices.Equal(all[1].WaitedOnBy, []int{3}) || !slices.Equal(all[2].WaitsOn, []int{2}) {
		t.Errorf("the listing reads %+v", all)
	}
}

func TestObjectiveWaitsOn_refusesWhatCannotBe(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Enclosure"),
		asked(ObjectiveNew, "title", "Firmware"),
		waitsOn("2", "1"),
	)
	for _, c := range []struct {
		ask  func() (Operation, Input)
		want error
	}{
		{waitsOn("1", "1"), objective.ErrWaitsOnItself},
		{waitsOn("1", "2"), objective.ErrWaitsOnItself},
		{waitsOn("1", "9"), ErrNoSuchObjective},
		{waitsOn("1", "x"), ErrWaitsOnRequired},
		{waitsOn("1", "=2", "+1"), ErrWaitsOnBothWays},
		{func() (Operation, Input) {
			return ObjectiveNew, Input{Given: map[string]string{"title": "Other"}, Listed: map[string][]string{"waits-on": {"+1"}}}
		}, ErrWaitsOnSignAtBirth},
	} {
		op, in := c.ask()
		if _, err := Do(op, path, in); !errors.Is(err, c.want) {
			t.Errorf("%s %v reads %v, want %v", op.Line(), in.Listed, err, c.want)
		}
	}
	ran(t, path, func() (Operation, Input) {
		return ObjectiveNew, Input{Given: map[string]string{"title": "Features"}, Listed: map[string][]string{"waits-on": {"2"}}}
	})
	if _, said := shownObjective(t, path, "3"); !slices.Equal(said.WaitsOn, []int{2}) {
		t.Errorf("#3 made waiting on #2 waits on %v", said.WaitsOn)
	}
}

func TestObjectiveWaitsOn_warnsWhenWorkStarts(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Home automation"),
		asked(ObjectiveNew, "title", "Firmware"),
		asked(ObjectiveNew, "title", "Features", "parent", "1"),
		waitsOn("1", "2"),
		asked(TaskNew, "title", "Dimmer scenes", "objective", "3"),
		asked(TaskNew, "title", "Presence", "objective", "3"),
		asked(TaskNew, "title", "Siren", "objective", "3"),
	)
	want := "#1 Home automation waits on #2 Firmware, not reached"
	moved, err := Do(TaskMove, path, given("number", "1", "state", "current"))
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(TaskMove, moved).(Tasks); !slices.Equal(said.Warnings, []string{want}) {
		t.Errorf("the move warns %q", said.Warnings)
	}
	if text := RenderText(TaskMove, moved); !strings.Contains(text, want) {
		t.Errorf("a person reads:\n%s", text)
	}
	again, _ := Do(TaskMove, path, given("number", "1", "state", "current"))
	if len(again.Warnings) != 0 {
		t.Errorf("a task already in current warns %q", again.Warnings)
	}
	edited, err := Do(TaskEdit, path, given("number", "2", "state", "current"))
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(TaskEdit, edited).(Task); !slices.Equal(said.Warnings, []string{want}) || !strings.Contains(RenderText(TaskEdit, edited), want) {
		t.Errorf("the edit warns %q", said.Warnings)
	}
	focused, err := Do(TaskFocus, path, given("number", "3"))
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(TaskFocus, focused).(Focus); !slices.Equal(said.Warnings, []string{want}) || !strings.Contains(RenderText(TaskFocus, focused), want) {
		t.Errorf("the focus warns %q", said.Warnings)
	}
	made, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "Lights", "state": "current"}, Listed: map[string][]string{"objective": {"3"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(made.Warnings, []string{want}) {
		t.Errorf("a task made in current warns %q", made.Warnings)
	}
	ran(t, path, asked(ObjectiveEdit, "number", "2", "state", "reached"), asked(TaskNew, "title", "Last", "objective", "3"))
	if quiet, _ := Do(TaskMove, path, given("number", "5", "state", "current")); len(quiet.Warnings) != 0 {
		t.Errorf("once #2 is reached the move warns %q", quiet.Warnings)
	}
}

func TestObjectiveWaitsOn_goesWithWhatIsDeleted(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Firmware"),
		asked(ObjectiveNew, "title", "Features"),
		waitsOn("2", "1"),
	)
	if _, err := Do(ObjectiveDelete, path, Agreed(given("number", "1"))); err != nil {
		t.Fatal(err)
	}
	if _, said := shownObjective(t, path, "2"); len(said.WaitsOn) != 0 {
		t.Errorf("#2 still waits on %v", said.WaitsOn)
	}
}

func TestObjectiveWaitsOn_aTaskMovedElsewhereThanCurrentWarnsOfNothing(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Home automation"),
		asked(ObjectiveNew, "title", "Firmware"),
		waitsOn("1", "2"),
		asked(TaskNew, "title", "Dimmer scenes", "objective", "1"),
	)
	moved, err := Do(TaskMove, path, given("number", "1", "state", "todo"))
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(TaskMove, moved).(Tasks); len(said.Warnings) != 0 {
		t.Errorf("a move into todo warns %q", said.Warnings)
	}
}
