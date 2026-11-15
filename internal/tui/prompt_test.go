package tui

import (
	"strings"
	"testing"
)

func TestACommandNobodyKnowsIsSaid(t *testing.T) {
	b := hit(standing(twoTasks()), ":", "z", "enter")
	if !strings.Contains(bare(b.Draw()), "not a command") {
		t.Fatalf("a command nobody knows was swallowed:\n%s", b.Draw())
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
