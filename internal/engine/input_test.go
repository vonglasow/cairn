package engine

import (
	"reflect"
	"strings"
	"testing"
)

func TestRead_handsOverWhatWasTypedTheWayTheOperationTakesIt(t *testing.T) {
	for _, tc := range []struct {
		what  string
		op    Operation
		words string
		want  Input
	}{
		{"a positional, then a flag", TaskEdit, "3 --priority high",
			Input{Given: map[string]string{"number": "3", "priority": "high"}}},
		{"a list flag, named twice", TaskEdit, "3 --label x --label y",
			Input{Given: map[string]string{"number": "3"}, Listed: map[string][]string{"label": {"x", "y"}}}},
		{"a positional list takes every word", TaskSearch, "foo bar",
			Input{Listed: map[string][]string{"named": {"foo", "bar"}}}},
		{"a yes on its own", TaskMove, "2 --first",
			Input{Given: map[string]string{"number": "2", "first": "true"}}},
		{"a yes said false", TaskMove, "2 --first=false",
			Input{Given: map[string]string{"number": "2", "first": "false"}}},
		{"a value after an equals sign", TaskEdit, "3 --priority=low",
			Input{Given: map[string]string{"number": "3", "priority": "low"}}},
	} {
		got, err := tc.op.Read(strings.Fields(tc.words))
		if err != nil {
			t.Errorf("%s: %v", tc.what, err)
			continue
		}
		if !sameInput(got, tc.want) {
			t.Errorf("%s: Read = %+v, want %+v", tc.what, got, tc.want)
		}
	}
}

func TestRead_refusesWhatItCannotPlace(t *testing.T) {
	for _, tc := range []struct {
		what  string
		op    Operation
		words string
		says  string
	}{
		{"a word with no positional left", TaskShow, "3 4", `"4"`},
		{"a flag waiting for a value", TaskEdit, "3 --priority", "--priority"},
		{"a name the operation does not take", TaskShow, "3 --numbre 4", "numbre"},
		{"a yes that is neither", TaskMove, "2 --first=perhaps", "perhaps"},
	} {
		_, err := tc.op.Read(strings.Fields(tc.words))
		if err == nil || !strings.Contains(err.Error(), tc.says) {
			t.Errorf("%s: Read = %v, want a refusal naming %s", tc.what, err, tc.says)
		}
	}
}

func TestFromJSON_handsOverWhatWasSentTheWayTheOperationTakesIt(t *testing.T) {
	for _, tc := range []struct {
		what string
		op   Operation
		sent string
		want Input
	}{
		{"a large number", TaskShow, `{"number": 1000000}`,
			Input{Given: map[string]string{"number": "1000000"}}},
		{"a list", TaskSearch, `{"named": ["spec", "team"]}`,
			Input{Listed: map[string][]string{"named": {"spec", "team"}}}},
		{"null empties a field", TaskEdit, `{"number": 3, "deadline": null}`,
			Input{Given: map[string]string{"number": "3", "deadline": ""}}},
		{"null empties a list", TaskEdit, `{"number": 3, "label": null}`,
			Input{Given: map[string]string{"number": "3"}, Listed: map[string][]string{"label": {}}}},
		{"a yes", TaskMove, `{"number": 2, "first": true}`,
			Input{Given: map[string]string{"number": "2", "first": "true"}}},
		{"nothing at all", TaskList, ``, Input{}},
	} {
		got, err := tc.op.FromJSON([]byte(tc.sent))
		if err != nil {
			t.Errorf("%s: %v", tc.what, err)
			continue
		}
		if !sameInput(got, tc.want) {
			t.Errorf("%s: FromJSON = %+v, want %+v", tc.what, got, tc.want)
		}
	}
}

func TestFromJSON_refusesWhatItCannotRead(t *testing.T) {
	for _, tc := range []struct {
		what string
		op   Operation
		sent string
		says string
	}{
		{"not JSON", TaskShow, `{`, ""},
		{"a list that is a string", TaskSearch, `{"named": "spec"}`, "named"},
		{"a number among strings", TaskSearch, `{"named": ["spec", 3]}`, "named"},
		{"a yes that is a string", TaskMove, `{"number": 2, "first": "yes"}`, "first"},
	} {
		_, err := tc.op.FromJSON([]byte(tc.sent))
		if err == nil || !strings.Contains(err.Error(), tc.says) {
			t.Errorf("%s: FromJSON = %v, want a refusal naming %q", tc.what, err, tc.says)
		}
	}
}

func TestCalled_findsTheOperationTwoWordsName(t *testing.T) {
	if op, found := Called("task", "show"); !found || op.Name() != TaskShow.Name() {
		t.Errorf("Called(task, show) = %s, %v", op.Name(), found)
	}
	if _, found := Called("task", "fly"); found {
		t.Error("Called(task, fly) found something")
	}
}

func sameInput(a, b Input) bool {
	norm := func(in Input) Input {
		if in.Given == nil {
			in.Given = map[string]string{}
		}
		if in.Listed == nil {
			in.Listed = map[string][]string{}
		}
		return in
	}
	return reflect.DeepEqual(norm(a), norm(b))
}
