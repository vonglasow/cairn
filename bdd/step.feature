Feature: Steps
  A task is split into steps: smaller pieces of it, each done or not. A step is
  not a task — it has no number on the board, no column, no focus, no time and
  no notes; all of that stays with its task. A step is named by its place in
  the task, the first being 1.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A step is added to a task, and read back with it

    @step-new-needs-a-title
    Scenario: a step needs a title
      Given I added a task titled "Ship it"
      When I run `cairn step new 1 "   "`
      Then it fails
      And it says a step needs a title

    @step-new-of-no-task
    Scenario: a step is added only to a task that exists
      When I run `cairn step new 9 "Write the spec"`
      Then it fails
      And it says there is no such task

  Rule: A step is ticked, renamed and taken away by its place

    @step-edit-done
    Scenario: a step is ticked
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step edit 1 1 --done`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write the spec" as done
