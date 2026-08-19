package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cairn/internal/domain"
)

// jsonFlag is the name of the flag that asks for machine-readable output.
const jsonFlag = "json"

// taskJSON is how a task crosses to a machine. The field names are the wire,
// so they are written here once rather than derived from the entity — an
// entity that knew its wire format would drag a protocol into every test.
type taskJSON struct {
	Number      int      `json:"number"`
	Title       string   `json:"title"`
	State       string   `json:"state"`
	Priority    string   `json:"priority"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Deadline    string   `json:"deadline"`
	Estimate    string   `json:"estimate"`
	TimeSpent   string   `json:"time_spent"`
}

func forWire(t domain.Task) taskJSON {
	labels := t.Labels
	if labels == nil {
		labels = []string{}
	}
	return taskJSON{
		Number: t.Number, Title: t.Title,
		State: string(t.State), Priority: string(t.Priority),
		Description: t.Description, Labels: labels,
		Deadline:  domain.FormatDay(t.Deadline),
		Estimate:  domain.FormatDuration(t.Estimate),
		TimeSpent: domain.FormatDuration(t.TimeSpent),
	}
}

// wantsJSON says whether the caller asked to be answered as a machine.
func wantsJSON(cmd *cobra.Command) bool {
	asked, err := cmd.Root().PersistentFlags().GetBool(jsonFlag)
	return err == nil && asked
}

// write answers with one task, as text or as JSON.
func write(cmd *cobra.Command, task domain.Task, text string) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, forWire(task))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}

// writeList answers with every task, saying so when there is none.
func writeList(cmd *cobra.Command, tasks []domain.Task) error {
	if wantsJSON(cmd) {
		wire := make([]taskJSON, 0, len(tasks))
		for _, t := range tasks {
			wire = append(wire, forWire(t))
		}
		return writeJSON(cmd, wire)
	}
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "there is no task")
		return err
	}
	var out strings.Builder
	for _, t := range tasks {
		out.WriteString(line(t))
		out.WriteString("\n")
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
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
