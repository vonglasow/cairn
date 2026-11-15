Feature: Ordering

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A new task takes the last place in the backlog

    @ordering-added-last
    Scenario: a task arrives behind the ones already there
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task new "Third"`
      Then it lists the tasks in the order "First", "Second", "Third"

  Rule: The column gives the coarse order, and a list is never four lists

    @ordering-columns-in-reading-order
    Scenario: the columns come back in the order they are read
      Given I added a task titled "Finished" with the state "done"
      And I added a task titled "Put off" with the state "paused"
      And I added a task titled "Delegated" with the state "watching"
      And I added a task titled "Waiting on somebody" with the state "blocked"
      And I added a task titled "Under way" with the state "current"
      And I added a task titled "Chosen" with the state "todo"
      And I added a task titled "Waiting"
      When I run `cairn task list`
      Then it lists the tasks in the order "Waiting", "Chosen", "Under way", "Waiting on somebody", "Delegated", "Put off", "Finished"

  Rule: A task is moved inside the column it is already in

    @ordering-moved-first
    Scenario: a task is moved to the first place
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task move 3 --first`
      Then it lists the tasks in the order "Third", "First", "Second"

  Rule: A task is moved to another column and a place in one act

    @ordering-moved-into-another-column-or-not-at-all
    Scenario: a place refused leaves the task in the column it was in
      Given I added a task titled "Coming from the backlog"
      And I added a task titled "Already to do" with the state "todo"
      When I run `cairn task move 1 --state todo --after 99`
      Then it fails
      And it says there is no such task
      When I run `cairn task list --state backlog`
      Then it lists task 1 titled "Coming from the backlog"

  Rule: A move that cannot be carried out is refused, and says why

    @ordering-moved-that-does-not-exist
    Scenario: a task that is not there cannot be moved
      Given I added a task titled "The only one"
      When I run `cairn task move 99 --first`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

  Rule: Several tasks can be in focus at once

    @ordering-focus-nothing-at-first
    Scenario: nothing is in focus to begin with
      Given I added a task titled "The only one"
      When I run `cairn task focus`
      Then it succeeds
      And it says nothing is in focus

  Rule: The focus brings waiting work into current, and moves nothing else

    @ordering-focus-a-task-that-does-not-exist
    Scenario: a task that is not there cannot be put in focus
      Given I added a task titled "The only one"
      When I run `cairn task focus 99`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

    Scenario: unfocusing nothing is not a failure
      Given I added a task titled "The only one"
      When I run `cairn task unfocus`
      Then it succeeds
      And it says nothing is in focus

  Rule: A task blocked lets go of its focus

    @focus-let-go-when-blocked
    Scenario: a task blocked lets go of the focus it held
      Given I added a task titled "Ship it"
      And I put task 1 in focus
      When I run `cairn task edit 1 --state blocked --because "the client's answer"`
      And I run `cairn task list --json`
      Then it says no task in it is in focus

  Rule: A focus says how long its session has been up, and past the length the configuration gives it asks to be renewed

    @focus-session-says-how-long-it-has-been-up
    Scenario: the focus says how long its session has been up
      Given I added a task titled "Ship it"
      And I put task 1 in focus
      When it is 3 hours later
      And I run `cairn task focus --json`
      Then the focus session of task 1 has been up "3h"
      And the focus session of task 1 asks nothing

  Rule: A task's steps keep the order they were written in

    @ordering-a-task-carries-its-steps
    Scenario: moving a task carries its steps with it
      Given I added a task titled "Ship it"
      And I added the step "A step" to task 1
      And I added a task titled "Another task"
      When I run `cairn task move 1 --last`
      Then it lists the tasks in the order "Another task", "Ship it"
      And the steps of task 1 read "A step"

  Rule: A place changes only when a task joins or leaves the ones it sits with

    @ordering-an-ordinary-edit-moves-nothing
    Scenario Outline: editing what a task carries leaves its place alone
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I edited task 1 with the <field> "<value>"
      When I run `cairn task list`
      Then it lists the tasks in the order "First", "Second", "Third"

      Examples: none of these is a change of the tasks it sits with
        | field       | value      |
        | description | the one    |
        | priority    | high       |
        | deadline    | 2026-09-01 |
        | estimate    | 2h         |
        | time spent  | 30m        |

  Rule: A place freed leaves no hole behind it

    @ordering-a-deletion-leaves-no-hole
    Scenario: a task deleted leaves no hole for a later move to fall into
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I added a task titled "Fourth"
      And I deleted task 3
      When I run `cairn task move 1 --last`
      Then it lists the tasks in the order "Second", "Fourth", "First"
