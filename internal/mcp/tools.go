package mcp

// What the server offers is what it was allowed to do: a tool per operation
// engine.All() names, left out entirely once its Level sits above s.Access —
// an agent cannot ask for what it was never shown.

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

// offered is every operation an agent at access may reach: those whose Level
// sits at or below it, in engine.All()'s order.
func offered(access Access) []engine.Operation {
	var ops []engine.Operation
	for _, op := range engine.All() {
		if op.Level <= engine.Level(access) {
			ops = append(ops, op)
		}
	}
	return ops
}

// toolFor describes what calling op would cost. readOnlyHint says it only
// looks; destructiveHint says it can lose work already done; openWorldHint
// is always false, because cairn is a store on this machine and goes looking
// nowhere else — the protocol assumes the opposite until told otherwise.
func toolFor(op engine.Operation) *sdk.Tool {
	destructive := op.Level == engine.Full
	openWorld := false
	return &sdk.Tool{
		Name:        op.Name(),
		Description: op.Does,
		InputSchema: schemaFor(op),
		Annotations: &sdk.ToolAnnotations{
			ReadOnlyHint:    op.Level == engine.Read,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		},
	}
}

// schemaFor is the JSON schema op's arguments make: each argument's name, its
// type, whether it is required, its OneOf as the schema's enum when there is
// a closed set, and its Means as the schema's description — the same
// sentence cobra's flag help reads, so the two readers cannot drift apart.
func schemaFor(op engine.Operation) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, arg := range op.Takes {
		prop := map[string]any{"type": jsonType(arg.Kind), "description": arg.Means}
		if len(arg.OneOf) > 0 {
			prop["enum"] = arg.OneOf
		}
		properties[arg.Name] = prop
		if arg.Required {
			required = append(required, arg.Name)
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// jsonType is the JSON schema type an argument's Kind reads as.
func jsonType(kind engine.Kind) string {
	switch kind {
	case engine.Number:
		return "integer"
	case engine.List:
		return "array"
	default:
		return "string"
	}
}

// handlerFor wires op to engine.Do as thinly as a tool can be. A refusal —
// an empty title, a store this cairn cannot use — comes back as IsError,
// with the reason as text; only what json.Unmarshal itself cannot read is a
// protocol error, since that never reached engine.Do at all.
//
// An answer fills both content and structuredContent, byte for byte what the
// command line would have printed as text and as `--json`: the same claim
// docs/mcp.md makes about an agent and a script never seeing two different
// cairns, and it costs nothing because both renderings already exist in
// internal/engine.
func handlerFor(op engine.Operation, path string) sdk.ToolHandler {
	return func(_ context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		given := map[string]any{}
		if len(req.Params.Arguments) > 0 {
			if err := json.Unmarshal(req.Params.Arguments, &given); err != nil {
				return nil, err
			}
		}
		in := engine.Input{Given: map[string]string{}}
		for name, value := range given {
			// JSON null is how a field is named as nothing, which is not the
			// same request as leaving it out: engine.Input models that as a
			// name present and holding the empty string, never fmt.Sprint's
			// "<nil>".
			if value == nil {
				in.Given[name] = ""
				continue
			}
			in.Given[name] = fmt.Sprint(value)
		}
		answer, err := engine.Do(op, path, in)
		if err != nil {
			// Something cairn understood and turned down — an empty title, a
			// store this cairn cannot use — is an answer, not a broken tool:
			// it comes back as IsError, with the reason as text an agent can
			// read and correct itself from, rather than a protocol error.
			return &sdk.CallToolResult{
				Content: []sdk.Content{&sdk.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil
		}
		return &sdk.CallToolResult{
			Content:           []sdk.Content{&sdk.TextContent{Text: engine.RenderText(op, answer)}},
			StructuredContent: engine.Of(answer),
		}, nil
	}
}
