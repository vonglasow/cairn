package board

import (
	"cmp"
	"fmt"
	"slices"

	"cairn/internal/engine"
	"cairn/internal/engine/agent"
)

// AgentsBy is how the agents page groups its sessions.
type AgentsBy int

// The three ways the agents page groups its sessions, the person's choice.
const (
	ByState AgentsBy = iota
	ByTask
	ByProject
)

// AgentsBySaid is what each grouping is called on the page, in its order.
var AgentsBySaid = []string{"State", "Task", "Project"}

// AgentsState is what the agents page shows: the sessions read last, and how they are grouped.
type AgentsState struct {
	Agents   agent.Agents
	AgentsBy AgentsBy
}

// AgentGroup is a heading on the agents page and the sessions under it.
type AgentGroup struct {
	Head   string
	Agents []agent.Agent
}

// NoTaskSays is the heading of the sessions holding no task.
const NoTaskSays = "No task"

var stateSaid = map[agent.State]string{agent.Waiting: "Waiting for you", agent.Working: "Working", agent.Idle: "Idle"}

// AgentGroups groups the sessions the way chosen: by state, waiting first; by task, in the order of the numbers, the ones with none last; by project, the last name of the folder.
func AgentGroups(listed []agent.Agent, by AgentsBy) []AgentGroup {
	type keyed struct {
		order int
		head  string
	}
	key := func(a agent.Agent) keyed {
		switch by {
		case ByTask:
			if a.Task == nil {
				return keyed{int(^uint(0) >> 1), NoTaskSays}
			}
			return keyed{a.Task.Number, fmt.Sprintf("#%d %s", a.Task.Number, a.Task.Title)}
		case ByProject:
			return keyed{0, a.Project}
		}
		return keyed{slices.Index([]agent.State{agent.Waiting, agent.Working, agent.Idle}, a.State), stateSaid[a.State]}
	}
	var keys []keyed
	held := map[keyed][]agent.Agent{}
	for _, a := range listed {
		k := key(a)
		if _, ok := held[k]; !ok {
			keys = append(keys, k)
		}
		held[k] = append(held[k], a)
	}
	slices.SortStableFunc(keys, func(x, y keyed) int { return cmp.Or(cmp.Compare(x.order, y.order), cmp.Compare(x.head, y.head)) })
	groups := make([]AgentGroup, 0, len(keys))
	for _, k := range keys {
		in := held[k]
		if by != ByState {
			slices.SortStableFunc(in, func(x, y agent.Agent) int { return cmp.Compare(rank(x.State), rank(y.State)) })
		}
		groups = append(groups, AgentGroup{Head: k.head, Agents: in})
	}
	return groups
}

func rank(s agent.State) int {
	return slices.Index([]agent.State{agent.Waiting, agent.Working, agent.Idle}, s)
}

// ReadAgents asks the engine for the Claude sessions running in the workspace in use, saying the window listens for the permissions they ask.
func ReadAgents(w Work) (agent.Agents, error) {
	answer, err := w.Ask(engine.AgentList, engine.Input{Given: map[string]string{"listening": "true"}})
	if err != nil || answer.Agents == nil {
		return agent.Agents{}, err
	}
	return *answer.Agents, nil
}

// TabAlert is the count a page's tab shows in red: the sessions waiting for the person, on the agents tab while the configuration counts them, and nothing elsewhere.
func (b Board) TabAlert(page Tab) int {
	if page != Agents || !b.Pages.Counts() {
		return 0
	}
	return b.Agents.Waiting
}

// AgentOn is what the sessions holding a task do, the one that most needs the person: waiting before working before idle; nothing when none holds it.
func (b Board) AgentOn(task int) agent.State {
	var on agent.State
	for _, a := range b.Agents.Agents {
		if a.Task != nil && a.Task.Number == task && (on == "" || rank(a.State) < rank(on)) {
			on = a.State
		}
	}
	return on
}
