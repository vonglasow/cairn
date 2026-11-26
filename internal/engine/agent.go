package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"time"

	"cairn/internal/engine/agent"
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"cairn/internal/store/files"
)

// ClaudeVariable names the folder Claude Code keeps its sessions in, ~/.claude when it is not set.
const ClaudeVariable = files.ClaudeVariable

var machine agent.Machine = processes{}

// AgentList lists the Claude sessions running in the workspace in use, for the agents page.
var AgentList = Operation{
	Thing:   "agent",
	Verb:    "list",
	Short:   "List the Claude sessions running in this workspace",
	Does:    "List the Claude sessions running on this machine that belong to the workspace in use: the ones holding one of its tasks, and the ones holding none whose folder answers from it. Each says what it is doing — working, waiting for the person, or idle — the task it holds, its branch, model and context, and the terminal it runs in. Read from what Claude Code keeps in its own folder.",
	Example: "cairn agent list",
	Level:   Read,
	Needs:   aStore,
	Answers: agent.Agents{},
	Takes:   []Argument{{Name: "listening", Kind: Bool, Means: "say a window listens for the permissions sessions ask, for the next 10 seconds"}},
	Run: func(on workedOn, in Input) (Answer, error) {
		asks := asksFolder(on.Home)
		if asks != "" && in.yes("listening") {
			if err := listened(listeningAt(asks, on.Path)); err != nil {
				return Answer{}, err
			}
		}
		if err := swept(asks, claudeFolder(on.Home)); err != nil {
			return Answer{}, err
		}
		read, err := configuredAt(on.Home, on.Path)
		if err != nil {
			return Answer{}, err
		}
		silent, err := time.ParseDuration(read.Agents().Silent)
		if err != nil {
			return Answer{}, fmt.Errorf("agents.silent %q: %w", read.Agents().Silent, files.ErrSilent)
		}
		listed, err := on.Engine.agents(on.Home, on.Notes, on.Path)
		if err != nil {
			return Answer{}, err
		}
		for i, a := range listed {
			if a.State == agent.Working && ticking.now().Sub(a.Silent) < silent {
				listed[i].Silent = time.Time{}
			}
		}
		if err := on.Engine.suggesting(listed); err != nil {
			return Answer{}, err
		}
		for i, a := range listed {
			if pending, ok := pendingOf(asks, a.ID); ok {
				listed[i].Asks = agent.AskOf(pending)
			}
		}
		counted := agent.Counted(listed)
		return Answer{Agents: &counted}, nil
	},
}

func claudeFolder(h Home) string {
	if said, ok := h.lookup()(ClaudeVariable); ok && said != "" {
		return said
	}
	return filepath.Join(h.Dir, ".claude")
}

func running(h Home) []files.ClaudeSession {
	alive := []files.ClaudeSession{}
	for _, s := range files.ClaudeSessions(claudeFolder(h)) {
		if machine.Alive(s.PID, s.ProcStart) {
			alive = append(alive, s)
		}
	}
	return alive
}

func sessionAgent(s files.ClaudeSession) agent.Agent {
	return agent.Agent{
		ID: s.SessionID, Short: agent.ShortOf(s.SessionID), Name: s.Name,
		State: agent.StateOf(s.Status, s.WaitingFor), Waiting: s.WaitingFor,
		Since: time.UnixMilli(s.StatusUpdatedAt), Folder: s.Cwd, Project: agent.ProjectOf(s.Cwd), Terminal: terminalOf(s), Resume: agent.ResumeOf(s.Cwd, s.SessionID),
	}
}

func terminalOf(s files.ClaudeSession) agent.Terminal {
	if program := agent.ProgramOf(s.Entrypoint); program != "" {
		return agent.Terminal{Program: program, PID: s.PID}
	}
	t := machine.Terminal(s.PID)
	t.PID = s.PID
	return t
}

func runningNamed(h Home, id string) (agent.Agent, error) {
	for _, s := range running(h) {
		if s.SessionID == id || agent.ShortOf(s.SessionID) == id {
			return sessionAgent(s), nil
		}
	}
	return agent.Agent{}, Refusing("no-such-session", fmt.Sprintf("%s: no Claude session by that id runs on this machine", id))
}

func (e *Engine) agents(h Home, notes, store string) ([]agent.Agent, error) {
	folder := claudeFolder(h)
	listed := []agent.Agent{}
	for _, s := range running(h) {
		a := sessionAgent(s)
		if n := files.SessionHeld(notes, s.SessionID); n != 0 {
			held, err := e.heldBy(n)
			if err != nil {
				return nil, err
			}
			a.Task = held
		}
		if a.Task == nil && !answersFrom(h, s.Cwd, store) {
			continue
		}
		said := files.ClaudeTranscript(folder, s.SessionID)
		a.Branch, a.Model, a.Tokens, a.Context = said.Branch, said.Model, said.Tokens, agent.ContextOf(said.Tokens)
		if a.State == agent.Working {
			a.Silent = time.UnixMilli(said.WrittenAt)
		}
		a.Doing = agent.DoingOf(said.Tool, said.On, said.Description)
		listed = append(listed, a)
	}
	return listed, nil
}

func (e *Engine) heldBy(n int) (*agent.Held, error) {
	t, err := e.show(n)
	if errors.Is(err, task.ErrNoSuchTask) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	shape := task.OfTask(t)
	held := &agent.Held{Number: n, Title: shape.Title}
	for _, step := range shape.Steps {
		if !step.Done {
			held.Open = append(held.Open, step.Title)
		}
	}
	return held, nil
}

func answersFrom(h Home, folder, store string) bool {
	if h.Dir == "" {
		return true
	}
	found, err := Locate(Home{Dir: h.Dir, Here: folder, Version: h.Version, Lookup: h.Lookup})
	return err == nil && found.Store == store
}

var wholeNumber = regexp.MustCompile(`\d+`)

func (e *Engine) suggesting(listed []agent.Agent) error {
	held, err := e.focus()
	if err != nil {
		return err
	}
	current := filter.Narrowing{}
	current.InColumn(string(item.Current))
	tasks, err := e.narrowed(&current)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if n := t.Carried().Number; !slices.Contains(held, n) {
			held = append(held, n)
		}
	}
	free := slices.DeleteFunc(held, func(n int) bool {
		return slices.ContainsFunc(listed, func(a agent.Agent) bool { return a.Task != nil && a.Task.Number == n })
	})
	for i, a := range listed {
		if a.Task != nil {
			continue
		}
		named := slices.IndexFunc(free, func(n int) bool { return slices.Contains(wholeNumber.FindAllString(a.Branch, -1), strconv.Itoa(n)) })
		switch {
		case len(free) == 1:
			named = 0
		case len(free) == 0 || named < 0:
			continue
		}
		if listed[i].Suggested, err = e.heldBy(free[named]); err != nil {
			return err
		}
	}
	return nil
}
