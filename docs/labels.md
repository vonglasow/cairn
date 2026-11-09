# Labels

Read this before touching how a label is written, how one is put on a task or
taken off it, or how the window offers them.

A label is a word a task carries, as many as it likes. It means nothing to
cairn: it is not a notion, nothing is decided by it, and a narrowing is the only
thing that reads it (see [listing.md](listing.md)). What matters is that the
same word is always written the same way — two spellings of one word are two
labels, and a narrowing by one misses every task carrying the other.

## One way to write a label

The engine writes every label it is given the same way, whichever surface gave
it — the command line, the MCP server, the terminal or the window:

1. lower case;
2. the accents of the Latin letters taken off — `modularité` is `modularite`,
   and `ç`, `ñ`, `ü` are `c`, `n`, `u`;
3. spaces, `_` and `/` become `-`, a run of `-` becomes one, and none is kept
   at either end — `Sp26 20` is `sp26-20`;
4. what is left may hold `a`–`z`, `0`–`9`, `-` and `.`, and nothing else.

Anything else is refused, and the refusal names what a label may hold. A `:` is
refused with a word of its own: a label has no family. `ado:77594` is two
labels, `ado` and `77594`, and a task carries as many as it likes.

The rule writes rather than refuses wherever a writing is beyond doubt — case,
accents, separators — because a person typing `Proxmox` means `proxmox`, and a
refusal there would be a rule that only made typing slower. It refuses where
any writing would be a guess.

## What was written before the rule

A migration writes the labels already in a store the same way, a backup set
aside first as every migration does. A label carrying a family is split in two
at its `:` — `chantier:proxmox` becomes `chantier` and `proxmox` — and each half
is written by the rule. A task left carrying the same label twice carries it
once.

Splitting rather than joining was chosen: `chantier-proxmox` would have kept a
family under another spelling, and a label written as two words carries both
without a convention to remember.

The migration is SQL, like every change to a store, and SQL knows the accents
of the Latin letters it is told about and nothing more: a label carrying any
other is kept as it was. Such a label is carried as it is and never stops a
change to its task — a label joining is written by the rule, the labels already
there are left alone. Labels were already kept in lower case before the rule,
which is why SQL's lower case, ASCII only, is enough.

The trace keeps what was written when it was written. A label put on a task
before the rule reads there as it was spelt then; the migration rewrote the
labels, not their history, and writes nothing to it.

**A label is looked for the way it is kept.** A narrowing writes the labels it
is given by the same rule, so `cairn task list --label SP26_20` finds what
carries `sp26-20` rather than nothing.

## Putting labels on a task, and taking them off

On a change, a label given with no sign **joins** the ones the task carries:

```sh
cairn task edit 1 --label urgent            # one more, the others kept
cairn task edit 1 --label +urgent           # the same, signed
cairn task edit 1 --label -draft            # one fewer
cairn task edit 1 --label =spec,team        # the set becomes exactly this
cairn task edit 1 --label =                 # every label taken off
cairn task edit 1 --label ""                # the same: an attribute given empty is cleared
```

An `=` before the first value says what the set becomes, and the values after
it join that set rather than the old one. `+` or `-` beside an `=` is refused:
the change either replaces the set or moves it, never both. A bare value and a
signed one mix freely, since both work from what is there.

It was a bare value that replaced the set, once. In a real session one call
meant to add a label took five off a task: an agent and a person both read a
bare label as one more, whatever the help said. The reading that loses nothing
is the one kept; replacing is the one that has to be asked for.

The answer says what moved — `labels added`, `labels taken off` under the card,
and `added` and `taken_off` in the JSON — so a change that took off more than
was meant is seen at once, not found later.

zsh reads a word starting with `=` as the path of a command, and `--label =spec`
fails there before cairn sees it (`spec not found`). Quote it —
`--label '=spec'` — or join it to the flag, `--label==spec`; bash and the MCP
server take it as it is.

At creation there is nothing to add to: bare or after an `=`, the labels given
are the set, and a signed one is refused. The same grammar holds for
`--blocker`, `--link` and `--objective`, and for `objective edit --label`; the
forms of both boards send an `=`, since a form says what the set becomes.

## Putting labels on a task in the window

The form's `labels` field opens a list, the one objectives are chosen from:

- every label some task carries, in alphabetical order, those the task carries
  ticked;
- typing narrows the list;
- when what is typed, written by the rule, is no label yet, the first row offers
  to create it, under the name it will have — typing `Sp26 20` offers
  `sp26-20`, and typing a label that exists, however it is spelt, offers the
  label that exists instead;
- `⏎` on the row that creates a label creates it ticked and clears the line,
  and the list stays open for the next one, as it does for objectives;
- what the rule refuses — a `:`, a `#` — is said on that row instead, and
  `⏎` there creates nothing;
- `⏎` on a label ticks it or unticks it; `⌘⏎` keeps what is ticked in the
  field, `esc` keeps nothing;
- nothing is written until the form is saved, like every other field.

A label exists while a task carries it, or once it was made on its own. Taken
off the last task that carried it, a label nobody made on its own is no longer
offered; one made with `label new` stays, listed as carried by nothing, until
it is deleted (`@label-new-stands-on-its-own`,
`@label-new-stays-once-carried-and-dropped`). Those are kept apart, in a table
of their own (`labels_kept`, migration 16); rename and delete reach them as
they reach what is carried (`@label-new-goes-with-delete-and-follows-rename`).

## Every label at once

`cairn label list` lists every label, archived tasks included, with how many
tasks and how many objectives carry it; `label rename` writes another name on
everything carrying one, the new name written by the rule above, and a label
renamed onto one already there becomes that one; `label delete` takes one off
everything, and asks first, since what it takes off is put back by nobody. Each
change is an edit of its task or objective, so a task's trace says it. The
window reaches the same four from `⇧L` and the File menu: the list, each label
a pill with what carries it at the right; `n` names a new one in a row of its
own at the top (`@gui-a-label-is-made-from-the-labels`); `r` renames the one
the keyboard is on in its own row, the panel keeping its name, `⏎` saving
(`@gui-a-label-is-renamed-in-its-row`); `⌘⌫` takes it off everything after
asking. The list opens with no label at all, to make the first. A label made
on its own is offered by a task's list of labels as any other
(`@gui-a-label-made-is-offered-in-a-form`).

Leaving these out was decided once, the labels being nothing but what the
tasks carry, and kept when managing them was first asked for. Amended
2026-11-01: a label could not be prepared before the work that would carry it
— a sprint's, a release's — so `label new` makes one on its own, and only
those are kept apart; every other label is still read off what carries it.

## A label wears a colour

A label is drawn in a colour of its own, one of fourteen: Catppuccin Mocha's
accents — rosewater, flamingo, pink, mauve, red, maroon, peach, yellow,
green, teal, sky, sapphire, blue, lavender (`domain.LabelColours`). They were
made to be read on a dark ground and are all of one weight, so no label
shouts over another; PICO-8's sixteen were the other candidate, and four of
them — black, night blue, dark purple, dark grey — vanish on the board's
ground. A colour is kept by its name, not its value, so a theme can draw the
same names otherwise.

A label nobody chose a colour for wears one read off its name
(`domain.ColourOf`): the same label wears the same colour on every board,
with nothing written down (`@label-colour-is-read-off-the-name`).
`label colour release peach` chooses one, kept in a table of its own
(`label_colours`, migration 17); a rename carries it to the new name, a
delete lets it go, and a colour that is not one of the fourteen is refused
(`@label-colour-is-chosen`, `@label-colour-is-one-of-fourteen`). `label list`
says each label's colour, and the board reads them with the rest
(`Snapshot.Colours`). The window draws every label pill — on a card, in its
bubble, in a form, in a list and in the list of labels — in its colour, the
name in the colour itself on a fifth of it over a card's ground
(`Skin.LabelPill`, `LabelChip`). Its name is centred in the pill by
the ink of letters that rise and stand on the line, measured by the shaper
(`pillLine`): centred by its line, it stood a point high, and centred by its
own ink, `gui` and `release` would not share a line. In the list of labels, `c` — or *Colour* — opens the
fourteen swatches under the label the keyboard is on, the one it wears
ringed: `←→` walk them and `⏎` takes one, a click takes the one clicked,
`esc` keeps the colour it had (`ColourSwatches`,
`@gui-a-label-is-coloured-from-the-labels`,
`@gui-a-colour-is-walked-to-with-the-keys`).

## A label is for tasks, for objectives, or for both

A label has a scope: what it may be put on. `label new area --scope
objectives` makes one for objectives only, `label scope chore tasks`
narrows one to tasks, `label scope area both` widens it again
(`@label-scope-made-with-the-label`, `@label-scope-is-changed`). A label
nobody scoped is for both, so every label already in a store is offered where
it was (`@label-scope-is-both-unless-said`). It is kept in a table of its own
(`label_scopes`, migration 19), holding only the labels narrowed; a rename
carries it, a delete lets it go, and rolling the migration back leaves every
label for both (`@label-scope-migration-rollback-leaves-both`).

Putting a label on what its scope leaves out is refused, naming the scope —
`"area" is a label for objectives only, not for a task` — whichever surface
asked (`@label-scope-refuses-a-task`, `@label-scope-refuses-an-objective`,
`@label-scope-mcp`). Only a label joining is checked: what a task or an
objective already carries never stops a change to it. Narrowing a label is
refused while what it would leave out carries it — `"doc" is carried by 1
task` — since a scope that its own carriers broke would mean nothing
(`@label-scope-refused-while-carried-outside-it`).

`label list` says each label's scope — `for objectives` — and a machine reads
`scope`. `label list --for objectives` keeps the labels that may be put on
objectives, those for both included (`@label-scope-narrows-what-is-offered`):
it is what a list a label is chosen from asks, so that an objective's form
offers only what is meant for objectives, and the engine, not a board, says
which those are. The boards do not ask it yet. An export carries a label's
scope beside its colour, and an import brings it for a label the store does
not hold yet (`@exchange-a-label-comes-with-its-scope`).

**Type labels.** What kind of thing an objective is — an area, a project, a
key result — is a label it carries, scoped to objectives; see
[objectives.md](objectives.md#two-levels). `area`
and `project` are the ones suggested, and nothing makes them: `label new area
--scope objectives` and `label new project --scope objectives` do, when they
are wanted.

## The Claude Code plugin

A task's Azure DevOps work items are its labels made only of digits, on a task
that also carries `ado` — see [claude-code.md](claude-code.md). The same list
opens from an objective's form, which carries labels too.

## Not here

The list in the terminal's form, and the list of every label in the terminal.
