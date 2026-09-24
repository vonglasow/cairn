package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/engine"
)

// jsonFlag is the name of the flag that asks for machine-readable output.
const jsonFlag = "json"

// wantsJSON says whether the caller asked to be answered as a machine.
func wantsJSON(cmd *cobra.Command) bool {
	asked, err := cmd.Root().PersistentFlags().GetBool(jsonFlag)
	return err == nil && asked
}

// write answers with what an operation answered, as text or as JSON.
func write(cmd *cobra.Command, op engine.Operation, a engine.Answer) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, engine.Of(a))
	}
	if a.Config != nil {
		_, err := fmt.Fprint(cmd.OutOrStdout(), renderConfig(*a.Config))
		return err
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), engine.RenderText(op, a))
	return err
}

// renderConfig lays out what `config show` answers: the file cairn read, then
// each of its two places with what named it — spelled out in full, since a
// person reading this is exactly someone who does not want to guess.
func renderConfig(c engine.Config) string {
	// A file cairn could not make sense of gets no emplacements printed: in
	// production this same file turns every other command away, so showing
	// "store → somewhere" would say the opposite of what is true. It gets the
	// file, the problem, and the warning that every other command carries —
	// nothing else.
	if c.Problem != "" {
		return fmt.Sprintf(
			"file    %s\n\nthe file cannot be made sense of: %s\nthe other commands will refuse until it is\n",
			c.File, c.Problem,
		)
	}
	// The two places sit in a column of their own, padded to the longer of
	// them: what a person compares here is the pair "place, and what named
	// it", and a ragged right edge makes them read the second column line by
	// line instead of at a glance. Padded to the longer rather than to a
	// fixed width, since a path has no length anyone can plan for.
	// max rather than a comparison written out. The two say the same thing,
	// and the written-out one generates a boundary mutant — > against >= —
	// that no test can kill: the two branches differ only when the lengths are
	// equal, and then both give the same width. An equivalent mutant is not
	// chased, it is designed out.
	wide := max(len(c.Store.At), len(c.Notes.At))
	return fmt.Sprintf(
		"file    %s\n\nstore   %-*s   named by %s\nnotes   %-*s   named by %s\n",
		c.File, wide, c.Store.At, c.Store.By, wide, c.Notes.At, c.Notes.By,
	)
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
	encoded, marshalErr := json.Marshal(engine.Refusal{Error: err.Error()})
	if marshalErr != nil {
		return marshalErr
	}
	_, writeErr := fmt.Fprintf(out, "%s\n", encoded)
	return writeErr
}
