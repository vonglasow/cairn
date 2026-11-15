package board

import (
	"cairn/internal/engine"
	"slices"
	"strings"
)

// Form is the modal over the board while it is open.
type Form struct {
	Adding       bool
	ForObjective bool
	ForSeries    bool
	Number       int
	Whole        bool
	Answered     bool

	Fields      []Field
	At          int
	Caret       int
	Held        int
	PartAt      int
	PartsTaken  []int
	PartsTop    int
	SpareParts  int
	WritingPart bool
	Named       Written
	Carried     int
	Opened      []string
	Behind      int
	Tracing     bool
	OnWork      bool
	WorkAt      int
}

// Changed says whether anything in the form differs from what it opened with.
func (f Form) Changed() bool {
	if len(f.Opened) != len(f.Fields) {
		return false
	}
	for i := range f.Fields {
		if f.Fields[i].Value != f.Opened[i] {
			return true
		}
	}
	return false
}

// AsOpened is the form as it opens: what each field holds is what it is changed against.
func AsOpened(f Form) Form {
	f.Opened = make([]string, len(f.Fields))
	for i := range f.Fields {
		f.Opened[i] = f.Fields[i].Value
	}
	return f
}

// Writing is the field the keyboard is on, as a text with its caret.
func (f Form) Writing() Written {
	if len(f.Fields) == 0 {
		return Written{}
	}
	return Written{Value: f.Fields[f.At].Value, At: f.Caret, Held: f.Held}
}

// Writes puts a written text back into the field the keyboard is on.
func (f Form) Writes(written Written) Form {
	if len(f.Fields) == 0 {
		return f
	}
	f = f.Set(f.Fields[f.At].Name, written.Reads())
	f.Caret, f.Held = written.At, written.Held
	return f
}

// Onto is the form with the keyboard on another field, the caret at the end of what it holds.
func (f Form) Onto(at int) Form {
	if f.WritingPart {
		f = f.Committing()
		f.WritingPart = false
	}
	arriving := at != f.At
	if arriving {
		f.PartAt = 0
	}
	f.At = min(len(f.Fields)-1, max(0, at))
	f.OnWork = false
	if _, sectional := SectionOf(f.Fields[f.At]); sectional {
		f.Tracing = false
	}
	if arriving && f.OnParts() {
		f = f.TakingTheOneOn()
	}
	f.Caret, f.Held = len([]rune(f.Fields[f.At].Value)), 0
	return f
}

// Set writes a field.
func (f Form) Set(name, value string) Form {
	fields := make([]Field, len(f.Fields))
	copy(fields, f.Fields)
	for i := range fields {
		if fields[i].Name == name {
			fields[i].Value = value
			f.At = i
		}
	}
	f.Fields = fields
	return f
}

// On is the field the keyboard is on.
func (f Form) On() int { return f.At }

// Argued is what the field standing for an argument holds, and nothing when no field does.
func (f Form) Argued(arg string) string {
	for _, field := range f.Fields {
		if field.Arg == arg {
			return field.Value
		}
	}
	return ""
}

// Values is what each field holds, in order.
func (f Form) Values() []string {
	var held []string
	for _, field := range f.Fields {
		held = append(held, field.Value)
	}
	return held
}

// Titled says whether the form has a title written, or asks for none.
func (f Form) Titled() bool {
	for _, field := range f.Fields {
		if field.Name == "title" {
			return strings.TrimSpace(field.Value) != ""
		}
	}
	return true
}

// Drafted is what the form holds that differs from what it opened with, by field name.
func (f Form) Drafted() map[string]string {
	values := map[string]string{}
	for _, field := range f.Fields {
		if field.Value != field.Was {
			values[field.Name] = field.Value
		}
	}
	return values
}

// PartsField is where the steps are among the fields, and whether the form has them.
func (f Form) PartsField() (int, bool) {
	at := slices.IndexFunc(f.Fields, func(field Field) bool { return field.Name == "steps" })
	return at, at != -1
}

// Parts is the steps the form holds.
func (f Form) Parts() []engine.Part {
	at, ok := f.PartsField()
	if !ok {
		return nil
	}
	return engine.PartsIn(f.Fields[at].Value)
}

// WithParts is the form holding these steps.
func (f Form) WithParts(parts []engine.Part) Form {
	if at, ok := f.PartsField(); ok {
		f.Fields = slices.Clone(f.Fields)
		f.Fields[at].Value = PartsWritten(parts)
	}
	return f
}

// OnParts says whether the keyboard is on the steps.
func (f Form) OnParts() bool {
	at, ok := f.PartsField()
	return ok && f.At == at
}

// Lines is how many lines the steps take: one each, and one to add another.
func (f Form) Lines() int { return len(f.Parts()) + 1 }

// WalkingParts is the form moved through its steps, onto the field above from the first.
func (f Form) WalkingParts(by int) Form {
	if f.PartAt+by < 0 {
		return f.Onto(f.At - 1)
	}
	f.PartAt = min(f.Lines()-1, f.PartAt+by)
	return f.TakingTheOneOn()
}

// TickedTogether is the steps a tick acts on: those taken when the one on is among them, otherwise the one on.
func (f Form) TickedTogether() []int {
	if taken := f.TakenParts(); slices.Contains(taken, f.PartAt) {
		return taken
	}
	return []int{f.PartAt}
}

// Ticking is the form with one step done or not.
func (f Form) Ticking(at int, done bool) Form {
	parts := f.Parts()
	parts[at].Done = done
	return f.WithParts(parts)
}

// OpenedWith is the form with one step held as it opened with as the engine now has it, so ticking it changes nothing to save.
func (f Form) OpenedWith(part engine.Part) Form {
	at, _ := f.PartsField()
	opened := engine.PartsIn(f.Opened[at])
	for i := range opened {
		if opened[i].Number == part.Number {
			opened[i].Done = part.Done
		}
	}
	f.Opened = slices.Clone(f.Opened)
	f.Opened[at] = PartsWritten(opened)
	return f
}

// Without is the steps with one taken out, those after it numbered one less.
func Without(parts []engine.Part, gone int) []engine.Part {
	kept := make([]engine.Part, 0, len(parts))
	for _, part := range parts {
		if part.Number == gone {
			continue
		}
		if part.Number > gone {
			part.Number--
		}
		kept = append(kept, part)
	}
	return kept
}

// Naming is the form writing the title of the step it is on, or of a new one.
func (f Form) Naming() Form {
	parts := f.Parts()
	f.WritingPart, f.Named = true, Written{}
	if f.PartAt < len(parts) {
		title := parts[f.PartAt].Title
		f.Named = Written{Value: title, At: len([]rune(title))}
	}
	return f
}

// Committing is the form with the step being written kept: a new one added at the end, or the one named renamed.
func (f Form) Committing() Form {
	if !f.WritingPart {
		return f
	}
	f.WritingPart = false
	if f.Named.Empty() {
		return f
	}
	parts := f.Parts()
	if f.PartAt < len(parts) {
		parts[f.PartAt].Title = f.Named.Reads()
		return f.WithParts(parts)
	}
	f = f.WithParts(append(parts, engine.Part{Title: f.Named.Reads()}))
	f.PartAt, f.WritingPart, f.Named = f.Lines()-1, true, Written{}
	return f
}

// Carrying is the form with one step carried to another place in the list.
func (f Form) Carrying(from, to int) Form {
	parts := f.Parts()
	if from >= len(parts) {
		return f
	}
	to = min(len(parts)-1, to)
	moved := parts[from]
	f.PartAt = to
	return f.WithParts(slices.Insert(slices.Delete(parts, from, from+1), to, moved))
}

// TakingPart is the form with a step taken: alone, gathered with the others taken, or every step from the one on to it.
func (f Form) TakingPart(at int, gathering, stretching bool) Form {
	parts := f.Parts()
	switch {
	case at >= len(parts):
		f.PartsTaken = nil
	case gathering:
		taken := slices.Clone(f.SelectedParts())
		if i := slices.Index(taken, at); i >= 0 {
			taken = slices.Delete(taken, i, i+1)
		} else {
			taken = append(taken, at)
		}
		slices.Sort(taken)
		f.PartsTaken = taken
	case stretching:
		from := min(f.PartAt, len(parts)-1)
		f.PartsTaken = nil
		for i := min(from, at); i <= max(from, at); i++ {
			f.PartsTaken = append(f.PartsTaken, i)
		}
	default:
		f.PartsTaken = []int{at}
	}
	f.PartAt = at
	return f
}

// SelectedParts is the steps taken, those still in the list.
func (f Form) SelectedParts() []int {
	n := len(f.Parts())
	return slices.DeleteFunc(slices.Clone(f.PartsTaken), func(i int) bool { return i >= n })
}

// TakenParts is the steps taken together: two and more, or nothing.
func (f Form) TakenParts() []int {
	if taken := f.SelectedParts(); len(taken) >= 2 {
		return taken
	}
	return nil
}

// TakingTheOneOn is the form with only the step it is on taken.
func (f Form) TakingTheOneOn() Form {
	f.PartsTaken = nil
	if f.PartAt < len(f.Parts()) {
		f.PartsTaken = []int{f.PartAt}
	}
	return f
}

// CarryingSeveral is the form with the steps taken together carried to a place, as one block.
func (f Form) CarryingSeveral(to int) Form {
	parts, taken := f.Parts(), f.TakenParts()
	if taken == nil || slices.Contains(taken, to) {
		return f
	}
	var block, rest []engine.Part
	for i, p := range parts {
		if slices.Contains(taken, i) {
			block = append(block, p)
			continue
		}
		rest = append(rest, p)
	}
	at := len(rest)
	if to < len(parts) {
		at = 0
		for i := range to {
			if !slices.Contains(taken, i) {
				at++
			}
		}
		if to > taken[0] {
			at++
		}
	}
	f.PartsTaken = nil
	for i := range block {
		f.PartsTaken = append(f.PartsTaken, at+i)
	}
	f.PartAt = at + len(block) - 1
	return f.WithParts(slices.Insert(rest, at, block...))
}

// TakingAll is the form with every step taken.
func (f Form) TakingAll() Form {
	n := len(f.Parts())
	if n < 2 {
		return f
	}
	f.PartsTaken = nil
	for i := range n {
		f.PartsTaken = append(f.PartsTaken, i)
	}
	f.PartAt = n - 1
	return f
}

// FormSections are the sections a form lays its fields under, each with the arguments it holds.
var FormSections = []struct {
	Head string
	args []string
}{
	{"PARENT", []string{"parent", "waits-on"}},
	{"STATE", []string{"state", "priority", "label"}},
	{"TIME", []string{"deadline", "estimate", "spent"}},
	{"BLOCKERS", []string{"blocker", "because"}},
	{"LINKS", []string{"link", "objective"}},
	{"SERIES", []string{"every", "from"}},
}

// SectionOf is the section of the form a field is laid out under, and whether it is under one.
func SectionOf(field Field) (int, bool) {
	for i, s := range FormSections {
		if slices.Contains(s.args, field.Arg) {
			return i, true
		}
	}
	return 0, false
}
