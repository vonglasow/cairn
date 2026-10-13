package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/engine"
)

const jsonFlag = "json"

func wantsJSON(cmd *cobra.Command) bool {
	asked, err := cmd.Root().PersistentFlags().GetBool(jsonFlag)
	return err == nil && asked
}

func write(cmd *cobra.Command, op engine.Operation, a engine.Answer) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, engine.Of(op, a))
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), engine.RenderText(op, a))
	return err
}

func writeJSON(cmd *cobra.Command, v any) error {
	encoded, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("writing JSON: %w", err)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", encoded)
	return err
}

func writeRefusal(out interface{ Write([]byte) (int, error) }, err error) error {
	encoded, marshalErr := json.Marshal(engine.RefusalOf(err))
	if marshalErr != nil {
		return marshalErr
	}
	_, writeErr := fmt.Fprintf(out, "%s\n", encoded)
	return writeErr
}
