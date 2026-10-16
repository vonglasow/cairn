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

    @ordering-outlives-the-command
    Scenario: the order one command writes, the next one reads
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task list`
      Then it lists the tasks in the order "First", "Second", "Third"

    @ordering-a-freed-place-is-not-reused
    Scenario: a deleted task does not leave its place behind
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I deleted task 2
      When I run `cairn task new "Fourth"`
      Then it lists the tasks in the order "First", "Third", "Fourth"

    @ordering-last-means-last-in-its-column
    Scenario: last means last in the backlog, not last of all
      Given I added a task titled "Already done" with the state "done"
      When I run `cairn task new "Just arrived"`
      Then it lists the tasks in the order "Just arrived", "Already done"

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

    @ordering-columns-do-not-interleave
    Scenario: tasks added in turn come back grouped by column
      Given I added a task titled "Backlog first"
      And I added a task titled "Todo first" with the state "todo"
      And I added a task titled "Backlog second"
      And I added a task titled "Todo second" with the state "todo"
      When I run `cairn task list`
      Then it lists the tasks in the order "Backlog first", "Backlog second", "Todo first", "Todo second"

    @ordering-a-column-change-moves-it
    Scenario: a task given another column leaves the section it was in
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I edited task 2 with the state "todo"
      When I run `cairn task list`
      Then it lists the tasks in the order "First", "Third", "Second"

    @ordering-a-column-change-lands-it-last
    Scenario: it lands behind what was already in that column
      Given I added a task titled "Coming from the backlog"
      And I added a task titled "Already to do" with the state "todo"
      And I edited task 1 with the state "todo"
      When I run `cairn task list`
      Then it lists the tasks in the order "Already to do", "Coming from the backlog"

  Rule: A task is moved inside the column it is already in

    @ordering-moved-first
    Scenario: a task is moved to the first place
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task move 3 --first`
      Then it lists the tasks in the order "Third", "First", "Second"

    @ordering-moved-after
    Scenario: a task is moved behind one that is named
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task move 1 --after 2`
      Then it lists the tasks in the order "Second", "First", "Third"

    @ordering-moved-last
    Scenario: a task is moved to the last place
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task move 1 --last`
      Then it lists the tasks in the order "Second", "Third", "First"

    @ordering-moved-after-another-column
    Scenario: a task cannot be moved behind one in another column
      Given I added a task titled "In the backlog"
      And I added a task titled "Already to do" with the state "todo"
      When I run `cairn task move 1 --after 2`
      Then it fails
      And it says a task is moved inside its own column
      And it says which column each of them is in

  Rule: A task is moved to another column and a place in one act

    @ordering-moved-into-another-column-at-a-place
    Scenario: a task is moved into another column, behind one already there
      Given I added a task titled "Coming from the backlog"
      And I added a task titled "Already to do" with the state "todo"
      And I added a task titled "Also to do" with the state "todo"
      When I run `cairn task move 1 --state todo --after 2`
      Then it lists the tasks in the order "Already to do", "Coming from the backlog", "Also to do"

    @ordering-moved-into-another-column-lands-last
    Scenario: a task moved into another column with no place lands last in it
      Given I added a task titled "Coming from the backlog"
      And I added a task titled "Already to do" with the state "todo"
      When I run `cairn task move 1 --state todo`
      Then it lists the tasks in the order "Already to do", "Coming from the backlog"

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

    @ordering-moved-after-one-that-does-not-exist
    Scenario: a task cannot be moved behind one that is not there
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task move 1 --after 99`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

    @ordering-moved-behind-itself
    Scenario: a task cannot be moved behind itself
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task move 1 --after 1`
      Then it fails
      And it says a task cannot be moved behind itself

    @ordering-moved-nowhere
    Scenario: a move needs somewhere to go
      Given I added a task titled "The only one"
      When I run `cairn task move 1`
      Then it fails
      And it says where a task can be moved

  Rule: Several tasks can be in focus at once

    @ordering-focus-nothing-at-first
    Scenario: nothing is in focus to begin with
      Given I added a task titled "The only one"
      When I run `cairn task focus`
      Then it succeeds
      And it says nothing is in focus

    @ordering-focus-a-task
    Scenario: a task is put in focus
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task focus 2`
      Then it reports the task titled "Second"

    @ordering-focus-keeps-the-one-before
    Scenario: putting another in focus keeps the one before
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn task list`
      Then it marks "First" as a task in focus
      And it marks "Second" as a task in focus
      And it marks 2 tasks in focus

    @ordering-focus-says-everything-it-holds
    Scenario: asking about the focus names every task in it
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn task focus`
      Then it succeeds
      And it lists the tasks in the order "First", "Second"

    @ordering-focus-twice-holds-it-once
    Scenario: a task already in focus is not held twice
      Given I added a task titled "The only one"
      And I put task 1 in focus
      And I put task 1 in focus
      When I run `cairn task list`
      Then it marks 1 task in focus

    @ordering-unfocus-one
    Scenario: one task is unfocused and the others stay
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn task unfocus 1`
      Then it succeeds
      When I run `cairn task list`
      Then it marks "Second" as a task in focus
      And it marks 1 task in focus

    @ordering-unfocus-one-that-is-not-held
    Scenario: unfocusing a task that was not in focus is not a failure
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      When I run `cairn task unfocus 2`
      Then it succeeds
      When I run `cairn task list`
      Then it marks 1 task in focus

    @ordering-unfocus
    Scenario: bare unfocus lets go of everything at once
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn task unfocus`
      Then it succeeds
      And it says nothing is in focus

  Rule: The focus names a task without moving it

    @ordering-focus-does-not-reorder
    Scenario: the task in focus keeps its place in the list
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I put task 3 in focus
      When I run `cairn task list`
      Then it lists the tasks in the order "First", "Second", "Third"
      And it marks "Third" as the task in focus

    @ordering-focus-cleared-by-deleting
    Scenario: deleting the task in focus leaves nothing in focus
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 2 in focus
      And I deleted task 2
      When I run `cairn task focus`
      Then it succeeds
      And it says nothing is in focus

    @ordering-focus-a-task-that-does-not-exist
    Scenario: a task that is not there cannot be put in focus
      Given I added a task titled "The only one"
      When I run `cairn task focus 99`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

    @ordering-focus-let-go-when-done
    Scenario: a task put in done is no longer in focus
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      When I run `cairn task list`
      Then it marks "Second" as a task in focus
      And it marks 1 task in focus

    @ordering-focus-let-go-when-moved-to-done
    Scenario: a task moved into done is no longer in focus
      Given I added a task titled "First"
      And I put task 1 in focus
      When I run `cairn task move 1 --state done --first`
      Then it succeeds
      When I run `cairn task focus`
      Then it says nothing is in focus

    @ordering-focus-refused-on-done
    Scenario: a task in done cannot be put in focus
      Given I added a task titled "First"
      And I edited task 1 with the state "done"
      When I run `cairn task focus 1`
      Then it fails
      And it says a task that is done cannot be in focus
      When I run `cairn task focus`
      Then it says nothing is in focus

    @ordering-several-moved-together
    Scenario: tasks moved together land one behind the other, in the order named
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Third"
      And I added a task titled "Fourth"
      When I run `cairn task move 3 --state todo --after 1 --also 4`
      Then it succeeds
      When I run `cairn task list --state todo`
      Then it lists the tasks in the order "First", "Third", "Fourth", "Second"

    Scenario: unfocusing nothing is not a failure
      Given I added a task titled "The only one"
      When I run `cairn task unfocus`
      Then it succeeds
      And it says nothing is in focus

  Rule: A sub-task takes its place among its sisters

    @ordering-sub-task-added-last
    Scenario: a new sub-task arrives behind its sisters
      Given I added a task titled "The parent"
      And I added a task titled "First sister" with the parent "1"
      And I added a task titled "Second sister" with the parent "1"
      When I run `cairn task new "Third sister" --parent 1`
      Then it lists the sub-tasks of task 1 in the order "First sister", "Second sister", "Third sister"

    @ordering-sub-task-moved-among-its-sisters
    Scenario: a sub-task is moved among its sisters
      Given I added a task titled "The parent"
      And I added a task titled "First sister" with the parent "1"
      And I added a task titled "Second sister" with the parent "1"
      When I run `cairn task move 3 --first`
      Then it lists the sub-tasks of task 1 in the order "Second sister", "First sister"

    @ordering-sub-task-not-moved-among-root-tasks
    Scenario: a sub-task cannot be moved behind a root task
      Given I added a task titled "The parent"
      And I added a task titled "A sister" with the parent "1"
      And I added a task titled "A root task"
      When I run `cairn task move 2 --after 3`
      Then it fails
      And it says a sub-task is moved among its sisters
      And it says whose sub-task it is

    @ordering-a-sub-task-given-another-parent
    Scenario: a sub-task given another parent arrives behind its new sisters
      Given I added a task titled "The first parent"
      And I added a task titled "Moving over" with the parent "1"
      And I added a task titled "Staying" with the parent "1"
      And I added a task titled "The second parent"
      And I added a task titled "Already there" with the parent "4"
      And I added a task titled "Also already there" with the parent "4"
      When I run `cairn task edit 2 --parent 4`
      Then it lists the sub-tasks of task 4 in the order "Already there", "Also already there", "Moving over"
      And it lists the sub-tasks of task 1 in the order "Staying"

    @ordering-a-parent-carries-its-sub-tasks
    Scenario: moving a parent carries its sub-tasks with it
      Given I added a task titled "The parent"
      And I added a task titled "A sister" with the parent "1"
      And I added a task titled "Another root task"
      When I run `cairn task move 1 --last`
      Then it lists the tasks in the order "Another root task", "The parent"
      And it lists the sub-tasks of task 1 in the order "A sister"

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

    @ordering-a-title-change-moves-nothing
    Scenario: renaming a task leaves its place alone
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I edited task 1 with the title "First, renamed"
      When I run `cairn task list`
      Then it lists the tasks in the order "First, renamed", "Second", "Third"

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

    @ordering-a-column-left-leaves-no-hole
    Scenario: a task given another column leaves no hole behind it
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I added a task titled "Fourth"
      And I edited task 2 with the state "todo"
      When I run `cairn task move 1 --last`
      Then it lists the tasks in the order "Third", "Fourth", "First", "Second"

    @ordering-a-parent-left-leaves-no-hole
    Scenario: a sub-task given another parent leaves no hole among its old sisters
      Given I added a task titled "The parent"
      And I added a task titled "First sister" with the parent "1"
      And I added a task titled "Second sister" with the parent "1"
      And I added a task titled "Third sister" with the parent "1"
      And I added a task titled "Fourth sister" with the parent "1"
      And I added a task titled "Another parent"
      And I edited task 4 with the parent "6"
      When I run `cairn task move 2 --last`
      Then it lists the sub-tasks of task 1 in the order "Second sister", "Fourth sister", "First sister"
