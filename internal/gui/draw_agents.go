package gui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/board"
	"cairn/internal/engine/agent"
	"cairn/internal/gui/geom"
)

// NoAgentsSays is what the agents page says while no Claude session runs in the workspace.
const NoAgentsSays = "No Claude session is running in this workspace."

var standingOfAgent = map[agent.State]string{agent.Waiting: "on-hold", agent.Working: "reached", agent.Idle: "dropped"}

var agentStateSaid = map[agent.State]string{agent.Waiting: "Waiting", agent.Working: "Working", agent.Idle: "Idle"}

func drawAgents(k Kit, w Window) {
	groups := w.AgentsGrouped()
	laid := w.AgentsLaid()
	var segments []geom.Box
	segments = append(segments, laid.By...)
	Segmented{Said: board.AgentsBySaid, At: segments, On: int(w.AgentsBy)}.Draw(k, geom.Box{X: laid.By[0].X - segmentInset, Y: laid.By[0].Y, W: laid.By[len(laid.By)-1].X + laid.By[len(laid.By)-1].W - laid.By[0].X + 2*segmentInset, H: laid.By[0].H})
	if len(groups) == 0 {
		Empty{Said: NoAgentsSays}.Draw(k, EmptyAt(geom.Box{X: laid.Run.X, Y: laid.Run.Y + geom.BarHigh + geom.Gap, W: laid.Run.W, H: geom.CardHeight}))
		return
	}
	in, clipped := k.within(laid.Shown)
	for i, g := range groups {
		Head{Said: g.Head, Count: strconv.Itoa(len(g.Agents))}.Draw(in, laid.Heads[i])
		for j, a := range g.Agents {
			AgentCard{Agent: a, Now: w.Now, Parts: w.AgentPartsOn(laid.Cards[i][j], a), Chosen: a.ID == w.sessions.on}.Draw(in, laid.Cards[i][j])
		}
	}
	clipped.Pop()
}

// AgentCard is a Claude session's card, laid out as its parts say: its state and for how long, its title or the task suggested, what it does and the answers to what it asks, the steps left, its branch, its context, where it runs, and its buttons.
type AgentCard struct {
	Agent  agent.Agent
	Now    time.Time
	Parts  AgentParts
	Chosen bool
}

func agentInside(at geom.Box) geom.Box {
	return geom.Box{X: at.X + 2*inset + 2, Y: at.Y + 2*inset + 2, W: at.W - 4*inset - 4, H: at.H - 4*inset - 4}
}

// Draw draws the card in at.
func (c AgentCard) Draw(k Kit, at geom.Box) geom.Box {
	k.put("agent", at)
	a, s, p, m := c.Agent, k.skin, c.Parts, k.measured()
	rounded(k.gtx, at, s.Radius.Card, s.Colour.Surface.Card)
	switch {
	case c.Chosen:
		Ring{Radius: s.Radius.Card}.Draw(k.inside("agent"), at)
	case a.State == agent.Waiting:
		ringed(k, at, s.Radius.Card, s.Colour.State.Late)
	}
	state := strings.Join(nonEmpty(agentStateSaid[a.State], a.Waiting), " · ")
	Pill{Said: state, Tone: ToneStanding, Standing: standingOfAgent[a.State]}.Draw(k, geom.Box{X: p.State.X, Y: p.State.Y, W: ends(m, state, small) + 2*pillPad, H: p.State.H})
	since := s.Colour.Ink.Low
	if !a.Silent.IsZero() {
		since = s.Colour.State.Late
	}
	sayIn(k.gtx, k.th, since, p.State, c.since(), small, s.Type.Mono, text.End)
	bold := font.Font{Weight: font.SemiBold}
	title, rest := sessionCalled(a), p.Title
	if a.Task != nil {
		title, rest = a.Task.Title, ahead(k, p.Title, "#"+strconv.Itoa(a.Task.Number)+" ", s.Colour.Accent.Lit, bold)
	}
	sayIn(k.gtx, k.th, s.Colour.Ink.Text, rest, fits(k.gtx, k.th, title, rest.W-6, small), small, bold, text.Start)
	if p.Tie.W > 0 {
		line := geom.Box{X: p.Suggestion.X, Y: p.Suggestion.Y + (p.Suggestion.H-lineHigh)/2, W: p.Suggestion.W, H: lineHigh}
		rest := ahead(k, ahead(k, line, "Maybe ", s.Colour.Ink.Faint, font.Font{}), "#"+strconv.Itoa(a.Suggested.Number)+" ", s.Colour.Accent.Lit, font.Font{})
		say(k.gtx, k.th, s.Colour.Ink.Faint, rest, fits(k.gtx, k.th, a.Suggested.Title+"?", rest.W, small), small)
		Button{Said: TieButtonSays, Small: true}.Draw(k, p.Tie)
	}
	if a.Asks != nil && len(a.Asks.Questions) == 0 {
		rest := ahead(k, p.Doing, a.Asks.Tool+": ", s.Colour.Ink.Faint, font.Font{})
		on := a.Asks.On
		if monoWide(on, small) > rest.W {
			on = cutToFit(on, rest.W, func(said string) int { return monoWide(said, small) })
		}
		sayIn(k.gtx, k.th, s.Colour.State.Late, rest, on, small, s.Type.Mono, text.Start)
	} else {
		say(k.gtx, k.th, s.Colour.Ink.Faint, p.Doing, fits(k.gtx, k.th, doingOf(a), p.Doing.W, small), small)
	}
	looks := map[string]buttonLook{AllowSays: buttonOn, DenySays: buttonDanger, GoToSessionSays: buttonOn, AnswerSays: buttonOn}
	for i, button := range p.Answers {
		Button{Said: p.AnswersSaid[i], Look: looks[p.AnswersSaid[i]], Small: true}.Draw(k, button)
	}
	for i, step := range p.Steps {
		say(k.gtx, k.th, s.Colour.Ink.Faint, step, fits(k.gtx, k.th, "○ "+a.Task.Open[i], step.W, small), small)
	}
	if a.Shared > 0 {
		say(k.gtx, k.th, s.Colour.State.Late, p.Shared, fits(k.gtx, k.th, "⚠ "+sharedSays(a.Shared), p.Shared.W, small), small)
	}
	say(k.gtx, k.th, s.Colour.Ink.Faint, p.Branch, fits(k.gtx, k.th, "⎇ "+a.Branch, p.Branch.W, small), small)
	model := strings.Join(nonEmpty(agent.ModelSaid(a.Model), fmt.Sprintf("%dk tok", a.Tokens/1000)), " · ")
	sayIn(k.gtx, k.th, s.Colour.Ink.Faint, p.Context, model, small, font.Font{}, text.End)
	say(k.gtx, k.th, s.Colour.Ink.Faint, p.Context, contextSays(a), small)
	rounded(k.gtx, p.Meter, 2, s.Colour.Surface.Control)
	fill := map[agent.ContextLevel]color.NRGBA{agent.ContextRoomy: s.Colour.State.Done, agent.ContextHigh: s.Colour.State.Late, agent.ContextFull: s.Colour.State.Wrong}[agent.ContextLevelOf(a.Context)]
	rounded(k.gtx, geom.Box{X: p.Meter.X, Y: p.Meter.Y, W: p.Meter.W * min(a.Context, 100) / 100, H: p.Meter.H}, 2, fill)
	say(k.gtx, k.th, s.Colour.Ink.Low, p.Meta, p.MetaSaid, small)
	for i, button := range p.Buttons {
		Button{Said: fits(k.gtx, k.th, p.ButtonsSaid[i], button.W-4, small), Look: looks[p.ButtonsSaid[i]], Small: true}.Draw(k, button)
	}
	return at
}

func ahead(k Kit, line geom.Box, said string, ink color.NRGBA, face font.Font) geom.Box {
	sayIn(k.gtx, k.th, ink, line, said, small, face, text.Start)
	wide := min(line.W, ends(k.measured(), said, small)+1)
	return geom.Box{X: line.X + wide, Y: line.Y, W: line.W - wide, H: line.H}
}

func (c AgentCard) since() string {
	if c.Agent.Since.IsZero() || c.Now.IsZero() {
		return ""
	}
	if !c.Agent.Silent.IsZero() {
		return "silent " + lasted(c.Now.Sub(c.Agent.Silent))
	}
	return lasted(c.Now.Sub(c.Agent.Since))
}

func lasted(d time.Duration) string {
	if d < time.Minute {
		return "for <1m"
	}
	return "for " + strings.TrimSuffix(d.Truncate(time.Minute).String(), "0s")
}

func sharedSays(others int) string {
	if others == 1 {
		return "1 other session holds this task"
	}
	return fmt.Sprintf("%d other sessions hold this task", others)
}

func nonEmpty(said ...string) []string {
	var kept []string
	for _, one := range said {
		if strings.TrimSpace(one) != "" {
			kept = append(kept, one)
		}
	}
	return kept
}

func drawQuestions(k Kit, w Window) {
	s, asked := k.skin, w.questionsAsked()
	for _, row := range w.QuestionRows() {
		q := asked[row.Question]
		if row.Option < 0 {
			rest := ahead(k, row.At, strings.ToUpper(q.Header)+"  ", s.Colour.Accent.Lit, font.Font{Weight: font.SemiBold})
			sayIn(k.gtx, k.th, s.Colour.Ink.Text, rest, fits(k.gtx, k.th, q.Question, rest.W-6, small), small, font.Font{Weight: font.SemiBold}, text.Start)
			continue
		}
		check := CheckAt(geom.Box{X: row.At.X + padding, Y: row.At.Y, W: row.At.W - padding, H: row.At.H})
		Check{Taken: row.Chosen, Round: !q.Multi}.Draw(k, check)
		o := q.Options[row.Option]
		line := geom.Box{X: check.X + check.W + 2*inset, Y: row.At.Y, W: row.At.X + row.At.W - check.X - check.W - 2*inset, H: row.At.H}
		rest := ahead(k, line, o.Label+"  ", s.Colour.Ink.Text, font.Font{})
		say(k.gtx, k.th, s.Colour.Ink.Faint, rest, fits(k.gtx, k.th, o.Description, rest.W, small), small)
	}
}
