package engine

// The three moments of a form both boards share: opening one empty, opening
// one on a task, and saving it. Each board drew its own form and wrote its own
// copy of these, and the two had already come apart — the window opened a new
// task with no column and no priority where the terminal filled both in.

import (
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// NewTaskFields is the form of a task about to be added on a column: that
// column already filled in — a task added there belongs there, and asking
// again would be asking something the board already knows — and normal
// priority, which is what a task carries when nothing is said.
func NewTaskFields(state string) []Field {
	return TaskFields(Card{State: state, Priority: string(domain.Normal)})
}

// NewGoalFields is the form of an objective about to be added: active, which
// is where an objective starts, at normal priority.
func NewGoalFields() []Field {
	return GoalFields(Goal{State: string(objective.Active), Priority: string(domain.Normal)})
}

// EditFields is the form of a task being changed, its parts included. They are
// reachable here and nowhere else: a part is off the board, so a task that
// carried one could neither be changed nor deleted once it did.
func EditFields(c Card, cards []Card) []Field {
	return append(TaskFields(c), partsField(PartsSaid(Hanging(cards, c.Number))))
}

// AddingFields is the form of a task about to be added in a window: the new
// task's fields and, last, its parts — empty, and the same field its edit form
// has — so a task whose parts were already written down goes in with them.
func AddingFields(state string) []Field {
	return append(NewTaskFields(state), partsField(""))
}

// partsField is the field a task's parts are written in, a line each.
func partsField(said string) Field {
	return Field{Name: "parts", Value: said, Long: true}
}

// Saved is what a saved form asks for: adding or editing — one form, so the
// difference is the operation and the number and nothing else — with the
// title trimmed on its way out, since what somebody called a task does not
// include the space they left after it.
func Saved(fields []Field, adding bool, number int, add, edit Operation) (Operation, Input) {
	op := edit
	if adding {
		op = add
	}
	asked := Asked(fields, op)
	title, _ := valueOf(fields, "title")
	asked.Given["title"] = strings.TrimSpace(title)
	if !adding {
		asked.Given["number"] = strconv.Itoa(number)
	}
	return op, asked
}

// valueOf is what a form holds under a name.
func valueOf(fields []Field, name string) (string, bool) {
	for _, f := range fields {
		if f.Name == name {
			return f.Value, true
		}
	}
	return "", false
}
