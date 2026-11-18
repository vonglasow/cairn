Feature: Objectives

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: An objective needs a title

    @objective-add
    Scenario: an objective is added and numbered
      When I run `cairn objective new "Ship cairn 1.0"`
      Then it succeeds
      And it numbers the objective 1

  Rule: An objective is numbered so it can be referred to later

    @objective-number-next
    Scenario: the next objective gets the next number
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective new "Learn Go properly"`
      Then it numbers the objective 2

  Rule: A task can be put under an objective

    @objective-task-added-under
    Scenario: a task can be added under an objective
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn task new "Write the spec" --objective 1`
      Then the task carries the objective "1"

  Rule: A task can serve more than one objective

    @objective-task-serves-two
    Scenario: a task can be given two objectives at once
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      When I run `cairn task new "Write the spec" --objective 1 --objective 2`
      Then it succeeds
      And `cairn task show 1` says it serves objectives 1 and 2

  Rule: An objective knows what was put under it

    @objective-carries-the-task
    Scenario: an objective names the tasks put under it
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective show 1`
      Then it lists task 1 titled "Write the spec"

  Rule: An objective says how much of it is done

    @objective-progress
    Scenario: an objective says how much of it is done
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      When I run `cairn objective show 1`
      Then it says 0 of 2 tasks are done

  Rule: An objective given to a task is one that exists

    @objective-not-a-number
    Scenario Outline: an objective that is not a number says so
      Given I added an objective titled "Ship cairn 1.0"
      When I run `<command>`
      Then it fails
      And it says an objective number is expected

      Examples: everywhere a number is read
        | command                                          |
        | cairn task new "Write the spec" --objective spec |
        | cairn objective show spec                        |
        | cairn objective edit spec --title "Nope"         |
        | cairn objective delete spec --yes                |

  Rule: An objective can be listed, read back and changed

    @objective-list
    Scenario: listing shows every objective with its number
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Learn Go properly"
      When I run `cairn objective list`
      Then it lists 2 objectives
      And it lists objective 1 titled "Ship cairn 1.0"

  Rule: A listing of objectives can be narrowed the same way

    @objective-list-narrowed-by-state
    Scenario: only the objectives standing where I asked
      Given I added an objective titled "Ship cairn 1.0" with the state "on-hold"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective list --state on-hold`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

  Rule: An objective is looked for by anything it carries

    @objective-search-by-a-word
    Scenario: a word from the title is the way back to a number
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective search cairn`
      Then it lists objective 1 titled "Ship cairn 1.0"

  Rule: An objective carries only what it was given

    @objective-defaults
    Scenario: a new objective carries nothing else
      When I run `cairn objective new "Ship cairn 1.0"`
      Then the objective has no description
      And the objective has no labels
      And the objective has no deadline
      And the objective has no estimate
      And the objective carries the priority "normal"
      And the objective is active

  Rule: An objective can be deleted

    @objective-delete-several
    Scenario: several objectives are deleted in one act
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Keep it honest"
      When I run `cairn objective delete 1 --also 2 --yes`
      Then it succeeds
      And `cairn objective list` says there is no objective

  Rule: An objective says the time spent on it and what is left of it

    @objective-says-the-time-spent-on-it
    Scenario: the time spent on an objective is the time spent on every task serving it
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I edited task 1 with the time spent "1h30m"
      And I edited task 2 with the time spent "45m"
      When I run `cairn objective show 1`
      Then the objective carries the time spent "2h15m"

  Rule: An objective is reviewed over a period

    @objective-report-says-the-time-month-by-month
    Scenario: the time spent on an objective's tasks in a period is said month by month
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I added a task titled "Elsewhere"
      And I ran `cairn timesheet add 1 2h --on 2026-07-30`
      And I ran `cairn timesheet add 1 1h30m --on 2026-08-04`
      And I ran `cairn timesheet add 2 45m --on 2026-09-10`
      And I ran `cairn timesheet add 3 3h --on 2026-09-10`
      When I run `cairn objective report 1 --since 2026-08-01 --until 2026-09-30`
      Then it succeeds
      And the objective carries the time spent "2h15m in the period, 4h15m in all"
      And the month "2026-08" says "1h30m"
      And the month "2026-09" says "45m"
      And it says nothing of the month "2026-07"

    @objective-report-reads-by-a-machine
    Scenario: a review can be read by a machine
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I ran `cairn timesheet add 1 1h30m --on 2026-08-04`
      And I edited task 1 with the state "done"
      When I run `cairn objective report 1 --since 2026-08-01 --json`
      Then the JSON report gives the month "2026-08" at "1h30m" and task 1 done

  Rule: An objective may hold sub-objectives, two levels deep and no more

    @objective-sub-made-under-a-parent
    Scenario: an objective is made under another
      Given I added an objective titled "Home automation"
      When I run `cairn objective new "Enclosure" --parent 1`
      Then it succeeds
      And the objective sits under "#1 Home automation"

    @objective-sub-named-under-its-parent
    Scenario: a task names the objective it serves under its parent
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Pin map" with the objective "2"
      When I run `cairn task show 1`
      Then the answer says "#2 Home automation › Enclosure"
      When I run `cairn task list`
      Then the answer says "◇ Home automation › Enclosure"

  Rule: A task is counted once in a tree

    @objective-sub-attached-to-a-parent-and-its-sub-objectives-refused
    Scenario: a task on two sub-objectives is not attached to their parent as well
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Mount the board" with the objectives "2" and "3"
      When I run `cairn task edit 1 --objective +1`
      Then it fails
      And the refusal says "#1 is already counted in #1 Home automation through #2 Enclosure and #3 Firmware; nothing was changed"
      And `cairn task show 1` says it serves objectives 2 and 3

  Rule: A parent counts over its own tasks and its sub-objectives', each task once

    @objective-sub-parent-counts-each-task-once
    Scenario: a parent's progress and time are over the union of its tree
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Mount the board" with the objectives "2" and "3"
      And I added a task titled "Wiring diagram" with the objective "1"
      And I added a task titled "Flash it" with the objective "3"
      And I edited task 1 with the state "done"
      And I edited task 1 with the time spent "1h"
      And I edited task 2 with the time spent "30m"
      And I edited task 3 with the time spent "15m"
      When I run `cairn objective show 1`
      Then the answer says "1 of 3 tasks are done"
      And the objective carries the time spent "1h45m"
      And the objective carries the left "2 tasks, no deadline"

  Rule: A parent is proposed as reached, never put there on its own

    @objective-sub-reached-proposed-by-the-last-one
    Scenario: reaching the last sub-objective proposes the parent
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I edited objective 2 with the state "reached"
      When I run `cairn objective edit 3 --state reached`
      Then it succeeds
      And the answer says "every sub-objective of #1 Home automation is reached: it may be put reached too"
      And `cairn objective show 1` says "every sub-objective is reached: it may be put reached too"
      And `cairn objective show 1` says "active"

  Rule: Deleting an objective with sub-objectives or tasks asks what goes with them

    @objective-sub-delete-takes-everything
    Scenario: a parent deleted with its sub-objectives and their tasks
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Print the case" with the objective "2"
      And I added a task titled "Wiring diagram" with the objective "1"
      When I run `cairn objective delete 1 --sub-objectives delete --tasks delete --yes`
      Then it succeeds
      And `cairn objective list` says there is no objective
      And `cairn task list` says there is no task

  Rule: An objective may wait on another, which warns and never blocks

    @objective-waits-on-is-said-on-both
    Scenario: an objective that waits on another says so, and the other names who waits on it
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Features" under objective 1
      When I run `cairn objective edit 3 --waits-on 2`
      Then it succeeds
      And the objective waits on "#2 Firmware"
      And `cairn objective show 2` says "#3 Features waits on it"

  Rule: An objective outlives the command that created it

    @objective-survives
    Scenario: an objective added by one command is found by the next
      Given I added an objective titled "Ship cairn 1.0"
      And cairn has exited
      When I run `cairn objective list`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

  Rule: An objective can be read by a machine

    @objective-json
    Scenario: asking for machine-readable output
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective show 1 --json`
      Then it writes JSON to standard output
      And it carries the title "Ship cairn 1.0"
      And it carries the tasks under it

  Rule: Getting the command wrong is told, never guessed

    @objective-usage-refused
    Scenario Outline: the command says what it needed
      When I run `<command>`
      Then it fails
      And it does not crash
      And it says how to use it

      Examples: what someone types the first time
        | command                    | what is wrong             |
        | cairn objective new        | no title at all           |
        | cairn objective show       | no number given           |
        | cairn objective            | the group with no command |
        | cairn objective addd "x"   | a typo in the subcommand  |
