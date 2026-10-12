Feature: The trace
  Every act performed on a task is written down: what it was, what it became,
  and when. Nothing is ever rewritten, so a task can be put back together as it
  stood at any past moment.

  Looking at a task is not an act. A trace records what changed, never what was
  read.

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

    @trace-a-change-says-what-it-was-and-became
    Scenario: a change says what it went from and what it became
      Given I added a task titled "Write the spec"
      And I edited task 1 with the priority "high"
      When I run `cairn trace show 1`
      Then it says the priority went from "normal" to "high"

    @trace-reading-is-not-an-act
    Scenario: looking at a task is not an act
      Given I added a task titled "Write the spec"
      And I looked at task 1
      And I listed the tasks
      And I looked for "spec"
      When I run `cairn trace show 1`
      Then it carries 1 act

    @trace-oldest-first
    Scenario: a trace reads oldest first
      Given I added a task titled "Write the spec"
      And I edited task 1 with the priority "high"
      And I edited task 1 with the state "todo"
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "priority", "column"

  Rule: Taking the focus and letting it go are acts

    Amended 2026-09-14. Taking up a second task used to close the stretch on
    the first, and the release named what had taken over so a cut could be told
    from a stop. Several tasks are now worked on at once, one per agent
    session, and their stretches overlap on purpose — so nothing is cut, and
    what closes a stretch is unfocusing that task and nothing else.

    @trace-focus-taken-and-released
    Scenario: taking the focus and letting it go are both written down
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      And I let the focus go
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "focus taken", "focus released"

    @trace-focus-taking-another-leaves-the-first-open
    Scenario: putting another task in focus leaves the first one open
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "focus taken"

    @trace-focus-letting-one-go-closes-only-that-one
    Scenario: letting one task go closes its stretch and no other
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      And I put task 2 in focus
      And I let task 1 go
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "focus taken", "focus released"
      When I run `cairn trace show 2`
      Then it lists the acts in the order "created", "focus taken"

    @trace-focus-closed-when-the-task-goes
    Scenario: deleting the task in focus closes the stretch, and its trace stays
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      And I deleted task 1
      When I run `cairn trace show 1`
      Then it lists the acts in the order "created", "focus taken", "focus released", "deleted"

    @trace-a-deletion-writes-down-what-went-with-it
    Scenario: what a deleted task carried is written down with the deletion
      Given I added a task titled "Write the spec"
      And I edited task 1 with the description "the client wants two columns"
      And I edited task 1 with the priority "high"
      And I deleted task 1
      When I run `cairn trace show 1`
      Then it says the deletion took "Write the spec" with it
      And it says the deletion took "the client wants two columns" with it
      And it says the deletion took "high" with it

  Rule: What blocks a task is written down like anything else it carries

    A blocker and the words beside it are attributes, so they are traced as
    attributes: what they were, what they became, when, and through what. No new
    kind of act — the one that says an attribute changed already says it.

    The column going back to todo when the last thing is taken away is traced
    too, and that matters more than it looks. It is a change cairn made rather
    than one anybody typed, so without it the trace would show a task in todo
    with nothing saying how it got there.

    @trace-a-blocker-added
    Scenario: naming what blocks a task is written down
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      When I run `cairn trace show 2`
      Then it says the blockers went from "—" to "#1"

    @trace-a-blocker-taken-away
    Scenario: taking a blocker away is written down too
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the blocker "-1"
      When I run `cairn trace show 2`
      Then it says the blockers went from "#1" to "—"

    @trace-the-words-that-block
    Scenario: what is not in cairn is written down in the words it was given
      Given I added a task titled "Ship it"
      And I edited task 1 with the because "the client's answer"
      When I run `cairn trace show 1`
      Then it says the because went from "—" to "the client's answer"

    @trace-the-objectives-served
    Scenario: what a task serves is written down like anything else it carries
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objective "1"
      And I edited task 1 with the objective "+2"
      When I run `cairn trace show 1`
      Then it says the objectives went from "#1" to "#1, #2"

    @trace-the-column-cairn-put-it-back-in
    Scenario: the return to todo is not a silent move
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      And I edited task 2 with the blocker "-1"
      When I run `cairn trace show 2`
      Then it lists the acts in the order "created", "blockers", "column", "blockers", "column"

  Rule: What changed nothing is not written down

    @trace-focus-again-on-the-same-task-changes-nothing
    Scenario: putting the focus where it already is does nothing at all
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      And I put task 1 in focus
      When I run `cairn trace show 1`
      Then it carries 2 acts

    @trace-an-edit-that-changes-nothing-is-not-an-act
    Scenario: giving an attribute the value it already holds is not a change
      Given I added a task titled "Write the spec"
      And I edited task 1 with the priority "normal"
      When I run `cairn trace show 1`
      Then it carries 1 act

    @trace-a-move-to-where-it-stands-is-not-an-act
    Scenario: moving a task to the place it already holds is not a move
      Given I added a task titled "Alone in its column"
      And I moved task 1 to the first place
      And I moved task 1 to the last place
      When I run `cairn trace show 1`
      Then it carries 1 act

  Rule: An act that is not about an attribute says what it did

    @trace-moved-says-where-it-went
    Scenario: moving a task is written down, with where it went
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I moved task 2 to the first place
      When I run `cairn trace show 2`
      Then it lists the acts in the order "created", "moved"
      And it says the move went to the first place

    @trace-labels-are-a-set
    Scenario: a set of labels says what it was and what it became
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      And I edited task 1 with the labels "spec" and "urgent"
      When I run `cairn trace show 1`
      Then it says the labels went from "spec, team" to "spec, urgent"

  Rule: What happened over a period is listed, with what can be worked out from it

    @trace-list-is-today-by-default
    Scenario: with no day named it lists what was done today
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 1 in focus
      When I run `cairn trace list`
      Then it lists the acts of 2 tasks

    @trace-list-since-a-day-nothing-reaches
    Scenario: a period nothing falls into says so
      Given I added a task titled "First"
      And I put task 1 in focus
      When I run `cairn trace list --since 2099-01-01`
      Then it says nothing was done

    @trace-letting-go-of-nothing-is-written-down-nowhere
    Scenario: unfocusing nothing is written down nowhere
      Given I added a task titled "Write the spec"
      And I let the focus go
      When I run `cairn trace list`
      Then it lists the acts of 1 task

  Rule: A trace that cannot be shown says why

    @trace-show-a-task-that-never-existed
    Scenario: a task that was never there has no trace to show
      Given I added a task titled "First"
      When I run `cairn trace show 99`
      Then it fails
      And it says there is no such task
      And it shows the number it could not find

    @trace-show-needs-a-number
    Scenario: a trace is asked for by number
      Given I added a task titled "First"
      When I run `cairn trace show`
      Then it fails
      And it says how to use it

    @trace-list-since-something-that-is-not-a-day
    Scenario: a period has to start on a day
      Given I added a task titled "First"
      When I run `cairn trace list --since pasunjour`
      Then it fails
      And it says what it expects instead of "pasunjour"
