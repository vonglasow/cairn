package gui

import (
	"cmp"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/engine/agent"
	"cairn/internal/engine/item"
	"cairn/internal/gui/geom"
)

// What the menu of a session, its card and the questions it opens say.
const (
	OpenTaskSays     = "Open task"
	GoToSessionSays  = "Go to session"
	CopyResumeSays   = "Copy claude --resume"
	NewTaskForSays   = "New task for this session"
	TieSays          = "Tie to a task…"
	AllowSays        = "Allow"
	DenySays         = "Deny"
	AllowAlwaysSays  = "Allow always…"
	AllowInProject   = "In this project's settings"
	AllowForUser     = "In my user settings"
	StartSessionSays = "Start a Claude session…"
	StartFolderSays  = "Folder…"
	CopyStartSays    = "Copy the command"
	TieButtonSays    = "Tie"
	AnswerSays       = "Answer…"
	SendSays         = "Send"
)

var answerOf = map[string]string{AllowSays: agent.Allow, DenySays: agent.Deny, AllowInProject: agent.AlwaysProject, AllowForUser: agent.AlwaysUser}

const tieWaits = "tie"

type sessionsState struct {
	on        string
	read      bool
	notices   []Notice
	startIn   string
	terminals []string
	scroll    int
	picked    map[[2]int]bool
}

// AgentCardWide is the least width of a session's card on the agents page: the cards of a row share what the run leaves.
const AgentCardWide = 380

const (
	agentButtonHigh = 22
	agentButtonPad  = 10
	meterHigh       = 5
)

// AgentsLaid is where the agents page puts each thing: the choice of grouping, and each group's head and cards scrolled by so much in the room shown under it, and how high they all stand.
type AgentsLaid struct {
	By    []geom.Box
	Heads []geom.Box
	Cards [][]geom.Box
	Run   geom.Box
	Shown geom.Box
	High  int
}

// AgentsRun is the room the agents page has under the tabs: the board's run, from the first column to the last.
func AgentsRun(l geom.Layout) geom.Box {
	last := l.Column[len(l.Column)-1]
	return geom.Box{X: l.Column[0].X, Y: l.Head[0].Y, W: last.X + last.W - l.Column[0].X, H: l.Column[0].Y + l.Column[0].H - l.Head[0].Y}
}

// LaidAgents lays the agents page out: the grouping chosen at the right of a strip, then each group's head over its cards, as many in a row as the run holds, scrolled up by scroll.
func LaidAgents(l geom.Layout, groups []board.AgentGroup, scroll int, high func(agent.Agent) int) AgentsLaid {
	run := AgentsRun(l)
	laid := AgentsLaid{Run: run}
	end := run.X + run.W
	for i := range board.AgentsBySaid {
		x := end - (len(board.AgentsBySaid)-i)*(rangeWide+between)
		laid.By = append(laid.By, geom.Box{X: x, Y: run.Y, W: rangeWide, H: geom.BarHigh})
	}
	room := run.W - l.Listed
	across := max(1, (room+geom.Gap)/(AgentCardWide+geom.Gap))
	wide := (room - (across-1)*geom.Gap) / across
	top := run.Y + geom.BarHigh + geom.Gap
	laid.Shown = geom.Box{X: run.X, Y: top, W: room, H: run.Y + run.H - top}
	y := top - scroll
	for _, g := range groups {
		laid.Heads = append(laid.Heads, geom.Box{X: run.X, Y: y, W: room, H: lineHigh + inset})
		y += lineHigh + inset + geom.Gap
		var cards []geom.Box
		for row := range slices.Chunk(g.Agents, across) {
			tallest := 0
			for _, a := range row {
				tallest = max(tallest, high(a))
			}
			for i := range row {
				cards = append(cards, geom.Box{X: run.X + i*(wide+geom.Gap), Y: y, W: wide, H: tallest})
			}
			y += tallest + geom.Gap
		}
		laid.Cards = append(laid.Cards, cards)
		y += geom.Gap
	}
	laid.High = y - top + scroll
	return laid
}

// AgentsLaid is the agents page laid out as the window shows it now.
func (w Window) AgentsLaid() AgentsLaid {
	return LaidAgents(w.Layout(), w.AgentsGrouped(), w.sessions.scroll, w.agentHigh)
}

func (w Window) agentHigh(a agent.Agent) int {
	return AgentPartsOf(geom.Box{W: AgentCardWide}, a, w.answers(a), w.measured()).High
}

func (w Window) settledAgents(want int) int {
	laid := LaidAgents(w.Layout(), w.AgentsGrouped(), 0, w.agentHigh)
	return min(max(0, laid.High-laid.Shown.H), max(0, want))
}

// AgentsGrouped is the sessions the page shows, in the groups chosen.
func (w Window) AgentsGrouped() []board.AgentGroup {
	return board.AgentGroups(w.Agents.Agents, w.AgentsBy)
}

func (w Window) readingAgents() Window {
	if !w.Pages.Shows(engine.PageAgents) {
		return w
	}
	found, err := board.ReadAgents(w.work)
	if err != nil {
		return w.failing(err)
	}
	if w.sessions.read && w.Pages.Notifies() {
		w.sessions.notices = append(slices.Clone(w.sessions.notices), noticesOf(w.Agents.Agents, found.Agents)...)
	}
	w.Agents, w.sessions.read = found, true
	w.sessions.scroll = w.settledAgents(w.sessions.scroll)
	return w
}

func noticesOf(was, now []agent.Agent) []Notice {
	var notices []Notice
	for _, a := range now {
		waited := slices.ContainsFunc(was, func(b agent.Agent) bool { return b.ID == a.ID && b.State == agent.Waiting })
		if a.State != agent.Waiting || waited {
			continue
		}
		said := a.Waiting
		if a.Task != nil {
			said = fmt.Sprintf("#%d %s", a.Task.Number, a.Task.Title)
		}
		notices = append(notices, Notice{Title: sessionCalled(a) + " waits for you", Said: said})
	}
	return notices
}

// Notices is the desktop notifications the window wants sent: one for each session that started waiting since the window last read them.
func (w Window) Notices() []Notice { return w.sessions.notices }

// NoticesTaken says the loop has sent them.
func (w Window) NoticesTaken() Window {
	w.sessions.notices = nil
	return w
}

// Notified takes what came of sending a notification: a command that could not start is said.
func (w Window) Notified(err error) Window {
	if err != nil {
		return w.failing(err)
	}
	return w
}

func sessionCalled(a agent.Agent) string { return cmp.Or(a.Name, filepath.Base(a.Folder)) }

func (w Window) agentAt(p geom.Point) (agent.Agent, geom.Box, bool) {
	groups, laid := w.AgentsGrouped(), w.AgentsLaid()
	for i, g := range groups {
		for j, a := range g.Agents {
			if laid.Cards[i][j].In(p) && laid.Shown.In(p) {
				return a, laid.Cards[i][j], true
			}
		}
	}
	return agent.Agent{}, geom.Box{}, false
}

func (w Window) sessionChosen() (agent.Agent, bool) {
	at := slices.IndexFunc(w.Agents.Agents, func(a agent.Agent) bool { return a.ID == w.sessions.on })
	if at < 0 {
		return agent.Agent{}, false
	}
	return w.Agents.Agents[at], true
}

func (w Window) answers(a agent.Agent) bool { return a.Asks != nil && w.Pages.AnswersPermissions() }

func (w Window) clickingAgents(p geom.Point) Window {
	for i, at := range w.AgentsLaid().By {
		if at.In(p) {
			w.AgentsBy = board.AgentsBy(i)
			return w
		}
	}
	a, box, ok := w.agentAt(p)
	if !ok {
		return w
	}
	w.sessions.on = a.ID
	parts := w.AgentPartsOn(box, a)
	if parts.Tie.In(p) {
		return w.asking(engine.AgentTie, tying(a, a.Suggested.Number))
	}
	in := func(b geom.Box) bool { return b.In(p) }
	if i := slices.IndexFunc(parts.Answers, in); i >= 0 {
		return w.onSession(parts.AnswersSaid[i])
	}
	if i := slices.IndexFunc(parts.Buttons, in); i >= 0 {
		return w.onSession(parts.ButtonsSaid[i])
	}
	if a.Task != nil {
		w.Chosen, w.ChosenGoal = a.Task.Number, false
		w = w.opening()
	}
	return w.reachingSession(a)
}

func sessionOf(a agent.Agent) engine.Input {
	return engine.Input{Given: map[string]string{"session": a.ID}}
}

func tying(a agent.Agent, n int) engine.Input {
	in := sessionOf(a)
	in.Given["task"] = strconv.Itoa(n)
	return in
}

func (w Window) reachingSession(a agent.Agent) Window {
	_, err := w.work.Ask(engine.AgentGo, sessionOf(a))
	switch {
	case err != nil && engine.RefusalOf(err).Kind == engine.KindRefused:
		return w.copied(a.Resume)
	case err != nil:
		return w.failing(err)
	}
	return w
}

func (w Window) sessionMenuAt(p geom.Point) Window {
	a, _, ok := w.agentAt(p)
	if !ok {
		return w
	}
	w.sessions.on = a.ID
	w.menu, w.rules = w.sessionMenu(a)
	w.menuAt, w.menuOn, w.menuSets, w.Screen = p, 0, "", board.OnMenu
	return w
}

func (w Window) sessionMenu(a agent.Agent) ([]string, []int) {
	var lines []string
	if a.Task != nil {
		lines = append(lines, OpenTaskSays)
	}
	lines = append(lines, GoToSessionSays, CopyResumeSays, NewTaskForSays, TieSays)
	if !w.answers(a) {
		return lines, nil
	}
	rules := []int{len(lines)}
	if len(a.Asks.Questions) > 0 {
		return append(lines, AnswerSays), rules
	}
	lines = append(lines, AllowSays, DenySays)
	if a.Asks.Always {
		lines = append(lines, AllowAlwaysSays)
	}
	return lines, rules
}

func (w Window) onSession(said string) Window {
	a, ok := w.sessionChosen()
	if !ok {
		return w
	}
	in := sessionOf(a)
	switch said {
	case OpenTaskSays:
		w.Chosen, w.ChosenGoal = a.Task.Number, false
		return w.opening()
	case GoToSessionSays:
		return w.reachingSession(a)
	case CopyResumeSays:
		return w.copied(a.Resume)
	case NewTaskForSays, NewTaskButtonSays:
		in.Given["title"] = sessionCalled(a)
		return w.asking(engine.AgentTie, in)
	case TieSays, TieButtonsSays:
		return waitingFor(tieWaits)(w)
	case AllowAlwaysSays:
		w.Screen, w.Question = board.OnAsking, board.AboutAllowingAlways
		return w
	case AnswerSays:
		w.Screen, w.Question, w.sessions.picked = board.OnAsking, board.AboutAnswering, nil
		return w
	}
	in.Given["answer"] = answerOf[said]
	return w.asking(engine.AgentAnswer, in)
}

func (w Window) tieRows(typed string) []PaletteRow {
	a, ok := w.sessionChosen()
	if !ok {
		return nil
	}
	var rows []PaletteRow
	for _, c := range w.Snap.Cards {
		if c.Archived || c.State == string(item.Done) || typed != "" && !board.ReachesCard(c, typed) {
			continue
		}
		in := tying(a, c.Number)
		rows = append(rows, PaletteRow{Group: paletteTasksGroup, Says: "#" + strconv.Itoa(c.Number) + "  " + c.Title, does: func(w Window) Window {
			return w.asking(engine.AgentTie, in)
		}})
	}
	return rows
}

func (w Window) startingSession() Window {
	got, err := w.work.Ask(engine.AgentTerminals, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	w.sessions.terminals = nil
	if got.Terminals != nil {
		w.sessions.terminals = got.Terminals.Terminals
	}
	w.sessions.startIn = w.Place().Folder
	for _, a := range w.Agents.Agents {
		if a.Task != nil && a.Task.Number == w.Chosen {
			w.sessions.startIn = a.Folder
		}
	}
	w.Screen, w.Question = board.OnAsking, board.AboutStarting
	return w
}

func (w Window) startCommand() string {
	return "cd " + agent.Quoted(w.sessions.startIn) + " && " + agent.CommandFor(w.Chosen)
}

func (w Window) agentQuestion() (head string, said []Said, offers []string, ok bool) {
	switch {
	case w.Screen != board.OnAsking:
		return "", nil, nil, false
	case w.Question == board.AboutAllowingAlways:
		a, _ := w.sessionChosen()
		var asked agent.Ask
		if a.Asks != nil {
			asked = *a.Asks
		}
		return strings.TrimSpace("Allow always: " + asked.Tool + " " + asked.On),
			[]Said{{Means: "Claude will not ask again: the rule goes in this project's settings, or in yours."}},
			[]string{AllowInProject, AllowForUser, "Cancel"}, true
	case w.Question == board.AboutAnswering:
		a, _ := w.sessionChosen()
		return sessionCalled(a) + " asks", make([]Said, w.questionLines()), []string{SendSays, "Cancel"}, true
	case w.Question == board.AboutStarting && len(w.sessions.terminals) == 0:
		return "Start a Claude session on #" + strconv.Itoa(w.Chosen),
			[]Said{{Means: "No terminal cairn can open is installed."}, {Means: "Copy the command, and run it in a terminal of your own."}},
			[]string{CopyStartSays, "Cancel"}, true
	case w.Question == board.AboutStarting:
		return "Start a Claude session on #" + strconv.Itoa(w.Chosen),
			[]Said{{Means: "In " + w.sessions.startIn}, {Means: "Claude is asked to work on cairn#" + strconv.Itoa(w.Chosen) + "."}},
			append(slices.Clone(w.sessions.terminals), StartFolderSays, "Cancel"), true
	}
	return "", nil, nil, false
}

func (w Window) choosingOnAgents(name string) (Window, bool) {
	if _, _, _, ok := w.agentQuestion(); !ok || name == "Cancel" {
		return w, false
	}
	switch {
	case name == StartFolderSays:
		w.fileAsk = &FileAsk{Folder: true}
		return w, true
	case name == SendSays && !w.answeredAll():
		return w, true
	case name == SendSays:
		a, _ := w.sessionChosen()
		in := sessionOf(a)
		in.Listed = map[string][]string{"choice": w.chosenLabels()}
		w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
		return w.asking(engine.AgentAnswer, in), true
	case w.Question == board.AboutAllowingAlways && answerOf[name] != "":
		w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
		return w.onSession(name), true
	case name == CopyStartSays:
		w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
		return w.copied(w.startCommand()), true
	case w.Question == board.AboutStarting && slices.Contains(w.sessions.terminals, name):
		w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
		return w.asking(engine.AgentStart, engine.Input{Given: map[string]string{"folder": w.sessions.startIn, "terminal": name, "task": strconv.Itoa(w.Chosen)}}), true
	}
	return w, false
}

func (w Window) startFolderChosen(got Filed) Window {
	switch {
	case got.Err != nil:
		return w.failing(got.Err)
	case !got.Declined:
		w.sessions.startIn = got.Chosen
	}
	return w
}

// AgentParts is where each row and button of a session's card stands, top to bottom, the buttons at its foot; a row the session has not is left empty and takes no room. High is how high the card must be to hold them.
type AgentParts struct {
	High                                                                       int
	State, Title, Suggestion, Tie, Doing, Shared, Branch, Context, Meter, Meta geom.Box
	Steps, Answers, Buttons                                                    []geom.Box
	AnswersSaid, ButtonsSaid                                                   []string
	MetaSaid                                                                   string
	Copy                                                                       geom.Box
}

// AgentPartsOn is where each row and button of a session's card in card stands.
func (w Window) AgentPartsOn(card geom.Box, a agent.Agent) AgentParts {
	return AgentPartsOf(card, a, w.answers(a), w.measured())
}

// AgentPartsOf lays a session's card out: its state, its title and the task suggested, what it does and the answers to what it asks, the steps left, its branch, its context, where it runs, and its buttons.
func AgentPartsOf(card geom.Box, a agent.Agent, answering bool, m Measure) AgentParts {
	in := agentInside(card)
	y := in.Y
	next := func(high int) geom.Box {
		at := geom.Box{X: in.X, Y: y, W: in.W, H: high}
		y += high
		return at
	}
	var p AgentParts
	p.State = next(lineHigh)
	y += 2 * inset
	p.Title = next(lineHigh)
	if a.Task == nil && a.Suggested != nil {
		p.Suggestion = next(agentButtonHigh)
		p.Tie = buttonsIn(p.Suggestion, []string{TieButtonSays}, m)[0]
		p.Tie.X = in.X + in.W - p.Tie.W
		p.Suggestion.W -= p.Tie.W + inset
	}
	y += inset
	if doingOf(a) != "" {
		p.Doing = next(lineHigh)
	}
	if answering {
		p.AnswersSaid = []string{AllowSays, AllowAlwaysSays, DenySays}
		switch {
		case len(a.Asks.Questions) > 0:
			p.AnswersSaid = []string{AnswerSays}
		case !a.Asks.Always:
			p.AnswersSaid = []string{AllowSays, DenySays}
		}
		p.Answers = buttonsIn(next(agentButtonHigh), p.AnswersSaid, m)
		y += inset
	}
	if a.Task != nil {
		for range min(2, len(a.Task.Open)) {
			p.Steps = append(p.Steps, next(lineHigh))
		}
		if a.Shared > 0 {
			p.Shared = next(lineHigh)
		}
	}
	if a.Branch != "" {
		y += inset
		p.Branch = next(lineHigh)
	}
	y += inset
	p.Context = next(lineHigh)
	p.Meter = next(meterHigh)
	y += inset
	p.Meta = next(lineHigh)
	p.MetaSaid = metaOf(a)
	if wide := func(said string) int { return ends(m, said, small) }; wide(p.MetaSaid) > p.Meta.W-copyRoom {
		p.MetaSaid = cutToFit(p.MetaSaid, p.Meta.W-copyRoom, wide)
	}
	p.Copy = markBefore(geom.Box{X: p.Meta.X + ends(m, p.MetaSaid, small) + copyRoom - copySide, Y: p.Meta.Y, H: p.Meta.H}, a.Resume).At
	y += 2 * inset
	if a.Terminal.Reachable {
		p.ButtonsSaid = []string{GoToSessionSays}
	}
	if a.Task != nil {
		p.ButtonsSaid = append(p.ButtonsSaid, OpenTaskSays)
	} else {
		p.ButtonsSaid = append(p.ButtonsSaid, NewTaskButtonSays, TieButtonsSays)
	}
	p.High = y + agentButtonHigh - card.Y + 2*inset + 2
	y = max(y, in.Y+in.H-agentButtonHigh)
	p.Buttons = buttonsIn(next(agentButtonHigh), p.ButtonsSaid, m)
	return p
}

func buttonsIn(row geom.Box, said []string, m Measure) []geom.Box {
	words, gaps := 0, (len(said)-1)*(inset+2)
	for _, one := range said {
		words += ends(m, one, small)
	}
	pad := max(0, min(agentButtonPad, (row.W-gaps-words)/(2*len(said))))
	room := row.W - gaps - 2*pad*len(said)
	var boxes []geom.Box
	x := row.X
	for _, one := range said {
		wide := ends(m, one, small)
		if words > room {
			wide = wide * room / words
		}
		boxes = append(boxes, geom.Box{X: x, Y: row.Y, W: wide + 2*pad, H: row.H})
		x += wide + 2*pad + inset + 2
	}
	return boxes
}

func metaOf(a agent.Agent) string {
	tty := cmp.Or(strings.Trim(a.Terminal.TTY, "?"), "no tty")
	return strings.Join(nonEmpty("⌂ "+a.Project, strings.Join(nonEmpty(a.Terminal.Program, tty), " · "), a.Short), "   ")
}

func (w Window) resumeMarks() []CopyMark {
	laid := w.AgentsLaid()
	var marks []CopyMark
	for i, g := range w.AgentsGrouped() {
		for j, a := range g.Agents {
			at := w.AgentPartsOn(laid.Cards[i][j], a).Copy
			if laid.Shown.In(geom.Point{X: at.X, Y: at.Y}) && laid.Shown.In(geom.Point{X: at.X + at.W, Y: at.Y + at.H}) {
				marks = append(marks, CopyMark{At: at, Said: a.Resume})
			}
		}
	}
	return marks
}

func sessionTitle(a agent.Agent) string {
	if a.Task != nil {
		return fmt.Sprintf("#%d %s", a.Task.Number, a.Task.Title)
	}
	return sessionCalled(a)
}

func doingOf(a agent.Agent) string {
	if a.Asks != nil && len(a.Asks.Questions) == 1 {
		return "Asks 1 question"
	}
	if a.Asks != nil && len(a.Asks.Questions) > 1 {
		return fmt.Sprintf("Asks %d questions", len(a.Asks.Questions))
	}
	if a.Asks != nil {
		return strings.TrimSpace(a.Asks.Tool + ": " + a.Asks.On)
	}
	return a.Doing
}

func (w Window) agentBubbleAt(p geom.Point) ([]string, geom.Box, int) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Agents {
		return nil, geom.Box{}, 0
	}
	a, card, ok := w.agentAt(p)
	if !ok {
		return nil, geom.Box{}, 0
	}
	parts := w.AgentPartsOn(card, a)
	var wrote []string
	if !a.Silent.IsZero() {
		wrote = []string{"Last written at " + a.Silent.Local().Format("15:04")}
	}
	for _, row := range []struct {
		at   geom.Box
		said []string
	}{
		{parts.State, wrote},
		{parts.Title, []string{sessionTitle(a)}},
		{parts.Doing, []string{doingOf(a)}},
		{parts.Branch, []string{a.Branch}},
		{parts.Context, []string{strings.Join(nonEmpty(a.Model, fmt.Sprintf("%d tokens", a.Tokens)), " · ")}},
		{parts.Meta, []string{a.Folder, strings.Join(nonEmpty(a.Name, a.ID), " · ")}},
	} {
		if row.at.In(p) && row.said != nil {
			return row.said, row.at, card.W
		}
	}
	return nil, geom.Box{}, 0
}

// QuestionRow is one line of the dialog answering the questions a session asks: a question, or one of its options, where it stands, and whether it is chosen.
type QuestionRow struct {
	Question, Option int
	At               geom.Box
	Chosen           bool
}

// QuestionRows is the lines of the dialog answering the questions a session asks, the option -1 on a question's own line.
func (w Window) QuestionRows() []QuestionRow {
	if w.Screen != board.OnAsking || w.Question != board.AboutAnswering {
		return nil
	}
	var rows []QuestionRow
	for i, q := range w.questionsAsked() {
		rows = append(rows, QuestionRow{Question: i, Option: -1})
		for j := range q.Options {
			rows = append(rows, QuestionRow{Question: i, Option: j, Chosen: w.sessions.picked[[2]int{i, j}]})
		}
	}
	for i := range rows {
		rows[i].At = w.SaidAt(i)
	}
	return rows
}

func (w Window) questionLines() int {
	lines := 0
	for _, q := range w.questionsAsked() {
		lines += 1 + len(q.Options)
	}
	return lines
}

func (w Window) questionsAsked() []agent.Question {
	a, ok := w.sessionChosen()
	if !ok || a.Asks == nil {
		return nil
	}
	return a.Asks.Questions
}

func (w Window) pickingOption(p geom.Point) (Window, bool) {
	rows := w.QuestionRows()
	at := slices.IndexFunc(rows, func(row QuestionRow) bool { return row.Option >= 0 && row.At.In(p) })
	if at < 0 {
		return w, false
	}
	row := rows[at]
	picked := maps.Clone(w.sessions.picked)
	if picked == nil {
		picked = map[[2]int]bool{}
	}
	if !w.questionsAsked()[row.Question].Multi {
		maps.DeleteFunc(picked, func(k [2]int, _ bool) bool { return k[0] == row.Question })
	}
	picked[[2]int{row.Question, row.Option}] = !row.Chosen
	w.sessions.picked = picked
	return w, true
}

func (w Window) chosenLabels() []string {
	var chosen []string
	for i, q := range w.questionsAsked() {
		var labels []string
		for j, o := range q.Options {
			if w.sessions.picked[[2]int{i, j}] {
				labels = append(labels, o.Label)
			}
		}
		chosen = append(chosen, strings.Join(labels, ", "))
	}
	return chosen
}

func (w Window) answeredAll() bool { return !slices.Contains(w.chosenLabels(), "") }

// What the buttons of a session holding no task say.
const (
	NewTaskButtonSays = "New task"
	TieButtonsSays    = "Tie…"
)

func contextSays(a agent.Agent) string {
	said := fmt.Sprintf("Context %d%%", a.Context)
	if agent.ContextLevelOf(a.Context) == agent.ContextFull {
		said += " · compact soon"
	}
	return said
}

var agentKeys = map[string]string{GoToSessionSays: "enter", OpenTaskSays: "o", AllowSays: "a", AnswerSays: "a", DenySays: "d", CopyResumeSays: "c"}

func (w Window) agentOffers() []string {
	a, ok := w.sessionChosen()
	if !ok {
		return nil
	}
	offers := []string{GoToSessionSays}
	if a.Task != nil {
		offers = append(offers, OpenTaskSays)
	}
	switch {
	case w.answers(a) && len(a.Asks.Questions) > 0:
		offers = append(offers, AnswerSays)
	case w.answers(a):
		offers = append(offers, AllowSays, DenySays)
	}
	return append(offers, CopyResumeSays)
}

var agentSteps = map[string][2]int{"left": {-1, 0}, "h": {-1, 0}, "right": {1, 0}, "l": {1, 0}, "up": {0, -1}, "k": {0, -1}, "down": {0, 1}, "j": {0, 1}}

func (w Window) agentsKey(name string) (Window, bool) {
	if name == "esc" && w.sessions.on != "" {
		w.sessions.on = ""
		return w, true
	}
	step, ok := agentSteps[name]
	if !ok {
		return w, false
	}
	return w.walkingSessions(step[0], step[1]), true
}

func (w Window) walkingSessions(across, down int) Window {
	laid := w.AgentsLaid()
	var ids []string
	var boxes []geom.Box
	for i, g := range w.AgentsGrouped() {
		for j, a := range g.Agents {
			ids, boxes = append(ids, a.ID), append(boxes, laid.Cards[i][j])
		}
	}
	if len(ids) == 0 {
		return w
	}
	at := slices.Index(ids, w.sessions.on)
	switch {
	case at < 0:
		at = 0
	case down == 0:
		at = min(len(ids)-1, max(0, at+across))
	default:
		from, best := boxes[at], -1
		far := func(b geom.Box) int { return abs(b.Y-from.Y)*100000 + abs(b.X-from.X) }
		for i, b := range boxes {
			if (b.Y-from.Y)*down > 0 && (best < 0 || far(b) < far(boxes[best])) {
				best = i
			}
		}
		at = cmp.Or(best+1, at+1) - 1
	}
	w.sessions.on = ids[at]
	box, shown := boxes[at], laid.Shown
	switch {
	case box.Y < shown.Y:
		w.sessions.scroll -= shown.Y - box.Y
	case box.Y+box.H > shown.Y+shown.H:
		w.sessions.scroll += box.Y + box.H - shown.Y - shown.H
	}
	w.sessions.scroll = w.settledAgents(w.sessions.scroll)
	return w
}
