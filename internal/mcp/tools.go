package mcp

// What the server offers is what it was allowed to do: a tool per operation
// engine.All() names, left out entirely once its Level sits above s.Access —
// an agent cannot ask for what it was never shown.

import (
	"context"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

// offered is every operation an agent at access may reach: those whose Level
// sits at or below it, in engine.All()'s order.
func offered(access engine.Level) []engine.Operation {
	var ops []engine.Operation
	for _, op := range engine.All() {
		if op.Level <= access {
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
		Name:         op.Name(),
		Description:  op.Does,
		InputSchema:  schemaFor(op),
		OutputSchema: answerSchemaFor(op),
		Annotations: &sdk.ToolAnnotations{
			ReadOnlyHint:    op.Level == engine.Read,
			DestructiveHint: &destructive,
			OpenWorldHint:   &openWorld,
		},
	}
}

// answerSchemaFor describes the shape an operation's answer takes, from what
// the operation declared it answers.
//
// Derived from the Go type rather than written out, so the description and the
// thing described cannot drift: the same struct that crosses the wire is the one
// the schema is read off.
//
// Every answer is one object, so every schema describes one object. That is not
// a convenience — a client validating what comes back needs somewhere to put it,
// and one that will not read "either of these" reads none of the tools at all,
// since a single tool it refuses costs the whole list. Nothing here enforces
// that: it is enforced where it is decided, in what each operation declares.
//
// A shape that cannot be read off its type is a programming mistake rather than
// something a caller did, so it comes back as no schema at all: a tool that
// cannot describe itself is better than a server that will not start.
func answerSchemaFor(op engine.Operation) any {
	if op.Answers == nil {
		return nil
	}
	shape, err := jsonschema.ForType(reflect.TypeOf(op.Answers), nil)
	if err != nil {
		return nil
	}
	return shape
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
	case engine.Bool:
		return "boolean"
	default:
		return "string"
	}
}

// throughTheServer is how a history names what an agent did. The same word as
// the command that starts it, so a line nobody explained still reads.
const throughTheServer = "mcp"

// handlerFor wires op to engine.DoFrom as thinly as a tool can be. A refusal —
// an empty title, a store this cairn cannot use — comes back as IsError,
// with the reason as text; only what cannot be read as the arguments the
// operation takes is a protocol error, since that never reached cairn at all.
//
// An answer fills both content and structuredContent, byte for byte what the
// command line would have printed as text and as `--json`: the same claim
// docs/mcp.md makes about an agent and a script never seeing two different
// cairns, and it costs nothing because both renderings already exist in
// internal/engine.
func handlerFor(op engine.Operation, home engine.Home) sdk.ToolHandler {
	return func(_ context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		// What the call sent is read the way the operation takes it, by the
		// engine: what cannot be read as what an argument carries never
		// reached cairn, so it is a protocol error rather than a refusal.
		in, err := op.FromJSON(req.Params.Arguments)
		if err != nil {
			return nil, err
		}
		// Named as the server's doing, so a history shows what an agent
		// changed as something the person reading it did not do themselves.
		answer, err := engine.DoFrom(op, home, in, throughTheServer)
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
			StructuredContent: engine.Of(op, answer),
		}, nil
	}
}
