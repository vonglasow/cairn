# What stops a task, and how it starts again

Three of the seven columns mean "not moving". This is about the one where
something else is in the way, what that something can be, and why nothing
starts again on its own.

## Three words for stopped, not one

`blocked`, `paused` and `watching` all mean the work is not advancing. They are
three columns rather than one because what gets each moving again differs, and
that is the only question worth asking about a stopped task:

| | it is waiting on | what starts it again |
|---|---|---|
| `blocked` | somebody else, or another task | a nudge, or the other task finishing |
| `watching` | somebody you handed it to | a look |
| `paused` | nothing at all | your own decision to pick it up |

Collapsed into one column, the list would be shorter and "what do I do about
this?" would have no answer in it. That is the whole of the trade: three words
of vocabulary bought against one lost question.

Their place in the reading order is in [ordering.md](ordering.md): after the
three you can act on, before what is history.

**A task watched is followed up with what it already carries.** Who or what
it waits on is its `because`; the day to look again is its deadline, which
both boards write in red once it has passed; what is due a look today is
`task list --state watching --due-before 2026-10-24`, on that day.

```sh
cairn task edit 7 --state watching --because "Bob, for the review" --deadline 2026-10-28
```

A follow-up day and a person awaited, as fields of their own, were weighed on
2026-10-24 and left out: they would have said again what these two say.

## Two kinds of blocker, two flags

A task is stopped by another task, or by something that is not in cairn at all.

```sh
cairn task edit 5 --state blocked --blocker 3
cairn task edit 5 --state blocked --because "the client's answer"
```

Both land in the same place — what blocks this — and the card shows them
together on one row.

**Two flags rather than one**, because a single flag taking either would have to
decide whether `3` is task 3 or the word "3". Guessing is the one thing cairn
does not do, and there is no separator that saves it: a `because` accepts any
words, digits included.

`--blocker` follows the grammar of `--label`: a plain number says what the set
becomes, `+3` adds, `-3` takes away. `domain.SplitSign` is the one place either
sign means anything, so the two cannot come to disagree.

## A blocker is a note; the column is the claim

A blocker can sit on a task in any column. It may block part of the work and
not the rest, and noting that is not the same as stopping.

When it blocks the whole thing, the column is what says so. The better move is
then to split the part that is stopped into a task of its own — a step carries
no blocker, so the part waited on is a task, not a step — and carry on with the
rest, which cairn suggests nowhere and refuses never. It is a working
habit, not a rule, and encoding it would be cairn deciding how you work.

## Two columns are shut while something is in the way

A task in `blocked` with something recorded against it cannot be moved to `todo`
or `current`:

```
$ cairn task edit 5 --state current
ERROR  task 5: it is still blocked, by #3 and by "the client's answer".
       take them away first: task 3, the reason it is waiting
```

A refusal and not a confirmation. Those two columns say the work is yours to do
now, and a task waiting on somebody else is not — a column changed anyway would
quietly make the record wrong, and the refusal is also the only moment cairn
gets to say what is in the way.

The refusal names each thing that has to go. A refusal that said only "it is
blocked" would leave somebody holding three blockers to work out which is which.

It names the things and **not the flags that clear them**. The engine is behind
four surfaces and only one of them has a flag to type: somebody on the board or
in the window was being told to type `--blocker -3`, which is not something
either of them can be told.

**Every other column stays reachable.** Deciding a blocked task is done, putting
it back in the backlog, pausing it, handing it over to watch — all of those are
things you can say about work that is waiting. Only claiming it is yours to do
now is refused.

**And a `blocked` task with nothing recorded is free.** That is the
blocked-by-something-cairn-was-never-told-about case: there is nothing to take
away, so there is nothing to refuse.

## Taking the last thing away is what starts it again

Nothing leaves `blocked` on its own — not even when every task blocking it is
done. What moves it is you taking the last thing out of its way, and then it
lands in `todo`.

```sh
cairn task edit 2 --state blocked --blocker 1 --because "the client's answer"
cairn task edit 1 --state done          # task 1 is finished. Task 2 has not moved.
cairn task edit 2 --because ""          # the client answered. Blocker 1 remains.
cairn task edit 2 --blocker -1          # the last thing goes. Task 2 is in todo.
```

Two commands to unstick a task that carried two reasons. What that buys: no task
changes column without somebody saying so, so the trace always knows who wanted
what.

`todo` and not `current`, because "nothing is in the way" and "I am doing this
now" are different sentences, and only the first was said.

## The mark, and the three times cairn says nothing

A blocked task would otherwise wait forever on somebody remembering to look. So
a listing marks the ones whose every named blocker is done:

```
#5  Ship it  (nothing blocks it)
```

That mark is a reminder, not a move. It says it is your turn.

It appears **only** when blockers were named and all of them are done. Three
cases get no mark, and all three for the same reason — claiming nothing is in
the way would be a guess:

- **blocked by words.** "the client's answer" is either in or out and cairn has
  no way to know which.
- **blocked by words as well as a task.** Something is still recorded, so
  nothing is in the way is simply false.
- **blocked by nothing named.** Something cairn was never told about is still
  something.

There is a deliberate asymmetry here, and it is about who is reading. Cairn
never concludes anything from words it cannot check. But somebody *erasing*
those words has said the thing is out of the way, so that erasure takes the task
back to `todo` like any other last thing. One is a guess cairn is not allowed to
make; the other is a sentence somebody typed.

## What cannot be recorded

- **A task cannot block itself.**
- **A ring cannot be closed**, however long the way round: 1 waiting on 2
  waiting on 3 waiting on 1 is a ring where nothing can ever start, so it is
  always somebody's mistake rather than a situation to record. The walk asks the
  store what blocks each task it reaches, so the cost is a graph walk before
  each write rather than a comparison of two numbers.
- **A blocker has to be a task that exists.** Refused, unlike `--objective 9`
  on a *listing*, which answers nothing: a filter says what is wanted, where a
  blocker writes a link, and a link to nothing is a fact that was never true.

## What it is written down as

Both are attributes, so both are traced as attributes — no new kind of act. A
history reads as a sentence, cause before consequence:

```
2026-08-28 17:50  column    backlog → blocked
2026-08-28 17:50  blockers  — → #1
2026-08-28 17:50  because   — → the client's answer
2026-08-28 17:50  because   the client's answer → —
2026-08-28 17:50  blockers  #1 → —
2026-08-28 17:50  column    blocked → todo
```

That last pair comes from one command, and the order is not free. The column
cairn chose is written as **its own change**, after the one you asked for, for
two reasons: nobody asked for that column, so folding it into your change would
put your name on cairn's decision — and a single recording pass lays attributes
out in the order a card does, where the column comes first, which would read the
consequence before its cause.

## Deliberately not there yet

- **`unblock` as a verb.** Weighed and left out: `--blocker -3` already does it,
  and the return to `todo` belongs to the edit rather than to a command, so it
  holds for an agent going through MCP that never types `unblock`.
- **The mark crossing to a machine.** A listing says `blocked_by` per task and
  the state of each, so an agent can work it out. Adding a derived field before
  anything asks for it would be a second thing to keep true.
- **Narrowing a listing by a blocker.** `task list --blocker 3` — everything
  waiting on task 3 — would fit the rule in [listing.md](listing.md), since a
  task number is something to match exactly. Nothing has asked for it.
- **Blocked by an objective.** Nothing has asked for it, and an objective is not
  work — it is what work is for.
- **Being told when a blocker finishes.** That is a notification, which cairn
  has none of. The mark is what stands in for it.
- **Deleting a blocker unmarks, and says so once.** Delete task 3 and the link
  goes with it, so a task that was marked stops being marked. The deletion is
  the one moment that can say it: `task delete` answers `task 5 was waiting on
  it` for each task it was in the way of (`was_blocking` to a machine). After
  that, the trace keeps the deletion; the listing does not.
