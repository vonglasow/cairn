# Backups and exports

```sh
cairn backup new                 # a new backup of the store as it stands
cairn backup list                # every backup, with its number and date
cairn backup show 2              # when it was made, and what it holds
                                 # `holds 4 task(s) and 2 objective(s)`
cairn backup delete 2 --yes
cairn backup restore 2 --yes     # restore one; without a number, the most recent

cairn task export ~/tasks.json   # every task and objective, never over a file
cairn task import ~/tasks.json   # add the tasks of such a file, as new tasks

cairn migration list             # every change, applied or waiting
cairn migration show 2           # what it does, and what rolling it back does
cairn migration apply            # apply what is waiting; `apply 2` for one
cairn migration rollback --yes   # unmake the last change applied
```

When cairn knows a change a store on disk does not, it says so and stops —
`the store is behind this cairn`, then `cairn migration apply` on a line of its
own, so it is copied whole; a machine reading the refusal as JSON gets it in
one sentence. The boards ask instead: opened on
such a store, they offer to apply what is waiting, or to quit. Nothing about the
shape of your store changes unless you asked for it. `apply` and `rollback` set
a backup aside first, and so does `backup restore` before replacing anything, so
picking the wrong one costs a command rather than the work it overwrote.

A backup is read in the shape it was written, never today's: one taken before
objectives existed says it holds none, rather than refusing to be read at all.

An export is a file of version 7. Beside each task and the objectives it
serves, it carries the time spent day by day (`"spent": {"2026-10-01": "2h"}`),
whether the task is in focus (`"focused": true`), every label with its
colour and, when it is not for both, its scope (`"scope": "objectives"`),
labels nothing carries included, and each objective's `parent` and
`waits_on`, the objectives they name written with it, and every series, its
rule, template, state and trace (see [series](series.md#carried-in-a-file)), and every
view, its name and its filters (see [views](listing.md#views)). All of it is optional, so a file
written by hand — fixtures, a board to show — says only what it needs, and may
count its days from the day it is read: `"deadline": "+3d"`, `"-1d"`, `"+2w"`,
the same in the keys of `spent`. An
export always writes the days as they are. A day that cannot be read refuses
the whole file before anything is added.

The time comes in as time added by hand on its day, so the Time page shows it
where it was spent. A file of version 3 carries the total alone, and it is
added on the day the file is read. A task in
focus comes back in focus, which puts it under way as focusing does. A label
already in the store keeps its colour and its scope, and a view already named
here keeps its filters: nothing here is replaced.
