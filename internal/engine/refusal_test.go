package engine

import (
	"cairn/internal/engine/migration"
	"cairn/internal/engine/task"
	"fmt"
	"testing"
)

func TestRefusalOf_readsWhatTheReadersRefuseAsNotUnderstood(t *testing.T) {
	_, typed := TaskShow.Read([]string{"4", "5"})
	_, flag := TaskNew.Read([]string{"x", "--title"})
	_, yes := TaskMove.Read([]string{"4", "--first=maybe"})
	_, sent := TaskNew.FromJSON([]byte(`{"label": "one"}`))
	_, yesSent := TaskMove.FromJSON([]byte(`{"first": "yes"}`))
	_, broken := TaskNew.FromJSON([]byte(`{"title": `))
	for what, err := range map[string]error{
		"a word too many": typed, "a flag given nothing": flag, "a yes that is not": yes,
		"a list sent as a string": sent, "a yes sent as a string": yesSent, "JSON that is not": broken,
	} {
		if got := RefusalOf(err); got.Kind != KindUsage {
			t.Errorf("%s: RefusalOf(%v).Kind = %q, want usage", what, err, got.Kind)
		}
	}
}

func TestAgreeingTo_aStepStillOpenIsAgreedToByTheSameActConfirmed(t *testing.T) {
	open := fmt.Errorf("task 4: %w", ErrStepStillOpen)
	finishing := Operation{Thing: "task", Verb: "move", Finishes: true}
	asked := Input{Given: map[string]string{"number": "4"}}

	got, ok := AgreeingTo(finishing, asked, open)
	if !ok || got.Reason != "step-still-open" || got.Op.Name() != finishing.Name() || !got.In.Confirmed || got.In.Given["number"] != "4" {
		t.Errorf("AgreeingTo(finishing, asked, open) = %+v, %v, want the same act confirmed for step-still-open", got, ok)
	}

	for what, c := range map[string]struct {
		op  Operation
		in  Input
		err error
	}{
		"an act that does not finish": {Operation{Thing: "task", Verb: "show"}, asked, open},
		"an act already confirmed":    {finishing, Agreed(asked), open},
		"another refusal":             {finishing, asked, task.ErrNoSuchTask},
		"no refusal":                  {finishing, asked, nil},
	} {
		if got, ok := AgreeingTo(c.op, c.in, c.err); ok || got.Reason != "" || got.In.Confirmed {
			t.Errorf("%s: AgreeingTo = %+v, %v, want nothing to agree to", what, got, ok)
		}
	}
}

func TestAgreeingTo_aStoreFromANewerCairnHasNoActToAgreeWith(t *testing.T) {
	newer := Told(fmt.Errorf("%w: written by a later cairn", migration.ErrNewerStore))

	got, ok := AgreeingTo(TaskList, Input{}, newer)
	if !ok || got.Reason != "store-newer" || got.Op.Run != nil || got.Op.Name() != (Operation{}).Name() {
		t.Errorf("AgreeingTo(TaskList, Input{}, newer) = %+v, %v, want a question for store-newer with no act", got, ok)
	}
}

func TestAgreeingTo_aStoreBehindIsAgreedToByApplying(t *testing.T) {
	behind := Told(fmt.Errorf("%w by 2 change(s)", migration.ErrStoreBehind))

	got, ok := AgreeingTo(TaskList, Input{}, behind)
	if !ok || got.Reason != "store-behind" || got.Op.Name() != MigrationApply.Name() || got.In.Given["number"] != "" {
		t.Errorf("AgreeingTo(TaskList, Input{}, behind) = %+v, %v, want migration apply for store-behind", got, ok)
	}
}
