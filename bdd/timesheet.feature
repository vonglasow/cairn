Feature: How long the work took, day by day

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A day is said task by task, to the minute, and comes to their sum

    @timesheet-says-the-minutes
    Scenario: each task's time is said to the minute, and the day adds those up
      Given I added a task titled "Write the spec"
      And I added a task titled "Review the PR"
      And I ran `cairn timesheet add 1 50m --on 2026-09-29`
      And I ran `cairn timesheet add 2 53m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "50m"
      And the day "2026-09-29" gives task 2 "53m"
      And the day "2026-09-29" comes to "1h43"

    @timesheet-hides-no-short-work
    Scenario: work shorter than a quarter of an hour is said, not left out
      Given I added a task titled "Answer the review"
      And I ran `cairn timesheet add 1 5m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "5m"

    @timesheet-a-minute-begun-counts
    Scenario: a minute begun counts whole
      Given I added a task titled "Answer the review"
      And I ran `cairn timesheet add 1 10s --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1m"

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

    @timesheet-time-added-by-hand-is-in-the-trace
    Scenario: time added by hand is written down where the rest of the task's trace is
      Given I added a task titled "Sprint review"
      When I run `cairn timesheet add 1 1h --on 2026-09-29`
      And I run `cairn trace show 1`
      Then it says 1h was added for "2026-09-29"

    @timesheet-time-added-by-hand-is-taken-back-the-same-way
    Scenario: too much added by hand is taken back by adding less than nothing
      Given I added a task titled "Sprint review"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn timesheet add 1 --on 2026-09-29 -- -30m`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "30m"

  Rule: A deleted task keeps its time, unless the configuration says it takes it with it

    @timesheet-a-deleted-task-keeps-its-time-and-says-it-is-gone
    Scenario: a task deleted keeps its time on the sheet, said to be deleted
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn task delete 1 --yes`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h"
      And `cairn timesheet show --since 2026-09-29 --until 2026-09-29` says "Write the spec (deleted)"

    @timesheet-a-task-deleted-with-its-time-leaves-the-sheet
    Scenario: a task deleted takes its time with it when the configuration says so
      Given a deleted task takes its time with it
      And I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn task delete 1 --yes`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" says nothing of task 1
      And `cairn trace show 1` says "time: taken"

  Rule: A machine reads it too

    @timesheet-travels-in-the-json
    Scenario: the timesheet reads as JSON, day by day and task by task
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h30m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29 --json`
      Then it writes JSON to standard output
      And the JSON gives the day "2026-09-29" task 1 at 90 minutes
