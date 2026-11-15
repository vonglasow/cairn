# Decisions

How cairn is built, one line per decision: what was decided — why. The
decisions a person using cairn meets are in [docs/decisions.md](docs/decisions.md).
A decision added, changed or undone is a line here, in the commit that does it.

## Shape

- One notion, one package under `engine/`; store and files under `store/` — a notion is read whole in one folder.
- The declarations of the operations stay in the catalogue; what reads several notions stays in `engine` — a notion declaring its own would import `Answer`, which imports every notion: a cycle.
- A notion hands over its facts as one record (`Carried`), not one method each — eighteen membranes copied the same fields, and each fact added touched all of them.
- Every change goes through `Ask`; a surface decides nothing — the rules live once.
- What the boards show is computed once in `internal/board` — the window and the terminal each worked out a card's facts, and drifted.
- A card's facts are one list, each a word, a mark and a tone — a fact one board showed and the other did not was a fact one surface forgot.
- Both surfaces show the same thing, made from one source — the user's rule, 2026-11-14.
- Menus, the palette and the help are made from the catalogue, what loses work last — a hand-written list missed new operations and kept turned-off ones.
- Forms and settings are made from an operation's fields and the configuration keys, with today's words and order — one list of fields, not one per board.
- What asks first is the engine's (`Operation.AsksFirst`); `--yes`, a question and an agent's access are how each surface confirms — the boards asked before a sweep and the command line did not.
- What kind of answer a failure is, is the engine's (`RefusalOf`) — a script's exit code and an agent's refusal read one list.
- Upgrade is asked from the command line and the window, never over MCP — it replaces the binary an agent runs from.
- Drafts, window sizes and recent workspaces are written through operations no command or tool lists — a surface writes nothing behind the engine.
- Files on disk are written by `store/files` only; their paths and formats stay.
- The Claude sessions holding a task are the hook's, not the engine's — cairn never reads a task's notes.
- `--json` is a flag of every command — one flag at the root reads the same everywhere.
- Several tasks are deleted or archived in one act, under `--also` — three asked one by one left two gone when the third was refused.
- `engine/item`, `engine/calendar` and `engine/change` keep their names; `domain` is gone — a name says what a package holds, never a layer.
- gui splits into `geom`, `design`, `motion` only; tui into `geom`, `design` — components and screens draw with the board's fields, and moving them moved the board.
- The window answers the pointer from the boxes it lays out (`Placed`, `CardRowOf`, `PillRowOf`, `FormPlace`), not from `Drawn` — `Drawn` names a part, not what it stands for, and waits on a frame.
- The window's menu bar holds cairn and Help — it repeated every key, the right click and the palette, and drifted.
- The band is three lanes under the last three columns, on both boards — one walk for both.
- The terminal has no Normal and Insert modes: hjkl walks with `board.Walk`, `^H` `^L` reach another column, `^J` `^K` the band — the window has no modes.
- The terminal wears the window's scheme, not the terminal's sixteen colours — both boards show the same.
- The terminal edits text with the board's text model — one way to write, on both boards.
- The form's blurred fields keep huh's white and black — huh hands over no style for them.
- A notion never imports another, held by depguard deny-lists; the leaf packages by allow-lists — a deny-list forbids only what somebody thought to name.
- A package added is a depguard rule in the same commit (`rule-check`) — `logo` and `macapp` once went in without one.
- No comments but one line per exported symbol; no prose in the scenarios — a quarter of the code was comments, much of it history.
- A Unicode Cf character is refused when written and replaced when shown — a list of characters let the rest through.
- A tool whose answer shape cannot be derived is offered without one — a test proves every tool derives one.
- Lines are not the first measure; a sound split comes first — a library may replace code of ours.

## Store

- SQLite in Go, one file — one static binary, no cgo for the store.
- WAL and a five-second wait — measured: without the wait four writers lost 14 writes of 20; without WAL six took the whole wait.
- Every write is one transaction, every migration one per change — an act half kept is worse than none.
- No index on the tasks table — a store holds hundreds of tasks; one is added the day a listing is measured slow.
- Integrity between tables is kept by the code, not foreign keys — SQLite leaves them off per connection.
- Closed sets are refused in `engine/item`, not by a `CHECK` — one place says what a column is.
- Opening never migrates — `task list` rewriting a database as a side effect is not a listing.
- A change of data alone is forgotten, not refused, when rolled back — refusing it would wall off every older shape.
- Migrations run in order, and a rollback reads the ledger again in its transaction — two rollbacks read at once both undid the same change.
- A database holding tables, none a store's, is refused before anything is written — a mistaken path leads to somebody else's data.
- A backup SQLite cannot read whole is refused as a restore — it trades a store that opens for one that does not.
- A held engine follows a store put back under it — measured: a board kept writing to the replaced file, without an error.
- Looking at a backup opens it read-only — opened the shared way, it was rewritten and its date moved.

## Tests

- Behaviour is tested once, in table tests in the engine and the board.
- A scenario is kept when it reaches a line nothing else reaches, plus one per rule — the scenarios repeated the engine's tests; the plugin's, session's and version's scenarios test shell and stay.
- A unit test is kept the same way, the engine's and the board's chosen first; a surface keeps what alone reaches its lines — the floors come before one contract test per surface.
- A scenario marked `@later` is gone — no suite ran it.
- Scenarios may be rewritten; what a person sees stays.
- A program a test runs is written by `test.WriteProgram`, under `syscall.ForkLock` — a child forked while the file was open kept it busy.
- The session hook's idle delay is a `sleep` the test stands in for — the scenarios slept it out, and a busy machine needed more.
- "is not told the tools changed" still waits 300 ms — no answer proves a notification is not coming.
- A form's commands are awaited in the terminal tests, all but the cursor's blink — racing each against 5 ms dropped one on a busy machine (cairn#298).
- `test` imports `testing` outside a test file; `harness-check` keeps it out of the binary — Go compiles a package's test files only for that package.

## Checks

- What the tools find stays current; the tools are pinned — a tool changing itself finds nothing.
- No threshold is lowered — a gate that bends is not a gate.
- Mutation is a report run by hand on the core, never on the surfaces, one package at a time — the sweep kept a push over an hour waiting, and the surfaces gave mutants no test tells apart.
- Linux is checked by hand, in a container, before a version — it needs Docker, which no hook may assume.
- The container image is named by tag, not digest; commitlint's dependencies are not locked — a check run by hand, and one hook not worth a lockfile.

## Release

- The version is read off the commits by svu, tagged by `make release` once every check passed — a number typed by hand misreads the history.
- A release runs every check but the mutation sweep — a release waiting an hour on it was a release nobody made.
- Each binary is built on the system it runs on; no source archive — the window needs cgo and the system's headers.
- The release stays on its scripts, not GoReleaser — gathering several runners' builds is GoReleaser Pro.
- The plugin is installed from GitHub at the binary's tag; cairn says the commands, never runs them — the plugin is the one the version was made with.
- A study and what it decides land in the same version — #222 and #226 landed a version apart.

## Dependencies

A dependency added is a line here, in the same commit; one whose reason no
longer holds is taken out.

- `modernc.org/sqlite` — the store: SQLite in Go, no cgo.
- `spf13/cobra`, `spf13/pflag` — the command tree read off the catalogue; pflag's own error tells `-30m` from three flags.
- `charmbracelet/fang` — styled help, versions and refusals over cobra.
- `modelcontextprotocol/go-sdk` — the MCP protocol: 35 lines against 250–350 by hand.
- `google/jsonschema-go` — each tool's answer schema, read off the Go type that crosses the wire.
- `pelletier/go-toml/v2` — the configuration file carries comments; strict decoding names an unknown key.
- `go.yaml.in/yaml/v3` — base16 schemes are YAML, written by others; only `internal/theme` imports it.
- `charm.land/bubbletea`, `lipgloss`, `huh`, `bubbles` — the terminal board: its loop, drawing in cells, its form, scrolling and keys.
- `gioui.org` — one vector-drawn window on every system, in the one binary; held at fd8b8788 past v0.10.3, whose key release spent a dead key on Linux.
- `gioui.org/x` (`explorer` only) — the file dialog on Windows and Linux; macOS opens AppKit's panels itself, to link `-lobjc` once.
- `golang.org/x/image` — the fixed-point numbers Gio measures text in.
- Figtree (SIL OFL 1.1), embedded — the window's face at three weights; figures stay in Go Mono.
- `yuin/goldmark` — the site's Markdown, its anchors and links checked; not in the binary.
- `cucumber/godog` — runs the scenarios of `bdd/`.
- `rsvg-convert`, `tc-hib/go-winres` — draw the logo when it changes, never in a build; what they make is committed.
- Not added: `teambition/rrule-go` — the words typed are not RRULE, the subset is two hundred lines, no release since 2023.

## Docs

- `docs/` is for the person using cairn: guides by task, a glossary, and a reference generated from cairn — a key described twice drifted.
- The developer reads `ARCHITECTURE.md` and this register; `.claude/docs` is gone — twelve thousand lines told the history of each choice, and drifted from the code.
- One word for each notion on every surface: series, rule, occurrence, trace, column, deadline, notes, theme, settings, tab, report (#292) — one notion was named three ways.
