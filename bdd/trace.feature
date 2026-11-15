Feature: The trace

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: Every act on a task is written down

    @trace-created-carries-the-initial-state
    Scenario: creation records what the task started as
      Given I added a task titled "Write the spec" with the priority "high"
      When I run `cairn trace show 1`
      Then it says the task was created
      And it says it started in the backlog at high priority
      And it says when

  Rule: Taking the focus and letting it go are acts

    @trace-focus-taken-and-released
    Scenario: taking the focus and letting it go are both written down
      Given I added a task titled "Write the spec" with the state "current"
      And I put task 1 in focus
      And I let the focus go
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "focus taken", "focus released"

  Rule: What blocks a task is written down like anything else it carries

    @trace-a-blocker-added
    Scenario: naming what blocks a task is written down
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      When I run `cairn trace show 2`
      Then it says the blockers went from "—" to "#1"

  Rule: A link is written down on both tasks it joins

    @trace-a-link-is-written-on-both-tasks
    Scenario: a link is written in the trace of both tasks
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the link "1"
      When I run `cairn trace show 2`
      Then it says the links went from "—" to "#1"
      When I run `cairn trace show 1`
      Then it says the links went from "—" to "#2"
      And it carries 2 acts

  Rule: What changed nothing is not written down

    @trace-focus-again-on-the-same-task-changes-nothing
    Scenario: putting the focus where it already is does nothing at all
      Given I added a task titled "Write the spec" with the state "current"
      And I put task 1 in focus
      And I put task 1 in focus
      When I run `cairn trace show 1`
      Then it carries 2 acts

  Rule: An act that is not about an attribute says what it did

    @trace-moved-says-where-it-went
    Scenario: moving a task is written down, with where it went
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I moved task 2 to the first place
      When I run `cairn trace show 2`
      Then it lists the acts in the order "created", "moved"
      And it says the move went to the first place

  Rule: What happened over a period is listed, with what can be worked out from it

    @trace-list-is-today-by-default
    Scenario: with no day named it lists what was done today
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      When I run `cairn trace list`
      Then it lists the acts of 2 tasks

  Rule: A trace that cannot be shown says why

    @trace-show-a-task-that-never-existed
    Scenario: a task that was never there has no trace to show
      Given I added a task titled "First"
      When I run `cairn trace show 99`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

  Rule: Every act says who did it, and through what

    @trace-says-who-acted
    Scenario: an act says who did it and the surface it came through
      Given the author is "sam"
      And I added a task titled "Write the spec"
      When I run `cairn trace show 1`
      Then every act says it was done by "sam" through "cli"
