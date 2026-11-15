Feature: Changes to the shape of the store

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A store behind the cairn that opens it is not touched

    @store-behind-refused
    Scenario: a command on a store in need of a change says what to run
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn task list`
      Then it exits with code 3
      And it says the store is behind
      And it says to run `cairn migration apply`

  Rule: A change is applied when it is asked for

    @migration-apply-in-order
    Scenario: a change is never made before one it was written after

      Given a store from an older cairn, holding a task titled "Write the spec"
      And I rolled the last change back
      When I apply the last change waiting, past the one before it
      Then it fails
      And it says an earlier change is still waiting

  Rule: A label is written one way, and has no family

    @migration-labels-split-at-their-family
    Scenario: the change splits every label with a family into labels of its own
      Given a store from before labels were written one way
      When I run `cairn migration apply`
      Then task 1 carries exactly the labels "77594, ado, chantier, proxmox"
      And objective 1 carries exactly the labels "chantier, doc-infra"

  Rule: The time written on a task is kept when the time spent is read from the trace

    @migration-time-spent-kept
    Scenario: the time written on a task before is still its time spent
      Given a store from before the time spent was read from the trace
      When I run `cairn migration apply`
      And I run `cairn task show 2`
      Then the task carries the time spent "2h30m"

  Rule: A store's sub-tasks become steps of the task they hung under

    @migration-sub-tasks-become-steps
    Scenario: every sub-task becomes a step of its parent, done when it was, what is left first
      Given a store from before tasks were split into steps
      When I run `cairn migration apply`
      Then `cairn step list 1` names step 1 "Write the tests" as not done
      And `cairn step list 1` names step 2 "Write the spec" as done
      And `cairn step list 4` names step 1 "Draft the post" as not done

  Rule: What a change will do can be read before it is made

    @migration-show-unknown
    Scenario: showing a change that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn migration show 99`
      Then it fails
      And it says there is no such change

  Rule: What was done to a store can be read back

    @migration-list
    Scenario: the store says what has been done to it, and when
      Given I added a task titled "Write the spec"
      When I run `cairn migration list`
      Then it succeeds
      And it names every change made to the store
      And it says when each was made

  Rule: The last change applied can be rolled back

    @migration-rollback
    Scenario: rolling back puts the store back the way it was
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      When I run `cairn migration rollback --yes`
      Then it succeeds
      And it names the change it undid

  Rule: An objective can sit under another, and rolling that back keeps every objective

    @objective-sub-migration-every-objective-level-one
    Scenario: every objective already in a store stands on its own once the change is applied
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I rolled back the change named "an objective can sit under another"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn objective show 1` says it sits under nothing
      And `cairn task show 1` says it serves objective 1 alone

  Rule: A label has a scope, and rolling that back leaves every label for both

    @label-scope-migration-rollback-leaves-both
    Scenario: rolled back, a label for objectives only is for both again
      Given I ran `cairn label new area --scope objectives`
      And I rolled back the change named "a label has a scope"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn label list` says "for both"

  Rule: An objective can wait on another, and rolling that back keeps every objective

    @objective-waits-on-migration-rollback
    Scenario: rolled back, an objective waits on nothing
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And I ran `cairn objective edit 2 --waits-on 1`
      And I rolled back the change named "an objective can wait on another"
      When I run `cairn migration apply`
      Then it succeeds
      When I run `cairn objective show 1`
      Then the answer says "Firmware"
      And the answer does not say "waits on it"

  Rule: The views are kept in the store, and rolling that back loses them alone

    @view-migration-rollback-loses-the-views
    Scenario: rolled back, the views are gone and every task stays
      Given I added a task titled "Write the spec"
      And I ran `cairn view new Next --state backlog`
      And I rolled back the change named "a view keeps the filters under a name"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn view list` says "there is no view"
      And `cairn task list` lists 1 task

  Rule: The series are kept in the store, and rolling that back keeps every task they made

    @series-migration-rollback-keeps-the-tasks
    Scenario: rolled back, the series are gone and every task they made stays
      Given recurrence is on for this store
      And I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I rolled back the change named "a task can come back, as a series"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn series list` says "there is no series"
      And `cairn task list` lists 1 task
