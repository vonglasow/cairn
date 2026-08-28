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
	_, err := fmt.Fprint(cmd.OutOrStdout(), engine.RenderText(op, a))
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
	encoded, marshalErr := json.Marshal(engine.Refusal{Error: err.Error()})
	if marshalErr != nil {
		return marshalErr
	}
	_, writeErr := fmt.Fprintf(out, "%s\n", encoded)
	return writeErr
}
