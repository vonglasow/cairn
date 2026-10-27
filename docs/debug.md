# Debug mode

Read this before touching what debug mode writes down, or where.

`bdd/debug.feature` says what happens; this file keeps the reasons.

## What it is for

A bug seen on somebody else's machine is described in words, and the words
leave out the one thing that mattered. Debug mode is for that person: run
cairn with it on, do the one thing that goes wrong, and send what was written
down. Nothing about it is for the person reading the code here; `make trace`
was, and asked for a clone, a Go toolchain and a script.

## It is turned on by the environment, at launch

```sh
CAIRN_DEBUG=1 cairn gui
```

The variable and nothing else (`@debug-on-writes-a-file-beside-the-store`,
`@debug-off-writes-nothing`). A flag would have to be passed to every command,
and the MCP server is started by Claude Code from the plugin's `.mcp.json`,
where an environment is what can be set. A key in the configuration would be
forgotten on, and the configuration is for lasting choices. And it is decided
at launch, never from a board while it runs: what went wrong before it was
turned on would be missing from the file, and so would the way the machine was
when it started.

## One file per launch, beside the store

Each launch writes a file of its own (`@debug-each-run-writes-its-own-file`)
in `debug/`, beside the store — `~/.cairn/debug/` by default, where the store,
the notes, the plugin, the draft and the window's sizes already are, so one
folder is all a person has to know. The file is named by the moment it opened
and the command it ran, `20261026-140304-task-list-….log`, and its lines are
text, read by eye.

It is held to a size: a file is cut at 10 MB, and says so on its last line;
past twenty files, the oldest goes (`@debug-keeps-the-last-twenty-files`). A
mode somebody turns on to show one thing is a mode somebody forgets to turn
off.

Where the file is, is said on standard error on the way out,
`debug: written down in …`. Not on the way in: the terminal board covers the
screen, and the MCP server's standard output is its protocol.

A folder that cannot be written in does not stop cairn, and is said on
standard error, `debug: cannot write down in …`
(`@debug-a-folder-that-cannot-be-made-is-said`): debug mode is there to show
what went wrong, not to be one more thing that does — and a person who thinks
it is on when it is not sends nothing worth reading.

## The file begins with what the machine is

The first lines, each `# name value`, say what nobody thinks to say in a bug
report (`@debug-the-file-begins-with-the-machine`, `engine.DebugHeader`,
`debug.Machine`): the build (version, commit, Go), the system, on a Mac its
version, on Windows its own, on Linux and the BSDs the distribution, the
display session, the desktop, the input methods and the keyboard — read from
the environment, and from `pgrep`, `setxkbmap` and `sw_vers` when they are
there, `not found` when they are not; then the language and the terminal.
These are what `make trace` wrote down and what found the bug where IBus
swallowed the first accented letter ([gui.md](gui.md)).

Then the configuration as `config show` says it, with the home written `~`,
so a user name does not travel with the file; the store in numbers — tasks per
column, archived, objectives, labels, how many changes of its shape are applied
— never what it holds; and the command, its words and flag names as they are
and what was typed after them masked.

## Every operation, written down once, by the engine

Every surface asks the engine, so the engine is where each act is written
down, once, whatever asked it (`@debug-an-operation-is-written-down`,
`engine.NoteTo`): `task move after=2 number=4 → ok tasks 3 12ms`. What it was
given is written with what it is — a number, a day, a duration, a value of a
fixed list such as a column or a priority are plain; anything else is masked —
and what it came to is a summary: `ok` and what the answer holds, counted
(a task's number and column, how many tasks, what was deleted), or the refusal
and its reason (`@debug-a-refusal-is-written-down`). A failure nobody foresaw
is written with the chain of its error types and its message masked: the
message may carry a title, and the types are often enough.

A board asks for the version of the store every two seconds, to see what
changed elsewhere. That is written only when it answers another version: the
same answer a thousand times would bury the rest.

## Each surface says what only it saw

The window writes every press, drag, release and scroll, with what it made of
it (`PointerTrace`) — never the pointer passing over, which is most of what a
pointer does — every key, and what it shows each time that changes: the
screen, the page, the card chosen (`Seen`). The terminal writes every key,
every resize and what it shows each time that changes. The MCP server writes
the client, once, and each method it is asked; the tools it runs are the
engine's to write. A panic is written with its stack, then left to stop cairn
the way it did: debug mode watches, it does not change what happens.

## Nothing a person wrote reaches the file

Every text a person wrote is masked, always: titles, descriptions, steps,
labels, searches, the command's values (`@debug-written-text-is-masked`). A
text is written `‹text 3f9a·14›`: the first two bytes of a SHA-256 of the text
and a salt drawn for that file, and its length. The same title reads the same
all through one file, so a reader can follow one task from act to act, and no
other file shares the salt, so two files cannot be laid side by side to read a
title back.

What is typed is written by its length alone, `‹typed 1›`. A fingerprint per
key would be the same for the same letter all through the file — a
substitution, and a substitution is read back by anybody patient.

There is no way to lift the mask. A file a person can send without reading it
first is the only kind a person will send.

## A report packs it all

`cairn debug report` (and the MCP tool `debug_report`) writes one zip archive
in the current folder, or the one named with `--to`: every file the debug
folder keeps, under `debug/`, and `header.txt`, what the machine and the store
are at the moment it is made (`@debug-report-packs-the-files`). A zip because
every system opens one without installing anything. It is an operation like
any other, so an agent that sees cairn misbehave can make it and say where it
is; it writes a file, so a server started to read only does not offer it.

## What it replaced

`make trace` and `CAIRN_GUI_TRACE` wrote the window's pointer and keyboard
events to standard error, through a script that needed a clone and Go to run.
Debug mode writes the same events, from the binary anybody installed, for
every surface — and the script and the variable are gone, so there is one way
to ask.
