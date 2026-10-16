Feature: Steps
  A task is split into steps: smaller pieces of it, each done or not. A step is
  not a task — it has no number on the board, no column, no focus, no time and
  no notes; all of that stays with its task. A step is named by its place in
  the task, the first being 1.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A step is added to a task, and read back with it

    @step-new
    Scenario: a step is added to a task, not done
      Given I added a task titled "Ship it"
      When I run `cairn step new 1 "Write the spec"`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write the spec" as not done

    @step-new-goes-last
    Scenario: a step added goes behind the ones already there
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step new 1 "Write the tests"`
      Then `cairn step list 1` names step 2 "Write the tests" as not done

    @step-new-done
    Scenario: a step can be added already done
      Given I added a task titled "Ship it"
      When I run `cairn step new 1 "Plan it" --done`
      Then `cairn step list 1` names step 1 "Plan it" as done

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

    @step-list-json
    Scenario: a task's steps can be read by a machine
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step list 1 --json`
      Then it writes JSON to standard output
      And it carries the step "Write the spec", not done

  Rule: A step is ticked, renamed and taken away by its place

    @step-edit-done
    Scenario: a step is ticked
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step edit 1 1 --done`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write the spec" as done

    @step-edit-not-done
    Scenario: a step ticked is unticked
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I ticked step 1 of task 1
      When I run `cairn step edit 1 1 --not-done`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write the spec" as not done

    @step-edit-title
    Scenario: a step is renamed
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step edit 1 1 --title "Write it"`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write it" as not done

    @step-edit-both-ways
    Scenario: a step is not ticked and unticked in one go
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step edit 1 1 --done --not-done`
      Then it fails
      And it says a step is done or not done, not both

    @step-delete
    Scenario: a step taken away leaves the next one in its place
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the tests" to task 1
      When I run `cairn step delete 1 1`
      Then it succeeds
      And `cairn step list 1` names step 1 "Write the tests" as not done
      And `cairn step list 1` does not mention "Write the spec"

    @step-unknown
    Scenario: a place the task has no step at is refused
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn step edit 1 2 --done`
      Then it fails
      And it says the task has no such step

    @step-traced
    Scenario: what is done to a step is written in its task's trace
      Given I added a task titled "Ship it"
      When I run `cairn step new 1 "Write the spec"`
      And I run `cairn trace show 1`
      Then it shows "steps"
