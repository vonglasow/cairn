# Follow an objective

```sh
cairn objective new "Ship v1" --deadline 2026-12-15             # what the work is for
cairn objective new "Write the docs" --parent 1                 # a part of it
cairn task new "Write the install guide" --objective 2          # work serving it
cairn task edit 3 --objective +1                                # serving one more
cairn objective show 1                                          # how far along it is
cairn task list --objective 1                                   # its work, sub-objectives included
cairn objective report 1 --since 2026-11-01                     # what moved, and the time it took
cairn objective edit 1 --state reached                          # done with it
```

An objective is what work is for. A task serves one or more objectives, and
an objective counts how far along the tasks serving it are, its
sub-objectives' tasks included.

An objective sits under a parent, two levels deep at most. It may wait on
another: work started on it while the other is not reached is warned about,
and never refused.

The report says what was done for an objective over a period and how long it
took.

See [objectives](objectives.md).
