# What a task carries from outside

Read this before touching what `task show` says about a task's outside context,
or before changing where that context is looked for.

A task carries more than cairn holds. Meeting notes, a supplier's contract, what
a working session decided — written elsewhere, in whatever form came to hand.
This file says how cairn points at that, and why it points rather than reads.

The scenarios are in `bdd/context.feature` — all but one run in `internal/cli`,
and the one about the server in `internal/mcp`. Where cairn finds the root they sit
under is a separate matter, in [config.md](config.md) and `bdd/config.feature`.
What is settled here is settled with its reason, because the reason is the part
a scenario cannot carry.

**Showing a task is what answers**, and no tool of its own was added. Which
surface answered stayed open while the answer was still going to be a listing,
because a listing is work and work deserves its own name. The answer turned out
to be a path composed from a number, so a tool of its own would cost a round trip
to say what showing a task already had every means to say. `task list` and `task
search` stay silent — a place on every line would drown what they are for.

## Cairn says where, and nothing else

Showing a task answers a location: the directory a task's outside context lives
in. Cairn does not open it, does not create it, and does not look to see whether
it is there.

The line is not fussiness. A cairn that read those files would decide what is
worth reading, how much of it, and in what order — and that is a harness's job,
not a task manager's. Cairn plugs into a harness over MCP, and any harness will
do. The moment it reads content, it is competing with the thing it plugs into.

| who | what |
|---|---|
| cairn | says where |
| the harness | looks, reads, decides what to open, writes |

The consequence worth stating: nothing here changes what cairn does with a task.
It changes what a caller can find out about one.

## The path is the link

There is no link to store. A task's context lives in a directory named after the
task — its number, the one cairn already gave it and never gives again — under a
root the person configures once. The path is derived rather than recorded.
Nothing can go stale, because nothing was written down.

`notes` is one of the two keys in `~/.config/cairn/config.toml`, and it is the
only thing cairn knows about any of this. A machine that never set it is not a
machine with no notes: said nowhere means `~/.cairn/notes`, beside the store, so
the tree has a place from the first command. [config.md](config.md) says how
that place is found, and what the environment does to it.

Material can exist before the task does. It does not need a holding pen for
that: a session that produces something worth keeping creates the task there and
then and writes into its directory, and a person writing by hand puts it in a
directory with a name that means something to them. Renaming that directory to
the task's id is what attaches it — a person's act, reversible, and visible in a
listing rather than buried in a field.

That was the second design, and the first is worth saying because it is the one
most people reach for. A link was going to be an opaque locator the writer chose
— a file, a note in some application, anything — with cairn storing the string
and resolving none of it. It is not needed. The convention turned out strong
enough that the link never had to exist as an object, and an object that need not
exist is one nothing has to keep true.

## It never looks, so there is nothing here that can fail

A place that holds nothing is answered exactly like a place that holds a year of
notes. Cairn cannot tell them apart, and does not try.

**Looking was the design before this one, and it was dropped.** It had cairn stat
the directory and list it, and it needed a table of seven answers to do it — a named root missing from
disk told apart from a default root nobody has written under yet, a directory
absent, a directory unreadable, a dreaming absent, a task closed, a task unknown
— each with its own refusal and its own scenario. All of it to spare the caller a
`readdir` it was about to perform anyway, because a caller told that a note exists
still has to open it. The work was real and what it bought was not.

What it would have cost beyond the scenarios is the part worth writing down: the
engine does no filesystem work today, and the store is the only thing it opens.
Reaching outside it for a `stat` puts a class of failure into an operation that
currently has none, and every one of those failures then has to mean something
to a person and to a machine both.

**The absence of a signal is not a gap to fill later.** It is tempting to answer
`exists: false` — it costs one syscall and it reads as helpful. It is a claim
about the disk made at the moment of the answer and consumed later, by which
time a parallel session may have written there. A caller that looks for itself
is never wrong; a caller told "nothing there" stops looking, and is wrong exactly
when it matters.

The path comes back whether or not anything is there, because that is where
whoever writes next will have to create. Answering "no such directory" would make
the caller compose the path itself, and then the layout would be known in two
places.

One refusal remains, and it is not about the disk: a task cairn does not hold has
no number to compose a path from.

## What makes an agent go and read is not cairn's to give

The reason for saying so here is that it is the first thing anyone asks of this,
and the answer is not obvious enough to leave unwritten.

Nothing in a tool description can oblige a model to open a path it was handed. A
schema describes; it does not command. So the guarantee that what is in the tree
gets read cannot come from cairn under any design — not from a listing, not from
a presence flag, not from wording. It comes from the harness, which can simply
look, every time, without being asked: a hook reads where a model chooses.

That is why the path is answered unconditionally rather than only when something
is there. A signal that appears on every task and pays on a few would teach a
model to stop following it, which is the failure this would have been built to
avoid. Answered to a harness that always looks, it costs nothing and teaches
nothing, because no judgement is being asked of the reader.

## One place per task, and three reserved names

Everything from here to the end is convention, and cairn knows none of it. It
composes a path and stops; the shape below is what a harness writing into that
path honours, and what the next one will have to honour to read what the last
one left. It is written in cairn's documentation rather than a harness's because
the path is the one thing the two share, and a convention kept beside only one
of its users is a convention the second one invents differently.

Nothing here can be enforced, and none of it needs to be: a tree that does not
follow it is still answered, still readable, still nobody's error.

```
<notes>/
├── 14/
│   ├── dreaming.md                     the entry point
│   ├── versions/
│   │   ├── dreaming-2026-08-21.md      what it said before
│   │   └── dreaming-2026-08-12.md
│   ├── 2026-08-11-meeting.md
│   └── supplier-contract.pdf
├── 27/
│   └── 2026-08-25-session.md           notes, never consolidated
└── _archive/
    └── 8/                              closed, intact
```

The directory is the task's number and nothing else. It is already unique, never
reused, and needs no prefix to be told from anything: `_archive` is the one name
at this level that is not a number, and that is what tells them apart.

`dreaming.md`, `versions/`, `_archive/`. Everything else is free — any name, any
type, PDF beside Markdown beside plain text. A rule on file names would have to
be known by everyone who ever drops a file in there, including a person in a
hurry, and a rule like that is broken the first time it is inconvenient.

`versions/` sits inside the task and `_archive/` sits at the root, and they were
nearly the same word. They are not the same thing: one holds what a dreaming used
to say, the other holds a task that is over. Naming them alike cost a mistake
during design, in a room where both had just been defined, which is the cheapest
possible demonstration that it would cost more later.

Flat files named `dreaming-2026-08-12.md.bak` would remove a reserved name and
were turned down for what a reader sees: dead versions would sit beside real
notes, and telling them apart would need a rule on file names — the one thing
that was meant to stay free.

## The dreaming is read, the rest is only seen

`dreaming.md` is a consolidation of what is known about a task: the description,
the notes, what sessions decided. It is an entry point, not a self-sufficient
artifact — when something is missing from it, the answer is further down.

So a harness opening the directory reads the dreaming and takes only the names of
what sits beside it. One that loaded every note would hold the dreaming and its
own sources at once, which makes the dreaming a duplicate of what it was meant to
replace. Seen and not opened, a `2026-08-11-meeting.md` is known to exist and is
gone to when it is needed.

That is the whole point of the shape: resume the work, **and know where to go
down**. A reader that finds a claim no longer matching what it sees looks for the
trace in the notes, and asks the person when it cannot find it.

There is exactly one dreaming per task. A session touching five tasks is five
dreamings, not one.

## Capturing and consolidating are different acts

**A capture is cheap and cannot be wrong.** A dated file, added, replacing
nothing. It is written at the end of a session by a hook, or during one when
something has just been established.

**A consolidation is asked for.** It reads the task's description, everything in
the directory, and the session it runs in; it reads code only when told to. It
writes a new `dreaming.md` and moves the previous one into `versions/`. It
touches no note, no capture, and nothing in cairn.

Letting the hook consolidate was turned down. A hook runs inside the session that
is ending, after whatever compaction that session went through — the worst view
of the work anyone will ever have of it. Capture is safe there because it adds;
consolidation is not, because it decides. Splitting them means a degraded session
can never damage what was already consolidated.

A new version is written beside the old rather than over it, so the previous one
is readable next to the current one and the person keeps or discards. Under git
it would be readable too, but behind a diff — audit rather than choice. Nothing
forbids the tree being a repository; it is just not what carries the versions.

## What was weighed and turned down

**Mempalace.** It does not replace the tree, it adds to it: files that are not
text need a directory anyway, so there would be two stores to attach, list,
ratify and archive in step. Its strongest card, search across everything, only
pays on knowledge that outlives a task — and that is somewhere else.

**Apple Notes.** No addressable files, no container per task, no versioning at
all, and a write that fails without a word when nothing interactive is watching.

**Notion, Confluence, work-item attachments.** Unreadable while the network is
down, and an authentication to keep alive inside something that runs at the end
of every session. What they offer that the tree does not — native versions and
attachments — the tree already has.

None of these were ruled out for being remote or for being someone else's. Each
lost on a requirement that was written down before the comparison started.

## Where it is said, and what it is for

`task show` says it as a line apart from the facts, `context  <notes>/<n>`, and
`--json` and `task_show` as `"context"`. The boards say it too: a `CONTEXT`
panel on the terminal's open task, and a line at the foot of the window's,
beside its offers — not among the fields, where it cost the box of parts a row
on a window somebody works in.

A surface finds the notes the way it finds the store, and hands both to the
engine together (`engine.Places`). With no home to put the default under and
nothing naming them, they are left unsaid: a place relative to wherever a
command happened to run would send a harness to write where nobody looks.

**This is the journal.** What an agent learns while working — what was tried,
what was decided, what is left — is a capture: a dated file in this directory,
added beside the others. For weeks the board's descriptions had been carrying
it instead, dated paragraph after dated paragraph, eight thousand characters in
the longest; an edit rewrites a description whole, so every entry was a read and
a rewrite of all the others, and one careless one lost them. It went there
because nothing said where else to put it: this contract was written down and
nothing ran it, so no answer ever carried the place. `task show`'s description,
which an agent reads as the tool's, now says it in so many words. That does not
oblige an agent to write there — nothing in a description can — but it is the
first thing the agent is told, at the moment it looks at the task.

## Deliberately not there yet

- **Knowledge that outlives a task.** What was learned on one task and serves the
  next is a different notion with a different lifetime, and it is not this.
- **Reaching into another tool.** Nothing here resolves a locator, and adding one
  adapter is adding all of them.
- **What a dreaming contains.** Its form is the harness's business. Cairn does
  not know it is Markdown.
- **Consolidation on a trigger.** It is asked for. A consolidation nobody asked
  for is one nobody reads.
- **Objectives.** A dreaming belongs to a task. An objective has no directory,
  and nothing has asked for one.
