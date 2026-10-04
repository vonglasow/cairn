Feature: How long the work took, day by day
  A timesheet is filled in from what cairn kept: how long something was in
  focus, day by day, task by task. Cairn says the real time — every hour of it,
  however many there were. Sorting it into the kinds of work a sheet asks for,
  and bringing a day down to what a sheet accepts, is done afterwards, by
  whoever fills the sheet in; cairn holds no kind of work and no ceiling.

  Focus is what is measured: taken when the work starts, let go when it stops,
  at the moment it is done and no other — a focus placed after the fact would
  be a timesheet anybody could write. What was done without it — a meeting, an
  afternoon nobody focused — is added by hand, and too much is taken back the
  same way. Every one of these is written down in the history, so the timing
  of a task can be read back long after the sheet was sent.

  The scenarios below are set on days already past, so their time is added by
  hand: focus moves when it is done, and nothing can place it on a day already
  gone. How focus becomes time — the minutes several tasks share split between
  them, a focus left open stopped at midnight and said so — is arithmetic over
  moments, and pinned where moments can be chosen: internal/trace.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A day is said task by task, each rounded, and comes to their sum

    @timesheet-rounds-to-the-quarter
    Scenario: each task's time is rounded to the nearest quarter of an hour, and the day adds those up
      Given I added a task titled "Write the spec"
      And I added a task titled "Review the PR"
      And I ran `cairn timesheet add 1 50m --on 2026-09-29`
      And I ran `cairn timesheet add 2 53m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "45m"
      And the day "2026-09-29" gives task 2 "1h"
      And the day "2026-09-29" comes to "1h45"

    @timesheet-says-the-real-time-however-long
    Scenario: a long day is said as it was, with no ceiling
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 11h30m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" comes to "11h30"

    @timesheet-a-day-is-told-apart-from-the-next
    Scenario: work over several days is said day by day, and a day nothing was done on says so
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h --on 2026-09-28`
      And I ran `cairn timesheet add 1 2h --on 2026-09-30`
      When I run `cairn timesheet show --since 2026-09-28 --until 2026-09-30`
      Then the day "2026-09-28" comes to "1h"
      And the day "2026-09-29" comes to "0m"
      And the day "2026-09-30" comes to "2h"

    @timesheet-with-no-day-named-is-today
    Scenario: with no day named it says today
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      When I run `cairn timesheet show`
      Then it says today

  Rule: What was done without focus is added by hand

    @timesheet-time-is-added-by-hand
    Scenario: a meeting nobody focused is added to the day it happened
      Given I added a task titled "Sprint review"
      When I run `cairn timesheet add 1 1h --on 2026-09-29`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h"

    @timesheet-time-added-by-hand-is-in-the-history
    Scenario: time added by hand is written down where the rest of the task's history is
      Given I added a task titled "Sprint review"
      When I run `cairn timesheet add 1 1h --on 2026-09-29`
      And I run `cairn trace show 1`
      Then it says 1h was added for "2026-09-29"

    # A length that begins with a minus is told apart from an option the way
    # every command line tells them apart: after `--`.

    @timesheet-time-added-by-hand-is-taken-back-the-same-way
    Scenario: too much added by hand is taken back by adding less than nothing
      Given I added a task titled "Sprint review"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn timesheet add 1 --on 2026-09-29 -- -30m`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "30m"

  Rule: A machine reads it too

    @timesheet-travels-in-the-json
    Scenario: the timesheet reads as JSON, day by day and task by task
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h30m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29 --json`
      Then it writes JSON to standard output
      And the JSON gives the day "2026-09-29" task 1 at 90 minutes
