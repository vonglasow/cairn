package site

import (
	"fmt"
	"strings"

	"cairn/internal/theme"
)

func home(said Said, schemes []theme.Scheme) page {
	var body strings.Builder
	body.WriteString(`<section class="hero">
<h1>Your objectives and tasks on one board, wherever you work.</h1>
<p class="lead">A board in a window, the same board in your terminal, a command line for scripts, and a way in for your agent. One file, on your machine.</p>
<div class="cta"><a class="btn" href="` + downloadPage + `">Download for macOS</a><a class="btn ghost" href="` + downloadPage + `">Linux and Windows</a></div>
<pre><code>brew install vonglasow/tap/cairn</code></pre>
</section>
<section class="band">`)
	if said.Window != nil {
		fmt.Fprintf(&body, `<img class="shot" data-shot src="%s" alt="The board in a window">`, windowPicture(darkFirst))
	}
	body.WriteString(`<div class="schemes">`)
	for _, s := range schemes {
		fmt.Fprintf(&body, `<button class="scheme" data-pick="%s"><span class="sw">`, s.Slug)
		for _, slot := range []int{0x0, 0x2, 0x5, 0xD, 0xB, 0xA, 0x8} {
			fmt.Fprintf(&body, `<i style="background:%s"></i>`, hex(s.Slots()[slot]))
		}
		fmt.Fprintf(&body, `</span><b>%s</b></button>`, e(s.Name))
	}
	body.WriteString(`</div>
<p class="faint">Every scheme cairn carries. The window, the terminal and this site wear the one you pick.</p>
</section>
<section class="band grid2">
<div class="tile door"><span class="k">For you</span><h2>See the week, choose the next task, count the time.</h2><span>Tasks stand in columns, in the order you set. Objectives say what the work is for and how far along it is. Focus a task and its time runs.</span><a href="#features">What the board does →</a></div>
<div class="tile door"><span class="k">For your agent</span><h2>Your agent works the board. You decide.</h2><span>With Claude Code or any MCP client, an agent focuses the task you name, reads its notes, writes what it learned, and proposes what to do next.</span><a href="plug-in-an-agent.html">Plug in an agent →</a></div>
</section>
<section class="band grid3">
<div><b>Yours.</b> <span class="faint">One file and a folder of notes on your machine. No account, no server, nothing sent anywhere.</span></div>
<div><b>Facts, not scores.</b> <span class="faint">Each task says what it blocks, when it is due, how long it has waited. cairn ranks nothing: the choice stays yours.</span></div>
<div><b>Nothing forgotten.</b> <span class="faint">Every change is written down, who made it and when. Ask what happened to a task, or how long the week took.</span></div>
</section>
<section class="band" id="features">`)
	feature(&body, "Plan the week.", "Move the week's tasks into todo, in the order you will work them. Focus one when you start: it moves to current, and its time runs.", "plan-a-week.html", "Plan a week", drawn(said, columns, "The todo and current columns in the window"))
	feature(&body, "See where the week went.", "The time of each focus is counted. The week view shows, day by day, which tasks took it and how it compares with their estimate.", "trace.html", "The trace and the time", drawn(said, "time", "The week in the window: the time each day took"))
	feature(&body, "Follow what the work is for.", "An objective counts how far along the tasks serving it are, its sub-objectives included. Its report says what moved over a month, and the time it took.", "follow-an-objective.html", "Follow an objective", drawn(said, "objectives", "The objectives in the window, each with how far along it is"))
	feature(&body, "The same board in your terminal.", "Open it with <code>cairn tui</code>. <code>hjkl</code> to walk, a key for every act, the same cards and colours as the window.", "tui.html", "In the terminal", `<pre class="term">`+terminalHTML(said.Terminal)+`</pre>`)
	body.WriteString(`<div class="grid3">
<a class="tile" href="blocking.html"><b>Say what stops a task.</b><span>A task in the way, or a reason said. The board shows what each task blocks, and filters the ones nothing stops.</span><span class="k">cairn task edit 7 --blocker 4</span></a>
<a class="tile" href="series.html"><b>Work that comes back.</b><span>A series makes the next task on its day: the stand-up every morning, the plants every ten days after done.</span><span class="k">--every "monday 9:00"</span></a>
<a class="tile" href="context.html"><b>Notes beside each task.</b><span>A folder per task for what was done and decided, readable in any editor, or in Obsidian.</span><span class="k">cairn task show 12</span></a>
<a class="tile" href="listing.html"><b>Narrow, and keep the view.</b><span>Filter by column, label, objective, deadline; search; keep the filters under a name.</span><span class="k">cairn task list --unblocked</span></a>
<a class="tile" href="config.html"><b>Several boards.</b><span>One workspace for work, one for home. Each keeps its tasks, its notes and its settings.</span><span class="k">cairn workspace use home</span></a>
<a class="tile" href="usage.html#what-cairn-answers-a-script"><b>Scripts welcome.</b><span>Every command answers in JSON with --json, and its exit code says why it refused.</span><span class="k">cairn task list --json</span></a>
</div>
</section>
<section class="band start">
<h2>Five minutes to start.</h2>
<pre><code>brew install vonglasow/tap/cairn
cairn task new "Write the spec"     # task 1, in the backlog
cairn task focus 1                  # its time runs
cairn gui                           # the board in a window</code></pre>
<div class="cta"><a class="btn" href="` + downloadPage + `">Download</a><a class="btn ghost" href="` + docsPage + `">Read the docs</a></div>
</section>`)
	return page{Name: homePage, Title: "cairn", Wide: true, Body: body.String()}
}

func drawn(said Said, of, shows string) string {
	if said.Window == nil {
		return ""
	}
	return drawing(of, shows)
}

func feature(body *strings.Builder, title, says, to, named, picture string) {
	fmt.Fprintf(body, `<div class="feature"><div><h2>%s</h2><p class="faint">%s</p><a href="%s">%s →</a></div>%s</div>`, title, says, to, named, picture)
}

func agents(version string) page {
	body := `<section class="hero">
<span class="k">For your agent</span>
<h1>Your agent works the board. You decide.</h1>
<p class="lead">Claude Code, or any MCP client, reads your tasks, focuses the one you name, writes what it learned beside it, and proposes what comes next. Nothing it proposes is written without your yes.</p>
<div class="cta"><a class="btn" href="#with-claude-code">With Claude Code</a><a class="btn ghost" href="#with-any-mcp-client">With another MCP client</a></div>
</section>
<section class="band">
<div class="chat" aria-label="A Claude Code session">
<p><b>you</b> let's work on cairn#12</p>
<p><b>claude</b> Task 12 is in focus and moved to current. From its notes: the export writes one file per objective; the import is left. Two steps open. Next: read the import format.</p>
<p><b>you</b> what should I work on after that?</p>
<p><b>claude</b> cairn#7 — Write the install guide: frees 3 tasks, due Friday, 2h estimated.<br>cairn#4 — Review the packaging spec: its objective is late.<br>Shall I move 7 to the head of todo?</p>
</div>
<p class="faint">The session holds the task in focus while you work: its time is counted, and the next session starts where this one stopped.</p>
</section>
<section class="band grid3">
<div><b>It starts with the context.</b> <span class="faint">Name a task and the agent reads its notes, its steps, where it stands. You do not repeat yourself.</span></div>
<div><b>It writes where the next one reads.</b> <span class="faint">What was done or decided goes in the task's notes, dated. Never lost in a chat.</span></div>
<div><b>It reaches what you allow.</b> <span class="faint">Read only, or also write, or also delete: you set it once. Everything it did is in the trace, in its own name.</span></div>
</section>
<section class="band grid2">
<div><h2 id="with-claude-code">With Claude Code</h2>
<p>cairn carries a Claude Code plugin. It ties a session to the task it serves, holds the task in focus while you work, and lets it go when you stop. It needs <code>jq</code> on your <code>PATH</code>.</p>
<p>Then say, in your own words:</p>
<ul><li>“let's work on cairn#12” — it focuses the task and reads its notes;</li><li>“what should I work on?” — it weighs the board and proposes;</li><li>“where do things stand?” — it reads back what is in focus and done;</li><li>“what did I do yesterday?” — it reads the trace and the time.</li></ul>
<a href="claude-code.html">With Claude Code, in the docs →</a></div>
<div><p>Print the commands for your version of cairn:</p>
<pre><code>cairn plugin show</code></pre>
<p>and paste them into Claude Code:</p>
<pre><code>/plugin marketplace add vonglasow/cairn@` + e(version) + `
/plugin install cairn@cairn</code></pre>
<p class="faint">After an upgrade, <code>cairn upgrade</code> prints the commands for the new version.</p></div>
</section>
<section class="band grid2">
<div><h2 id="with-any-mcp-client">With any MCP client</h2>
<p>Claude Desktop, an editor, your own client: it starts <code>cairn mcp serve</code> and speaks to it directly. No port is opened, nothing leaves your machine.</p>
<p>Every command is a tool of the same name: <code>cairn task new</code> is <code>task_new</code>. A refusal comes with its reason, so the agent corrects itself.</p>
<a href="` + toolsPage + `">Every tool, in the reference →</a></div>
<pre><code>{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "write"],
    "env": { "CAIRN_WORKSPACE": "work" }
  }
}</code></pre>
</section>
<section class="band">
<h2 id="how-far-it-reaches">How far it reaches</h2>
<p class="faint">Set once, when the server starts. A tool past the level is not offered at all: an agent cannot be talked into what it does not have.</p>
<div class="tw"><table><tr><th>--access</th><th>the agent may</th></tr>
<tr><td><code>read</code></td><td>list and show, nothing else — the default</td></tr>
<tr><td><code>write</code></td><td>also add and change tasks, objectives and steps; export, import, make a backup</td></tr>
<tr><td><code>full</code></td><td>also what can lose work: delete, restore a backup</td></tr></table></div>
</section>
<section class="band start">
<h2>Plug it in now.</h2>
<div class="cta"><a class="btn" href="` + downloadPage + `">Download cairn</a><a class="btn ghost" href="plug-in-an-agent.html">Plug in an agent, step by step</a></div>
</section>`
	return page{Name: agentsPage, Title: "For your agent", Top: "agents", Wide: true, Body: body}
}

func docsHome(pages []page) page {
	by := map[string]page{}
	for _, p := range pages {
		by[p.Name] = p
	}
	var body strings.Builder
	body.WriteString(`<p class="lead">How to do each thing with cairn, at the command line, on the board, or through your agent.</p>`)
	for _, group := range groups {
		fmt.Fprintf(&body, `<h2 id="%s">%s</h2>`, headingID(group), e(group))
		switch group {
		case "Guides":
			body.WriteString(`<div class="grid3">`)
			for _, s := range published {
				if s.group == group {
					fmt.Fprintf(&body, `<a class="tile" href="%s"><b>%s</b><span>%s</span></a>`, nameOf(s.from), e(by[nameOf(s.from)].Title), e(s.says))
				}
			}
			body.WriteString(`</div>`)
		case "Reference":
			body.WriteString(`<div class="tile"><span>Read off cairn itself, so it is always the version you run.</span><p>`)
			for i, name := range []string{commandsPage, toolsPage, keysPage, themesPage} {
				if i > 0 {
					body.WriteString(" · ")
				}
				fmt.Fprintf(&body, `<a href="%s">%s</a>`, name, e(by[name].Title))
			}
			body.WriteString(`</p></div>`)
		default:
			body.WriteString(`<ul class="rows">`)
			for _, s := range published {
				if s.group != group {
					continue
				}
				fmt.Fprintf(&body, `<li><a href="%s">%s</a>`, nameOf(s.from), e(by[nameOf(s.from)].Title))
				if s.says != "" {
					fmt.Fprintf(&body, ` <span class="faint">— %s</span>`, e(s.says))
				}
				body.WriteString(`</li>`)
			}
			body.WriteString(`</ul>`)
		}
	}
	return page{Name: docsPage, Title: "cairn docs", Top: "docs", Body: body.String()}
}
