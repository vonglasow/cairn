# Plan a week

```sh
cairn task list --state backlog                        # what is waiting
cairn task move 4 --state todo                         # chosen for the week
cairn task move 7 --state todo --first                 # and the one to start with
cairn task edit 7 --deadline 2026-11-20 --estimate 3h  # when, and how long
cairn task list --state todo                           # the week, in its order
cairn task list --deadline-before 2026-11-20           # what is wanted by Friday
cairn task focus 7                                     # the work starts
cairn task edit 7 --state done                         # and ends
cairn timesheet show --since 2026-11-16 --until 2026-11-20  # how long the week took, once it is over
```

A week is the tasks you move into `todo`, in the order you work them. A task
moved with `--first` goes to the head of its column; `--after 4` puts it
behind task 4.

A deadline is the day a task is wanted by; the estimate, how long it should
take. A card shows both, and the time spent beside the estimate.

Focus a task when you start on it: it moves into `current`, and the time runs
until you unfocus it or put it `done`. The timesheet adds up each day.

A task that comes back every week is a series:

```sh
cairn config set series.on true                        # series are off until turned on
cairn task new "Plan the week" --every "monday 9:00"
```

See [the order](ordering.md), [series](series.md) and
[the trace and the time](trace.md).
