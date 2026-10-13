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
// looks; destructiveHint says it can lose work already done or replace what
// was written — the protocol's sense of the word, wider than the access
// level's; idempotentHint says a second call does nothing more than the
// first; openWorldHint is always false, because cairn is a store on this
// machine and goes looking nowhere else — the protocol assumes the opposite
// until told otherwise.
func toolFor(op engine.Operation) *sdk.Tool {
	destructive := op.Level == engine.Full || op.Overwrites
	openWorld := false
	return &sdk.Tool{
		Name:         op.Name(),
		Description:  op.Does,
		InputSchema:  schemaFor(op),
		OutputSchema: answerSchemaFor(op),
		Annotations: &sdk.ToolAnnotations{
			Title:           op.Short,
			ReadOnlyHint:    op.Level == engine.Read,
			DestructiveHint: &destructive,
			IdempotentHint:  op.Level == engine.Read || op.Again == engine.ChangesNothingMore,
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
		if arg.Kind == engine.List {
			prop["items"] = map[string]any{"type": "string"}
		}
		properties[arg.Name] = prop
		if arg.Required {
			required = append(required, arg.Name)
		}
	}
	// A name the operation does not take is refused by the engine either
	// way; said here, a client knows before it calls.
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
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

// refusalKey is where a refusal's kind and reason travel, under _meta. The
// protocol reserves unprefixed names for itself, so the key carries cairn's.
const refusalKey = "cairn/refusal"

// throughTheServer is how a history names what an agent did. The same word as
// the command that starts it, so a line nobody explained still reads.
const throughTheServer = "mcp"

// handlerFor wires op to engine.DoFrom as thinly as a tool can be. A refusal —
// an empty title, a store this cairn cannot use, arguments that cannot be read
// as what the operation takes — comes back as IsError, with the reason as text
// and its kind under _meta. Only a tool that was never offered fails as a call.
//
// An answer fills both content and structuredContent, byte for byte what the
// command line would have printed as text and as `--json`: the same claim
// docs/mcp.md makes about an agent and a script never seeing two different
// cairns, and it costs nothing because both renderings already exist in
// internal/engine.
func handlerFor(op engine.Operation, home engine.Home) sdk.ToolHandler {
	return func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		// A call the client gave up on before it ran is not run. One that
		// has started runs to its end: it is one transaction of a few
		// milliseconds, and stopping it half way is what the transaction is
		// there to prevent.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// What the call sent is read the way the operation takes it, by the
		// engine. What cannot be read is refused like a name the operation
		// does not take: both are a request not understood, and an agent
		// handed two shapes for one kind of mistake could branch on only one.
		in, err := op.FromJSON(req.Params.Arguments)
		if err != nil {
			return refused(err), nil
		}
		// The access the server was started with is how an agent was asked:
		// a tool it was given is one it may use.
		in.Confirmed = true
		// Named as the server's doing, so a history shows what an agent
		// changed as something the person reading it did not do themselves.
		answer, err := engine.DoFrom(op, home, in, throughTheServer)
		if err != nil {
			// Something cairn understood and turned down — an empty title, a
			// store this cairn cannot use — is an answer, not a broken tool:
			// it comes back as IsError, with the reason as text an agent can
			// read and correct itself from, rather than a protocol error.
			//
			// The kind and the reason travel beside the sentence, under
			// _meta, so a program branches on words that do not change when
			// the sentence is reworded. Not as structuredContent: the client
			// measures that against the answer's shape even on a refusal,
			// and a refusal is not an answer of that shape (docs/mcp.md).
			return refused(err), nil
		}
		return &sdk.CallToolResult{
			Content:           []sdk.Content{&sdk.TextContent{Text: engine.RenderText(op, answer)}},
			StructuredContent: engine.Of(op, answer),
		}, nil
	}
}

// refused is how a refusal is answered: IsError, the sentence as text, and
// its kind and reason under _meta.
func refused(err error) *sdk.CallToolResult {
	return &sdk.CallToolResult{
		Meta:    sdk.Meta{refusalKey: engine.RefusalOf(err)},
		Content: []sdk.Content{&sdk.TextContent{Text: err.Error()}},
		IsError: true,
	}
}
