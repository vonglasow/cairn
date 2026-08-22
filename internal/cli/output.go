package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"cairn/internal/domain"
)

// jsonFlag is the name of the flag that asks for machine-readable output.
const jsonFlag = "json"

// taskJSON is how a task crosses to a machine. The field names are the wire,
// so they are written here once rather than derived from the entity — an
// entity that knew its wire format would drag a protocol into every test.
type taskJSON struct {
	Number      int           `json:"number"`
	Parent      int           `json:"parent"`
	Title       string        `json:"title"`
	State       string        `json:"state"`
	Priority    string        `json:"priority"`
	Description string        `json:"description"`
	Labels      []string      `json:"labels"`
	Deadline    string        `json:"deadline"`
	Estimate    string        `json:"estimate"`
	TimeSpent   string        `json:"time_spent"`
	Objective   int           `json:"objective"`
	Progress    *progressJSON `json:"progress"`
}

// progressJSON is how far along something is, for a machine. It is a pointer
// on the way out and null when nothing is under it: zeroes would say none of
// the work is done, where the truth is that there is no work.
//
// The share is published rather than left to be divided, so a client and a
// person reading the same objective never disagree about the rounding.
type progressJSON struct {
	Tasks    *tallyJSON `json:"tasks,omitempty"`
	SubTasks *tallyJSON `json:"sub_tasks,omitempty"`
	Percent  int        `json:"percent"`
}

// tallyJSON is how many of one kind of thing there are, and how many are done.
type tallyJSON struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

// forWireProgress turns a progress into what crosses, or nil when there is
// nothing under it. A tally of nothing is left out entirely rather than
// written as zeroes: a task has no tasks under it, and saying "0 of 0 tasks"
// invents a kind of thing it cannot have.
func forWireProgress(p domain.Progress) *progressJSON {
	if !p.Any() {
		return nil
	}
	out := progressJSON{Percent: p.Percent()}
	if p.Tasks.Total > 0 {
		out.Tasks = &tallyJSON{Total: p.Tasks.Total, Done: p.Tasks.Done}
	}
	if p.SubTasks.Total > 0 {
		out.SubTasks = &tallyJSON{Total: p.SubTasks.Total, Done: p.SubTasks.Done}
	}
	return &out
}

// objectiveJSON is how an objective crosses to a machine. tasks holds what was
// put directly under it, which is why it can hold three numbers while progress
// counts seven things: the sub-tasks under those tasks are counted, not listed.
type objectiveJSON struct {
	Number      int           `json:"number"`
	Title       string        `json:"title"`
	State       string        `json:"state"`
	Priority    string        `json:"priority"`
	Description string        `json:"description"`
	Labels      []string      `json:"labels"`
	Deadline    string        `json:"deadline"`
	Estimate    string        `json:"estimate"`
	Tasks       []int         `json:"tasks"`
	Progress    *progressJSON `json:"progress"`
}

func forWire(t domain.Task, p domain.Progress) taskJSON {
	labels := t.Labels
	if labels == nil {
		labels = []string{}
	}
	return taskJSON{
		Number: t.Number, Parent: t.Parent, Title: t.Title,
		State: string(t.State), Priority: string(t.Priority),
		Description: t.Description, Labels: labels,
		Deadline:  domain.FormatDay(t.Deadline),
		Estimate:  domain.FormatDuration(t.Estimate),
		TimeSpent: domain.FormatDuration(t.TimeSpent),
		Objective: t.Objective,
		Progress:  forWireProgress(p),
	}
}

// forWireObjective turns an objective into what crosses.
func forWireObjective(o domain.Objective, under []domain.Task, p domain.Progress) objectiveJSON {
	labels := o.Labels
	if labels == nil {
		labels = []string{}
	}
	numbers := make([]int, 0, len(under))
	for _, t := range under {
		numbers = append(numbers, t.Number)
	}
	return objectiveJSON{
		Number: o.Number, Title: o.Title, State: string(o.State),
		Priority: string(o.Priority), Description: o.Description, Labels: labels,
		Deadline: domain.FormatDay(o.Deadline),
		Estimate: domain.FormatDuration(o.Estimate),
		Tasks:    numbers,
		Progress: forWireProgress(p),
	}
}

// wantsJSON says whether the caller asked to be answered as a machine.
func wantsJSON(cmd *cobra.Command) bool {
	asked, err := cmd.Root().PersistentFlags().GetBool(jsonFlag)
	return err == nil && asked
}

// write answers with one task, as text or as JSON.
func write(cmd *cobra.Command, task domain.Task, p domain.Progress, text string) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, forWire(task, p))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}

// writeList answers with every task, saying so when there is none.
func writeList(cmd *cobra.Command, tasks []domain.Task) error {
	if wantsJSON(cmd) {
		wire := make([]taskJSON, 0, len(tasks))
		for _, t := range tasks {
			wire = append(wire, forWire(t, domain.Progress{}))
		}
		return writeJSON(cmd, wire)
	}
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "there is no task")
		return err
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), tree(tasks))
	return err
}

// writeDeleted confirms a deletion.
func writeDeleted(cmd *cobra.Command, n int) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{"deleted": n})
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "task %d was deleted\n", n)
	return err
}

// writeObjective answers with one objective, as text or as JSON.
func writeObjective(cmd *cobra.Command, o domain.Objective, under []domain.Task, p domain.Progress, text string) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, forWireObjective(o, under, p))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}

// writeObjectives answers with every objective, saying so when there is none.
// No progress is worked out here: a listing that counted every task under
// every objective would ask the store once per task to print one line each.
func writeObjectives(cmd *cobra.Command, all []domain.Objective) error {
	if wantsJSON(cmd) {
		wire := make([]objectiveJSON, 0, len(all))
		for _, o := range all {
			wire = append(wire, forWireObjective(o, nil, domain.Progress{}))
		}
		return writeJSON(cmd, wire)
	}
	if len(all) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "there is no objective")
		return err
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), objectiveList(all))
	return err
}

// writeObjectiveDeleted confirms an objective was removed.
func writeObjectiveDeleted(cmd *cobra.Command, n int) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{"deleted": n})
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "objective %d was deleted\n", n)
	return err
}

// backupJSON is how a backup crosses to a machine.
type backupJSON struct {
	Number int       `json:"number"`
	Made   string    `json:"made"`
	Holds  *heldJSON `json:"holds,omitempty"`
}

// heldJSON is what a store or a backup contains. An object rather than one
// number, because tasks and objectives are counted apart: a single total would
// hide which of the two somebody is about to overwrite.
type heldJSON struct {
	Tasks      int `json:"tasks"`
	Objectives int `json:"objectives"`
}

// changeJSON is how a change to the store crosses to a machine.
type changeJSON struct {
	Number  int    `json:"number"`
	Name    string `json:"name"`
	Applied bool   `json:"applied"`
	Made    string `json:"made,omitempty"`
}

// madeAt writes a moment the way a person reads one.
func madeAt(t time.Time) string { return t.Local().Format("2006-01-02 15:04") }

// writeBackup answers with one backup. holds is written only when it was counted:
// making a backup does not open it, and a zero nobody measured would read as an
// empty store.
func writeBackup(cmd *cobra.Command, one domain.Backup, holds domain.Held) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, backupJSON{
			Number: one.Number, Made: madeAt(one.Made),
			Holds: &heldJSON{Tasks: holds.Tasks, Objectives: holds.Objectives},
		})
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "#%d  made %s  %s\n",
		one.Number, madeAt(one.Made), heldReads(holds))
	return err
}

// heldReads writes both counts. The zero is written rather than left out: this
// is an inventory, and a line missing at zero reads as a kind of thing the
// backup could not have held at all.
func heldReads(held domain.Held) string {
	return fmt.Sprintf("holds %d task(s) and %d objective(s)", held.Tasks, held.Objectives)
}

// writeBackups answers with every backup, saying so when there is none.
func writeBackups(cmd *cobra.Command, all []domain.Backup) error {
	if wantsJSON(cmd) {
		wire := make([]backupJSON, 0, len(all))
		for _, one := range all {
			wire = append(wire, backupJSON{Number: one.Number, Made: madeAt(one.Made)})
		}
		return writeJSON(cmd, wire)
	}
	if len(all) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "there is no backup")
		return err
	}
	var out strings.Builder
	for _, one := range all {
		_, _ = fmt.Fprintf(&out, "#%d  made %s\n", one.Number, madeAt(one.Made))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}

// writeChanges answers with every change, applied or waiting. Both are listed:
// a list that only showed what was done could not tell you what to apply.
func writeChanges(cmd *cobra.Command, changes []domain.Migration, held domain.Held) error {
	if wantsJSON(cmd) {
		wire := make([]changeJSON, 0, len(changes))
		for _, c := range changes {
			wire = append(wire, changeJSON{
				Number: c.Number, Name: c.Name, Applied: c.Applied, Made: appliedAt(c),
			})
		}
		return writeJSON(cmd, wire)
	}
	waiting := 0
	var out strings.Builder
	for _, c := range changes {
		if !c.Applied {
			waiting++
		}
		_, _ = fmt.Fprintf(&out, "#%d  %s  %s\n", c.Number, state(c), c.Name)
	}
	_, _ = fmt.Fprintf(&out, "\nthe store %s, %d change(s) waiting\n", heldReads(held), waiting)
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}

// state is how a change reads in a listing.
func state(c domain.Migration) string {
	if c.Applied {
		return "applied " + madeAt(c.Made)
	}
	return "waiting          "
}

// appliedAt is when a change was made, and nothing when it has not been.
func appliedAt(c domain.Migration) string {
	if !c.Applied {
		return ""
	}
	return madeAt(c.Made)
}

// writeChange answers with one change: what it is, what it does, and what
// undoing it does. The SQL is shown because that is what somebody wants to read
// before letting a tool touch their data.
func writeChange(cmd *cobra.Command, c domain.Migration) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{
			"number": c.Number, "name": c.Name, "applied": c.Applied,
			"made": appliedAt(c), "does": c.Does, "undoes": c.Undoes, "loses": c.Loses,
		})
	}
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "#%d  %s  %s\n\n", c.Number, state(c), c.Name)
	_, _ = fmt.Fprintf(&out, "what it does\n%s\n\n", orNothing(strings.TrimSpace(c.Does)))
	_, _ = fmt.Fprintf(&out, "what undoing it does\n%s\n\n", orNothing(strings.TrimSpace(c.Undoes)))
	_, _ = fmt.Fprintf(&out, "what undoing it loses\n%s\n", orNothing(c.Loses))
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}

// writeDone confirms an act on something numbered — a backup deleted, a backup put
// back. Answering in the form the caller asked for is the whole point of the
// flag: a script that gets a sentence where it expected a document has no way
// to tell success from anything else.
func writeDone(cmd *cobra.Command, what string, number int) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{what: number})
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "backup %d was %s\n", number, what)
	return err
}

// writeRolledBack names the change a rollback unmade, and what that cost. The
// cost is said here rather than left to be worked out from the SQL: somebody
// who rolls back to try something is exactly somebody who has not read the
// schema, and the backup set aside is only useful to whoever knows to reach for
// it.
func writeRolledBack(cmd *cobra.Command, one domain.Migration) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{"rolled_back": one.Name, "lost": one.Loses})
	}
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "rolled back %q\n", one.Name)
	if one.Loses != "" {
		_, _ = fmt.Fprintf(&out, "lost: %s — `cairn backup restore --yes` puts it back\n", one.Loses)
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}

// writeNothingWaiting says that there was nothing to apply — which is an
// answer, not a refusal, so it goes to standard output like any other.
func writeNothingWaiting(cmd *cobra.Command) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{"applied": []string{}})
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), "there is nothing waiting to be applied")
	return err
}

// writeApplied names what a change command just made.
func writeApplied(cmd *cobra.Command, made []string) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, map[string]any{"applied": made})
	}
	var out strings.Builder
	for _, name := range made {
		_, _ = fmt.Fprintf(&out, "applied %q\n", name)
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}

// writeJSON writes one JSON document and a newline.
func writeJSON(cmd *cobra.Command, v any) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", encoded)
	return err
}

// writeRefusal answers a failure the same way it was asked for: a machine that
// asked for JSON gets JSON, on the one stream it reads, even when the answer is
// a refusal it cannot otherwise parse.
func writeRefusal(out interface{ Write([]byte) (int, error) }, err error) error {
	encoded, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
	if marshalErr != nil {
		return marshalErr
	}
	_, writeErr := fmt.Fprintf(out, "%s\n", encoded)
	return writeErr
}
