# Labels

A label is a word a task or an objective carries, as many as it likes. It means
nothing to cairn: a listing narrows by it, and that is all.

## How a label is written

Every surface writes a label the same way, so one word is always one label:

1. lower case;
2. the accents of the Latin letters taken off — `modularité` is `modularite`;
3. spaces, `_` and `/` become `-`, a run of `-` becomes one, and none is kept
   at either end — `Sp26 20` is `sp26-20`;
4. what is left holds `a`–`z`, `0`–`9`, `-` and `.`, and nothing else.

Anything else is refused, naming what a label may hold. A `:` is refused too: a
label has no family, so `ado:77594` is two labels, `ado` and `77594`. A listing
writes the labels it is given the same way, so `--label SP26_20` finds
`sp26-20`.

## Putting labels on and taking them off

On a change, a label given with no sign joins the ones already there:

```sh
cairn task edit 1 --label urgent            # one more, the others kept
cairn task edit 1 --label +urgent           # the same, signed
cairn task edit 1 --label -draft            # one fewer
cairn task edit 1 --label =spec,team        # the set becomes exactly this
cairn task edit 1 --label =                 # every label taken off
```

`--label spec --label team` and `--label spec,team` say the same thing. An `=`
is not mixed with `+` or `-`. The answer says which labels were added and which
were taken off. On a new task, the labels given are the set, and a sign is
refused. Objectives take the same grammar.

In zsh, quote a value starting with `=`: `--label '=spec'`.

## Every label at once

```sh
cairn label list                                  # each with how many tasks and objectives carry it
cairn label new release-train                     # kept and offered, carried or not, until deleted
cairn label colour release-train peach            # one of fourteen
cairn label rename sp26-20 sprint-20              # on everything; onto one there, the two become one
cairn label delete sp26-20 --yes                  # off everything
```

A label exists while something carries it. One made with `label new` stays,
carried by nothing, until it is deleted — so a sprint's label can be made
before its tasks. Renaming or deleting a label is a change to each task and
objective carrying it, and shows in their trace.

In the window, `⇧L` (or the File menu) opens the list of labels: `n` makes one,
`r` renames the one chosen, `c` chooses its colour, `⌘⌫` takes it off
everything after asking.

## Colours

A label wears one of fourteen colours: rosewater, flamingo, pink, mauve, red,
maroon, peach, yellow, green, teal, sky, sapphire, blue, lavender. One nobody
chose wears a colour read off its name, the same on every board.
`label colour` chooses one; a rename keeps it.

## For tasks, for objectives, or both

A label may be put on tasks, on objectives, or on both — both unless said:

```sh
cairn label new area --scope objectives
cairn label scope chore tasks
cairn label list --for objectives                 # those that may go on an objective
```

Putting a label where its scope does not allow is refused, naming the scope.
Narrowing a label is refused while something outside the new scope carries it.

What kind of thing an objective is — an area, a project — is a label scoped to
objectives; see [objectives](objectives.md#two-levels).

## In the window's form

The `labels` field opens a list of every label, those the task carries ticked.
Typing narrows it; typing a label that does not exist yet offers to create it,
under the name it will have. `⏎` ticks or unticks, `⌘⏎` keeps what is ticked,
`esc` keeps nothing. Nothing is written until the form is saved.

A view that narrows by a label follows it when it is renamed; see
[views](listing.md#views).
