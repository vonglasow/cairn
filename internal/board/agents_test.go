package board

import (
	"slices"
	"testing"

	"cairn/internal/engine/agent"
)

func TestAgentGroups(t *testing.T) {
	held := func(n int, title string) *agent.Held { return &agent.Held{Number: n, Title: title} }
	listed := []agent.Agent{
		{ID: "a", State: agent.Idle, Folder: "/w/cairn", Project: "cairn", Task: held(2, "Overrun")},
		{ID: "b", State: agent.Working, Folder: "/w/ops", Project: "ops"},
		{ID: "c", State: agent.Waiting, Folder: "/w/cairn/.claude/worktrees/spec", Project: "cairn", Task: held(1, "Spec")},
		{ID: "d", State: agent.Working, Folder: "/w/cairn", Project: "cairn", Task: held(2, "Overrun")},
	}
	heads := func(groups []AgentGroup) []string {
		var said []string
		for _, g := range groups {
			said = append(said, g.Head)
		}
		return said
	}
	ids := func(g AgentGroup) []string {
		var said []string
		for _, a := range g.Agents {
			said = append(said, a.ID)
		}
		return said
	}
	tests := []struct {
		name       string
		by         AgentsBy
		heads      []string
		firstHolds []string
	}{
		{"by state, waiting first, then working, then idle", ByState, []string{"Waiting for you", "Working", "Idle"}, []string{"c"}},
		{"by task, in the order of their numbers, the ones with none last", ByTask, []string{"#1 Spec", "#2 Overrun", "No task"}, []string{"c"}},
		{"by project, its worktrees with it, in order", ByProject, []string{"cairn", "ops"}, []string{"c", "d", "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := AgentGroups(listed, tt.by)
			if got := heads(groups); !slices.Equal(got, tt.heads) {
				t.Fatalf("heads %v, want %v", got, tt.heads)
			}
			if got := ids(groups[0]); !slices.Equal(got, tt.firstHolds) {
				t.Fatalf("first group holds %v, want %v", got, tt.firstHolds)
			}
		})
	}
	t.Run("a state no session is in has no group", func(t *testing.T) {
		if got := heads(AgentGroups(listed[1:2], ByState)); !slices.Equal(got, []string{"Working"}) {
			t.Fatalf("%v", got)
		}
	})
}

func TestAgentOn(t *testing.T) {
	held := func(n int) *agent.Held { return &agent.Held{Number: n} }
	b := Board{AgentsState: AgentsState{Agents: agent.Agents{Agents: []agent.Agent{
		{State: agent.Idle, Task: held(1)}, {State: agent.Waiting, Task: held(1)}, {State: agent.Working, Task: held(2)}, {State: agent.Working},
	}}}}
	for _, tt := range []struct {
		task int
		want agent.State
	}{{1, agent.Waiting}, {2, agent.Working}, {3, ""}} {
		if got := b.AgentOn(tt.task); got != tt.want {
			t.Errorf("task %d: %q, want %q", tt.task, got, tt.want)
		}
	}
}
