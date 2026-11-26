package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// The answers a permission takes from the window.
const (
	Allow         = "allow"
	Deny          = "deny"
	AlwaysProject = "always-project"
	AlwaysUser    = "always-user"
)

// Answers is every answer a permission takes from the window.
var Answers = []string{Allow, Deny, AlwaysProject, AlwaysUser}

// Request is the PermissionRequest Claude Code hands its hook.
type Request struct {
	Session     string          `json:"session_id"`
	Tool        string          `json:"tool_name"`
	Input       json.RawMessage `json:"tool_input"`
	Suggestions json.RawMessage `json:"permission_suggestions"`
}

// Pending is what waits on the window, as cairn keeps it: the tool, what it acts on, the suggestions Claude made, kept as they came, and the questions it asks.
type Pending struct {
	Tool        string          `json:"tool_name"`
	On          string          `json:"on"`
	Suggestions json.RawMessage `json:"permission_suggestions,omitempty"`
	Questions   []Question      `json:"questions,omitempty"`
}

// PendingOf is a request as cairn keeps it while it waits.
func PendingOf(r Request) Pending {
	if r.Tool == AskUserQuestion {
		var asked struct{ Questions []Question }
		_ = json.Unmarshal(r.Input, &asked)
		return Pending{Tool: r.Tool, Questions: asked.Questions}
	}
	return Pending{Tool: r.Tool, On: onOf(r.Input), Suggestions: r.Suggestions}
}

// AskOf is a pending permission as a card shows it.
func AskOf(p Pending) *Ask {
	var suggested []json.RawMessage
	_ = json.Unmarshal(p.Suggestions, &suggested)
	return &Ask{Tool: p.Tool, On: p.On, Always: len(suggested) > 0, Questions: p.Questions}
}

func onOf(input json.RawMessage) string {
	var named struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
	}
	if json.Unmarshal(input, &named) == nil && named.Command+named.FilePath != "" {
		return firstLine(named.Command + named.FilePath)
	}
	d := json.NewDecoder(bytes.NewReader(input))
	if open, err := d.Token(); err != nil || open != json.Delim('{') {
		return ""
	}
	for d.More() {
		var value any
		if _, err := d.Token(); err != nil || d.Decode(&value) != nil {
			return ""
		}
		if said, ok := value.(string); ok {
			return firstLine(said)
		}
	}
	return ""
}

func firstLine(said string) string {
	line, _, _ := strings.Cut(said, "\n")
	return line
}

// HookSaying is what the hook prints for Claude Code to read an answer from: for a permission, allow, deny, or allow with the suggested rules written to the project's or the user's settings; for questions, the questions given back with their answers.
func HookSaying(answer string, r Request) (json.RawMessage, error) {
	if r.Tool == AskUserQuestion {
		return answering(answer, r.Input)
	}
	suggestions := r.Suggestions
	decision := map[string]any{"behavior": "allow"}
	switch answer {
	case Allow:
	case Deny:
		decision = map[string]any{"behavior": "deny", "message": "Denied from the cairn window"}
	case AlwaysProject, AlwaysUser:
		var rules []map[string]any
		if len(suggestions) > 0 {
			if err := json.Unmarshal(suggestions, &rules); err != nil {
				return nil, err
			}
		}
		for _, rule := range rules {
			rule["destination"] = map[string]string{AlwaysProject: "projectSettings", AlwaysUser: "userSettings"}[answer]
		}
		if len(rules) > 0 {
			decision["updatedPermissions"] = rules
		}
	default:
		return nil, fmt.Errorf("%q is not an answer: %s", answer, strings.Join(Answers, ", "))
	}
	return json.Marshal(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PermissionRequest", "decision": decision}})
}

// Decision is what a permission was answered: the session, the answer, and what the hook prints for Claude Code, nothing when the window did not answer.
type Decision struct {
	Session string          `json:"session,omitempty"`
	Answer  string          `json:"answer,omitempty"`
	Hook    json.RawMessage `json:"hook,omitempty"`
}

// DecisionText is a decision as it is printed: the hook's JSON for Claude Code to read, nothing when there is none, or the answer given.
func DecisionText(d Decision) string {
	switch {
	case d.Hook != nil:
		return string(d.Hook) + "\n"
	case d.Answer != "":
		return fmt.Sprintf("%s answered %s\n", ShortOf(d.Session), d.Answer)
	}
	return ""
}

// StartedText is a session started, as a person reads it.
func StartedText(s Started) string {
	return fmt.Sprintf("%s runs %s in %s\n", s.Terminal, s.Command, s.Folder)
}

// TerminalsText is the terminals installed, one a line.
func TerminalsText(t Terminals) string {
	if len(t.Terminals) == 0 {
		return "no terminal cairn starts a session in is installed\n"
	}
	return strings.Join(t.Terminals, "\n") + "\n"
}

func answering(answer string, input json.RawMessage) (json.RawMessage, error) {
	var answers map[string]string
	if err := json.Unmarshal([]byte(answer), &answers); err != nil {
		return nil, fmt.Errorf("%q answers no question: %w", answer, err)
	}
	updated := map[string]any{}
	if err := json.Unmarshal(input, &updated); err != nil {
		return nil, err
	}
	updated["answers"] = answers
	return json.Marshal(map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": "PreToolUse", "permissionDecision": "allow", "updatedInput": updated}})
}

// Answered is questions answered by the choices given, one to each question in its order, the labels of a question taking several joined by commas, as the hook hands them to Claude Code.
func Answered(questions []Question, chosen []string) (string, error) {
	if len(chosen) != len(questions) {
		return "", fmt.Errorf("one choice for each of the %d questions, %d given", len(questions), len(chosen))
	}
	answers := map[string]string{}
	for i, q := range questions {
		labels := []string{chosen[i]}
		if q.Multi {
			labels = strings.Split(chosen[i], ",")
		}
		for j, label := range labels {
			labels[j] = strings.TrimSpace(label)
			if !slices.ContainsFunc(q.Options, func(o Option) bool { return o.Label == labels[j] }) {
				return "", fmt.Errorf("%q is not an option of %q", labels[j], q.Question)
			}
		}
		answers[q.Question] = strings.Join(labels, ", ")
	}
	said, err := json.Marshal(answers)
	return string(said), err
}
