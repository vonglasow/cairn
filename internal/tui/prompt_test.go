package tui

import (
	"strings"
	"testing"

	"cairn/internal/engine"
)

// The prompt at the foot of the board: a search, a command the engine knows,
// and what is typed into either.

func TestACommandNobodyKnowsIsSaid(t *testing.T) {
	b := hit(standing(twoTasks()), ":", "z", "enter")
	if !strings.Contains(bare(b.Draw()), "not a command") {
		t.Fatalf("a command nobody knows was swallowed:\n%s", b.Draw())
	}
}

func TestASearchIsKeptWhenItIsAccepted(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "i", "r", "enter")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("the board carries %d tasks after a search, want 1", got)
	}
}

func TestACapitalCanBeTyped(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "shift+f")
	if !strings.Contains(bare(b.Draw()), "/F") {
		t.Fatalf("a capital never reached the prompt:\n%s", b.Draw())
	}
}

func TestBackspaceAtAPrompt(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "i", "backspace")
	if !strings.Contains(bare(b.Draw()), "/F▏") {
		t.Fatalf("backspace did not take the last letter off:\n%s", b.Draw())
	}
}

func TestAKeyThatMeansNothingAtAPromptIsIgnored(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F", "f1")
	if !strings.Contains(bare(b.Draw()), "/F▏") {
		t.Fatalf("an unknown key changed what was typed:\n%s", b.Draw())
	}
}

func TestSearchingReachesTheDescriptionAndTheReason(t *testing.T) {
	f := newFake()
	f.cards = []engine.Card{
		{Number: 1, Title: "One", State: "todo", Description: "about the spec"},
		{Number: 2, Title: "Two", State: "todo", Because: "the client"},
		{Number: 3, Title: "Three", State: "todo"},
	}
	b := hit(standing(f), "/", "s", "p", "e", "c")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("a search of the description reached %d tasks, want 1", got)
	}
	b = hit(standing(f), "/", "c", "l", "i")
	if got := b.Where().Showing; got != 1 {
		t.Fatalf("a search of the reason reached %d tasks, want 1", got)
	}
}

func TestSearchingTheObjectives(t *testing.T) {
	f := newFake()
	f.goals = []engine.Goal{
		{Number: 1, Title: "Ship the desktop board", State: "active"},
		{Number: 2, Title: "Cut the onboarding time", State: "active"},
	}
	b := hit(standing(f), "tab", "/", "s", "h", "i", "p")
	if !strings.Contains(b.Draw(), "Ship the desktop") {
		t.Fatalf("the search hid what it should have found:\n%s", b.Draw())
	}
	if strings.Contains(b.Draw(), "onboarding") {
		t.Fatalf("the search kept what it should have dropped:\n%s", b.Draw())
	}
}

// The bar offers ":" and the prompt knew two words out of the whole engine. A
// command typed there was answered "unknown" however well the command line knew it.
func TestTheCommandPromptKnowsWhatTheEngineKnows(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), ":")
	for _, r := range "task focus 1" {
		b = hit(b, string(r))
	}
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 || f.asked[0].op.Verb != "focus" {
		t.Fatalf("the prompt asked %v, want a task focus", f.asked)
	}
}

// What the prompt refuses, and where it refuses it: a word where no argument
// goes, and a required one left out. Both are said rather than run.
func TestTheCommandPromptRefusesWhatTheEngineWouldNotUnderstand(t *testing.T) {
	for _, c := range []struct {
		what string
		line string
	}{
		{"a word past the last argument", "task show 1 2"},
		{"a flag with nothing after it", "task show --number"},
		{"an argument that does not exist", "task show --nope 1"},
		{"a required argument left out", "task new"},
	} {
		t.Run(c.what, func(t *testing.T) {
			f := twoTasks()
			b := hit(standing(f), ":")
			for _, r := range c.line {
				if r == ' ' {
					b = hit(b, "space")
					continue
				}
				b = hit(b, string(r))
			}
			if _, cmd := b.pressed("enter"); cmd != nil {
				cmd()
			}
			if len(f.asked) != 0 {
				t.Fatalf("%q was run: %v", c.line, f.asked)
			}
		})
	}
}

// A sweep the store refuses says so rather than redrawing as though a column
// had gone off the board.
func TestASweepTheStoreRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), "shift+x")
	_, cmd := b.pressed("enter")
	if cmd == nil {
		t.Fatal("agreeing to the sweep asked for nothing")
	}
	looked, ok := cmd().(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a sweep that was refused answered %v", looked)
	}
}

// A list typed at the prompt reaches the list the operation reads, and a yes
// is named on its own — the words the command line takes. The prompt used to
// put every --name value into the plain map, so a label given here was
// dropped without a word and --first asked for a value.
func TestTheCommandPromptHandsOverListsAndYeses(t *testing.T) {
	for _, c := range []struct {
		line  string
		check func(engine.Input) bool
	}{
		{"task edit 1 --label x", func(in engine.Input) bool { return len(in.Listed["label"]) == 1 && in.Listed["label"][0] == "x" }},
		{"task move 2 --first", func(in engine.Input) bool { return in.Given["first"] == "true" }},
	} {
		f := twoTasks()
		b := hit(standing(f), ":")
		for _, r := range c.line {
			if r == ' ' {
				b = hit(b, "space")
				continue
			}
			b = hit(b, string(r))
		}
		if _, cmd := b.pressed("enter"); cmd != nil {
			cmd()
		}
		if len(f.asked) == 0 || !c.check(f.asked[0].in) {
			t.Errorf("%q asked %+v", c.line, f.asked)
		}
	}
}

// Two words are a command: the thing and the verb, with nothing after them.
func TestACommandOfTwoWordsRuns(t *testing.T) {
	f := twoTasks()
	b := hit(standing(f), ":")
	for _, r := range "task list" {
		if r == ' ' {
			b = hit(b, "space")
			continue
		}
		b = hit(b, string(r))
	}
	if _, cmd := b.pressed("enter"); cmd != nil {
		cmd()
	}
	if len(f.asked) == 0 || f.asked[0].op.Verb != "list" {
		t.Fatalf("the prompt asked %v, want a task list", f.asked)
	}
}

// What the prompt runs and the store refuses is said, the way every other
// gesture's refusal is.
func TestACommandTheStoreRefusesIsSaid(t *testing.T) {
	f := twoTasks()
	f.failOp = errRefuse
	b := hit(standing(f), ":")
	for _, r := range "task focus 1" {
		if r == ' ' {
			b = hit(b, "space")
			continue
		}
		b = hit(b, string(r))
	}
	_, cmd := b.pressed("enter")
	if cmd == nil {
		t.Fatal("the prompt asked for nothing")
	}
	looked, ok := cmd().(Looked)
	if !ok || looked.Err == nil {
		t.Fatalf("a command that was refused answered %v", looked)
	}
}
