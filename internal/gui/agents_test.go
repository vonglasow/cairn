package gui

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/engine/agent"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
)

func TestTheAgentsPageGroupsTheWayClicked(t *testing.T) {
	w := homeWindow(1200, 800)
	w.Tab = board.Agents
	w.Agents = agent.Agents{Agents: []agent.Agent{{ID: "a", State: agent.Waiting, Folder: "/w/cairn"}}, Waiting: 1}
	for i, said := range board.AgentsBySaid {
		at := w.AgentsLaid().By[i]
		if got := w.Click(geom.Point{X: at.X + 4, Y: at.Y + 4}).AgentsBy; got != board.AgentsBy(i) {
			t.Errorf("clicking %s groups by %d", said, got)
		}
	}
	if got := w.TabAlert(board.Agents); got != 1 {
		t.Errorf("the agents tab says %d waiting", got)
	}
}

var (
	asking = agent.Agent{ID: "a3f9c1d2", Short: "a3f9c1", Name: "agents page", State: agent.Waiting, Waiting: "permission", Folder: "/w/cairn",
		Task: &agent.Held{Number: 3, Title: "Buy the relay"}, Resume: "cd '/w/cairn' && claude --resume a3f9c1d2",
		Asks: &agent.Ask{Tool: "Bash", On: "make test", Always: true}}
	loose = agent.Agent{ID: "7be210aa", Short: "7be210", State: agent.Working, Folder: "/w/shop", Resume: "cd '/w/shop' && claude --resume 7be210aa"}
)

func pagesWith(notify, count, permissions bool) engine.Pages {
	return engine.PagesOf(workspace.Config{Agents: &workspace.AgentsConfig{Notify: notify, Count: count, Permissions: permissions}})
}

func agentsWindow(f *fake, p engine.Pages, listed ...agent.Agent) Window {
	counted := agent.Counted(listed)
	f.agents = &counted
	w := holding(f, card(3, "Buy the relay", "current"), card(4, "Pin map", "todo"), card(5, "Old", "done"))
	w.Pages, w.Tab = p, board.Agents
	return w.readingAgents()
}

func sessionBox(w Window, id string) geom.Box {
	laid := w.AgentsLaid()
	for i, g := range w.AgentsGrouped() {
		for j, a := range g.Agents {
			if a.ID == id {
				return laid.Cards[i][j]
			}
		}
	}
	return geom.Box{}
}

func onSession(w Window, id string) geom.Point {
	box := sessionBox(w, id)
	return geom.Point{X: box.X + box.W - 10, Y: box.Y + box.H - 10}
}

func TestTheWindowReadingTheSessionsSaysItListens(t *testing.T) {
	f := newFake()
	agentsWindow(f, pagesWith(true, true, false))
	if in, ok := f.gave("agent", "list"); !ok || in.Given["listening"] != "true" {
		t.Errorf("the sessions were read with %v", in.Given)
	}
}

func TestClickingASessionOpensItsTaskAndBringsItsTerminalForward(t *testing.T) {
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, false), asking, loose)
	w = w.Click(onSession(w, asking.ID))
	if in, ok := f.gave("agent", "go"); !ok || in.Given["session"] != asking.ID {
		t.Errorf("the terminal was asked for with %v", in.Given)
	}
	if w.Screen != board.OnForm || w.Chosen != 3 {
		t.Errorf("the task is not open: screen %v, task %d", w.Screen, w.Chosen)
	}
	f.refuseGo = engine.Refusing("out-of-reach", "cairn cannot bring VS Code to the front")
	w = agentsWindow(f, pagesWith(true, true, false), asking, loose).Click(onSession(w, loose.ID))
	if said, ok := w.Copying(); !ok || said != loose.Resume || w.Telling != "copied "+loose.Resume {
		t.Errorf("a terminal out of reach copied %q and said %q", said, w.Telling)
	}
}

func TestASessionsMenuOffersWhatCanBeDoneToIt(t *testing.T) {
	for _, c := range []struct {
		name  string
		pages engine.Pages
		one   agent.Agent
		want  []string
	}{
		{"asking, answered from the window", pagesWith(true, true, true), asking,
			[]string{"Open task", "Go to session", "Copy claude --resume", "New task for this session", "Tie to a task…", "Allow", "Deny", "Allow always…"}},
		{"asking, answered in the terminal", pagesWith(true, true, false), asking,
			[]string{"Open task", "Go to session", "Copy claude --resume", "New task for this session", "Tie to a task…"}},
		{"holding no task", pagesWith(true, true, true), loose,
			[]string{"Go to session", "Copy claude --resume", "New task for this session", "Tie to a task…"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := agentsWindow(newFake(), c.pages, c.one)
			w = w.ContextAt(onSession(w, c.one.ID))
			if w.Screen != board.OnMenu || !slices.Equal(w.Menu(), c.want) {
				t.Errorf("the menu offers %q", w.Menu())
			}
		})
	}
}

func TestASessionsMenuLinesAskTheEngine(t *testing.T) {
	for _, c := range []struct {
		line, verb string
		one        agent.Agent
		want       map[string]string
	}{
		{"New task for this session", "tie", asking, map[string]string{"session": asking.ID, "title": "agents page"}},
		{"New task for this session", "tie", loose, map[string]string{"session": loose.ID, "title": "shop"}},
		{"Allow", "answer", asking, map[string]string{"session": asking.ID, "answer": agent.Allow}},
		{"Deny", "answer", asking, map[string]string{"session": asking.ID, "answer": agent.Deny}},
		{"Go to session", "go", asking, map[string]string{"session": asking.ID}},
	} {
		t.Run(c.line, func(t *testing.T) {
			f := newFake()
			w := agentsWindow(f, pagesWith(true, true, true), c.one)
			w = w.ContextAt(onSession(w, c.one.ID))
			w.fromMenu(slices.Index(w.Menu(), c.line))
			in, ok := f.gave("agent", c.verb)
			if !ok {
				t.Fatalf("%q asked nothing of agent %s", c.line, c.verb)
			}
			for arg, said := range c.want {
				if in.Given[arg] != said {
					t.Errorf("%q gave %s=%q", c.line, arg, in.Given[arg])
				}
			}
		})
	}
}

func TestCopyingTheResumeCommandSaysItWasCopied(t *testing.T) {
	w := agentsWindow(newFake(), pagesWith(true, true, true), asking)
	w = w.ContextAt(onSession(w, asking.ID))
	w = w.fromMenu(slices.Index(w.Menu(), "Copy claude --resume"))
	if said, ok := w.Copying(); !ok || said != asking.Resume || w.Telling != "copied "+asking.Resume {
		t.Errorf("copied %q and said %q", said, w.Telling)
	}
}

func TestAllowingForGoodAsksWhereTheRuleIsWritten(t *testing.T) {
	for _, c := range []struct{ offer, answer string }{
		{AllowInProject, agent.AlwaysProject},
		{AllowForUser, agent.AlwaysUser},
	} {
		f := newFake()
		w := agentsWindow(f, pagesWith(true, true, true), asking)
		w = w.ContextAt(onSession(w, asking.ID))
		w = w.fromMenu(slices.Index(w.Menu(), "Allow always…"))
		if w.Screen != board.OnAsking || Heads(w) != "Allow always: Bash make test" || !slices.Contains(w.Offers(), c.offer) {
			t.Fatalf("the window asks %q offering %q", Heads(w), w.Offers())
		}
		w = w.Choose(c.offer)
		if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != c.answer || in.Given["session"] != asking.ID {
			t.Errorf("%q answered %v", c.offer, in.Given)
		}
		if w.Screen != board.OnTheBoard {
			t.Errorf("the question stays open: %v", w.Screen)
		}
	}
}

func centre(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }

func partsOf(w Window, a agent.Agent) AgentParts { return w.AgentPartsOn(sessionBox(w, a.ID), a) }

func TestTheButtonsOfASessionAskingAnswerIt(t *testing.T) {
	once := asking
	once.Asks = &agent.Ask{Tool: "Bash", On: "make test"}
	for _, c := range []struct {
		name  string
		one   agent.Agent
		said  string
		check func(*testing.T, *fake, Window)
	}{
		{"allow", asking, AllowSays, func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != agent.Allow || in.Given["session"] != asking.ID {
				t.Errorf("answered %v", in.Given)
			}
		}},
		{"deny", asking, DenySays, func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != agent.Deny {
				t.Errorf("answered %v", in.Given)
			}
		}},
		{"allow always asks where", asking, AllowAlwaysSays, func(t *testing.T, _ *fake, w Window) {
			if w.Screen != board.OnAsking || Heads(w) != "Allow always: Bash make test" {
				t.Errorf("opens %v asking %q", w.Screen, Heads(w))
			}
		}},
		{"deny, when it cannot be allowed for good", once, DenySays, func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != agent.Deny {
				t.Errorf("answered %v", in.Given)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFake()
			w := agentsWindow(f, pagesWith(true, true, true), c.one)
			parts := partsOf(w, c.one)
			want := []string{AllowSays, AllowAlwaysSays, DenySays}
			if !c.one.Asks.Always {
				want = []string{AllowSays, DenySays}
			}
			if !slices.Equal(parts.AnswersSaid, want) || len(parts.Answers) != len(want) {
				t.Fatalf("the card answers %q", parts.AnswersSaid)
			}
			c.check(t, f, w.Click(centre(parts.Answers[slices.Index(want, c.said)])))
		})
	}
	w := agentsWindow(newFake(), pagesWith(true, true, false), asking)
	if got := partsOf(w, asking).Answers; len(got) != 0 {
		t.Errorf("a window not answering shows %d buttons", len(got))
	}
}

func TestTheButtonsOfASessionDoWhatItsMenuDoes(t *testing.T) {
	near := asking
	near.Terminal = agent.Terminal{Program: "iTerm2", TTY: "ttys004", Reachable: true}
	nearLoose := loose
	nearLoose.Terminal = near.Terminal
	for _, c := range []struct {
		name  string
		one   agent.Agent
		said  []string
		click string
		check func(*testing.T, *fake, Window)
	}{
		{"go to a session in reach", near, []string{GoToSessionSays, OpenTaskSays}, GoToSessionSays, func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "go"); !ok || in.Given["session"] != near.ID {
				t.Errorf("went with %v", in.Given)
			}
		}},
		{"open its task", near, []string{GoToSessionSays, OpenTaskSays}, OpenTaskSays, func(t *testing.T, f *fake, w Window) {
			if _, ok := f.gave("agent", "go"); ok || w.Screen != board.OnForm || w.Chosen != 3 {
				t.Errorf("screen %v, task %d", w.Screen, w.Chosen)
			}
		}},
		{"a session out of reach has its task to open", asking, []string{OpenTaskSays}, OpenTaskSays, func(t *testing.T, _ *fake, w Window) {
			if w.Screen != board.OnForm || w.Chosen != 3 {
				t.Errorf("screen %v, task %d", w.Screen, w.Chosen)
			}
		}},
		{"a session holding no task gets a task made for it", nearLoose, []string{GoToSessionSays, NewTaskButtonSays, TieButtonsSays}, NewTaskButtonSays, func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "tie"); !ok || in.Given["session"] != loose.ID || in.Given["title"] != "shop" {
				t.Errorf("tied with %v", in.Given)
			}
		}},
		{"a session holding no task is tied to a task picked", nearLoose, []string{GoToSessionSays, NewTaskButtonSays, TieButtonsSays}, TieButtonsSays, func(t *testing.T, _ *fake, w Window) {
			if w.Screen != board.OnPalette || len(w.PaletteRows()) != 2 {
				t.Errorf("the chooser is %v with %d rows", w.Screen, len(w.PaletteRows()))
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFake()
			w := agentsWindow(f, pagesWith(true, true, false), c.one)
			parts := partsOf(w, c.one)
			if !slices.Equal(parts.ButtonsSaid, c.said) || len(parts.Buttons) != len(c.said) {
				t.Fatalf("the card offers %q", parts.ButtonsSaid)
			}
			card, last := sessionBox(w, c.one.ID), parts.Buttons[len(parts.Buttons)-1]
			if last.X+last.W > card.X+card.W-2*inset {
				t.Errorf("the buttons run to %d, past the card's %d", last.X+last.W, card.X+card.W)
			}
			c.check(t, f, w.Click(centre(parts.Buttons[slices.Index(c.said, c.click)])))
		})
	}
}

func TestTheTaskSuggestedToASessionIsTiedInOneClick(t *testing.T) {
	suggested := loose
	suggested.Suggested = &agent.Held{Number: 4, Title: "Pin map"}
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, false), suggested)
	parts := partsOf(w, suggested)
	if parts.Tie.W == 0 {
		t.Fatal("the card has no Tie")
	}
	w.Click(centre(parts.Tie))
	if in, ok := f.gave("agent", "tie"); !ok || in.Given["task"] != "4" || in.Given["session"] != loose.ID {
		t.Errorf("the tie gave %v", in.Given)
	}
	if got := partsOf(w, loose).Tie; got.W != 0 {
		t.Errorf("a session with nothing suggested shows a Tie at %v", got)
	}
}

func TestRestingOnASessionsCardSaysItWhole(t *testing.T) {
	one := asking
	one.Name, one.ID, one.Folder = "packaging", "a3f9c1d2-0000-4000", "/Users/me/work/cairn"
	one.Branch, one.Model, one.Tokens = "feature/85621-registry-pvc-size-and-retention", "claude-opus-5-5", 144000
	w := agentsWindow(newFake(), pagesWith(true, true, true), one)
	parts := partsOf(w, one)
	for _, c := range []struct {
		name string
		at   geom.Box
		want string
	}{
		{"its title", parts.Title, "#3 Buy the relay"},
		{"what it does", parts.Doing, "Bash: make test"},
		{"its branch", parts.Branch, one.Branch},
		{"its model", parts.Context, "claude-opus-5-5 · 144000 tokens"},
		{"where it runs", parts.Meta, "/Users/me/work/cairn\npackaging · a3f9c1d2-0000-4000"},
	} {
		if got := w.Hovering(centre(c.at)).Bubble(); got != c.want {
			t.Errorf("%s: the bubble says %q", c.name, got)
		}
	}
}

func TestASessionIsTiedToATaskPickedInTheCommandBar(t *testing.T) {
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, true), loose)
	w = w.ContextAt(onSession(w, loose.ID))
	w = w.fromMenu(slices.Index(w.Menu(), "Tie to a task…"))
	var says []string
	for _, row := range w.PaletteRows() {
		says = append(says, row.Says)
	}
	if w.Screen != board.OnPalette || !slices.Equal(says, []string{"#3  Buy the relay", "#4  Pin map"}) {
		t.Fatalf("the command bar offers %q", says)
	}
	w.runningPalette(w.PaletteRows()[1])
	if in, ok := f.gave("agent", "tie"); !ok || in.Given["task"] != "4" || in.Given["session"] != loose.ID {
		t.Errorf("the tie gave %v", in.Given)
	}
}

func TestASessionIsStartedOnATaskInATerminalChosen(t *testing.T) {
	f := newFake()
	f.terminals = []string{"iTerm2", "Terminal"}
	w := agentsWindow(f, pagesWith(true, true, true), asking)
	w.Tab, w.Chosen = board.Tasks, 3
	if !slices.Contains(w.menuFor(), StartSessionSays) {
		t.Fatalf("the menu of a task offers %q", w.menuFor())
	}
	w = w.doingLine(StartSessionSays)
	if w.Screen != board.OnAsking || Heads(w) != "Start a Claude session on #3" || !slices.Equal(w.Offers(), []string{"iTerm2", "Terminal", StartFolderSays, "Cancel"}) {
		t.Fatalf("the window asks %q offering %q", Heads(w), w.Offers())
	}
	w = w.Choose("Terminal")
	in, ok := f.gave("agent", "start")
	if !ok || in.Given["terminal"] != "Terminal" || in.Given["folder"] != "/w/cairn" || in.Given["task"] != "3" {
		t.Errorf("the start gave %v", in.Given)
	}
}

func TestASessionNoTerminalStartsIsCopied(t *testing.T) {
	f := newFake()
	f.shown = &workspace.Shown{Workspace: "home", Folder: "/w/home"}
	w := agentsWindow(f, pagesWith(true, true, true))
	w.Tab, w.Chosen = board.Tasks, 4
	w = w.readingPlace().doingLine(StartSessionSays)
	if !slices.Equal(w.Offers(), []string{CopyStartSays, "Cancel"}) {
		t.Fatalf("with no terminal the window offers %q", w.Offers())
	}
	w = w.Choose(CopyStartSays)
	want := "cd '/w/home' && claude 'let'\\''s work on cairn#4'"
	if said, _ := w.Copying(); said != want || w.Screen != board.OnTheBoard {
		t.Errorf("copied %q", said)
	}
}

func TestASessionStartingToWaitIsNotified(t *testing.T) {
	idle := asking
	idle.State, idle.Waiting, idle.Asks = agent.Idle, "", nil
	for _, c := range []struct {
		name        string
		pages       engine.Pages
		first, then agent.Agent
		want        []Notice
	}{
		{"one starting to wait", pagesWith(true, true, false), idle, asking, []Notice{{Title: "agents page waits for you", Said: "#3 Buy the relay"}}},
		{"one waiting when the window opened", pagesWith(true, true, false), asking, asking, nil},
		{"notifications off", pagesWith(false, true, false), idle, asking, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFake()
			w := agentsWindow(f, c.pages, c.first)
			counted := agent.Counted([]agent.Agent{c.then})
			f.agents = &counted
			if got := w.readingAgents().Notices(); !slices.Equal(got, c.want) {
				t.Errorf("notified %+v", got)
			}
		})
	}
}

func TestTheAgentsTabCountsOnlyWhenAskedTo(t *testing.T) {
	if got := agentsWindow(newFake(), pagesWith(true, false, false), asking).TabAlert(board.Agents); got != 0 {
		t.Errorf("the agents tab counts %d", got)
	}
}

func TestANoticeIsSentByTheSystemsOwnCommand(t *testing.T) {
	n := Notice{Title: `"x" waits for you`, Said: `#3 a\b`}
	for _, c := range []struct {
		goos, name string
		args       []string
	}{
		{"darwin", "osascript", []string{"-e", `display notification "#3 a\\b" with title "\"x\" waits for you"`}},
		{"linux", "notify-send", []string{`"x" waits for you`, `#3 a\b`}},
	} {
		if name, args := NoticeCommand(c.goos, n); name != c.name || !slices.Equal(args, c.args) {
			t.Errorf("%s sends %s %q", c.goos, name, args)
		}
	}
}

func TestTheWheelScrollsTheAgentsPageAndClicksFollow(t *testing.T) {
	var many []agent.Agent
	for i := range 30 {
		one := loose
		one.ID, one.Terminal.Reachable = fmt.Sprintf("s%02d", i), true
		many = append(many, one)
	}
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, false), many...)
	last := many[len(many)-1]
	shown := w.AgentsLaid().Shown
	if box := sessionBox(w, last.ID); box.Y+box.H <= shown.Y+shown.H {
		t.Fatalf("the last card %v fits in %v unscrolled", box, shown)
	}
	w = w.Scroll(centre(shown), 10000)
	box := sessionBox(w, last.ID)
	if box.Y < shown.Y || box.Y+box.H > shown.Y+shown.H {
		t.Fatalf("scrolled to its end, the last card stands at %v out of %v", box, shown)
	}
	w.Click(centre(partsOf(w, last).Buttons[0]))
	if in, ok := f.gave("agent", "go"); !ok || in.Given["session"] != last.ID {
		t.Errorf("Go to session went with %v", in.Given)
	}
}

var questioning = agent.Agent{ID: "e1b2c3d4", Short: "e1b2c3", Name: "migration plan", State: agent.Waiting, Waiting: "question", Folder: "/w/cairn",
	Resume: "cd '/w/cairn' && claude --resume e1b2c3d4", Asks: &agent.Ask{Tool: agent.AskUserQuestion, Questions: []agent.Question{
		{Question: "Which store?", Header: "Store", Options: []agent.Option{{Label: "Postgres", Description: "the server"}, {Label: "SQLite"}}},
		{Question: "What to cover?", Header: "Plan", Multi: true, Options: []agent.Option{{Label: "Rollback"}, {Label: "Backups"}, {Label: "Timings"}}},
	}}}

func optionAt(w Window, question, option int) geom.Point {
	for _, row := range w.QuestionRows() {
		if row.Question == question && row.Option == option {
			return centre(row.At)
		}
	}
	return geom.Point{X: -1, Y: -1}
}

func TestTheQuestionsASessionAsksAreAnsweredInADialog(t *testing.T) {
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, true), questioning)
	parts := partsOf(w, questioning)
	if !slices.Equal(parts.AnswersSaid, []string{AnswerSays}) {
		t.Fatalf("the card answers %q", parts.AnswersSaid)
	}
	if got := w.Hovering(centre(parts.Doing)).Bubble(); got != "Asks 2 questions" {
		t.Errorf("the card says it does %q", got)
	}
	w = w.Click(centre(parts.Answers[0]))
	if w.Screen != board.OnAsking || Heads(w) != "migration plan asks" || !slices.Equal(w.Offers(), []string{SendSays, "Cancel"}) || len(w.QuestionRows()) != 7 {
		t.Fatalf("the dialog asks %q offering %q in %d rows", Heads(w), w.Offers(), len(w.QuestionRows()))
	}
	if w.OfferLook(0, SendSays) != buttonFaint {
		t.Errorf("Send is on before every question has an answer")
	}
	if w = w.Choose(SendSays); w.Screen != board.OnAsking {
		t.Fatalf("Send closed the dialog unanswered")
	}
	for _, at := range [][2]int{{0, 1}, {0, 0}, {1, 0}, {1, 2}, {1, 1}, {1, 1}} {
		w = w.Click(optionAt(w, at[0], at[1]))
	}
	if _, ok := f.gave("agent", "answer"); ok {
		t.Fatal("the session was answered before Send")
	}
	var chosen []string
	for _, row := range w.QuestionRows() {
		if row.Chosen {
			chosen = append(chosen, fmt.Sprint(row.Question, row.Option))
		}
	}
	if !slices.Equal(chosen, []string{"0 0", "1 0", "1 2"}) {
		t.Errorf("chosen %q", chosen)
	}
	w = w.Choose(SendSays)
	in, ok := f.gave("agent", "answer")
	if !ok || in.Given["session"] != questioning.ID || !slices.Equal(in.Listed["choice"], []string{"Postgres", "Rollback, Timings"}) {
		t.Errorf("the answer gave %v %v", in.Given, in.Listed)
	}
	if w.Screen != board.OnTheBoard {
		t.Errorf("the dialog stays open: %v", w.Screen)
	}
}

func TestQuestionsAreAnsweredFromTheMenuAndCancelledForTheTerminal(t *testing.T) {
	f := newFake()
	w := agentsWindow(f, pagesWith(true, true, true), questioning)
	w = w.ContextAt(onSession(w, questioning.ID))
	if slices.Contains(w.Menu(), AllowSays) || !slices.Contains(w.Menu(), AnswerSays) {
		t.Fatalf("the menu offers %q", w.Menu())
	}
	w = w.fromMenu(slices.Index(w.Menu(), AnswerSays))
	w = w.Click(optionAt(w, 0, 0)).Choose("Cancel")
	if _, ok := f.gave("agent", "answer"); ok || w.Screen != board.OnTheBoard {
		t.Errorf("Cancel answered, or left %v", w.Screen)
	}
}

func TestACardIsAsTallAsTheTallestOfItsRow(t *testing.T) {
	tall := asking
	tall.Task = &agent.Held{Number: 3, Title: "Buy the relay", Open: []string{"one", "two"}}
	tall.Branch, tall.Doing = "main", "Reads card.go"
	short := loose
	var listed []agent.Agent
	for i, one := range []agent.Agent{tall, short, short, short, short, short} {
		one.ID, one.State, one.Asks, one.Waiting = fmt.Sprintf("s%d", i), agent.Working, nil, ""
		listed = append(listed, one)
	}
	w := agentsWindow(newFake(), pagesWith(true, true, false), listed...)
	laid := w.AgentsLaid()
	cards := laid.Cards[0]
	across := 0
	for across < len(cards) && cards[across].Y == cards[0].Y {
		across++
	}
	if across < 2 || across == len(cards) {
		t.Fatalf("%d cards across of %d", across, len(cards))
	}
	for i, box := range cards {
		row := i / across
		if box.H != cards[row*across].H {
			t.Errorf("card %d is %d high, its row's first %d", i, box.H, cards[row*across].H)
		}
	}
	first, second := cards[0], cards[across]
	if second.H >= first.H {
		t.Errorf("a row of short cards is %d high, the row of a tall one %d", second.H, first.H)
	}
	for i, one := range listed {
		parts := w.AgentPartsOn(cards[i], one)
		if end := parts.Buttons[0].Y + parts.Buttons[0].H; end > cards[i].Y+cards[i].H || cards[i].Y+cards[i].H-end > 3*geom.Gap {
			t.Errorf("card %d ends at %d, its buttons at %d", i, cards[i].Y+cards[i].H, end)
		}
	}
	if gap := cards[1].X - cards[0].X - cards[0].W; gap != second.Y-first.Y-first.H {
		t.Errorf("the cards stand %d apart across and %d down", gap, second.Y-first.Y-first.H)
	}
}

func TestACardSaysForHowLongEvenUnderAMinute(t *testing.T) {
	now := time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC)
	for since, want := range map[time.Duration]string{20 * time.Second: "for <1m", 2 * time.Minute: "for 2m", 3*time.Hour + 22*time.Minute: "for 3h22m"} {
		if got := (AgentCard{Agent: agent.Agent{Since: now.Add(-since)}, Now: now}).since(); got != want {
			t.Errorf("%v: %q, want %q", since, got, want)
		}
	}
}

func resumeMark(w Window, a agent.Agent) (CopyMark, bool) {
	for _, mark := range w.CopyMarks() {
		if mark.Said == a.Resume {
			return mark, true
		}
	}
	return CopyMark{}, false
}

func TestTheMarkAfterASessionsIdCopiesTheCommandThatResumesIt(t *testing.T) {
	var many []agent.Agent
	for i := range 30 {
		one := loose
		one.ID, one.Short, one.Resume = fmt.Sprintf("s%02d", i), fmt.Sprintf("s%02d", i), fmt.Sprintf("claude --resume s%02d", i)
		many = append(many, one)
	}
	w := agentsWindow(newFake(), pagesWith(true, true, false), many...)
	first, last := many[0], many[len(many)-1]
	mark, ok := resumeMark(w, first)
	meta := partsOf(w, first).Meta
	if !ok || mark.At.Y < meta.Y || mark.At.Y+mark.At.H > meta.Y+meta.H || mark.At.X+mark.At.W > meta.X+meta.W {
		t.Fatalf("the mark of the first card stands at %v, its meta row at %v (%v)", mark.At, meta, ok)
	}
	if _, ok := resumeMark(w, last); ok {
		t.Error("the mark of a card out of sight can be clicked")
	}
	w = w.Scroll(centre(w.AgentsLaid().Shown), 10000)
	if mark, ok = resumeMark(w, last); !ok {
		t.Fatal("the last card has no mark, scrolled to it")
	}
	w = w.Click(centre(mark.At))
	if said, ok := w.Copying(); !ok || said != last.Resume || w.Telling != "copied "+last.Resume {
		t.Errorf("copied %q and said %q", said, w.Telling)
	}
}

func sessionsNamed(n int, from agent.Agent) []agent.Agent {
	var many []agent.Agent
	for i := range n {
		one := from
		one.ID, one.Short, one.Resume = fmt.Sprintf("s%02d", i), fmt.Sprintf("s%02d", i), fmt.Sprintf("claude --resume s%02d", i)
		many = append(many, one)
	}
	return many
}

func keyed(w Window, keys ...string) Window {
	for _, key := range keys {
		w = w.Key(key)
	}
	return w
}

func TestTheKeysWalkTheAgentsPage(t *testing.T) {
	w := agentsWindow(newFake(), pagesWith(true, true, false), sessionsNamed(6, loose)...)
	across := 0
	for _, box := range w.AgentsLaid().Cards[0] {
		if box.Y == w.AgentsLaid().Cards[0][0].Y {
			across++
		}
	}
	for _, c := range []struct {
		keys []string
		want string
	}{
		{[]string{"right"}, "s00"},
		{[]string{"right", "right"}, "s01"},
		{[]string{"l", "l", "h"}, "s00"},
		{[]string{"right", "down"}, fmt.Sprintf("s%02d", across)},
		{[]string{"j", "k"}, "s00"},
		{[]string{"right", "esc"}, ""},
	} {
		if got := keyed(w, c.keys...).sessions.on; got != c.want {
			t.Errorf("%q chooses %q, want %q", c.keys, got, c.want)
		}
	}
}

func TestTheChosenSessionIsKeptInSight(t *testing.T) {
	w := agentsWindow(newFake(), pagesWith(true, true, false), sessionsNamed(30, loose)...)
	w = w.Key("right")
	for range 30 {
		w = w.Key("down")
	}
	box, shown := sessionBox(w, w.sessions.on), w.AgentsLaid().Shown
	if w.sessions.on != "s27" && w.sessions.on != "s28" && w.sessions.on != "s29" || box.Y < shown.Y || box.Y+box.H > shown.Y+shown.H {
		t.Errorf("%s stands at %v out of %v", w.sessions.on, box, shown)
	}
	for range 30 {
		w = w.Key("up")
	}
	if box, shown := sessionBox(w, w.sessions.on), w.AgentsLaid().Shown; box.Y < shown.Y {
		t.Errorf("%s stands at %v above %v", w.sessions.on, box, shown)
	}
}

func TestTheKeysActOnTheChosenSession(t *testing.T) {
	near := asking
	near.Terminal.Reachable = true
	for _, c := range []struct {
		name  string
		pages engine.Pages
		one   agent.Agent
		key   string
		check func(*testing.T, *fake, Window)
	}{
		{"enter goes to it", pagesWith(true, true, false), near, "enter", func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "go"); !ok || in.Given["session"] != near.ID {
				t.Errorf("went with %v", in.Given)
			}
		}},
		{"o opens its task", pagesWith(true, true, false), near, "o", func(t *testing.T, _ *fake, w Window) {
			if w.Screen != board.OnForm || w.Chosen != 3 {
				t.Errorf("screen %v, task %d", w.Screen, w.Chosen)
			}
		}},
		{"c copies the command that resumes it", pagesWith(true, true, false), near, "c", func(t *testing.T, _ *fake, w Window) {
			if said, ok := w.Copying(); !ok || said != near.Resume {
				t.Errorf("copied %q", said)
			}
		}},
		{"a allows", pagesWith(true, true, true), near, "a", func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != agent.Allow {
				t.Errorf("answered %v", in.Given)
			}
		}},
		{"d denies", pagesWith(true, true, true), near, "d", func(t *testing.T, f *fake, _ Window) {
			if in, ok := f.gave("agent", "answer"); !ok || in.Given["answer"] != agent.Deny {
				t.Errorf("answered %v", in.Given)
			}
		}},
		{"a answers nothing while the window answers no permission", pagesWith(true, true, false), near, "a", func(t *testing.T, f *fake, _ Window) {
			if _, ok := f.gave("agent", "answer"); ok {
				t.Error("answered")
			}
		}},
		{"a opens the questions", pagesWith(true, true, true), questioning, "a", func(t *testing.T, _ *fake, w Window) {
			if w.Screen != board.OnAsking || w.Question != board.AboutAnswering {
				t.Errorf("screen %v, question %v", w.Screen, w.Question)
			}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := newFake()
			w := agentsWindow(f, c.pages, c.one).Key("right")
			c.check(t, f, w.Key(c.key))
		})
	}
}

func TestTheBarAndTheHelpSayTheKeysOfTheAgentsPage(t *testing.T) {
	w := agentsWindow(newFake(), pagesWith(true, true, true), asking).Key("right")
	want := map[string]string{GoToSessionSays: "enter", OpenTaskSays: "o", AllowSays: "a", DenySays: "d", CopyResumeSays: "c"}
	for offer, key := range want {
		if !slices.Contains(w.Offers(), offer) || w.shortcut(offer) != key || w.BarKey(offer) != w.spelled(key) {
			t.Errorf("the bar offers %q, %s on %q", w.Offers(), offer, w.shortcut(offer))
		}
	}
	if slices.Contains(agentsWindow(newFake(), pagesWith(true, true, false), asking).Key("right").Offers(), AllowSays) {
		t.Error("the bar offers Allow while the window answers no permission")
	}
	var helped []string
	for _, line := range w.Helping() {
		helped = append(helped, line.Key)
	}
	for _, key := range []string{"hjkl", "⏎", "o", "a d", "c"} {
		if !slices.Contains(helped, key) {
			t.Errorf("the help says no %q: %q", key, helped)
		}
	}
}

func TestASilentSessionSaysSoAndWhenItLastWrote(t *testing.T) {
	now := time.Date(2026, 11, 26, 9, 30, 0, 0, time.Local)
	silent := loose
	silent.Since, silent.Silent = now.Add(-20*time.Minute), now.Add(-12*time.Minute)
	if got := (AgentCard{Agent: silent, Now: now}).since(); got != "silent for 12m" {
		t.Errorf("the card says %q", got)
	}
	w := agentsWindow(newFake(), pagesWith(true, true, false), silent)
	w.Now = now
	if got := w.Hovering(centre(partsOf(w, silent).State)).Bubble(); got != "Last written at 09:18" {
		t.Errorf("the bubble says %q", got)
	}
}

func TestTheContextSaysWhenToCompact(t *testing.T) {
	for percent, want := range map[int]string{72: "Context 72%", 87: "Context 87% · compact soon"} {
		if got := contextSays(agent.Agent{Context: percent}); got != want {
			t.Errorf("%d%%: %q, want %q", percent, got, want)
		}
	}
}
