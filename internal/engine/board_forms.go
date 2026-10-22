package engine

import (
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// NewTaskFields is the form of a task about to be added on a column: that column already filled in — a task added there belongs there, and asking again would be asking something the board already knows — and normal priority, which is what a task carries when nothing is said.
func NewTaskFields(state string) []Field {
	return TaskFields(Card{State: state, Priority: string(domain.Normal)})
}

// LinkedTaskFields is the form of a task about to be added linked to another: in todo, where a task that came of one done starts again, and the link already written.
func LinkedTaskFields(number int) []Field {
	fields := AddingFields(string(domain.Todo))
	for i := range fields {
		if fields[i].Name == "links" {
			fields[i].Value = strconv.Itoa(number)
		}
	}
	return fields
}

// NewGoalFields is the form of an objective about to be added: active, which is where an objective starts, at normal priority.
func NewGoalFields() []Field {
	return GoalFields(Goal{State: string(objective.Active), Priority: string(domain.Normal)})
}

// EditFields is the form of a task being changed, its steps included.
func EditFields(c Card) []Field {
	return append(TaskFields(c), stepsField(PartsWritten(Hanging(c))))
}

// AddingFields is the form of a task about to be added in a window: the new task's fields and, last, its steps — empty, and the same field its edit form has — so a task whose steps were already written down goes in with them.
func AddingFields(state string) []Field {
	return append(NewTaskFields(state), stepsField(""))
}

func stepsField(said string) Field {
	return Field{Name: "steps", Arg: "steps", Value: said, Was: said, Means: meaning(TaskEdit, "steps"), Long: true,
		Example: "[ ] Draft the columns"}
}

// Saved is what a saved form asks for: adding or editing — one form, so the difference is the operation and the number and nothing else — with the title trimmed on its way out, since what somebody called a task does not include the space they left after it.
func Saved(fields []Field, adding bool, number int, add, edit Operation) (Operation, Input) {
	op, sent := edit, changedOf(fields)
	if adding {
		op, sent = add, fields
	}
	asked := Asked(sent, op)
	title, _ := valueOf(fields, "title")
	asked.Given["title"] = strings.TrimSpace(title)
	if !adding {
		asked.Given["number"] = strconv.Itoa(number)
	}
	return op, asked
}

func changedOf(fields []Field) []Field {
	var changed []Field
	for _, f := range fields {
		if f.Value != f.Was {
			changed = append(changed, f)
		}
	}
	return changed
}

func valueOf(fields []Field, name string) (string, bool) {
	for _, f := range fields {
		if f.Name == name {
			return f.Value, true
		}
	}
	return "", false
}
