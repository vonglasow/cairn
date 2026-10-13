package mcp

import (
	"context"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"cairn/internal/engine"
)

func offered(access engine.Level) []engine.Operation {
	var ops []engine.Operation
	for _, op := range engine.All() {
		if op.Level <= access {
			ops = append(ops, op)
		}
	}
	return ops
}

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
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

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

const refusalKey = "cairn/refusal"

const throughTheServer = "mcp"

func handlerFor(op engine.Operation, home engine.Home) sdk.ToolHandler {
	return func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		in, err := op.FromJSON(req.Params.Arguments)
		if err != nil {
			return refused(err), nil
		}
		in.Confirmed = true
		answer, err := engine.DoFrom(op, home, in, throughTheServer)
		if err != nil {
			return refused(err), nil
		}
		return &sdk.CallToolResult{
			Content:           []sdk.Content{&sdk.TextContent{Text: engine.RenderText(op, answer)}},
			StructuredContent: engine.Of(op, answer),
		}, nil
	}
}

func refused(err error) *sdk.CallToolResult {
	return &sdk.CallToolResult{
		Meta:    sdk.Meta{refusalKey: engine.RefusalOf(err)},
		Content: []sdk.Content{&sdk.TextContent{Text: err.Error()}},
		IsError: true,
	}
}
