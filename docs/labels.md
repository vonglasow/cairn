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

A label exists while a task carries it. Taken off the last task that carried
it, it is no longer offered: there is no list of labels to keep apart from the
tasks, and no gesture that deletes a label from every task at once.

## The Claude Code plugin

A task's Azure DevOps work items are its labels made only of digits, on a task
that also carries `ado` — see [claude-code.md](claude-code.md). The same list
opens from an objective's form, which carries labels too.

## Not here

The list in the terminal's form; deleting a label from every task; renaming a
label.
