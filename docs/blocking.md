# What stops a task, and how it starts again

Three of the seven columns mean the work is not moving. They are three rather
than one because what gets each moving again differs:

| | it is waiting on | what starts it again |
|---|---|---|
| `blocked` | somebody else, or another task | a nudge, or the other task finishing |
| `watching` | somebody you handed it to | a look |
| `paused` | nothing at all | your own decision to pick it up |

## What is in the way

A task is stopped by another task, or by something cairn does not hold. Each
has its flag:

```sh
cairn task edit 5 --state blocked --blocker 3
cairn task edit 5 --state blocked --because "the client's answer"
```

`--blocker` takes a task number: a plain number or `+3` adds one, `-3` takes one
away, `=3` says what the set becomes. `--because` takes words. Both show
together on the card, on one row.

A task cannot block itself, a ring of blockers — 1 waiting on 2 waiting on 1 —
is refused, and a blocker has to be a task that exists.

A blocker can sit on a task in any column: it may hold up part of the work and
not the rest. When part of a task is stopped and the rest is not, split the
stopped part into a task of its own and carry on with the rest.

## Two columns are shut while something is in the way

A task in `blocked` with something recorded against it cannot be moved to
`todo` or `current`, since those say the work is yours to do now:

```
$ cairn task edit 5 --state current
ERROR  task 5: it is still blocked, by #3 and by "the client's answer".
       take them away first: task 3, the reason it is waiting
```

Every other column stays open to it: `done`, `backlog`, `paused`, `watching`.
A `blocked` task with nothing recorded against it moves freely.

## Taking the last thing away starts it again

Nothing leaves `blocked` on its own, not even when every task blocking it is
done. Taking the last thing out of its way puts it in `todo`:

```sh
cairn task edit 2 --state blocked --blocker 1 --because "the client's answer"
cairn task edit 1 --state done          # task 1 is finished. Task 2 has not moved.
cairn task edit 2 --because ""          # the client answered. Blocker 1 remains.
cairn task edit 2 --blocker -1          # the last thing goes. Task 2 is in todo.
```

It goes to `todo`, not `current`: "nothing is in the way" is not "I am doing it
now".

## When it is your turn

A listing marks a blocked task whose every named blocker is done:

```
#5  Ship it  (nothing blocks it)
```

The mark is a reminder; the task does not move. It does not appear when the
task is also blocked by words, or blocked with nothing named, since cairn
cannot tell whether those are out of the way.

Deleting a task that was in the way says so once: `task delete` answers
`task 5 was waiting on it` for each task it was blocking.

## What a task holds up

Read from the other side, a task says how many open tasks wait on it: `⊳2` on
a card of either board and on a listing's row. `task show` names them —
`blocks  ⊳2  #2, #3` — and `task list --blocker 1` lists them. A task done or
archived is not counted as waiting.

## A task you are watching

A task handed to somebody is followed up with what it already carries: who it
waits on is its `because`, the day to look again is its deadline — written in
red on both boards once it has passed.

```sh
cairn task edit 7 --state watching --because "Bob, for the review" --deadline 2026-10-28
cairn task list --state watching --deadline-before 2026-10-29   # due a look
```

## Not there

- No `unblock` command: `--blocker -3` does it.
- An objective does not block a task. An objective may wait on another
  objective, which warns and shuts nothing; see
  [objectives](objectives.md#one-objective-waiting-on-another).
- No notification when a blocker finishes: the mark stands in for it.
