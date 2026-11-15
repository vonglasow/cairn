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

  Rule: What was done without focus is added by hand

    @timesheet-time-is-added-by-hand
    Scenario: a meeting nobody focused is added to the day it happened
      Given I added a task titled "Sprint review"
      When I run `cairn timesheet add 1 1h --on 2026-09-29`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h"

  Rule: A deleted task keeps its time, unless the configuration says it takes it with it

    @timesheet-a-deleted-task-keeps-its-time-and-says-it-is-gone
    Scenario: a task deleted keeps its time on the sheet, said to be deleted
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn task delete 1 --yes`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h"
      And `cairn timesheet show --since 2026-09-29 --until 2026-09-29` says "Write the spec (deleted)"

  Rule: A machine reads it too

    @timesheet-travels-in-the-json
    Scenario: the timesheet reads as JSON, day by day and task by task
      Given I added a task titled "Write the spec"
      And I ran `cairn timesheet add 1 1h30m --on 2026-09-29`
      When I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29 --json`
      Then it writes JSON to standard output
      And the JSON gives the day "2026-09-29" task 1 at 90 minutes
